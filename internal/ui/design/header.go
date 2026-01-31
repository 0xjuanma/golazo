// Package design provides reusable UI design components.
package design

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/lucasb-eyer/go-colorful"
)

const diag = `╱`

// RenderHeader renders a header with gradient text followed by diagonal fill.
// text is the header text to display.
// width is the total width to fill.
// Returns a styled header string with gradient text and diagonal lines.
func RenderHeader(text string, width int) string {
	return renderHeaderWithFocus(text, width, true)
}

// RenderHeaderDim renders a dimmed header for unfocused state.
// text is the header text to display.
// width is the total width to fill.
func RenderHeaderDim(text string, width int) string {
	return renderHeaderWithFocus(text, width, false)
}

// renderHeaderWithFocus renders header with gradient or dim styling based on focus.
func renderHeaderWithFocus(text string, width int, focused bool) string {
	startHex, endHex := AdaptiveGradientColors()

	var title string
	var diagColor string

	if focused {
		title = applyHeaderGradient(text, startHex, endHex)
		diagColor = startHex
	} else {
		// Dim style for unfocused
		dimStyle := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#666666", Dark: "#555555"}).Bold(true)
		title = dimStyle.Render(text)
		diagColor = "#555555"
	}

	remainingWidth := width - lipgloss.Width(text) - 2
	if remainingWidth > 0 {
		lines := strings.Repeat(diag, remainingWidth)
		styledLines := lipgloss.NewStyle().Foreground(lipgloss.Color(diagColor)).Render(lines)
		title = fmt.Sprintf("%s %s", title, styledLines)
	}
	return title
}

// RenderHeaderCentered renders a header with diagonal fills on both sides.
// text is the header text to display centered.
// width is the total width to fill.
func RenderHeaderCentered(text string, width int) string {
	startHex, endHex := AdaptiveGradientColors()
	title := applyHeaderGradient(text, startHex, endHex)

	textWidth := lipgloss.Width(text)
	remainingWidth := width - textWidth - 2 // 2 for spaces around text
	if remainingWidth <= 0 {
		return title
	}

	leftWidth := remainingWidth / 2
	rightWidth := remainingWidth - leftWidth

	leftLines := strings.Repeat(diag, leftWidth)
	rightLines := strings.Repeat(diag, rightWidth)

	styledLeft := lipgloss.NewStyle().Foreground(lipgloss.Color(startHex)).Render(leftLines)
	styledRight := lipgloss.NewStyle().Foreground(lipgloss.Color(endHex)).Render(rightLines)

	return fmt.Sprintf("%s %s %s", styledLeft, title, styledRight)
}

// applyHeaderGradient applies a gradient to a single line of text.
func applyHeaderGradient(text string, startHex, endHex string) string {
	startColor, err1 := colorful.Hex(startHex)
	endColor, err2 := colorful.Hex(endHex)
	if err1 != nil || err2 != nil {
		return text
	}

	runes := []rune(text)
	if len(runes) == 0 {
		return text
	}

	var result strings.Builder
	for i, char := range runes {
		if char == ' ' {
			result.WriteRune(' ')
			continue
		}
		ratio := float64(i) / float64(max(len(runes)-1, 1))
		color := startColor.BlendLab(endColor, ratio)
		hexColor := color.Hex()
		charStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(hexColor)).Bold(true)
		result.WriteString(charStyle.Render(string(char)))
	}

	return result.String()
}
