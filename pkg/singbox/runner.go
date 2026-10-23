package singbox

import (
	"bufio"
	"fmt"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"archvpn/pkg/node"
)

type ConnectionStatus struct {
	IsConnected bool       `json:"is_connected"`
	ActiveNode  *node.Node `json:"active_node,omitempty"`
	StartedAt   time.Time  `json:"started_at,omitempty"`
	PID         int        `json:"pid,omitempty"`
	RecentLogs  []string   `json:"recent_logs,omitempty"`
}

type Runner struct {
	mu           sync.RWMutex
	cmd          *exec.Cmd
	status       ConnectionStatus
	logChan      chan string
	recentLogs   []string
	maxLogs      int
	tunInterface string
}

func NewRunner() *Runner {
	return &Runner{
		logChan:      make(chan string, 100),
		recentLogs:   make([]string, 0, 50),
		maxLogs:      50,
		tunInterface: "tun0",
	}
}

// LogChannel returns a read-only channel of live log lines
func (r *Runner) LogChannel() <-chan string {
	return r.logChan
}

// Status returns a copy of current status, checking persistent state file if local runner is idle
func (r *Runner) Status() ConnectionStatus {
	r.mu.RLock()
	st := r.status
	logs := append([]string{}, r.recentLogs...)
	r.mu.RUnlock()

	if st.IsConnected {
		st.RecentLogs = logs
		return st
	}

	// Fallback to cross-process disk state
	if state, active := LoadState(); active && state != nil {
		return ConnectionStatus{
			IsConnected: true,
			ActiveNode:  state.ActiveNode,
			StartedAt:   state.StartedAt,
			PID:         state.PID,
			RecentLogs:  logs,
		}
	}

	return ConnectionStatus{IsConnected: false}
}

// CheckConfig validates a config file using sing-box check
func CheckConfig(cfgPath string) error {
	cmd := exec.Command("sing-box", "check", "-c", cfgPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("sing-box check failed: %s (%w)", string(out), err)
	}
	return nil
}

// Start launches sing-box with the given config file
func (r *Runner) Start(n *node.Node, cfgPath string) error {
	r.mu.Lock()
	if r.status.IsConnected && r.cmd != nil {
		r.mu.Unlock()
		return fmt.Errorf("VPN is already running (PID %d)", r.status.PID)
	}
	r.mu.Unlock()

	// Validate config first
	if err := CheckConfig(cfgPath); err != nil {
		return err
	}

	// Clean up any stale tun interface or stray sing-box processes
	_ = DisconnectAll()

	// Try running directly using Linux capabilities first
	cmd := exec.Command("sing-box", "run", "-c", cfgPath)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true, // run in process group for clean tree termination
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout: %w", err)
	}
	cmd.Stderr = cmd.Stdout

	if err := cmd.Start(); err != nil {
		// Fallback to sudo if unprivileged launch fails
		cmd = exec.Command("sudo", "sing-box", "run", "-c", cfgPath)
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setpgid: true,
		}
		stdout, err = cmd.StdoutPipe()
		if err != nil {
			return fmt.Errorf("failed to get stdout: %w", err)
		}
		cmd.Stderr = cmd.Stdout
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("failed to start sing-box: %w", err)
		}
	}

	pid := cmd.Process.Pid
	now := time.Now()

	r.mu.Lock()
	r.cmd = cmd
	r.status = ConnectionStatus{
		IsConnected: true,
		ActiveNode:  n,
		StartedAt:   now,
		PID:         pid,
	}
	r.mu.Unlock()

	// Save cross-process persistent state
	_ = SaveState(&State{
		PID:          pid,
		StartedAt:    now,
		ActiveNode:   n,
		ConfigPath:   cfgPath,
		TunInterface: r.tunInterface,
	})

	// Read logs and monitor process in background
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			r.appendLog(line)
			select {
			case r.logChan <- line:
			default:
			}
		}

		_ = cmd.Wait()
		r.mu.Lock()
		r.status.IsConnected = false
		r.cmd = nil
		r.mu.Unlock()
		ClearState()
	}()

	// Wait 400ms without holding lock to verify process didn't crash
	time.Sleep(400 * time.Millisecond)

	r.mu.RLock()
	connected := r.status.IsConnected
	r.mu.RUnlock()

	if !connected || !isProcessAlive(pid) {
		ClearState()
		return fmt.Errorf("sing-box process exited immediately. Check logs or permissions")
	}

	return nil
}

// Stop cleanly terminates the sing-box process
func (r *Runner) Stop() error {
	r.mu.Lock()
	r.status.IsConnected = false
	r.cmd = nil
	r.mu.Unlock()

	return DisconnectAll()
}

func (r *Runner) appendLog(line string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.recentLogs) >= r.maxLogs {
		r.recentLogs = r.recentLogs[1:]
	}
	r.recentLogs = append(r.recentLogs, line)
}
