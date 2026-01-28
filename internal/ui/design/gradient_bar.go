// Package design provides reusable visual design components.
package design

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/lucasb-eyer/go-colorful"
)

// GradientBarConfig configures how a gradient comparison bar is rendered.
type GradientBarConfig struct {
	Width      int     // Total width of the bar
	HomeValue  float64 // Home team value (for ratio calculation)
	AwayValue  float64 // Away team value (for ratio calculation)
	StartColor string  // Hex color for start (home side), e.g., "#00FFFF"
	EndColor   string  // Hex color for end (away side), e.g., "#FF0055"
	FilledChar string  // Character for filled portion (default: "█")
	EmptyChar  string  // Character for empty portion (default: "░")
}

// DefaultGradientBarConfig returns a default configuration with theme colors.
func DefaultGradientBarConfig(width int, homeVal, awayVal float64) GradientBarConfig {
	startHex, endHex := AdaptiveGradientColors()
	return GradientBarConfig{
		Width:      width,
		HomeValue:  homeVal,
		AwayValue:  awayVal,
		StartColor: startHex,
		EndColor:   endHex,
		FilledChar: "█",
		EmptyChar:  "░",
	}
}

// AdaptiveGradientColors returns the appropriate gradient start/end hex colors
// based on the terminal background (light or dark).
// Dark terminals get bright vibrant colors, light terminals get darker saturated colors.
func AdaptiveGradientColors() (startHex, endHex string) {
	if lipgloss.HasDarkBackground() {
		// Dark terminal: bright cyan to bright red
		return "#00FFFF", "#FF0000"
	}
	// Light terminal: darker cyan (30% darker) to darker red for better visibility
	return "#006161", "#8B0000"
}

// RenderGradientBar creates a comparison bar with gradient coloring.
// The bar shows proportional representation of two values with smooth color transition.
func RenderGradientBar(cfg GradientBarConfig) string {
	if cfg.FilledChar == "" {
		cfg.FilledChar = "█"
	}
	if cfg.EmptyChar == "" {
		cfg.EmptyChar = "░"
	}
	if cfg.Width <= 0 {
		cfg.Width = 20
	}

	// Calculate proportions
	total := cfg.HomeValue + cfg.AwayValue
	if total == 0 {
		total = 1 // Avoid division by zero
	}

	halfWidth := cfg.Width / 2
	homeFilledWidth := int((cfg.HomeValue / total) * float64(cfg.Width))
	awayFilledWidth := int((cfg.AwayValue / total) * float64(cfg.Width))

	// Ensure at least 1 if value > 0
	if cfg.HomeValue > 0 && homeFilledWidth == 0 {
		homeFilledWidth = 1
	}
	if cfg.AwayValue > 0 && awayFilledWidth == 0 {
		awayFilledWidth = 1
	}

	// Cap at half width for each side
	if homeFilledWidth > halfWidth {
		homeFilledWidth = halfWidth
	}
	if awayFilledWidth > halfWidth {
		awayFilledWidth = halfWidth
	}

	// Parse colors
	startColor, err1 := colorful.Hex(cfg.StartColor)
	endColor, err2 := colorful.Hex(cfg.EndColor)
	if err1 != nil || err2 != nil {
		// Fallback to simple bars without gradient
		homeBar := strings.Repeat(cfg.FilledChar, homeFilledWidth) + strings.Repeat(cfg.EmptyChar, halfWidth-homeFilledWidth)
		awayBar := strings.Repeat(cfg.FilledChar, awayFilledWidth) + strings.Repeat(cfg.EmptyChar, halfWidth-awayFilledWidth)
		return homeBar + "│" + awayBar
	}

	// Build home side bar with gradient (left to center)
	var homeBar strings.Builder
	for i := 0; i < halfWidth; i++ {
		ratio := float64(i) / float64(halfWidth-1)
		midColor := startColor.BlendLab(endColor, ratio*0.5) // Blend to middle
		hexColor := midColor.Hex()
		charStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(hexColor))

		if i < homeFilledWidth {
			homeBar.WriteString(charStyle.Render(cfg.FilledChar))
		} else {
			// Empty portion - dim
			homeBar.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#444444")).Render(cfg.EmptyChar))
		}
	}

	// Build away side bar with gradient (center to right)
	var awayBar strings.Builder
	for i := 0; i < halfWidth; i++ {
		ratio := 0.5 + (float64(i) / float64(halfWidth-1) * 0.5) // 0.5 to 1.0
		midColor := startColor.BlendLab(endColor, ratio)
		hexColor := midColor.Hex()
		charStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(hexColor))

		if i < awayFilledWidth {
			awayBar.WriteString(charStyle.Render(cfg.FilledChar))
		} else {
			// Empty portion - dim
			awayBar.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#444444")).Render(cfg.EmptyChar))
		}
	}

	separator := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render("│")
	return homeBar.String() + separator + awayBar.String()
}

// RenderSimpleGradientBar creates a single-direction gradient bar.
// Useful for percentage displays like possession.
func RenderSimpleGradientBar(value float64, width int) string {
	startHex, endHex := AdaptiveGradientColors()
	startColor, _ := colorful.Hex(startHex)
	endColor, _ := colorful.Hex(endHex)

	filledWidth := int(value * float64(width))
	if filledWidth > width {
		filledWidth = width
	}

	var result strings.Builder
	for i := 0; i < width; i++ {
		ratio := float64(i) / float64(width-1)
		color := startColor.BlendLab(endColor, ratio)
		hexColor := color.Hex()
		charStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(hexColor))

		if i < filledWidth {
			result.WriteString(charStyle.Render("█"))
		} else {
			result.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#444444")).Render("░"))
		}
	}

	return result.String()
}
