package singbox

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"archvpn/pkg/node"
)

type State struct {
	PID          int        `json:"pid"`
	StartedAt    time.Time  `json:"started_at"`
	ActiveNode   *node.Node `json:"active_node"`
	ConfigPath   string     `json:"config_path"`
	TunInterface string     `json:"tun_interface"`
}

func stateFilePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "archvpn", "state.json")
}

// SaveState writes the active VPN connection state to disk
func SaveState(st *State) error {
	path := stateFilePath()
	_ = os.MkdirAll(filepath.Dir(path), 0755)
	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// LoadState reads the active state and verifies if the process is actually still running
func LoadState() (*State, bool) {
	path := stateFilePath()
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}

	var st State
	if err := json.Unmarshal(data, &st); err != nil {
		_ = os.Remove(path)
		return nil, false
	}

	// Check if PID is alive
	if st.PID <= 0 || !isProcessAlive(st.PID) {
		_ = os.Remove(path)
		return nil, false
	}

	return &st, true
}

// ClearState deletes the state file
func ClearState() {
	_ = os.Remove(stateFilePath())
}

// isProcessAlive sends signal 0 to test process liveness
func isProcessAlive(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

// DisconnectAll stops any active VPN process recorded in state or running sing-box
func DisconnectAll() error {
	st, active := LoadState()
	if active && st != nil && st.PID > 0 {
		// Terminate process group
		_ = syscall.Kill(-st.PID, syscall.SIGTERM)
		time.Sleep(200 * time.Millisecond)
		if isProcessAlive(st.PID) {
			_ = syscall.Kill(-st.PID, syscall.SIGKILL)
		}
	}

	// Also kill any remaining sing-box instances
	_ = exec.Command("sudo", "killall", "-15", "sing-box").Run()
	time.Sleep(100 * time.Millisecond)

	ClearState()

	// Clean up TUN interface if lingering
	tun := "tun0"
	if st != nil && st.TunInterface != "" {
		tun = st.TunInterface
	}
	_ = exec.Command("sudo", "ip", "link", "delete", tun).Run()

	return nil
}
