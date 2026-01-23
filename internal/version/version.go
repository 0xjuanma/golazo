package version

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/0xjuanma/golazo/internal/constants"
	"github.com/0xjuanma/golazo/internal/ui"
	"github.com/charmbracelet/lipgloss"
	"github.com/lucasb-eyer/go-colorful"
)

// IsOlder returns true if versionA is older than versionB
// Supports semantic versions like "v1.2.3" or "1.2.3"
func IsOlder(versionA, versionB string) bool {
	// Remove 'v' prefix for comparison
	va := strings.TrimPrefix(versionA, "v")
	vb := strings.TrimPrefix(versionB, "v")

	// Split by dots
	partsA := strings.Split(va, ".")
	partsB := strings.Split(vb, ".")

	// Compare each part numerically
	for i := 0; i < len(partsA) && i < len(partsB); i++ {
		numA, errA := strconv.Atoi(partsA[i])
		numB, errB := strconv.Atoi(partsB[i])

		if errA != nil || errB != nil {
			// If parsing fails, fall back to string comparison
			return va < vb
		}

		if numA < numB {
			return true
		}
		if numA > numB {
			return false
		}
	}

	// If all compared parts are equal, longer version is considered newer
	return len(partsA) < len(partsB)
}

// Print displays the ASCII logo with gradient and version information.
func Print(version string) {
	// Render ASCII title with gradient (same as main view, adaptive)
	title := ui.RenderGradientText(constants.ASCIITitle)

	// Render version with gradient color (use the end color - red, adaptive)
	_, endHex := ui.AdaptiveGradientColors()
	endColor, _ := colorful.Hex(endHex)
	versionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(endColor.Hex()))
	versionText := versionStyle.Render(version)

	// Concatenate version after the last line of ASCII art
	fmt.Println(title + "" + versionText)
}
