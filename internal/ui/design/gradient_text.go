package design

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/lucasb-eyer/go-colorful"
)

// ApplyGradientToText applies a gradient color to text, character by character.
func ApplyGradientToText(text string) string {
	startHex, endHex := AdaptiveGradientColors()
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
		ratio := float64(i) / float64(max(len(runes)-1, 1))
		color := startColor.BlendLab(endColor, ratio)
		hexColor := color.Hex()
		charStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(hexColor)).Bold(true)
		result.WriteString(charStyle.Render(string(char)))
	}

	return result.String()
}

// ApplyGradientToMultilineText applies a gradient color to multi-line text.
// Each line gets a color based on its position in the text (line-by-line gradient).
func ApplyGradientToMultilineText(text string) string {
	lines := strings.Split(text, "\n")
	if len(lines) == 0 {
		return text
	}

	startHex, endHex := AdaptiveGradientColors()
	startColor, err1 := colorful.Hex(startHex)
	endColor, err2 := colorful.Hex(endHex)
	if err1 != nil || err2 != nil {
		return text
	}

	var result strings.Builder
	for i, line := range lines {
		if line == "" {
			result.WriteString("\n")
			continue
		}

		// Calculate gradient position for this line (0.0 to 1.0)
		ratio := float64(i) / float64(max(len(lines)-1, 1))

		// Blend colors based on line position
		color := startColor.BlendLab(endColor, ratio)
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
