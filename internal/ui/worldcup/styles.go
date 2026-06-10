// Package worldcup provides UI rendering components for the World Cup 2026
// beta feature. All code in this package is isolated here to allow for easy
// iteration and eventual graduation or removal.
package worldcup

import "github.com/charmbracelet/lipgloss"

// Color palette — mirrors the app's neon theme without importing internal/ui
// to avoid circular imports.
var (
	colorRed     = lipgloss.AdaptiveColor{Light: "124", Dark: "196"}
	colorCyan    = lipgloss.AdaptiveColor{Light: "23", Dark: "51"}
	colorWhite   = lipgloss.AdaptiveColor{Light: "235", Dark: "255"}
	colorDim     = lipgloss.AdaptiveColor{Light: "243", Dark: "244"}
	colorDarkDim = lipgloss.AdaptiveColor{Light: "249", Dark: "239"}
	colorGold    = lipgloss.AdaptiveColor{Light: "136", Dark: "220"} // trophy / champion gold
	colorGreen   = lipgloss.AdaptiveColor{Light: "22", Dark: "82"}   // qualified highlight
)

// Base component styles
var (
	LoadingStyle = lipgloss.NewStyle().
			Foreground(colorDim).
			Italic(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(colorRed)

	HelpStyle = lipgloss.NewStyle().
			Foreground(colorDim).
			Align(lipgloss.Center)

	PanelStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(colorRed).
			Padding(0, 1)

	// Table header — cyan bold
	HeaderStyle = lipgloss.NewStyle().
			Foreground(colorCyan).
			Bold(true)

	// Separator line
	SepStyle = lipgloss.NewStyle().
			Foreground(colorDarkDim)

	// Phase badge rendered in header area
	PhaseBadgeStyle = lipgloss.NewStyle().
			Foreground(colorGold).
			Bold(true).
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorGold)
)

// Table row styles
var (
	QualifiedStyle = lipgloss.NewStyle().Foreground(colorCyan)
	QualPtsStyle   = lipgloss.NewStyle().Foreground(colorCyan).Bold(true)

	// Third-place row (possibly advances in WC 2026 format)
	ThirdStyle    = lipgloss.NewStyle().Foreground(colorGold)
	ThirdPtsStyle = lipgloss.NewStyle().Foreground(colorGold).Bold(true)

	EliminatedStyle = lipgloss.NewStyle().Foreground(colorDim)
	EliminPtsStyle  = lipgloss.NewStyle().Foreground(colorDim)
)

// Bracket styles
var (
	RoundHeaderStyle = lipgloss.NewStyle().
				Foreground(colorCyan).
				Bold(true)

	MatchLineStyle = lipgloss.NewStyle().
			Foreground(colorWhite)

	WinnerStyle = lipgloss.NewStyle().
			Foreground(colorCyan).
			Bold(true)

	ScoreStyle = lipgloss.NewStyle().
			Foreground(colorRed).
			Bold(true)

	PenStyle = lipgloss.NewStyle().
			Foreground(colorDim).
			Italic(true)

	ConnectorStyle = lipgloss.NewStyle().
			Foreground(colorDim)

	ChampionStyle = lipgloss.NewStyle().
			Foreground(colorGold).
			Bold(true)

	TrophyCardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorGold).
			Padding(0, 2)
)

// Group grid styles
var (
	GridGroupHeaderStyle = lipgloss.NewStyle().
				Foreground(colorRed).
				Bold(true)

	GridTeamQualStyle  = lipgloss.NewStyle().Foreground(colorCyan)
	GridTeamThirdStyle = lipgloss.NewStyle().Foreground(colorGold)
	GridTeamDimStyle   = lipgloss.NewStyle().Foreground(colorDim)

	GridSelectedGroupStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(colorRed)

	GridNormalGroupStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(colorDim)
)

// alignRight renders text right-aligned in a fixed-width column.
func alignRight(width int, text string) string {
	return lipgloss.NewStyle().Width(width).Align(lipgloss.Right).Render(text)
}

// alignLeft renders text left-aligned in a fixed-width column.
func alignLeft(width int, text string) string {
	return lipgloss.NewStyle().Width(width).Align(lipgloss.Left).Render(text)
}
