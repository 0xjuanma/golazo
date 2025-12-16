package ui

import (
	"fmt"
	"strings"

	"github.com/0xjuanma/golazo/internal/api"
	"github.com/0xjuanma/golazo/internal/constants"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
)

var (
	// Panel styles - Neon design with thick red borders
	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.ThickBorder()).
			BorderForeground(lipgloss.Color("196")). // neon red
			Padding(0, 1)

	// Header style - Neon with red accent
	panelTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")). // neon red
			Bold(true).
			PaddingBottom(0).
			BorderBottom(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("239")). // dark dim
			MarginBottom(0)

	// Selection styling - Neon with red highlight
	matchListItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("255")). // neon white
				Padding(0, 1)

	matchListItemSelectedStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("196")). // neon red
					Bold(true).
					Padding(0, 1)

	// Match details styles - Neon typography
	matchTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")). // neon white
			Bold(true).
			MarginBottom(0)

	matchScoreStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")). // neon red for scores
			Bold(true).
			Margin(0, 0).
			Background(lipgloss.Color("0")).
			Padding(0, 0)

	matchStatusStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("196")). // neon red for live
				Bold(true)

	// Event styles - Neon readable
	eventMinuteStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("244")). // dim
				Bold(true).
				Width(4).
				Align(lipgloss.Right).
				MarginRight(0)

	eventTextStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")). // neon white
			MarginLeft(0)

	eventGoalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")). // neon red
			Bold(true)

	eventCardStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")). // neon red
			Bold(true)

	// Live update styles - Neon
	liveUpdateStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")). // neon white
			Padding(0, 0)

	spinnerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("51")) // neon cyan
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

	// Create neon vertical separator - red accent
	separatorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")). // neon red
		Height(height).
		Padding(0, 1)
	separator := separatorStyle.Render("┃")

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
	title := panelTitleStyle.Width(width - 6).Render(constants.PanelLiveMatches)

	items := make([]string, 0, len(matches))
	contentWidth := width - 6 // Account for border and padding

	if len(matches) == 0 {
		emptyMessage := lipgloss.NewStyle().
			Foreground(dimColor).
			Align(lipgloss.Center).
			Width(contentWidth).
			PaddingTop(1).
			Render(constants.EmptyNoLiveMatches)

		items = append(items, emptyMessage)
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
		content,
	)

	panel := panelStyle.
		Width(width).
		Height(height).
		Render(panelContent)

	return panel
}

func renderMatchListItem(match MatchDisplay, selected bool, width int) string {
	// Compact status indicator with neon colors
	var statusIndicator string
	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Width(4).Align(lipgloss.Left) // dim
	if match.Status == api.MatchStatusLive {
		liveTime := constants.StatusLive
		if match.LiveTime != nil {
			liveTime = *match.LiveTime
		}
		statusIndicator = matchStatusStyle.Render(liveTime)
	} else if match.Status == api.MatchStatusFinished {
		finishedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("51")).Width(4).Align(lipgloss.Left) // cyan for FT
		statusIndicator = finishedStyle.Render(constants.StatusFinished)
	} else {
		statusIndicator = statusStyle.Render(constants.StatusNotStarted)
	}

	// Teams - neon display with cyan for teams
	homeTeamStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("51")) // cyan
	awayTeamStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("51")) // cyan
	if selected {
		homeTeamStyle = homeTeamStyle.Foreground(lipgloss.Color("196")).Bold(true) // red when selected
		awayTeamStyle = awayTeamStyle.Foreground(lipgloss.Color("196")).Bold(true) // red when selected
	}

	homeTeam := homeTeamStyle.Render(match.HomeTeam.ShortName)
	awayTeam := awayTeamStyle.Render(match.AwayTeam.ShortName)

	// Score - red for emphasis
	var scoreText string
	scoreStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true) // neon red
	if match.HomeScore != nil && match.AwayScore != nil {
		scoreText = scoreStyle.Render(fmt.Sprintf("%d-%d", *match.HomeScore, *match.AwayScore))
	} else {
		scoreText = lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Render("vs") // dim
	}

	// League name - subtle dim
	leagueName := lipgloss.NewStyle().
		Foreground(lipgloss.Color("244")). // dim
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
	return renderMatchDetailsPanelWithTitle(width, height, details, liveUpdates, sp, loading, true)
}

