package ui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

// Neon colors used across delegates.
var (
	delegateNeonRed   = neonRed
	delegateNeonCyan  = neonCyan
	delegateNeonWhite = neonWhite
	delegateNeonGray  = lipgloss.Color("244") // neonDim
	delegateNeonDim   = lipgloss.Color("238") // unique, but close to neonDarkDim
)

// NewMatchListDelegate creates a custom list delegate for match items.
// Height is set to 3 to accommodate title + 2-line description (with KO time).
// Uses Neon Gradient styling: red title, cyan description on selection.
func NewMatchListDelegate() list.DefaultDelegate {
	d := list.NewDefaultDelegate()

	// Set height to 3 lines: title (1) + description with KO time (2)
	d.SetHeight(3)

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

	// Dimmed items (non-matching during filter): very dim
	d.Styles.DimmedTitle = lipgloss.NewStyle().
		Foreground(neonDim).
		Padding(0, 1)
	d.Styles.DimmedDesc = lipgloss.NewStyle().
		Foreground(neonDim).
		Padding(0, 1)

	// Filter match highlight: cyan bold for matched text
	d.Styles.FilterMatch = lipgloss.NewStyle().
		Foreground(neonCyan).
		Bold(true).
		Underline(true)

	return d
}

// NewLeagueListDelegate creates a custom list delegate for league selection.
// Height is set to 2 to show league name (with checkbox) and country.
// Uses same red/cyan neon styling as match delegate for consistency.
func NewLeagueListDelegate() list.DefaultDelegate {
	d := list.NewDefaultDelegate()

	// Set height to 2 lines: title with checkbox (1) + country (1)
	d.SetHeight(2)

	// Selected items: Neon red title, cyan description, red left border
	// Matches the match list delegate exactly
	d.Styles.SelectedTitle = lipgloss.NewStyle().
		Foreground(delegateNeonRed).
		Bold(true).
		Padding(0, 1).
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(delegateNeonRed)
	d.Styles.SelectedDesc = lipgloss.NewStyle().
		Foreground(delegateNeonCyan).
		Padding(0, 1).
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(delegateNeonRed)

	// Normal items: White title, gray description
	d.Styles.NormalTitle = lipgloss.NewStyle().
		Foreground(delegateNeonWhite).
		Padding(0, 1)
	d.Styles.NormalDesc = lipgloss.NewStyle().
		Foreground(delegateNeonGray).
		Padding(0, 1)

	// Dimmed items (non-matching during filter): very dim
	d.Styles.DimmedTitle = lipgloss.NewStyle().
		Foreground(delegateNeonDim).
		Padding(0, 1)
	d.Styles.DimmedDesc = lipgloss.NewStyle().
		Foreground(delegateNeonDim).
		Padding(0, 1)

	// Filter match highlight: cyan bold for matched text
	d.Styles.FilterMatch = lipgloss.NewStyle().
		Foreground(delegateNeonCyan).
		Bold(true).
		Underline(true)

	return d
}
