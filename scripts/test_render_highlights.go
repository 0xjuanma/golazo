package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Define neon colors locally for testing - adaptive for light/dark terminals
var (
	neonCyan    = lipgloss.AdaptiveColor{Light: "30", Dark: "51"}   // Dark cyan / Electric cyan
	neonWhite   = lipgloss.AdaptiveColor{Light: "235", Dark: "255"} // Dark gray / White
	neonDarkDim = lipgloss.Color("239")
)

func main() {
	fmt.Println("Testing highlights rendering...")

	// Test the highlights rendering logic
	width := 80
	contentWidth := width - 6

	var content strings.Builder

	// Test highlights section (simulating the condition being met)
	fmt.Println("✅ Simulating highlights condition met")

	highlightsTitle := lipgloss.NewStyle().
		Foreground(neonCyan).
		Bold(true).
		PaddingTop(0).
		BorderBottom(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(neonDarkDim).
		Width(width - 6).
		Render("Highlights")
	content.WriteString(highlightsTitle)
	content.WriteString("\n")
	fmt.Printf("Title rendered: %q\n", highlightsTitle)

	// Simulate hyperlink (just return the text for now)
	highlightText := "📹 Test Match Highlights"
	highlightLink := highlightText // ui.Hyperlink(highlightText, "https://example.com")

	highlightLine := lipgloss.NewStyle().
		Foreground(neonWhite).
		Width(contentWidth).
		Render(highlightLink)
	content.WriteString(highlightLine)
	content.WriteString("\n\n")

	fmt.Printf("Highlight line rendered: %q\n", highlightLine)
	fmt.Printf("Full content:\n%s\n", content.String())
}
