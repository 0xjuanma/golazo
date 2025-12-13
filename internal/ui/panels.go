package ui

import (
	"fmt"
	"strings"

	"github.com/0xjuanma/golazo/internal/api"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
)

var (
	// Modern Neon panel styles - rounded borders with cyan
	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Padding(1, 2).
			Margin(0, 0)

	// Header style - modern with cyan accent
	panelTitleStyle = lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true).
			PaddingBottom(1).
			BorderBottom(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(borderColor).
			MarginBottom(1)

	// Selection styling - modern neon with text color highlight
	matchListItemStyle = lipgloss.NewStyle().
				Foreground(textColor).
				Padding(0, 2)

	matchListItemSelectedStyle = lipgloss.NewStyle().
					Foreground(highlightColor).
					Bold(true).
					Padding(0, 2)

	// Match details styles - refined typography
	matchTitleStyle = lipgloss.NewStyle().
			Foreground(textColor).
			Bold(true).
			MarginBottom(1)

	matchScoreStyle = lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true).
			Margin(0, 1).
			Background(lipgloss.Color("0")).
			Padding(0, 1)

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

	// Live update styles
	liveUpdateStyle = lipgloss.NewStyle().
			Foreground(textColor).
			Padding(0, 1)

	spinnerStyle = lipgloss.NewStyle().
			Foreground(accentColor)
)

// RenderMultiPanelView renders a minimal two-panel layout for live matches.
func RenderMultiPanelView(width, height int, matches []MatchDisplay, selected int, details *api.MatchDetails, liveUpdates []string, sp spinner.Model, loading bool) string {
	// Calculate panel dimensions
	// Left side: 35% width (matches list)
	// Right side: 65% width (match details + live updates)
	leftWidth := width * 35 / 100
	if leftWidth < 25 {
		leftWidth = 25 // Minimum width
	}
	rightWidth := width - leftWidth - 1 // -1 for border separator
	if rightWidth < 35 {
		rightWidth = 35
		leftWidth = width - rightWidth - 1
	}

	// Render left panel (matches list)
	leftPanel := renderMatchesListPanel(leftWidth, height, matches, selected)

	// Render right panel (match details with live updates)
	rightPanel := renderMatchDetailsPanel(rightWidth, height, details, liveUpdates, sp, loading)

	// Create modern neon vertical separator
	separatorStyle := lipgloss.NewStyle().
		Foreground(borderColor).
		Height(height).
		Padding(0, 1)
	separator := separatorStyle.Render("│")

	// Combine left and right panels horizontally
	content := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftPanel,
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
		// Nice empty state with icon and helpful message
		emptyIcon := lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true).
			Align(lipgloss.Center).
			Width(contentWidth).
			PaddingTop(2).
			Render("⚽")

		emptyMessage := lipgloss.NewStyle().
			Foreground(textColor).
			Align(lipgloss.Center).
			Width(contentWidth).
			PaddingTop(1).
			Render("No live matches")

		emptySubtext := lipgloss.NewStyle().
			Foreground(dimColor).
			Italic(true).
			Align(lipgloss.Center).
			Width(contentWidth).
			PaddingTop(1).
			Render("Check back later for live action")

		items = append(items, lipgloss.JoinVertical(
			lipgloss.Center,
			emptyIcon,
			emptyMessage,
			emptySubtext,
		))
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
	// Compact status indicator
	var statusIndicator string
	statusStyle := lipgloss.NewStyle().Foreground(dimColor).Width(5).Align(lipgloss.Left)
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

	// Teams - compact display
	homeTeamStyle := lipgloss.NewStyle().Foreground(textColor)
	awayTeamStyle := lipgloss.NewStyle().Foreground(textColor)
	if selected {
		homeTeamStyle = homeTeamStyle.Foreground(highlightColor).Bold(true)
		awayTeamStyle = awayTeamStyle.Foreground(highlightColor).Bold(true)
	}

	homeTeam := homeTeamStyle.Render(match.HomeTeam.ShortName)
	awayTeam := awayTeamStyle.Render(match.AwayTeam.ShortName)

	// Score - prominent display
	var scoreText string
	scoreStyle := lipgloss.NewStyle().Foreground(accentColor).Bold(true)
	if match.HomeScore != nil && match.AwayScore != nil {
		scoreText = scoreStyle.Render(fmt.Sprintf("%d-%d", *match.HomeScore, *match.AwayScore))
	} else {
		scoreText = lipgloss.NewStyle().Foreground(dimColor).Render("vs")
	}

	// League name - subtle
	leagueName := lipgloss.NewStyle().
		Foreground(dimColor).
		Italic(true).
		Render(Truncate(match.League.Name, 20))

	// Build compact match line
	line := lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.JoinHorizontal(
			lipgloss.Left,
			statusIndicator,
			" ",
			homeTeam,
			" ",
			scoreText,
			" ",
			awayTeam,
		),
		" "+leagueName,
	)

	// Truncate if needed
	if len(line) > width {
		line = Truncate(line, width)
	}

	// Apply selection style
	if selected {
		return matchListItemSelectedStyle.
			Width(width).
			Render(line)
	}
	return matchListItemStyle.
		Width(width).
		Render(line)
}

