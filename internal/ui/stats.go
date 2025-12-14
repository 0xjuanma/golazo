package ui

import (
	"fmt"
	"strings"

	"github.com/0xjuanma/golazo/internal/api"
	"github.com/0xjuanma/golazo/internal/constants"
	"github.com/charmbracelet/lipgloss"
)

// RenderStatsView renders the stats view with finished matches and high-level statistics.
func RenderStatsView(width, height int, matches []MatchDisplay, selected int, details *api.MatchDetails) string {
	// Calculate panel dimensions
	// Left side: 40% width (matches list)
	// Right side: 60% width (match stats)
	leftWidth := width * 40 / 100
	if leftWidth < 30 {
		leftWidth = 30
	}
	rightWidth := width - leftWidth - 1
	if rightWidth < 40 {
		rightWidth = 40
		leftWidth = width - rightWidth - 1
	}

	panelHeight := height - 2

	// Render left panel (finished matches list)
	leftPanel := renderFinishedMatchesPanel(leftWidth, panelHeight, matches, selected)

	// Render right panel (match stats)
	rightPanel := renderMatchStatsPanel(rightWidth, panelHeight, details)

	// Create modern neon vertical separator (matching live matches view style)
	separatorStyle := lipgloss.NewStyle().
		Foreground(borderColor).
		Height(panelHeight).
		Padding(0, 1)
	separator := separatorStyle.Render("│")

	// Combine panels horizontally
	combined := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftPanel,
		separator,
		rightPanel,
	)

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, combined)
}

// renderFinishedMatchesPanel renders the left panel with finished matches list.
func renderFinishedMatchesPanel(width, height int, matches []MatchDisplay, selected int) string {
	title := panelTitleStyle.Width(width - 6).Render(constants.PanelFinishedMatches)

	items := make([]string, 0, len(matches))
	contentWidth := width - 6

	if len(matches) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(dimColor).
			Padding(1, 0).
			Align(lipgloss.Center).
			Width(contentWidth)
		items = append(items, emptyStyle.Render(constants.EmptyNoFinishedMatches))
	} else {
		for i, match := range matches {
			item := renderFinishedMatchListItem(match, i == selected, contentWidth)
			items = append(items, item)
		}
	}

	content := lipgloss.JoinVertical(lipgloss.Left, items...)

	panelContent := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		content,
	)

	panel := panelStyle.
		Width(width).
		Height(height).
		Render(panelContent)

	return panel
}

// renderFinishedMatchListItem renders a single finished match list item.
func renderFinishedMatchListItem(match MatchDisplay, selected bool, width int) string {
	style := matchListItemStyle
	if selected {
		style = matchListItemSelectedStyle
	}

	// Format: "Team A 2 - 1 Team B"
	scoreText := "vs"
	if match.HomeScore != nil && match.AwayScore != nil {
		scoreText = fmt.Sprintf("%d - %d", *match.HomeScore, *match.AwayScore)
	}

	homeTeam := Truncate(match.HomeTeam.ShortName, 12)
	awayTeam := Truncate(match.AwayTeam.ShortName, 12)

	matchLine := fmt.Sprintf("%s %s %s", homeTeam, scoreText, awayTeam)

	// Add league name below
	leagueName := Truncate(match.League.Name, width-4)

	// Add date if available
	dateText := ""
	if match.MatchTime != nil {
		dateText = match.MatchTime.Format("Jan 2")
	}

	item := matchLine
	if leagueName != "" {
		item += "\n" + lipgloss.NewStyle().
			Foreground(dimColor).
			Render(leagueName)
	}
	if dateText != "" {
		item += " " + lipgloss.NewStyle().
			Foreground(dimColor).
			Render(dateText)
	}

	return style.Width(width).Render(item)
}

