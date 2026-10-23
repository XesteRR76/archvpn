package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"archvpn/internal/ui"
	"archvpn/pkg/config"
	"archvpn/pkg/node"
	"archvpn/pkg/ping"
	"archvpn/pkg/singbox"
	"archvpn/pkg/sub"
)

func Run(args []string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	mgr := sub.NewManager(cfg)
	runner := singbox.NewRunner()

	if len(args) == 0 {
		return runTUI(cfg, mgr, runner)
	}

	cmd := args[0]
	switch cmd {
	case "tui":
		return runTUI(cfg, mgr, runner)

	case "connect", "up":
		var target *node.Node
		nodes := mgr.GetNodes()
		if len(nodes) == 0 {
			fmt.Println("[*] No cached servers found, updating subscriptions...")
			if err := mgr.Update(); err != nil {
				return err
			}
			nodes = mgr.GetNodes()
		}

		if len(args) > 1 {
			target = mgr.FindNode(args[1])
			if target == nil {
				return fmt.Errorf("server %q not found", args[1])
			}
		} else if cfg.LastSelectedNodeID != "" {
			target = mgr.FindNode(cfg.LastSelectedNodeID)
		}

		if target == nil && len(nodes) > 0 {
			target = nodes[0]
		}
		if target == nil {
			return fmt.Errorf("no servers available")
		}

		return runConnect(cfg, runner, target)

	case "disconnect", "down":
		fmt.Println("[*] Stopping VPN connection...")
		if err := runner.Stop(); err != nil {
			return err
		}
		fmt.Println("[+] VPN disconnected cleanly.")
		return nil

	case "status":
		asJSON := false
		if len(args) > 1 && args[1] == "--json" {
			asJSON = true
		}
		return printStatus(cfg, runner, asJSON)

	case "update":
		if len(cfg.SubscriptionURLs) == 0 {
			return fmt.Errorf("no subscription URLs configured. Run 'archvpn add-sub <url>' first")
		}
		fmt.Println("[*] Updating subscriptions...")
		if err := mgr.Update(); err != nil {
			return err
		}
		fmt.Printf("[+] Successfully updated! Cached %d servers.\n", len(mgr.GetNodes()))
		return nil

	case "ping":
		return runPing(mgr)

	case "list", "ls":
		return printList(mgr)

	case "add-sub":
		if len(args) < 2 {
			return fmt.Errorf("usage: archvpn add-sub <url>")
		}
		newURL := args[1]
		cfg.SubscriptionURLs = append(cfg.SubscriptionURLs, newURL)
		if err := cfg.Save(); err != nil {
			return err
		}
		fmt.Printf("[+] Added subscription: %s\n", newURL)
		return mgr.Update()

	case "help", "-h", "--help":
		printHelp()
		return nil

	default:
		return fmt.Errorf("unknown command %q. Run 'archvpn help' for usage", cmd)
	}
}

func runTUI(cfg *config.Config, mgr *sub.Manager, runner *singbox.Runner) error {
	nodes := mgr.GetNodes()
	if len(nodes) == 0 {
		fmt.Println("[*] Initial launch: downloading subscriptions...")
		_ = mgr.Update()
	}

	model := ui.NewModel(cfg, mgr, runner)
	p := tea.NewProgram(model, tea.WithAltScreen())

	// Handle OS signals to cleanly stop singbox if user terminates
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		<-sigChan
		_ = runner.Stop()
		p.Kill()
	}()

	_, err := p.Run()
	_ = runner.Stop()
	return err
}

