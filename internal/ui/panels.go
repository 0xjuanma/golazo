package ui

import (
	"fmt"
	"strings"

	"github.com/0xjuanma/golazo/internal/api"
	"github.com/charmbracelet/lipgloss"
)

var (
	// Elegant panel styles with subtle borders
	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Padding(1, 2).
			Margin(0, 0)

	// Header style - elegant and refined
	panelTitleStyle = lipgloss.NewStyle().
			Foreground(textColor).
			Bold(true).
			PaddingBottom(1).
			BorderBottom(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(borderColor).
			MarginBottom(1)

	// Selection styling - elegant text color change (no background)
	matchListItemStyle = lipgloss.NewStyle().
				Foreground(textColor).
				Padding(0, 1)

	matchListItemSelectedStyle = lipgloss.NewStyle().
					Foreground(selectedColor).
					Bold(true).
					Padding(0, 1).
					BorderLeft(true).
					BorderStyle(lipgloss.ThickBorder()).
					BorderForeground(selectedColor).
					PaddingLeft(2)

	// Match details styles - refined typography
	matchTitleStyle = lipgloss.NewStyle().
			Foreground(textColor).
			Bold(true).
			MarginBottom(1)

	matchScoreStyle = lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true).
			Margin(0, 1)

	matchStatusStyle = lipgloss.NewStyle().
				Foreground(liveColor).
				Bold(true)

	// Event styles - elegant and readable
	eventMinuteStyle = lipgloss.NewStyle().
				Foreground(dimColor).
				Bold(true).
				Width(5).
				Align(lipgloss.Right).
				MarginRight(1)

	eventTextStyle = lipgloss.NewStyle().
			Foreground(textColor).
			MarginLeft(1)

	eventGoalStyle = lipgloss.NewStyle().
			Foreground(goalColor).
			Bold(true)

	eventCardStyle = lipgloss.NewStyle().
			Foreground(cardColor).
			Bold(true)
)

// RenderMultiPanelView renders the three-panel layout for live matches.
func RenderMultiPanelView(width, height int, matches []MatchDisplay, selected int, details *api.MatchDetails) string {
	// Calculate panel dimensions
	// Left side: 40% width, split into two panels (top: matches list, bottom: match details)
	// Right side: 60% width (minute-by-minute)
	leftWidth := width * 40 / 100
	if leftWidth < 30 {
		leftWidth = 30 // Minimum width
	}
	rightWidth := width - leftWidth - 1 // -1 for border separator
	if rightWidth < 30 {
		rightWidth = 30
		leftWidth = width - rightWidth - 1
	}

	topHeight := height * 50 / 100
	if topHeight < 5 {
		topHeight = 5
	}
	bottomHeight := height - topHeight - 1 // -1 for border separator
	if bottomHeight < 5 {
		bottomHeight = 5
		topHeight = height - bottomHeight - 1
	}

	// Render left top panel (live matches list)
	leftTop := renderMatchesListPanel(leftWidth, topHeight, matches, selected)

	// Render left bottom panel (match details/stats)
	leftBottom := renderMatchDetailsPanel(leftWidth, bottomHeight, details)

	// Render right panel (minute-by-minute)
	rightPanel := renderMinuteByMinutePanel(rightWidth, height, details)

	// Create elegant vertical separator
	separatorStyle := lipgloss.NewStyle().
		Foreground(borderColor).
		Height(height).
		Padding(0, 1)
	separator := separatorStyle.Render("│")

	// Combine left panels vertically
	leftPanels := lipgloss.JoinVertical(
		lipgloss.Left,
		leftTop,
		leftBottom,
	)

	// Combine left and right panels horizontally
	content := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftPanels,
		separator,
		rightPanel,
	)

	return content
}

