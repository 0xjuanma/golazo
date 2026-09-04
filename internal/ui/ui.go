package ui

import (
	"strings"

	"github.com/0xjuanma/golazo/internal/constants"
	"github.com/0xjuanma/golazo/internal/data"
	"github.com/0xjuanma/golazo/internal/ui/design"
	"github.com/charmbracelet/lipgloss"
)

// Truncate truncates text to fit the specified display width, appending "..."
// if truncated. Width is measured in terminal cells (via lipgloss.Width) so that
// multi-byte glyphs such as the arrow keys in help strings (↑/↓/←/→) are counted
// by how wide they render, not by their byte length.
func Truncate(text string, width int) string {
	if lipgloss.Width(text) <= width {
		return text
	}
	// Not enough room for the ellipsis itself: return what fits, cell by cell.
	if width <= 3 {
		return truncateToWidth(text, width)
	}
	return truncateToWidth(text, width-3) + "..."
}

// truncateToWidth returns the longest prefix of text whose display width does
// not exceed width, counting in terminal cells.
func truncateToWidth(text string, width int) string {
	if width <= 0 {
		return ""
	}
	var b strings.Builder
	used := 0
	for _, r := range text {
		rw := lipgloss.Width(string(r))
		if used+rw > width {
			break
		}
		b.WriteRune(r)
		used += rw
	}
	return b.String()
}

// renderStatusBanner renders a status banner based on the specified type.
// Returns an empty string if no banner should be displayed.
// The banner is styled with cyan color, bold text, and center alignment.
// The new version banner uses a gradient effect.
func renderStatusBanner(bannerType constants.StatusBannerType, width int) string {
	var message string

	switch bannerType {
	case constants.StatusBannerDebug:
		message = "[DEBUG MODE] Logs: " + data.DebugLogPath()
	case constants.StatusBannerNewVersion:
		message = "New Version Available! Run 'golazo --update'"
	case constants.StatusBannerDev:
		message = "[DEV BUILD] This is a development version"
	case constants.StatusBannerNone:
		fallthrough
	default:
		return "" // No banner for None or unknown types
	}

	var styledMessage string

	if bannerType == constants.StatusBannerNewVersion {
		// Apply gradient to new version banner (cyan → red, adaptive)
		styledMessage = design.ApplyGradientToText(message)
	} else {
		// Use simple cyan styling for other banners
		bannerStyle := lipgloss.NewStyle().
			Foreground(neonCyan).
			Bold(true)
		styledMessage = bannerStyle.Render(message)
	}

	// Center the banner in the available width
	containerStyle := lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center)

	return containerStyle.Render(styledMessage)
}
