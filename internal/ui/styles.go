package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Brand colors: Arch cyan and dark accents
	ColorArchCyan = lipgloss.Color("#1793D1")
	ColorEmerald  = lipgloss.Color("#10B981")
	ColorYellow   = lipgloss.Color("#F59E0B")
	ColorRose     = lipgloss.Color("#F43F5E")
	ColorMuted    = lipgloss.Color("#64748B")
	ColorSlate    = lipgloss.Color("#334155")
	ColorDarkBg   = lipgloss.Color("#0F172A")
	ColorWhite    = lipgloss.Color("#F8FAFC")

	// Styles
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorWhite).
			Background(ColorArchCyan).
			Padding(0, 1).
			MarginRight(1)

	StatusConnectedStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorWhite).
				Background(ColorEmerald).
				Padding(0, 1)

	StatusDisconnectedStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorWhite).
				Background(ColorSlate).
				Padding(0, 1)

	StatusConnectingStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorWhite).
				Background(ColorYellow).
				Padding(0, 1)

	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorArchCyan).
			Padding(0, 1)

	SelectedRowStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorArchCyan).
				Background(ColorSlate).
				Padding(0, 1)

	NormalRowStyle = lipgloss.NewStyle().
			Padding(0, 1)

	LatencyFastStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorEmerald)

	LatencyMediumStyle = lipgloss.NewStyle().
				Foreground(ColorYellow)

	LatencySlowStyle = lipgloss.NewStyle().
				Foreground(ColorRose)

	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			MarginTop(1)

	MutedTextStyle = lipgloss.NewStyle().Foreground(ColorMuted)
	CyanTextStyle  = lipgloss.NewStyle().Foreground(ColorArchCyan)

	LogBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorSlate).
			Foreground(ColorMuted).
			Padding(0, 1)
)
