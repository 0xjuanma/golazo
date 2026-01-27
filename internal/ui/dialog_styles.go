package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// Dialog-specific styles using existing adaptive colors from neon_styles.go.
// All colors are adaptive and work on both light and dark terminal backgrounds.
var (
	// dialogBorderStyle applies a rounded border with cyan accent.
	dialogBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(neonCyan).
				Padding(1, 2)

	// dialogTitleStyle styles the dialog title with red accent.
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
)

// RenderDialogFrame wraps content in a dialog frame with title.
func RenderDialogFrame(title, content string, width, height int) string {
	titleRendered := dialogTitleStyle.Render(title)

	innerContent := lipgloss.JoinVertical(lipgloss.Left, titleRendered, content)

	return dialogBorderStyle.
		Width(width).
		MaxWidth(width).
		Height(height).
		MaxHeight(height).
		Render(innerContent)
}

// RenderDialogFrameWithHelp wraps content in a dialog frame with title and help text.
func RenderDialogFrameWithHelp(title, content, help string, width, height int) string {
	titleRendered := dialogTitleStyle.Render(title)
	helpRendered := dialogHelpStyle.Render(help)

	innerContent := lipgloss.JoinVertical(lipgloss.Left, titleRendered, content, helpRendered)

	return dialogBorderStyle.
		Width(width).
		MaxWidth(width).
		Height(height).
		MaxHeight(height).
		Render(innerContent)
}
