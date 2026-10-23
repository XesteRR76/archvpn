package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"archvpn/pkg/config"
	"archvpn/pkg/node"
	"archvpn/pkg/ping"
	"archvpn/pkg/singbox"
	"archvpn/pkg/sub"
)

// Messages
type pingFinishedMsg struct{}
type pingResultMsg struct {
	result ping.BenchmarkResult
}
type updateFinishedMsg struct {
	err error
}
type connectionResultMsg struct {
	err error
}
type tickMsg time.Time

type Model struct {
	cfg          *config.Config
	mgr          *sub.Manager
	runner       *singbox.Runner
	nodes        []*node.Node
	cursor       int
	scrollOffset int
	height       int
	width        int

	isPinging    bool
	isUpdating   bool
	isConnecting bool
	statusText   string
	errMessage   string

	filterInput textinput.Model
	isFiltering bool
}

func NewModel(cfg *config.Config, mgr *sub.Manager, runner *singbox.Runner) Model {
	ti := textinput.New()
	ti.Placeholder = "Filter servers..."
	ti.Prompt = "🔍 "

	m := Model{
		cfg:         cfg,
		mgr:         mgr,
		runner:      runner,
		nodes:       mgr.GetNodes(),
		filterInput: ti,
		width:       80,
		height:      24,
	}

	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		tickCmd(),
	)
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m Model) filtered() []*node.Node {
	query := strings.ToLower(strings.TrimSpace(m.filterInput.Value()))
	if query == "" {
		return m.nodes
	}

	var res []*node.Node
	for _, n := range m.nodes {
		if strings.Contains(strings.ToLower(n.Remark), query) ||
			strings.Contains(strings.ToLower(n.Server), query) ||
			strings.Contains(strings.ToLower(n.InferCountry()), query) {
			res = append(res, n)
		}
	}
	return res
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tickMsg:
		cmds = append(cmds, tickCmd())

	case pingResultMsg:
		for _, n := range m.nodes {
			if n.ID == msg.result.Node.ID {
				n.LatencyMs = msg.result.LatencyMs
				break
			}
		}
		_ = m.mgr.SaveNodes()

	case pingFinishedMsg:
		m.isPinging = false
		m.nodes = m.mgr.GetNodes()
		m.statusText = "Ping benchmark complete"

	case updateFinishedMsg:
		m.isUpdating = false
		if msg.err != nil {
			m.errMessage = fmt.Sprintf("Update error: %v", msg.err)
		} else {
			m.nodes = m.mgr.GetNodes()
			m.statusText = fmt.Sprintf("Updated! Loaded %d servers", len(m.nodes))
			m.errMessage = ""
		}

	case connectionResultMsg:
		m.isConnecting = false
		if msg.err != nil {
			m.errMessage = fmt.Sprintf("Connect failed: %v", msg.err)
		} else {
			m.errMessage = ""
			m.statusText = "Connected!"
		}

	case tea.KeyMsg:
		if m.isFiltering {
			switch msg.String() {
			case "enter", "esc":
				m.isFiltering = false
				m.filterInput.Blur()
			default:
				var cmd tea.Cmd
				m.filterInput, cmd = m.filterInput.Update(msg)
				cmds = append(cmds, cmd)
				m.cursor = 0
			}
			return m, tea.Batch(cmds...)
		}

		switch msg.String() {
		case "q", "ctrl+c":
			if m.runner.Status().IsConnected {
				_ = m.runner.Stop()
			}
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			list := m.filtered()
			if m.cursor < len(list)-1 {
				m.cursor++
			}

		case "/":
			m.isFiltering = true
			m.filterInput.Focus()
			return m, textinput.Blink

		case "p":
			if !m.isPinging && len(m.nodes) > 0 {
				m.isPinging = true
				m.statusText = "Benchmarking latency for all nodes..."
				cmds = append(cmds, m.benchmarkCmd())
			}

		case "u":
			if !m.isUpdating {
				m.isUpdating = true
				m.statusText = "Fetching subscriptions..."
				cmds = append(cmds, m.updateCmd())
			}

		case "d":
			_ = m.runner.Stop()
			m.statusText = "Disconnected"
			m.errMessage = ""

		case "enter", " ":
			list := m.filtered()
			if len(list) > 0 && m.cursor < len(list) {
				selected := list[m.cursor]
				m.isConnecting = true
				m.statusText = fmt.Sprintf("Connecting to %s...", selected.DisplayName())
				m.errMessage = ""
				cmds = append(cmds, m.connectCmd(selected))
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m Model) connectCmd(target *node.Node) tea.Cmd {
	return func() tea.Msg {
		cfgData, err := singbox.GenerateConfig(target, m.cfg.TunInterface, m.cfg.DNSServers)
		if err != nil {
			return connectionResultMsg{err: err}
		}

		if err := m.cfg.Save(); err != nil {
			return connectionResultMsg{err: err}
		}

		cfgPath := filepath.Join(m.cfg.ConfigDir, "active_singbox.json")
		if err := os.WriteFile(cfgPath, cfgData, 0644); err != nil {
			return connectionResultMsg{err: fmt.Errorf("failed to save singbox config: %w", err)}
		}

		m.cfg.LastSelectedNodeID = target.ID
		_ = m.cfg.Save()

		err = m.runner.Start(target, cfgPath)
		return connectionResultMsg{err: err}
	}
}

func (m Model) benchmarkCmd() tea.Cmd {
	return func() tea.Msg {
		resChan := ping.BenchmarkNodes(m.nodes, 2*time.Second, 20)
		for res := range resChan {
			// Update in real-time
			for _, n := range m.nodes {
				if n.ID == res.Node.ID {
					n.LatencyMs = res.LatencyMs
					break
				}
			}
		}
		_ = m.mgr.SaveNodes()
		return pingFinishedMsg{}
	}
}

func (m Model) updateCmd() tea.Cmd {
	return func() tea.Msg {
		err := m.mgr.Update()
		return updateFinishedMsg{err: err}
	}
}

func (m Model) View() string {
	var b strings.Builder

	// Header
	title := TitleStyle.Render("⚡ ARCH-VPN")
	st := m.runner.Status()

	var statusBadge string
	if m.isConnecting {
		statusBadge = StatusConnectingStyle.Render("CONNECTING...")
	} else if st.IsConnected {
		uptime := time.Since(st.StartedAt).Round(time.Second)
		statusBadge = StatusConnectedStyle.Render(fmt.Sprintf("● CONNECTED: %s (%s)", st.ActiveNode.DisplayName(), uptime))
	} else {
		statusBadge = StatusDisconnectedStyle.Render("○ DISCONNECTED")
	}

	header := lipgloss.JoinHorizontal(lipgloss.Center, title, statusBadge)
	b.WriteString(header + "\n\n")

	// Search / Filter row
	if m.isFiltering {
		b.WriteString(m.filterInput.View() + "\n\n")
	} else if m.filterInput.Value() != "" {
		b.WriteString(MutedTextStyle.Render(fmt.Sprintf("Filter: %q (press / to edit, esc to clear)", m.filterInput.Value())) + "\n\n")
	}

	// Status line or error message
	if m.errMessage != "" {
		b.WriteString(LatencySlowStyle.Render("✖ "+m.errMessage) + "\n\n")
	} else if m.statusText != "" {
		b.WriteString(CyanTextStyle.Render("ℹ "+m.statusText) + "\n\n")
	}

	// Node List
	list := m.filtered()
	if len(list) == 0 {
		b.WriteString(MutedTextStyle.Render("  No servers available. Press [u] to fetch subscriptions.\n\n"))
	} else {
		// Viewport windowing
		visibleCount := m.height - 10
		if visibleCount < 5 {
			visibleCount = 8
		}

		startIdx := 0
		if m.cursor >= visibleCount {
			startIdx = m.cursor - visibleCount + 1
		}
		endIdx := startIdx + visibleCount
		if endIdx > len(list) {
			endIdx = len(list)
		}

		for i := startIdx; i < endIdx; i++ {
			n := list[i]
			isCurrent := (i == m.cursor)
			isConnected := st.IsConnected && st.ActiveNode != nil && st.ActiveNode.ID == n.ID

			country := n.InferCountry()
			name := n.DisplayName()
			if len(name) > 35 {
				name = name[:32] + "..."
			}

			// Latency badge
			var latStr string
			if n.LatencyMs > 0 {
				latText := fmt.Sprintf("%3d ms", n.LatencyMs)
				if n.LatencyMs < 150 {
					latStr = LatencyFastStyle.Render(latText)
				} else if n.LatencyMs < 350 {
					latStr = LatencyMediumStyle.Render(latText)
				} else {
					latStr = LatencySlowStyle.Render(latText)
				}
			} else if n.LatencyMs == -1 {
				latStr = LatencySlowStyle.Render("timeout")
			} else {
				latStr = MutedTextStyle.Render("  ---  ")
			}

			prefix := "  "
			if isConnected {
				prefix = "● "
			} else if isCurrent {
				prefix = "❯ "
			}

			rowContent := fmt.Sprintf("%s %-10s %-36s  [%s]  %s", prefix, country, name, n.NetworkType, latStr)

			if isCurrent {
				b.WriteString(SelectedRowStyle.Render(rowContent) + "\n")
			} else {
				b.WriteString(NormalRowStyle.Render(rowContent) + "\n")
			}
		}
	}

	// Footer / Shortcuts
	help := HelpStyle.Render("[↑/↓] Navigate  [Enter/Space] Connect  [d] Disconnect  [p] Ping All  [u] Update Subs  [/] Filter  [q] Quit")
	b.WriteString("\n" + help)

	return b.String()
}
