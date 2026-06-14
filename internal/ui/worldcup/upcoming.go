package worldcup

import (
	"fmt"
	"strings"
	"time"

	"github.com/0xjuanma/golazo/internal/api"
	"github.com/0xjuanma/golazo/internal/ui/design"
	"github.com/charmbracelet/lipgloss"
)

// wcUpcomingDateHeaderFormat is the format used for the per-date headers in
// the upcoming-matches view. Exposed at package level so tests can derive
// expected headers without duplicating layout strings.
const wcUpcomingDateHeaderFormat = "Mon 02 Jan"

// RenderUpcoming renders the World Cup upcoming-matches sub-view. Matches are
// grouped by local kickoff date with one header per day; under each header
// fixtures are listed in ascending kickoff order with home/away short names
// and local HH:MM time.
//
// When loading, the loading style is shown; when an error occurred, lastErr
// is rendered in the error style. An empty match slice renders a friendly
// "no matches" message.
func RenderUpcoming(width, height int, matches []api.Match, loading bool, lastErr, statusBanner string) string {
	if width <= 0 {
		return ""
	}

	header := design.RenderHeader("Upcoming Matches", width-2)
	help := HelpStyle.Width(width).Render("Esc: back to grid  q: quit")

	var body string
	switch {
	case loading:
		body = LoadingStyle.Render("Loading upcoming matches…")
	case lastErr != "":
		body = ErrorStyle.Render(lastErr)
	case len(matches) == 0:
		body = lipgloss.NewStyle().Foreground(colorDim).Render("No matches in the next 4 days")
	default:
		body = renderWCUpcomingMatches(matches)
	}

	parts := []string{}
	if statusBanner != "" {
		parts = append(parts, statusBanner)
	}
	parts = append(parts, header, "", body, help)
	return padToHeight(lipgloss.JoinVertical(lipgloss.Left, parts...), height)
}

// renderWCUpcomingMatches groups matches by local kickoff date and renders
// each group under a date header. Matches must already be sorted ascending
// by MatchTime — RenderUpcoming relies on the model providing sorted data
// (see app.sortAndDedupeWCUpcoming).
func renderWCUpcomingMatches(matches []api.Match) string {
	dateHeaderStyle := lipgloss.NewStyle().Foreground(colorCyan).Bold(true)
	timeStyle := lipgloss.NewStyle().Foreground(colorGold)
	teamStyle := lipgloss.NewStyle().Foreground(colorWhite)
	vsStyle := lipgloss.NewStyle().Foreground(colorDim)
	roundStyle := lipgloss.NewStyle().Foreground(colorDim).Italic(true)

	var (
		lines       []string
		currentDate string
	)

	for _, m := range matches {
		if m.MatchTime == nil {
			continue
		}
		local := m.MatchTime.Local()
		dateKey := local.Format(wcUpcomingDateHeaderFormat)

		if dateKey != currentDate {
			if currentDate != "" {
				lines = append(lines, "")
			}
			lines = append(lines, dateHeaderStyle.Render(dateKey))
			currentDate = dateKey
		}

		timeStr := timeStyle.Render(local.Format("15:04"))
		home := teamStyle.Render(TeamLabel(m.HomeTeam))
		away := teamStyle.Render(TeamLabel(m.AwayTeam))
		line := fmt.Sprintf("  %s  %s %s %s", timeStr, home, vsStyle.Render("vs"), away)

		if m.Round != "" {
			line += "  " + roundStyle.Render(roundLabel(m.Round))
		}
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// upcomingFormatDateHeader is exposed for tests to assert the date header
// format matches the layout used by RenderUpcoming.
func upcomingFormatDateHeader(t time.Time) string {
	return t.Format(wcUpcomingDateHeaderFormat)
}