func runConnect(cfg *config.Config, runner *singbox.Runner, target *node.Node) error {
	fmt.Printf("[*] Generating sing-box config for %s (%s)...\n", target.DisplayName(), target.Server)
	cfgData, err := singbox.GenerateConfig(target, cfg.TunInterface, cfg.DNSServers)
	if err != nil {
		return err
	}

	cfgPath := filepath.Join(cfg.ConfigDir, "active_singbox.json")
	if err := os.WriteFile(cfgPath, cfgData, 0644); err != nil {
		return err
	}

	cfg.LastSelectedNodeID = target.ID
	_ = cfg.Save()

	fmt.Println("[*] Starting TUN tunnel via sing-box...")
	if err := runner.Start(target, cfgPath); err != nil {
		return err
	}

	fmt.Printf("[+] Connected to %s! (Interface: %s)\n", target.DisplayName(), cfg.TunInterface)
	fmt.Println("[i] Press Ctrl+C to disconnect...")

	// Listen for logs
	go func() {
		for log := range runner.LogChannel() {
			fmt.Printf(" [sing-box] %s\n", log)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	<-sigChan

	fmt.Println("\n[*] Shutting down tunnel...")
	_ = runner.Stop()
	fmt.Println("[+] Done. Routes restored.")
	return nil
}

func runPing(mgr *sub.Manager) error {
	nodes := mgr.GetNodes()
	if len(nodes) == 0 {
		return fmt.Errorf("no nodes cached. Run 'archvpn update' first")
	}

	fmt.Printf("[*] Benchmarking %d servers concurrently...\n", len(nodes))
	resChan := ping.BenchmarkNodes(nodes, 2*time.Second, 20)

	for res := range resChan {
		latStr := "timeout"
		if res.LatencyMs > 0 {
			latStr = fmt.Sprintf("%d ms", res.LatencyMs)
		}
		fmt.Printf("  %-10s %-32s -> %s\n", res.Node.InferCountry(), res.Node.DisplayName(), latStr)
	}

	_ = mgr.SaveNodes()
	fmt.Println("[+] Latency benchmark complete and saved.")
	return nil
}

func printList(mgr *sub.Manager) error {
	nodes := mgr.GetNodes()
	if len(nodes) == 0 {
		fmt.Println("No servers available. Run 'archvpn update' first.")
		return nil
	}

	// Sort nodes by country and latency
	sorted := append([]*node.Node{}, nodes...)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].LatencyMs > 0 && sorted[j].LatencyMs > 0 {
			return sorted[i].LatencyMs < sorted[j].LatencyMs
		}
		return sorted[i].LatencyMs > sorted[j].LatencyMs
	})

	fmt.Printf("%-8s %-12s %-38s %-10s %-10s\n", "ID", "COUNTRY", "REMARK", "TYPE", "LATENCY")
	fmt.Println(strings.Repeat("-", 82))
	for _, n := range sorted {
		lat := "---"
		if n.LatencyMs > 0 {
			lat = fmt.Sprintf("%d ms", n.LatencyMs)
		} else if n.LatencyMs == -1 {
			lat = "timeout"
		}
		remark := n.DisplayName()
		if len(remark) > 36 {
			remark = remark[:33] + "..."
		}
		fmt.Printf("%-8s %-12s %-38s %-10s %-10s\n", n.ID, n.InferCountry(), remark, n.NetworkType, lat)
	}
	return nil
}

func printStatus(cfg *config.Config, runner *singbox.Runner, asJSON bool) error {
	st := runner.Status()
	if asJSON {
		waybarOutput := map[string]interface{}{
			"text":    "VPN: OFF",
			"class":   "disconnected",
			"tooltip": "ArchVPN Disconnected",
		}
		if st.IsConnected && st.ActiveNode != nil {
			waybarOutput["text"] = fmt.Sprintf("VPN: %s", st.ActiveNode.InferCountry())
			waybarOutput["class"] = "connected"
			waybarOutput["tooltip"] = fmt.Sprintf("Connected to %s (%s)", st.ActiveNode.DisplayName(), st.ActiveNode.Server)
		}
		data, _ := json.Marshal(waybarOutput)
		fmt.Println(string(data))
		return nil
	}

	if st.IsConnected && st.ActiveNode != nil {
		fmt.Printf("[+] Status: CONNECTED\n")
		fmt.Printf("    Server:  %s (%s:%d)\n", st.ActiveNode.DisplayName(), st.ActiveNode.Server, st.ActiveNode.Port)
		fmt.Printf("    Uptime:  %s\n", time.Since(st.StartedAt).Round(time.Second))
		fmt.Printf("    PID:     %d\n", st.PID)
	} else {
		fmt.Println("[-] Status: DISCONNECTED")
	}
	return nil
}

func printHelp() {
	fmt.Println(`⚡ ArchVPN — High Performance Terminal VPN Client for Arch Linux

USAGE:
    archvpn                       Launch interactive TUI dashboard
    archvpn connect [id|remark]   Connect to a specific or last active server
    archvpn disconnect            Stop active VPN connection
    archvpn status [--json]       Display current connection status (Waybar ready)
    archvpn update                Fetch latest nodes from all subscriptions
    archvpn ping                  Benchmark latency for all servers
    archvpn list                  Print list of all servers sorted by latency
    archvpn add-sub <url>         Add a new subscription link
    archvpn help                  Show this help message

TUI SHORTCUTS:
    [↑ / ↓] or [k / j]            Navigate server list
    [Enter] or [Space]            Connect to selected server
    [d]                           Disconnect
    [p]                           Run ping benchmark on all nodes
    [u]                           Update subscriptions
    [/]                           Search / filter servers
    [q]                           Quit`)
}