// renderMatchDetailsPanelWithTitle renders the right panel with optional title.
// Uses Neon design with Golazo red/cyan theme.
func renderMatchDetailsPanelWithTitle(width, height int, details *api.MatchDetails, liveUpdates []string, sp spinner.Model, loading bool, showTitle bool) string {
	// Neon color constants
	neonRed := lipgloss.Color("196")
	neonCyan := lipgloss.Color("51")
	neonDim := lipgloss.Color("244")
	neonWhite := lipgloss.Color("255")

	// Use cyan border for details panel
	detailsPanelStyle := lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(neonCyan).
		Padding(0, 1)

	if details == nil {
		emptyMessage := lipgloss.NewStyle().
			Foreground(neonDim).
			Align(lipgloss.Center).
			Width(width - 6).
			PaddingTop(1).
			Render(constants.EmptySelectMatch)

		content := emptyMessage
		if showTitle {
			title := panelTitleStyle.Width(width - 6).Render(constants.PanelMinuteByMinute)
			content = lipgloss.JoinVertical(
				lipgloss.Left,
				title,
				emptyMessage,
			)
		}

		return detailsPanelStyle.
			Width(width).
			Height(height).
			Render(content)
	}

	// Panel title (only if showTitle is true)
	var title string
	if showTitle {
		title = panelTitleStyle.Width(width - 6).Render(constants.PanelMinuteByMinute)
	}

	var content strings.Builder

	// Score section - neon red for scores
	scoreSection := lipgloss.NewStyle().
		Foreground(neonRed).
		Bold(true).
		Align(lipgloss.Center).
		Padding(0, 0)

	if details.HomeScore != nil && details.AwayScore != nil {
		scoreText := fmt.Sprintf("%d - %d", *details.HomeScore, *details.AwayScore)
		content.WriteString(scoreSection.Render(scoreText))
	} else {
		content.WriteString(scoreSection.Render("vs"))
	}
	content.WriteString("\n")

	// Status and league info with neon colors
	infoStyle := lipgloss.NewStyle().Foreground(neonDim)
	var statusText string
	if details.Status == api.MatchStatusLive {
		liveTime := constants.StatusLive
		if details.LiveTime != nil {
			liveTime = *details.LiveTime
		}
		statusText = lipgloss.NewStyle().Foreground(neonRed).Bold(true).Render(liveTime)
	} else if details.Status == api.MatchStatusFinished {
		statusText = lipgloss.NewStyle().Foreground(neonCyan).Render(constants.StatusFinished)
	} else {
		statusText = infoStyle.Render(constants.StatusNotStartedShort)
	}

	leagueText := infoStyle.Italic(true).Render(details.League.Name)
	content.WriteString(lipgloss.JoinHorizontal(
		lipgloss.Left,
		statusText,
		" ",
		leagueText,
	))
	content.WriteString("\n")

	// Teams section with cyan accent
	teamStyle := lipgloss.NewStyle().Foreground(neonCyan)
	vsStyle := lipgloss.NewStyle().Foreground(neonWhite)
	teamsDisplay := lipgloss.JoinHorizontal(lipgloss.Left,
		teamStyle.Render(details.HomeTeam.ShortName),
		vsStyle.Render(" vs "),
		teamStyle.Render(details.AwayTeam.ShortName),
	)
	content.WriteString(teamsDisplay)
	content.WriteString("\n\n")

	// For finished matches, show detailed match information
	// For live matches, show live updates
	if details.Status == api.MatchStatusFinished {
		// Match Information section
		var infoSection []string

		// Venue
		if details.Venue != "" {
			infoSection = append(infoSection, details.Venue)
		}

		// Half-time score
		if details.HalfTimeScore != nil && details.HalfTimeScore.Home != nil && details.HalfTimeScore.Away != nil {
			htText := fmt.Sprintf("HT: %d - %d", *details.HalfTimeScore.Home, *details.HalfTimeScore.Away)
			infoSection = append(infoSection, infoStyle.Render(htText))
		}

		// Match duration
		if details.ExtraTime {
			infoSection = append(infoSection, infoStyle.Render("AET"))
		}
		if details.Penalties != nil && details.Penalties.Home != nil && details.Penalties.Away != nil {
			penText := fmt.Sprintf("Pens: %d - %d", *details.Penalties.Home, *details.Penalties.Away)
			infoSection = append(infoSection, infoStyle.Render(penText))
		}

		if len(infoSection) > 0 {
			content.WriteString(strings.Join(infoSection, " | "))
			content.WriteString("\n\n")
		}

		// Goals Timeline section with neon styling
		var goals []api.MatchEvent
		for _, event := range details.Events {
			if event.Type == "goal" {
				goals = append(goals, event)
			}
		}

		if len(goals) > 0 {
			goalsTitle := lipgloss.NewStyle().
				Foreground(neonCyan).
				Bold(true).
				PaddingTop(0).
				BorderBottom(true).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("239")).
				Width(width - 6).
				Render("Goals")
			content.WriteString(goalsTitle)
			content.WriteString("\n")

			minuteStyle := lipgloss.NewStyle().Foreground(neonRed).Bold(true)
			for _, goal := range goals {
				player := "Unknown"
				if goal.Player != nil {
					player = *goal.Player
				}
				teamName := goal.Team.ShortName
				assistText := ""
				if goal.Assist != nil && *goal.Assist != "" {
					assistText = fmt.Sprintf(" (assist: %s)", *goal.Assist)
				}
				goalLine := lipgloss.JoinHorizontal(lipgloss.Left,
					minuteStyle.Render(fmt.Sprintf("%d'", goal.Minute)),
					lipgloss.NewStyle().Foreground(neonWhite).Render(fmt.Sprintf(" %s - %s%s", teamName, player, assistText)),
				)
				content.WriteString(goalLine)
				content.WriteString("\n")
			}
			content.WriteString("\n")
		}

		// Cards section with neon styling
		homeYellow := 0
		homeRed := 0
		awayYellow := 0
		awayRed := 0
		for _, event := range details.Events {
			if event.Type == "card" {
				isHome := event.Team.ID == details.HomeTeam.ID
				if event.EventType != nil {
					if *event.EventType == "yellow" {
						if isHome {
							homeYellow++
						} else {
							awayYellow++
						}
					} else if *event.EventType == "red" {
						if isHome {
							homeRed++
						} else {
							awayRed++
						}
					}
				}
			}
		}

		if homeYellow > 0 || homeRed > 0 || awayYellow > 0 || awayRed > 0 {
			cardsTitle := lipgloss.NewStyle().
				Foreground(neonCyan).
				Bold(true).
				PaddingTop(0).
				BorderBottom(true).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("239")).
				Width(width - 6).
				Render("Cards")
			content.WriteString(cardsTitle)
			content.WriteString("\n")

			// Visual card indicators - cyan for yellow, red for red
			yellowStyle := lipgloss.NewStyle().Foreground(neonCyan)
			redStyle := lipgloss.NewStyle().Foreground(neonRed)

			var homeCards []string
			if homeYellow > 0 {
				homeCards = append(homeCards, yellowStyle.Render(strings.Repeat("▪", homeYellow)))
			}
			if homeRed > 0 {
				homeCards = append(homeCards, redStyle.Render(strings.Repeat("▪", homeRed)))
			}
			var awayCards []string
			if awayYellow > 0 {
				awayCards = append(awayCards, yellowStyle.Render(strings.Repeat("▪", awayYellow)))
			}
			if awayRed > 0 {
				awayCards = append(awayCards, redStyle.Render(strings.Repeat("▪", awayRed)))
			}

			if len(homeCards) > 0 || len(awayCards) > 0 {
				homeCardsStr := "-"
				if len(homeCards) > 0 {
					homeCardsStr = strings.Join(homeCards, " ")
				}
				awayCardsStr := "-"
				if len(awayCards) > 0 {
					awayCardsStr = strings.Join(awayCards, " ")
				}
				cardsLine := lipgloss.JoinHorizontal(lipgloss.Left,
					teamStyle.Render(details.HomeTeam.ShortName),
					lipgloss.NewStyle().Foreground(neonDim).Render(": "),
					lipgloss.NewStyle().Foreground(neonWhite).Render(homeCardsStr),
					lipgloss.NewStyle().Foreground(neonDim).Render(" | "),
					teamStyle.Render(details.AwayTeam.ShortName),
					lipgloss.NewStyle().Foreground(neonDim).Render(": "),
					lipgloss.NewStyle().Foreground(neonWhite).Render(awayCardsStr),
				)
				content.WriteString(cardsLine)
				content.WriteString("\n\n")
			}
		}

		// Match Events section with neon styling
		eventsTitle := lipgloss.NewStyle().
			Foreground(neonCyan).
			Bold(true).
			PaddingTop(0).
			BorderBottom(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("239")).
			Width(width - 6).
			Render("All Events")
		content.WriteString(eventsTitle)
		content.WriteString("\n")

		// Display match events (goals, cards, substitutions)
		if len(details.Events) == 0 {
			emptyEvents := lipgloss.NewStyle().
				Foreground(neonDim).
				Padding(0, 0).
				Render("No events recorded")
			content.WriteString(emptyEvents)
		} else {
			// Show events in chronological order (oldest first)
			var eventsList []string
			for _, event := range details.Events {
				eventLine := formatMatchEventForDisplay(event, details.HomeTeam.ShortName, details.AwayTeam.ShortName)
				eventsList = append(eventsList, eventLine)
			}
			content.WriteString(strings.Join(eventsList, "\n"))
		}
	} else {
		// Live Updates section for live/upcoming matches with neon styling
		updatesTitle := lipgloss.NewStyle().
			Foreground(neonCyan).
			Bold(true).
			PaddingTop(0).
			BorderBottom(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("239")).
			Width(width - 6).
			Render(constants.PanelUpdates)
		content.WriteString(updatesTitle)
		content.WriteString("\n")

		// Show spinner if loading
		if loading {
			spinnerText := spinnerStyle.Render(sp.View() + " " + constants.LoadingFetching)
			content.WriteString(spinnerText)
			content.WriteString("\n")
		}

		// Display live updates (newest first)
		if len(liveUpdates) == 0 && !loading {
			emptyUpdates := lipgloss.NewStyle().
				Foreground(neonDim).
				Padding(0, 0).
				Render(constants.EmptyNoUpdates)
			content.WriteString(emptyUpdates)
		} else if len(liveUpdates) > 0 {
			// Show updates in reverse order (newest at top)
			var updatesList []string
			for i := len(liveUpdates) - 1; i >= 0; i-- {
				updateLine := liveUpdateStyle.Render(liveUpdates[i])
				updatesList = append(updatesList, updateLine)
			}
			content.WriteString(strings.Join(updatesList, "\n"))
		}
	}

	// Combine title and content (only include title if showTitle is true)
	panelContent := content.String()
	if showTitle && title != "" {
		panelContent = lipgloss.JoinVertical(
			lipgloss.Left,
			title,
			content.String(),
		)
	}

	panel := detailsPanelStyle.
		Width(width).
		Height(height).
		Render(panelContent)

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
		eventText = eventGoalStyle.Render(fmt.Sprintf("Goal: %s%s", player, assistText))
	case "card":
		player := "Unknown"
		if event.Player != nil {
			player = *event.Player
		}
		cardType := "Yellow"
		if event.EventType != nil {
			if *event.EventType == "red" {
				cardType = "Red"
			}
		}
		eventText = eventCardStyle.Render(fmt.Sprintf("Card (%s): %s", cardType, player))
	case "substitution":
		player := "Unknown"
		if event.Player != nil {
			player = *event.Player
		}
		subType := "Sub"
		if event.EventType != nil {
			if *event.EventType == "in" {
				subType = "In"
			} else if *event.EventType == "out" {
				subType = "Out"
			}
		}
		eventText = eventTextStyle.Render(fmt.Sprintf("%s: %s", subType, player))
	default:
		eventText = eventTextStyle.Render(event.Type)
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

