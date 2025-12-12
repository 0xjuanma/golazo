package ui

import (
	"strings"

	"github.com/0xjuanma/golazo/internal/api"
	"github.com/charmbracelet/lipgloss"
)

// MatchDisplay wraps a match with display information for rendering.
type MatchDisplay struct {
	api.Match
}

var (
	// Legacy match list styles (kept for backward compatibility)
	matchHeaderStyle = lipgloss.NewStyle().
		Foreground(accentColor).
		Bold(true).
		Padding(1, 0).
		BorderBottom(true).
		BorderForeground(borderColor)
)

// RenderLiveMatches renders the live matches view with a list of matches.
// width and height specify the terminal dimensions.
// matches is the list of matches to display.
// selected indicates which match is currently selected (0-indexed).
func RenderLiveMatches(width, height int, matches []MatchDisplay, selected int) string {
	lines := make([]string, 0, len(matches)+3) // +3 for header, empty line, help

	// Header
	header := matchHeaderStyle.Width(width - 2).Render("Live Matches")
	lines = append(lines, header)
	lines = append(lines, "")

	if len(matches) == 0 {
		noMatches := lipgloss.NewStyle().
			Foreground(dimColor).
			Align(lipgloss.Center).
			Padding(2, 0).
			Render("No matches available")
		lines = append(lines, noMatches)
	} else {
		// Render each match
		for i, match := range matches {
			line := renderMatchItem(match, i == selected, width-4)
			lines = append(lines, line)
		}
	}

	// Help text
	help := menuHelpStyle.Render("↑/↓: navigate  Esc: back  q: quit")
	lines = append(lines, "")
	lines = append(lines, help)

	content := strings.Join(lines, "\n")

	// Add padding
	paddedContent := lipgloss.NewStyle().
		Padding(1, 2).
		Render(content)

	return lipgloss.Place(
		width,
		height,
		lipgloss.Left,
		lipgloss.Top,
		paddedContent,
	)
}

// renderMatchItem is kept for backward compatibility but uses styles from panels.go
func renderMatchItem(match MatchDisplay, selected bool, width int) string {
	// This function is deprecated - use renderMatchListItem from panels.go instead
	return renderMatchListItem(match, selected, width)
}
