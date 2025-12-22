package ui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

// NewMatchListDelegate creates a custom list delegate for match items.
// Height is set to 3 to accommodate title + 2-line description (with KO time).
// Uses Neon Gradient styling: red title, cyan description on selection.
func NewMatchListDelegate() list.DefaultDelegate {
	d := list.NewDefaultDelegate()

	// Set height to 3 lines: title (1) + description with KO time (2)
	d.SetHeight(3)

	// Neon colors
	neonRed := lipgloss.Color("196")
	neonCyan := lipgloss.Color("51")
	neonWhite := lipgloss.Color("255")
	neonGray := lipgloss.Color("244")

	// Selected items: Neon red title, cyan description, red left border
	d.Styles.SelectedTitle = lipgloss.NewStyle().
		Foreground(neonRed).
		Bold(true).
		Padding(0, 1).
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(neonRed)
	d.Styles.SelectedDesc = lipgloss.NewStyle().
		Foreground(neonCyan).
		Padding(0, 1).
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(neonRed)

	// Normal items: White title, gray description
	d.Styles.NormalTitle = lipgloss.NewStyle().
		Foreground(neonWhite).
		Padding(0, 1)
	d.Styles.NormalDesc = lipgloss.NewStyle().
		Foreground(neonGray).
		Padding(0, 1)

	return d
}

// MatchListStyles returns styles for the list component.
func MatchListStyles() lipgloss.Style {
	return panelStyle
}