// renderMatchesListPanel renders the top-left panel with the list of live matches.
func renderMatchesListPanel(width, height int, matches []MatchDisplay, selected int) string {
	title := panelTitleStyle.Width(width - 6).Render("Live Matches")

	items := make([]string, 0, len(matches))
	contentWidth := width - 6 // Account for border and padding

	if len(matches) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(dimColor).
			Italic(true).
			Padding(2, 0).
			Align(lipgloss.Center).
			Width(contentWidth)
		items = append(items, emptyStyle.Render("No matches available"))
	} else {
		for i, match := range matches {
			item := renderMatchListItem(match, i == selected, contentWidth)
			items = append(items, item)
		}
	}

	content := strings.Join(items, "\n")

	panelContent := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		content,
	)

	panel := panelStyle.
		Width(width).
		Height(height).
		Render(panelContent)

	return panel
}

func renderMatchListItem(match MatchDisplay, selected bool, width int) string {
	// Status indicator with elegant styling
	var statusIndicator string
	statusStyle := lipgloss.NewStyle().Foreground(dimColor).Width(6).Align(lipgloss.Left)
	if match.Status == api.MatchStatusLive {
		liveTime := "LIVE"
		if match.LiveTime != nil {
			liveTime = *match.LiveTime
		}
		statusIndicator = matchStatusStyle.Render("● " + liveTime)
	} else if match.Status == api.MatchStatusFinished {
		statusIndicator = statusStyle.Render("FT")
	} else {
		statusIndicator = statusStyle.Render("VS")
	}

	// Teams with better spacing
	homeTeamStyle := lipgloss.NewStyle().Foreground(textColor)
	awayTeamStyle := lipgloss.NewStyle().Foreground(textColor)
	if selected {
		homeTeamStyle = homeTeamStyle.Foreground(selectedColor).Bold(true)
		awayTeamStyle = awayTeamStyle.Foreground(selectedColor).Bold(true)
	}

	homeTeam := homeTeamStyle.Render(match.HomeTeam.ShortName)
	awayTeam := awayTeamStyle.Render(match.AwayTeam.ShortName)

	// Score with elegant styling
	var scoreText string
	scoreStyle := lipgloss.NewStyle().Foreground(accentColor).Bold(true)
	if match.HomeScore != nil && match.AwayScore != nil {
		scoreText = scoreStyle.Render(fmt.Sprintf("%d - %d", *match.HomeScore, *match.AwayScore))
	} else {
		scoreText = lipgloss.NewStyle().Foreground(dimColor).Render("vs")
	}

	// Build line with proper spacing
	line := lipgloss.JoinHorizontal(
		lipgloss.Left,
		statusIndicator,
		"  ",
		homeTeam,
		"  ",
		scoreText,
		"  ",
		awayTeam,
	)

	// Truncate if needed
	if len(line) > width {
		line = Truncate(line, width)
	}

	// Apply selection style - elegant text color change
	if selected {
		return matchListItemSelectedStyle.
			Width(width).
			Render(line)
	}
	return matchListItemStyle.
		Width(width).
		Render(line)
}

