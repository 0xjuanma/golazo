package worldcup

import (
	"github.com/0xjuanma/golazo/internal/api"
	"github.com/0xjuanma/golazo/internal/ui/design"
	"github.com/charmbracelet/lipgloss"
)

// RenderUpcoming renders the World Cup upcoming-matches sub-view. This is a
// placeholder rendering used while phase 2 lands the MVU wiring; the final
// grouped-by-date layout is added in phase 3.
func RenderUpcoming(width, height int, matches []api.Match, loading bool, lastErr, statusBanner string) string {
	if width <= 0 {
		return ""
	}

	header := design.RenderHeader("Upcoming Matches", width-2)
	help := HelpStyle.Width(width).Render("Esc: back to groups  q: quit")

	var body string
	switch {
	case loading:
		body = LoadingStyle.Render("Loading upcoming matches…")
	case lastErr != "":
		body = ErrorStyle.Render(lastErr)
	case len(matches) == 0:
		body = lipgloss.NewStyle().Foreground(colorDim).Render("No matches in the next 3 days")
	default:
		// Placeholder: render only the count; full layout lands in phase 3.
		body = lipgloss.NewStyle().Foreground(colorDim).Render(
			"Loaded fixtures: placeholder view (final layout coming in phase 3)",
		)
	}

	parts := []string{}
	if statusBanner != "" {
		parts = append(parts, statusBanner)
	}
	parts = append(parts, header, "", body, help)
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}
