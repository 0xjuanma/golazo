// Package ui provides rendering functions for the terminal user interface.
package ui

import (
	"strings"

	"github.com/0xjuanma/golazo/internal/constants"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/lucasb-eyer/go-colorful"
)

var (
	// Menu styles
	menuItemStyle = lipgloss.NewStyle().
			Foreground(textColor).
			Padding(0, 0)

	menuItemSelectedStyle = lipgloss.NewStyle().
				Foreground(highlightColor).
				Bold(true).
				Padding(0, 0)

	menuTitleStyle = lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true).
			Align(lipgloss.Center).
			Padding(0, 0)

	menuHelpStyle = lipgloss.NewStyle().
			Foreground(dimColor).
			Align(lipgloss.Center).
			Padding(0, 0)
)

// RenderMainMenu renders the main menu view with navigation options.
// width and height specify the terminal dimensions.
// selected indicates which menu item is currently selected (0-indexed).
// sp is the spinner model to display when loading (for other views).
// randomSpinner is the random character spinner for main view.
// loading indicates if the spinner should be shown.
// bannerType determines what status banner (if any) to display at the top.
func RenderMainMenu(width, height, selected int, sp spinner.Model, randomSpinner *RandomCharSpinner, loading bool, bannerType constants.StatusBannerType) string {
	menuItems := []string{
		constants.MenuStats,
		constants.MenuLiveMatches,
		constants.MenuSettings,
	}

	items := make([]string, 0, len(menuItems))
	for i, item := range menuItems {
		if i == selected {
			items = append(items, menuItemSelectedStyle.Render(item))
		} else {
			items = append(items, menuItemStyle.Render(item))
		}
	}

	menuContent := strings.Join(items, "\n")

	// Apply gradient to ASCII title (cyan to red, same as spinner)
	title := renderGradientText(constants.ASCIITitle)
	help := menuHelpStyle.Render(constants.HelpMainMenu)

	// Spinner with fixed spacing - always reserve space to prevent movement
	// Use multiple spinner instances for a longer, more prominent animation
	spinnerStyle := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Height(1).
		Padding(0, 0)

	var spinnerContent string
	if loading && randomSpinner != nil {
		// Use random character spinner for main view
		spinnerContent = spinnerStyle.Render(randomSpinner.View())
	} else {
		// Reserve space even when not loading to prevent menu movement
		spinnerContent = spinnerStyle.Render("")
	}

	// Add status banner if needed
	statusBanner := renderStatusBanner(bannerType, width)
	if statusBanner != "" {
		statusBanner += "\n"
	}

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		"\n",
		statusBanner,
		spinnerContent,
		"\n",
		menuContent,
		"\n\n",
		help,
	)

	return lipgloss.Place(
		width,
		height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}

// RenderGradientText applies a gradient (cyan to red) to multi-line text.
// Exported wrapper for external use.
func RenderGradientText(text string) string {
	return renderGradientText(text)
}

// renderGradientText applies a gradient (cyan to red) to multi-line text.
// Uses adaptive colors based on terminal background for visibility.
func renderGradientText(text string) string {
	lines := strings.Split(text, "\n")
	if len(lines) == 0 {
		return text
	}

	// Create gradient colors - adaptive based on terminal background
	startHex, endHex := AdaptiveGradientColors()
	startColor, _ := colorful.Hex(startHex) // Cyan (adaptive)
	endColor, _ := colorful.Hex(endHex)     // Red (adaptive)

	var result strings.Builder
	for i, line := range lines {
		if line == "" {
			result.WriteString("\n")
			continue
		}

		// Calculate gradient position for this line (0.0 to 1.0)
		ratio := float64(i) / float64(len(lines)-1)

		// Blend colors based on line position
		color := startColor.BlendLab(endColor, ratio)

		// Convert to hex for lipgloss
		hexColor := color.Hex()

		// Style the line with gradient color
		lineStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(hexColor))
		result.WriteString(lineStyle.Render(line))
		if i < len(lines)-1 {
			result.WriteString("\n")
		}
	}

	return result.String()
}