// renderMatchDetailsPanel renders the right panel with match details and live updates.
func renderMatchDetailsPanel(width, height int, details *api.MatchDetails, liveUpdates []string, sp spinner.Model, loading bool) string {
	if details == nil {
		title := panelTitleStyle.Width(width - 6).Render("Minute-by-minute")

		// Nice empty state for match details
		emptyIcon := lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true).
			Align(lipgloss.Center).
			Width(width - 6).
			PaddingTop(2).
			Render("📊")

		emptyMessage := lipgloss.NewStyle().
			Foreground(textColor).
			Align(lipgloss.Center).
			Width(width - 6).
			PaddingTop(1).
			Render("Select a match")

		emptySubtext := lipgloss.NewStyle().
			Foreground(dimColor).
			Italic(true).
			Align(lipgloss.Center).
			Width(width - 6).
			PaddingTop(1).
			Render("to view live updates and details")

		return panelStyle.
			Width(width).
			Height(height).
			Render(lipgloss.JoinVertical(
				lipgloss.Left,
				title,
				"",
				lipgloss.JoinVertical(
					lipgloss.Center,
					emptyIcon,
					emptyMessage,
					emptySubtext,
				),
			))
	}

	// Panel title
	title := panelTitleStyle.Width(width - 6).Render("Minute-by-minute")

	var content strings.Builder

	// Score section - prominent
	scoreSection := lipgloss.NewStyle().
		Foreground(accentColor).
		Bold(true).
		Align(lipgloss.Center).
		Padding(1, 0)

	if details.HomeScore != nil && details.AwayScore != nil {
		scoreText := fmt.Sprintf("%d  -  %d", *details.HomeScore, *details.AwayScore)
		content.WriteString(scoreSection.Render(scoreText))
	} else {
		content.WriteString(scoreSection.Render("vs"))
	}
	content.WriteString("\n\n")

	// Status and league info
	infoStyle := lipgloss.NewStyle().Foreground(dimColor)
	var statusText string
	if details.Status == api.MatchStatusLive {
		liveTime := "LIVE"
		if details.LiveTime != nil {
			liveTime = *details.LiveTime
		}
		statusText = matchStatusStyle.Render("● " + liveTime)
	} else if details.Status == api.MatchStatusFinished {
		statusText = infoStyle.Render("Finished")
	} else {
		statusText = infoStyle.Render("Not started")
	}

	leagueText := infoStyle.Italic(true).Render(details.League.Name)
	content.WriteString(lipgloss.JoinHorizontal(
		lipgloss.Left,
		statusText,
		"  •  ",
		leagueText,
	))
	content.WriteString("\n\n")

	// Live Updates section
	updatesTitle := lipgloss.NewStyle().
		Foreground(accentColor).
		Bold(true).
		PaddingTop(1).
		BorderBottom(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(borderColor).
		Width(width - 6).
		Render("Live Updates")
	content.WriteString(updatesTitle)
	content.WriteString("\n")

	// Show spinner if loading
	if loading {
		spinnerText := spinnerStyle.Render(sp.View() + " Fetching updates...")
		content.WriteString(spinnerText)
		content.WriteString("\n")
	}

	// Display live updates (newest first)
	if len(liveUpdates) == 0 && !loading {
		emptyUpdates := lipgloss.NewStyle().
			Foreground(dimColor).
			Italic(true).
			Padding(1, 0).
			Render("No updates yet")
		content.WriteString(emptyUpdates)
	} else if len(liveUpdates) > 0 {
		// Show updates in reverse order (newest at top)
		updatesList := make([]string, 0, len(liveUpdates))
		for i := len(liveUpdates) - 1; i >= 0; i-- {
			updateLine := liveUpdateStyle.Render("• " + liveUpdates[i])
			updatesList = append(updatesList, updateLine)
		}
		content.WriteString(strings.Join(updatesList, "\n"))
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
	// Minute - compact
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
		eventText = eventGoalStyle.Render(fmt.Sprintf("⚽ %s%s", player, assistText))
	case "card":
		player := "Unknown"
		if event.Player != nil {
			player = *event.Player
		}
		cardType := "yellow"
		if event.EventType != nil {
			cardType = *event.EventType
		}
		cardEmoji := "🟨"
		if cardType == "red" {
			cardEmoji = "🟥"
		}
		eventText = eventCardStyle.Render(fmt.Sprintf("%s %s", cardEmoji, player))
	case "substitution":
		player := "Unknown"
		if event.Player != nil {
			player = *event.Player
		}
		subType := "sub"
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

	// Team name - subtle
	teamName := lipgloss.NewStyle().
		Foreground(dimColor).
		Render(event.Team.ShortName)

	line := lipgloss.JoinHorizontal(
		lipgloss.Left,
		minute,
		" ",
		eventText,
		" ",
		teamName,
	)

	// Truncate if needed
	if len(line) > width {
		line = Truncate(line, width)
	}

	return line
}