// renderMatchDetailsPanel renders the bottom-left panel with match details and stats.
func renderMatchDetailsPanel(width, height int, details *api.MatchDetails) string {
	var title string
	if details == nil {
		title = panelTitleStyle.Width(width - 6).Render("Details")
	} else {
		title = panelTitleStyle.Width(width - 6).Render(fmt.Sprintf("%s vs %s",
			details.HomeTeam.ShortName,
			details.AwayTeam.ShortName))
	}

	var content strings.Builder

	if details == nil {
		emptyStyle := lipgloss.NewStyle().
			Foreground(dimColor).
			Italic(true).
			Padding(2, 0).
			Align(lipgloss.Center)
		content.WriteString(emptyStyle.Render("Select a match to view details"))
	} else {
		// Score with elegant styling
		if details.HomeScore != nil && details.AwayScore != nil {
			score := matchScoreStyle.Render(
				fmt.Sprintf("%d - %d", *details.HomeScore, *details.AwayScore))
			content.WriteString(score)
			content.WriteString("\n\n")
		}

		// Status with elegant indicator
		var statusText string
		if details.Status == api.MatchStatusLive {
			liveTime := "LIVE"
			if details.LiveTime != nil {
				liveTime = *details.LiveTime
			}
			statusText = matchStatusStyle.Render("● " + liveTime)
		} else if details.Status == api.MatchStatusFinished {
			statusText = lipgloss.NewStyle().
				Foreground(dimColor).
				Render("Finished")
		} else {
			statusText = lipgloss.NewStyle().
				Foreground(dimColor).
				Render("Not started")
		}
		content.WriteString(statusText)
		content.WriteString("\n\n")

		// League with elegant styling
		leagueText := lipgloss.NewStyle().
			Foreground(dimColor).
			Italic(true).
			Render(details.League.Name)
		content.WriteString(leagueText)
	}

	panel := panelStyle.
		Width(width).
		Height(height).
		Render(lipgloss.JoinVertical(
			lipgloss.Left,
			title,
			"",
			content.String(),
		))

	return panel
}

// renderMinuteByMinutePanel renders the right panel with minute-by-minute events.
func renderMinuteByMinutePanel(width, height int, details *api.MatchDetails) string {
	title := panelTitleStyle.Width(width - 6).Render("Details")

	var content strings.Builder

	if details == nil || len(details.Events) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(dimColor).
			Italic(true).
			Padding(2, 0).
			Align(lipgloss.Center)
		content.WriteString(emptyStyle.Render("No events available"))
	} else {
		// Render events in reverse chronological order (newest first)
		for i := len(details.Events) - 1; i >= 0; i-- {
			event := details.Events[i]
			eventLine := renderEvent(event, width-6)
			content.WriteString(eventLine)
			if i > 0 {
				content.WriteString("\n")
			}
		}
	}

	panel := panelStyle.
		Width(width).
		Height(height).
		Render(lipgloss.JoinVertical(
			lipgloss.Left,
			title,
			"",
			content.String(),
		))

	return panel
}

func renderEvent(event api.MatchEvent, width int) string {
	// Minute
	minute := eventMinuteStyle.Render(fmt.Sprintf("%d'", event.Minute))

	// Event text based on type
	var eventText string
	switch event.Type {
	case "goal":
		player := "Unknown"
		if event.Player != nil {
			player = *event.Player
		}
		assistText := ""
		if event.Assist != nil {
			assistText = fmt.Sprintf(" (assist: %s)", *event.Assist)
		}
		eventText = eventGoalStyle.Render(fmt.Sprintf("⚽ Goal! %s%s", player, assistText))
	case "card":
		player := "Unknown"
		if event.Player != nil {
			player = *event.Player
		}
		cardType := "card"
		if event.EventType != nil {
			cardType = *event.EventType
		}
		cardEmoji := "🟨"
		if cardType == "red" {
			cardEmoji = "🟥"
		}
		eventText = eventCardStyle.Render(fmt.Sprintf("%s %s - %s", cardEmoji, player, cardType))
	case "substitution":
		player := "Unknown"
		if event.Player != nil {
			player = *event.Player
		}
		subType := "substitution"
		if event.EventType != nil {
			if *event.EventType == "in" {
				subType = "in"
			} else if *event.EventType == "out" {
				subType = "out"
			}
		}
		arrow := "→"
		if subType == "in" {
			arrow = "←"
		}
		eventText = eventTextStyle.Render(fmt.Sprintf("🔄 %s %s", arrow, player))
	default:
		eventText = eventTextStyle.Render(fmt.Sprintf("• %s", event.Type))
	}

	// Team name
	teamName := event.Team.ShortName

	line := fmt.Sprintf("%s %s [%s]", minute, eventText, teamName)

	// Truncate if needed
	if len(line) > width {
		line = Truncate(line, width)
	}

	return line
}