// formatMatchEventForDisplay formats a match event for display in the stats view
// Uses neon styling with red/cyan theme and no emojis
func formatMatchEventForDisplay(event api.MatchEvent, homeTeam, awayTeam string) string {
	// Neon colors
	neonRed := lipgloss.Color("196")
	neonCyan := lipgloss.Color("51")
	neonWhite := lipgloss.Color("255")

	minuteStr := fmt.Sprintf("%d'", event.Minute)
	minuteStyle := lipgloss.NewStyle().Foreground(neonRed).Bold(true).Width(4).Align(lipgloss.Right).Render(minuteStr)

	var eventText string
	switch event.Type {
	case "goal":
		teamName := homeTeam
		if event.Team.ShortName != homeTeam {
			teamName = awayTeam
		}
		playerName := "Unknown"
		if event.Player != nil {
			playerName = *event.Player
		}
		assistText := ""
		if event.Assist != nil && *event.Assist != "" {
			assistText = fmt.Sprintf(" (assist: %s)", *event.Assist)
		}
		goalStyle := lipgloss.NewStyle().Foreground(neonRed).Bold(true)
		eventText = goalStyle.Render(fmt.Sprintf("GOAL %s - %s%s", teamName, playerName, assistText))
	case "card":
		teamName := homeTeam
		if event.Team.ShortName != homeTeam {
			teamName = awayTeam
		}
		playerName := "Unknown"
		if event.Player != nil {
			playerName = *event.Player
		}
		cardType := "yellow"
		if event.EventType != nil {
			cardType = *event.EventType
		}
		cardIndicator := lipgloss.NewStyle().Foreground(neonCyan).Render("▪") // cyan for yellow
		if cardType == "red" {
			cardIndicator = lipgloss.NewStyle().Foreground(neonRed).Render("▪") // red for red
		}
		cardStyle := lipgloss.NewStyle().Foreground(neonWhite)
		eventText = lipgloss.JoinHorizontal(lipgloss.Left, cardIndicator, cardStyle.Render(fmt.Sprintf(" %s - %s", teamName, playerName)))
	case "substitution":
		teamName := homeTeam
		if event.Team.ShortName != homeTeam {
			teamName = awayTeam
		}
		playerName := "Unknown"
		if event.Player != nil {
			playerName = *event.Player
		}
		subStyle := lipgloss.NewStyle().Foreground(neonWhite)
		eventText = subStyle.Render(fmt.Sprintf("SUB %s - %s", teamName, playerName))
	default:
		teamName := homeTeam
		if event.Team.ShortName != homeTeam {
			teamName = awayTeam
		}
		playerName := ""
		if event.Player != nil {
			playerName = *event.Player
		}
		defaultStyle := lipgloss.NewStyle().Foreground(neonWhite)
		if playerName != "" {
			eventText = defaultStyle.Render(fmt.Sprintf("%s - %s", teamName, playerName))
		} else {
			eventText = defaultStyle.Render(teamName)
		}
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		minuteStyle,
		" ",
		eventText,
	)
}