// renderMatchStatsPanel renders the right panel with high-level match statistics.
func renderMatchStatsPanel(width, height int, details *api.MatchDetails) string {
	title := panelTitleStyle.Width(width - 6).Render(constants.PanelMatchStatistics)

	if details == nil {
		emptyStyle := lipgloss.NewStyle().
			Foreground(dimColor).
			Padding(1, 0).
			Align(lipgloss.Center).
			Width(width - 6)
		content := lipgloss.JoinVertical(
			lipgloss.Center,
			emptyStyle.Render(constants.EmptySelectMatch),
		)
		panelContent := lipgloss.JoinVertical(
			lipgloss.Left,
			title,
			content,
		)
		panel := panelStyle.
			Width(width).
			Height(height).
			Render(panelContent)
		return panel
	}

	contentWidth := width - 6
	stats := make([]string, 0)

	// Match header
	matchHeader := renderMatchHeader(details, contentWidth)
	stats = append(stats, matchHeader)

	// Score
	if details.HomeScore != nil && details.AwayScore != nil {
		scoreText := fmt.Sprintf("%d - %d", *details.HomeScore, *details.AwayScore)
		stats = append(stats, lipgloss.NewStyle().
			Width(contentWidth).
			Align(lipgloss.Center).
			MarginTop(0).
			MarginBottom(0).
			Render(matchScoreStyle.Render(scoreText)))
	}

	// League and date
	info := make([]string, 0)
	if details.League.Name != "" {
		info = append(info, lipgloss.NewStyle().
			Foreground(dimColor).
			Render(constants.LabelLeague+details.League.Name))
	}
	if details.MatchTime != nil {
		info = append(info, lipgloss.NewStyle().
			Foreground(dimColor).
			Render(constants.LabelDate+details.MatchTime.Format("Jan 2, 2006")))
	}
	if len(info) > 0 {
		stats = append(stats, strings.Join(info, " | "))
	}

	// Status
	statusDisplay := lipgloss.NewStyle().
		Foreground(secondaryColor).
		Bold(true).
		Render(constants.LabelStatus + constants.StatusFinishedText)
	stats = append(stats, statusDisplay)

	// Events summary
	if len(details.Events) > 0 {
		// Count events by type
		goals := 0
		cards := 0
		for _, event := range details.Events {
			if event.Type == "goal" {
				goals++
			} else if event.Type == "card" {
				cards++
			}
		}

		if goals > 0 || cards > 0 {
			stats = append(stats, "")
			summary := make([]string, 0)
			if goals > 0 {
				summary = append(summary, lipgloss.NewStyle().
					Foreground(secondaryColor).
					Render(fmt.Sprintf("Goals: %d", goals)))
			}
			if cards > 0 {
				summary = append(summary, lipgloss.NewStyle().
					Foreground(secondaryColor).
					Render(fmt.Sprintf("Cards: %d", cards)))
			}
			stats = append(stats, strings.Join(summary, " | "))
		}

		// Show recent events (last 5)
		recentEvents := details.Events
		if len(recentEvents) > 5 {
			recentEvents = recentEvents[len(recentEvents)-5:]
		}

		if len(recentEvents) > 0 {
			stats = append(stats, "")
			for _, event := range recentEvents {
				eventText := renderStatsEvent(event)
				stats = append(stats, eventText)
			}
		}
	}

	content := lipgloss.JoinVertical(lipgloss.Left, stats...)

	panelContent := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		content,
	)

	panel := panelStyle.
		Width(width).
		Height(height).
		Render(panelContent)

	return panel
}

// renderMatchHeader renders the match header with team names.
func renderMatchHeader(details *api.MatchDetails, width int) string {
	homeTeam := Truncate(details.HomeTeam.ShortName, 20)
	awayTeam := Truncate(details.AwayTeam.ShortName, 20)

	header := fmt.Sprintf("%s vs %s", homeTeam, awayTeam)
	return matchTitleStyle.Width(width).Render(header)
}

// renderStatsEvent renders a single match event for the stats view.
func renderStatsEvent(event api.MatchEvent) string {
	minute := eventMinuteStyle.Render(fmt.Sprintf("%d'", event.Minute))

	eventText := ""
	switch event.Type {
	case "goal":
		player := ""
		if event.Player != nil {
			player = *event.Player
		}
		teamName := event.Team.ShortName
		eventText = eventGoalStyle.Render(fmt.Sprintf("Goal: %s - %s", teamName, player))
	case "card":
		cardType := "Yellow"
		if event.EventType != nil && *event.EventType == "red" {
			cardType = "Red"
		}
		player := ""
		if event.Player != nil {
			player = *event.Player
		}
		teamName := event.Team.ShortName
		eventText = eventCardStyle.Render(fmt.Sprintf("Card (%s): %s - %s", cardType, teamName, player))
	default:
		eventText = eventTextStyle.Render(fmt.Sprintf("%s - %s", event.Team.ShortName, event.Type))
	}

	return lipgloss.JoinHorizontal(lipgloss.Left, minute, eventText)
}
