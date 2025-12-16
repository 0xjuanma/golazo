package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// Neon design styles - Golazo red/cyan theme
// Bold, vibrant design with thick borders and high contrast.

var (
	// Neon color palette - Golazo brand
	neonRed     = lipgloss.Color("196") // Bright red
	neonCyan    = lipgloss.Color("51")  // Electric cyan
	neonWhite   = lipgloss.Color("255") // Pure white
	neonDark    = lipgloss.Color("236") // Dark background
	neonDim     = lipgloss.Color("244") // Gray dim text
	neonDarkDim = lipgloss.Color("239") // Slightly lighter dark

	// Neon panel style - thick red border
	neonPanelStyle = lipgloss.NewStyle().
			Border(lipgloss.ThickBorder()).
			BorderForeground(neonRed).
			Padding(0, 1)

	// Neon panel style - cyan variant
	neonPanelCyanStyle = lipgloss.NewStyle().
				Border(lipgloss.ThickBorder()).
				BorderForeground(neonCyan).
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

// NeonPanelStyle returns the neon panel style.
func NeonPanelStyle() lipgloss.Style {
	return neonPanelStyle
}

// NeonPanelCyanStyle returns the cyan variant neon panel style.
func NeonPanelCyanStyle() lipgloss.Style {
	return neonPanelCyanStyle
}

// NeonSectionStyle returns the neon section style for inner bordered sections.
func NeonSectionStyle() lipgloss.Style {
	return neonSectionStyle
}

// NeonHeaderStyle returns the neon header style.
func NeonHeaderStyle() lipgloss.Style {
	return neonHeaderStyle
}

// NeonTeamStyle returns the neon team name style.
func NeonTeamStyle() lipgloss.Style {
	return neonTeamStyle
}

// NeonScoreStyle returns the neon score style.
func NeonScoreStyle() lipgloss.Style {
	return neonScoreStyle
}

// NeonDimStyle returns the neon dim text style.
func NeonDimStyle() lipgloss.Style {
	return neonDimStyle
}

// NeonValueStyle returns the neon value text style.
func NeonValueStyle() lipgloss.Style {
	return neonValueStyle
}

// NeonLabelStyle returns the neon label style.
func NeonLabelStyle() lipgloss.Style {
	return neonLabelStyle
}

// NeonSeparatorStyle returns the neon separator style.
func NeonSeparatorStyle() lipgloss.Style {
	return neonSeparatorStyle
}

// NeonPanelTitleStyle returns the neon panel title style.
func NeonPanelTitleStyle() lipgloss.Style {
	return neonPanelTitleStyle
}

// NeonEmptyStyle returns the neon empty state style.
func NeonEmptyStyle() lipgloss.Style {
	return neonEmptyStyle
}

// NeonLiveStyle returns the neon live status style.
func NeonLiveStyle() lipgloss.Style {
	return neonLiveStyle
}

// NeonFinishedStyle returns the neon finished status style.
func NeonFinishedStyle() lipgloss.Style {
	return neonFinishedStyle
}

// NeonSubtitleStyle returns the neon subtitle style.
func NeonSubtitleStyle() lipgloss.Style {
	return neonSubtitleStyle
}

// NeonListItemStyle returns the neon list item style.
func NeonListItemStyle() lipgloss.Style {
	return neonListItemStyle
}

// NeonListItemSelectedStyle returns the neon selected list item style.
func NeonListItemSelectedStyle() lipgloss.Style {
	return neonListItemSelectedStyle
}

// NeonDateSelectedStyle returns the selected date option style.
func NeonDateSelectedStyle() lipgloss.Style {
	return neonDateSelectedStyle
}

// NeonDateUnselectedStyle returns the unselected date option style.
func NeonDateUnselectedStyle() lipgloss.Style {
	return neonDateUnselectedStyle
}
