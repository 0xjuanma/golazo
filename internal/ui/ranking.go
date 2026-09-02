package ui

import (
	"fmt"
	"strings"

	"github.com/0xjuanma/golazo/internal/api"
	"github.com/0xjuanma/golazo/internal/constants"
	"github.com/charmbracelet/lipgloss"
)

// RenderFIFARanking renders a scrollable official FIFA ranking table.
func RenderFIFARanking(width, height int, ranking *api.FIFARanking, gender string, loading bool, errText string, offset, selected int, bannerType constants.StatusBannerType) string {
	panelWidth := max(min(width-4, 72), 20)
	panelHeight := max(height-5, 5)
	// neonPanelCyanStyle adds one cell of padding on each side.
	contentWidth := max(panelWidth-2, 18)

	activeTab := 0
	if gender == "women" {
		activeTab = 1
	}
	title := neonHeaderStyle.Width(contentWidth).Align(lipgloss.Center).
		Render("FIFA World Ranking")
	tabs := renderTabBar([]string{"Men", "Women"}, activeTab, contentWidth)

	var body string
	switch {
	case loading:
		body = neonDimStyle.Render(constants.LoadingFetching)
	case errText != "":
		body = neonEmptyStyle.Render(errText + "\n\n" + constants.ErrorRetryHint)
	case ranking == nil || len(ranking.Teams) == 0:
		body = neonEmptyStyle.Render("No ranking data available")
	default:
		visible := max(height-13, 1)
		offset = max(0, min(offset, max(len(ranking.Teams)-visible, 0)))
		end := min(offset+visible, len(ranking.Teams))
		headerStyle := lipgloss.NewStyle().Foreground(neonDim).Bold(true)
		lines := []string{headerStyle.Render(formatRankingRow(
			contentWidth, "Rank", "Country", "Points", "+/-",
		))}
		selectedStyle := lipgloss.NewStyle().Foreground(neonCyan).Bold(true)
		for i, team := range ranking.Teams[offset:end] {
			change := fmt.Sprintf("%+.2f", team.PointsDiff)
			if team.PointsDiff == 0 {
				change = "—"
			}
			row := formatRankingRow(
				contentWidth,
				fmt.Sprintf("%d", team.Rank),
				team.Name,
				fmt.Sprintf("%.2f", team.TotalPoints),
				change,
			)
			if offset+i == selected {
				lines = append(lines, selectedStyle.Render(row))
			} else {
				lines = append(lines, neonValueStyle.Render(row))
			}
		}
		body = strings.Join(lines, "\n")
		title += "\n" + neonDimStyle.Width(contentWidth).Align(lipgloss.Center).
			Render("Released "+ranking.Period.Name)
	}

	panel := neonPanelCyanStyle.Width(panelWidth).Height(panelHeight).
		Render(lipgloss.JoinVertical(lipgloss.Center, title, tabs, "", body))
	help := neonDimStyle.Width(width).Align(lipgloss.Center).
		Render("↑/↓: navigate  ←/→: switch tabs  r: refresh  Esc: back  q: quit")
	banner := renderStatusBanner(bannerType, width)
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center,
		lipgloss.JoinVertical(lipgloss.Center, banner, panel, help))
}

// formatRankingRow keeps every logical team on exactly one terminal line.
// Columns are progressively removed on very narrow terminals.
func formatRankingRow(width int, rank, country, points, change string) string {
	switch {
	case width >= 42:
		countryWidth := width - 26 // rank(4), points(10), change(9), separators(3)
		return fmt.Sprintf("%-4s %-*s %10s %9s",
			rank, countryWidth, truncateString(country, countryWidth), points, change)
	case width >= 28:
		countryWidth := width - 16 // rank(4), points(10), separators(2)
		return fmt.Sprintf("%-4s %-*s %10s",
			rank, countryWidth, truncateString(country, countryWidth), points)
	default:
		countryWidth := max(width-5, 1)
		return fmt.Sprintf("%-4s %s", rank, truncateString(country, countryWidth))
	}
}
