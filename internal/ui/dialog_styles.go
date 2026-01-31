package ui

import (
	"strings"

	"github.com/0xjuanma/golazo/internal/ui/design"
	"github.com/charmbracelet/lipgloss"
)

// Dialog-specific styles using existing adaptive colors from neon_styles.go.
// All colors are adaptive and work on both light and dark terminal backgrounds.
var (
	// dialogBorderStyle applies padding without border for a cleaner look.
	dialogBorderStyle = lipgloss.NewStyle().
				Padding(1, 2)

	// dialogTitleBarStyle styles the title bar with inverted colors.
	dialogTitleBarStyle = lipgloss.NewStyle().
				Background(neonRed).
				Foreground(neonWhite).
				Bold(true).
				Padding(0, 2).
				MarginBottom(1)

	// dialogTitleStyle styles plain dialog titles (fallback).
	dialogTitleStyle = lipgloss.NewStyle().
				Foreground(neonRed).
				Bold(true).
				MarginBottom(1)

	// dialogContentStyle styles the main dialog content.
	dialogContentStyle = lipgloss.NewStyle().
				Foreground(neonWhite)

	// dialogDimStyle styles secondary/muted text.
	dialogDimStyle = lipgloss.NewStyle().
			Foreground(neonDim)

	// dialogHeaderStyle styles column headers in tables.
	dialogHeaderStyle = lipgloss.NewStyle().
				Foreground(neonCyan).
				Bold(true)

	// dialogHighlightStyle highlights important rows (e.g., current teams).
	dialogHighlightStyle = lipgloss.NewStyle().
				Foreground(neonRed).
				Bold(true)

	// dialogValueStyle styles numeric values.
	dialogValueStyle = lipgloss.NewStyle().
				Foreground(neonWhiteAlt)

	// dialogLabelStyle styles labels with fixed width.
	dialogLabelStyle = lipgloss.NewStyle().
				Foreground(neonDim).
				Width(12)

	// dialogTeamStyle styles team names.
	dialogTeamStyle = lipgloss.NewStyle().
			Foreground(neonCyan).
			Bold(true)

	// dialogPositionStyle styles position indicators.
	dialogPositionStyle = lipgloss.NewStyle().
				Foreground(neonWhite).
				Width(3).
				Align(lipgloss.Right)

	// dialogSeparatorStyle styles horizontal separators.
	dialogSeparatorStyle = lipgloss.NewStyle().
				Foreground(neonDarkDim)

	// dialogHelpStyle styles help text at the bottom.
	dialogHelpStyle = lipgloss.NewStyle().
			Foreground(neonDim).
			Italic(true).
			MarginTop(1)

	// dialogBadgeStyle provides subtle background for values.
	dialogBadgeStyle = lipgloss.NewStyle().
				Background(neonDark).
				Foreground(neonWhite).
				Padding(0, 1)

	// dialogBadgeHighlightStyle provides highlighted background for winning values.
	dialogBadgeHighlightStyle = lipgloss.NewStyle().
					Background(neonRed).
					Foreground(neonWhite).
					Bold(true).
					Padding(0, 1)
)

// RenderDialogTitleBar creates a full-width title bar with background.
func RenderDialogTitleBar(title string, width int) string {
	// Center the title and fill the width
	titleLen := len(title)
	if titleLen >= width-4 {
		return dialogTitleBarStyle.Width(width).Render(title)
	}

	// Add padding characters to fill the bar
	padding := (width - titleLen - 4) / 2
	leftPad := strings.Repeat(" ", padding)
	rightPad := strings.Repeat(" ", width-titleLen-4-padding)

	fullTitle := leftPad + title + rightPad
	return dialogTitleBarStyle.Width(width).Align(lipgloss.Center).Render(fullTitle)
}

// DialogBadge wraps a value with a subtle background.
func DialogBadge(value string) string {
	return dialogBadgeStyle.Render(value)
}

// DialogBadgeHighlight wraps a value with a highlighted background.
func DialogBadgeHighlight(value string) string {
	return dialogBadgeHighlightStyle.Render(value)
}

// RenderDialogFrame wraps content in a dialog frame with title bar.
func RenderDialogFrame(title, content string, width, height int) string {
	titleBar := design.RenderHeader(title, width-6) // Use compact header with gradient

	innerContent := lipgloss.JoinVertical(lipgloss.Left, titleBar, "", content)

	return dialogBorderStyle.
		Width(width).
		MaxWidth(width).
		Height(height).
		MaxHeight(height).
		Render(innerContent)
}

// RenderDialogFrameWithHelp wraps content in a dialog frame with title bar and help text.
func RenderDialogFrameWithHelp(title, content, help string, width, height int) string {
	titleBar := design.RenderHeader(title, width-6) // Use compact header with gradient
	helpRendered := dialogHelpStyle.Width(width - 6).Align(lipgloss.Center).Render(help)

	innerContent := lipgloss.JoinVertical(lipgloss.Left, titleBar, "", content, helpRendered)

	return dialogBorderStyle.
		Width(width).
		MaxWidth(width).
		Height(height).
		MaxHeight(height).
		Render(innerContent)
}
