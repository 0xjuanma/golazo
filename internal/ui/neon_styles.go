package ui

import (
	"github.com/0xjuanma/golazo/internal/ui/design"
	"github.com/charmbracelet/lipgloss"
)

// Neon design styles - Golazo red/cyan theme
// Bold, vibrant design with thick borders and high contrast.

// Card symbols - consistent across all views
const (
	CardSymbolYellow = "▪" // Small square for yellow cards
	CardSymbolRed    = "■" // Filled square for red cards
)

var (
	// Neon color palette - Golazo brand
	// Primary colors - adaptive for light/dark terminals
	neonRed    = lipgloss.AdaptiveColor{Light: "124", Dark: "196"} // Dark red / Bright red
	neonCyan   = lipgloss.AdaptiveColor{Light: "23", Dark: "51"}   // Darker cyan / Electric cyan
	neonYellow = lipgloss.AdaptiveColor{Light: "136", Dark: "226"} // Dark gold / Bright yellow for cards
	// Adaptive white - dark gray on light terminals, white on dark terminals
	neonWhite = lipgloss.AdaptiveColor{Light: "235", Dark: "255"} // Adaptive text color
	// Adaptive white alt - slightly different shades for variety
	neonWhiteAlt = lipgloss.AdaptiveColor{Light: "236", Dark: "15"} // Standard adaptive text
	neonBlack    = lipgloss.Color("0")                              // Black

	// Gray scale - adaptive for light/dark terminals
	neonDark    = lipgloss.AdaptiveColor{Light: "252", Dark: "236"} // Light gray / Dark background
	neonDarkDim = lipgloss.AdaptiveColor{Light: "249", Dark: "239"} // Light gray / Slightly lighter dark
	neonGray    = lipgloss.AdaptiveColor{Light: "245", Dark: "240"} // Medium gray (visible on both)
	neonDim     = lipgloss.AdaptiveColor{Light: "243", Dark: "244"} // Gray dim text
	neonDimGray = lipgloss.AdaptiveColor{Light: "246", Dark: "238"} // Dim gray (for delegates)

	// Card styles - reusable across all views
	neonYellowCardStyle = lipgloss.NewStyle().Foreground(neonYellow).Bold(true)
	neonRedCardStyle    = lipgloss.NewStyle().Foreground(neonRed).Bold(true)

	// Neon panel style - thick red border
	neonPanelStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(neonRed).
			Padding(0, 1)

	// Neon panel style - cyan variant (no border for right panels)
	neonPanelCyanStyle = lipgloss.NewStyle().
				Padding(0, 1)

	// Neon panel title style - red accent
	neonPanelTitleStyle = lipgloss.NewStyle().
				Foreground(neonRed).
				Bold(true).
				PaddingBottom(0).
				BorderBottom(true).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(neonDarkDim).
				MarginBottom(0)

	// Neon section style - cyan borders for inner sections
	neonSectionStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(neonCyan).
				Padding(0, 1)

	// Neon header style - cyan
	neonHeaderStyle = lipgloss.NewStyle().
			Foreground(neonCyan).
			Bold(true)

	// Neon subtitle style - cyan italic
	neonSubtitleStyle = lipgloss.NewStyle().
				Foreground(neonCyan).
				Italic(true)

	// Neon team style - cyan for team names
	neonTeamStyle = lipgloss.NewStyle().
			Foreground(neonCyan).
			Bold(true)

	// Neon score style - red for emphasis
	neonScoreStyle = lipgloss.NewStyle().
			Foreground(neonRed).
			Bold(true)

	// Neon value style - white text
	neonValueStyle = lipgloss.NewStyle().
			Foreground(neonWhite)

	// Neon dim style - gray text
	neonDimStyle = lipgloss.NewStyle().
			Foreground(neonDim)

	// Neon label style - dim with fixed width
	neonLabelStyle = lipgloss.NewStyle().
			Foreground(neonDim).
			Width(14)

	// Neon list item styles
	neonListItemStyle = lipgloss.NewStyle().
				Foreground(neonWhite).
				Padding(0, 1)

	neonListItemSelectedStyle = lipgloss.NewStyle().
					Foreground(neonRed).
					Bold(true).
					Padding(0, 1)

	// Neon status styles
	neonLiveStyle = lipgloss.NewStyle().
			Foreground(neonRed).
			Bold(true)

	neonFinishedStyle = lipgloss.NewStyle().
				Foreground(neonCyan)

	// Neon separator style
	neonSeparatorStyle = lipgloss.NewStyle().
				Foreground(neonRed).
				Padding(0, 1)

	// Neon empty state style
	neonEmptyStyle = lipgloss.NewStyle().
			Foreground(neonDim).
			Padding(2, 2).
			Align(lipgloss.Center)

	// Neon date selector styles
	neonDateSelectedStyle = lipgloss.NewStyle().
				Foreground(neonRed).
				Bold(true).
				Padding(0, 1)

	neonDateUnselectedStyle = lipgloss.NewStyle().
				Foreground(neonDim).
				Padding(0, 1)
)

// FilterInputStyles returns cursor and prompt styles for list filter input.
// Cursor: neon cyan (solid color), Prompt: neon red to match theme.
func FilterInputStyles() (cursorStyle, promptStyle lipgloss.Style) {
	cursorStyle = lipgloss.NewStyle().
		Foreground(neonCyan).
		Bold(true)
	promptStyle = lipgloss.NewStyle().
		Foreground(neonRed).
		Bold(true)
	return cursorStyle, promptStyle
}

// AdaptiveGradientColors returns the appropriate gradient start/end hex colors
// based on the terminal background (light or dark).
// This is a convenience wrapper around design.AdaptiveGradientColors.
func AdaptiveGradientColors() (startHex, endHex string) {
	return design.AdaptiveGradientColors()
}
