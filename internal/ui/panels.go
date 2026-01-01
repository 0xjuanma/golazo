package ui

import (
	"fmt"
	"strings"

	"github.com/0xjuanma/golazo/internal/api"
	"github.com/0xjuanma/golazo/internal/constants"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/lucasb-eyer/go-colorful"
)

var (
	// Panel styles - Neon design with thick red borders
	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
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

	// Live update styles - Neon
	liveUpdateStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")). // neon white
			Padding(0, 0)

	spinnerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("51")) // neon cyan
)

// renderMatchDetailsPanel renders the right panel with match details and live updates.
func renderMatchDetailsPanel(width, height int, details *api.MatchDetails, liveUpdates []string, sp spinner.Model, loading bool) string {
	return renderMatchDetailsPanelFull(width, height, details, liveUpdates, sp, loading, true, nil, false)
}

// RenderMatchDetailsHeader renders only the static header portion of match details
// (status, teams, and score) that should not scroll.
func RenderMatchDetailsHeader(width int, details *api.MatchDetails) string {
	if details == nil {
		return ""
	}

	neonRed := lipgloss.Color("196")
	neonCyan := lipgloss.Color("51")
	neonDim := lipgloss.Color("244")

	contentWidth := width - 6

	var content strings.Builder

	// 1. Status/Minute and League info (centered)
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
	statusLine := lipgloss.NewStyle().
		Width(contentWidth).
		Align(lipgloss.Center).
		Render(statusText + " • " + leagueText)
	content.WriteString(statusLine)
	content.WriteString("\n")

	// 2. Teams section (centered)
	teamStyle := lipgloss.NewStyle().Foreground(neonCyan).Bold(true)
	vsStyle := lipgloss.NewStyle().Foreground(neonDim)
	teamsDisplay := lipgloss.JoinHorizontal(lipgloss.Center,
		teamStyle.Render(details.HomeTeam.ShortName),
		vsStyle.Render("  vs  "),
		teamStyle.Render(details.AwayTeam.ShortName),
	)
	teamsLine := lipgloss.NewStyle().
		Width(contentWidth).
		Align(lipgloss.Center).
		Render(teamsDisplay)
	content.WriteString(teamsLine)
	content.WriteString("\n\n")

	// 3. Large Score section (centered, prominent)
	if details.HomeScore != nil && details.AwayScore != nil {
		largeScore := renderLargeScore(*details.HomeScore, *details.AwayScore, contentWidth)
		content.WriteString(largeScore)
	} else {
		vsText := lipgloss.NewStyle().
			Foreground(neonDim).
			Bold(true).
			Width(contentWidth).
			Align(lipgloss.Center).
			Render("vs")
		content.WriteString(vsText)
	}
	content.WriteString("\n")

	return content.String()
}

// RenderMatchDetailsScrollableContent renders only the scrollable content portion
// of match details (events, goals, cards, etc.) without the header.
func RenderMatchDetailsScrollableContent(width int, details *api.MatchDetails, liveUpdates []string, pollingSpinner *RandomCharSpinner, isPolling bool, loading bool) string {
	if details == nil {
		return ""
	}

	neonRed := lipgloss.Color("196")
	neonCyan := lipgloss.Color("51")
	neonDim := lipgloss.Color("244")
	neonWhite := lipgloss.Color("255")

	contentWidth := width - 6
	infoStyle := lipgloss.NewStyle().Foreground(neonDim)

	var content strings.Builder

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
		var cardEvents []api.MatchEvent
		for _, event := range details.Events {
			if event.Type == "card" {
				cardEvents = append(cardEvents, event)
			}
		}

		if len(cardEvents) > 0 {
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

			for _, card := range cardEvents {
				player := "Unknown"
				if card.Player != nil {
					player = *card.Player
				}
				teamName := card.Team.ShortName

				cardSymbol := CardSymbolYellow
				cardStyle := neonYellowCardStyle
				if card.EventType != nil && *card.EventType == "red" {
					cardSymbol = CardSymbolRed
					cardStyle = neonRedCardStyle
				}

				cardLine := lipgloss.JoinHorizontal(lipgloss.Left,
					cardStyle.Render(cardSymbol),
					lipgloss.NewStyle().Foreground(neonRed).Bold(true).Render(fmt.Sprintf(" %d' ", card.Minute)),
					lipgloss.NewStyle().Foreground(neonWhite).Render(fmt.Sprintf("%s (%s)", player, teamName)),
				)
				content.WriteString(cardLine)
				content.WriteString("\n")
			}
			content.WriteString("\n")
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

		if len(details.Events) == 0 {
			emptyEvents := lipgloss.NewStyle().
				Foreground(neonDim).
				Padding(0, 0).
				Render("No events recorded")
			content.WriteString(emptyEvents)
		} else {
			var eventsList []string
			for _, event := range details.Events {
				eventLine := formatMatchEventForDisplay(event, details.HomeTeam.ShortName, details.AwayTeam.ShortName)
				eventsList = append(eventsList, eventLine)
			}
			content.WriteString(strings.Join(eventsList, "\n"))
		}
	} else {
		// Live Updates section for live/upcoming matches
		var titleText string
		if isPolling && loading && pollingSpinner != nil {
			pollingView := pollingSpinner.View()
			titleText = "Updating...  " + pollingView
		} else {
			titleText = constants.PanelUpdates
		}
		updatesTitle := lipgloss.NewStyle().
			Foreground(neonCyan).
			Bold(true).
			PaddingTop(0).
			BorderBottom(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("239")).
			Width(contentWidth).
			Render(titleText)
		content.WriteString(updatesTitle)
		content.WriteString("\n")

		if len(liveUpdates) == 0 && !loading && !isPolling {
			emptyUpdates := lipgloss.NewStyle().
				Foreground(neonDim).
				Padding(0, 0).
				Render(constants.EmptyNoUpdates)
			content.WriteString(emptyUpdates)
		} else if len(liveUpdates) > 0 {
			var updatesList []string
			for _, update := range liveUpdates {
				updateLine := renderStyledLiveUpdate(update)
				updatesList = append(updatesList, updateLine)
			}
			content.WriteString(strings.Join(updatesList, "\n"))
		}
	}

	return content.String()
}

// RenderMatchDetailsPanelWithViewport renders the right panel with viewport scrolling.
// The header (teams + score) stays fixed, while the content below scrolls.
func RenderMatchDetailsPanelWithViewport(width, height int, details *api.MatchDetails, liveUpdates []string, sp spinner.Model, loading bool, pollingSpinner *RandomCharSpinner, isPolling bool, viewportView string, showTitle bool) string {
	neonDim := lipgloss.Color("244")

	// Details panel - no border, just padding for clean look
	detailsPanelStyle := lipgloss.NewStyle().
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
			MaxHeight(height).
			Render(content)
	}

	// Render the static header
	header := RenderMatchDetailsHeader(width, details)

	// Combine header with viewport content
	panelContent := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		viewportView,
	)

	// Add title if needed
	if showTitle {
		title := panelTitleStyle.Width(width - 6).Render(constants.PanelMinuteByMinute)
		panelContent = lipgloss.JoinVertical(
			lipgloss.Left,
			title,
			panelContent,
		)
	}

	panel := detailsPanelStyle.
		Width(width).
		Height(height).
		MaxHeight(height).
		Render(panelContent)

	return panel
}

// renderMatchDetailsPanelWithPolling renders the right panel with polling spinner support.
func renderMatchDetailsPanelWithPolling(width, height int, details *api.MatchDetails, liveUpdates []string, sp spinner.Model, loading bool, pollingSpinner *RandomCharSpinner, isPolling bool) string {
	return renderMatchDetailsPanelFull(width, height, details, liveUpdates, sp, loading, true, pollingSpinner, isPolling)
}

// renderMatchDetailsPanelFull renders the right panel with optional title and polling spinner.
// Uses Neon design with Golazo red/cyan theme.
func renderMatchDetailsPanelFull(width, height int, details *api.MatchDetails, liveUpdates []string, sp spinner.Model, loading bool, showTitle bool, pollingSpinner *RandomCharSpinner, isPolling bool) string {
	// Neon color constants
	neonRed := lipgloss.Color("196")
	neonCyan := lipgloss.Color("51")
	neonDim := lipgloss.Color("244")
	neonWhite := lipgloss.Color("255")

	// Details panel - no border, just padding for clean look
	detailsPanelStyle := lipgloss.NewStyle().
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
			MaxHeight(height).
			Render(content)
	}

	// Panel title (only if showTitle is true)
	var title string
	if showTitle {
		title = panelTitleStyle.Width(width - 6).Render(constants.PanelMinuteByMinute)
	}

	var content strings.Builder
	contentWidth := width - 6

	// 1. Status/Minute and League info (centered)
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
	statusLine := lipgloss.NewStyle().
		Width(contentWidth).
		Align(lipgloss.Center).
		Render(statusText + " • " + leagueText)
	content.WriteString(statusLine)
	content.WriteString("\n")

	// 2. Teams section (centered)
	teamStyle := lipgloss.NewStyle().Foreground(neonCyan).Bold(true)
	vsStyle := lipgloss.NewStyle().Foreground(neonDim)
	teamsDisplay := lipgloss.JoinHorizontal(lipgloss.Center,
		teamStyle.Render(details.HomeTeam.ShortName),
		vsStyle.Render("  vs  "),
		teamStyle.Render(details.AwayTeam.ShortName),
	)
	teamsLine := lipgloss.NewStyle().
		Width(contentWidth).
		Align(lipgloss.Center).
		Render(teamsDisplay)
	content.WriteString(teamsLine)
	content.WriteString("\n\n")

	// 3. Large Score section (centered, prominent)
	if details.HomeScore != nil && details.AwayScore != nil {
		largeScore := renderLargeScore(*details.HomeScore, *details.AwayScore, contentWidth)
		content.WriteString(largeScore)
	} else {
		vsText := lipgloss.NewStyle().
			Foreground(neonDim).
			Bold(true).
			Width(contentWidth).
			Align(lipgloss.Center).
			Render("vs")
		content.WriteString(vsText)
	}
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

		// Cards section with neon styling - detailed list with player, minute, team
		var cardEvents []api.MatchEvent
		for _, event := range details.Events {
			if event.Type == "card" {
				cardEvents = append(cardEvents, event)
			}
		}

		if len(cardEvents) > 0 {
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

			for _, card := range cardEvents {
				player := "Unknown"
				if card.Player != nil {
					player = *card.Player
				}
				teamName := card.Team.ShortName

				// Determine card type and apply appropriate color (using shared styles)
				cardSymbol := CardSymbolYellow
				cardStyle := neonYellowCardStyle
				if card.EventType != nil && *card.EventType == "red" {
					cardSymbol = CardSymbolRed
					cardStyle = neonRedCardStyle
				}

				// Format: ▪ 28' PlayerName (Team)
				cardLine := lipgloss.JoinHorizontal(lipgloss.Left,
					cardStyle.Render(cardSymbol),
					lipgloss.NewStyle().Foreground(neonRed).Bold(true).Render(fmt.Sprintf(" %d' ", card.Minute)),
					lipgloss.NewStyle().Foreground(neonWhite).Render(fmt.Sprintf("%s (%s)", player, teamName)),
				)
				content.WriteString(cardLine)
				content.WriteString("\n")
			}
			content.WriteString("\n")
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
		// Build title - show "Updating..." with spinner only during poll API calls
		var titleText string
		if isPolling && loading && pollingSpinner != nil {
			// Poll API call in progress - show "Updating..." with spinner
			pollingView := pollingSpinner.View()
			titleText = "Updating...  " + pollingView
		} else {
			// Not polling or not loading - just show "Updates"
			titleText = constants.PanelUpdates
		}
		updatesTitle := lipgloss.NewStyle().
			Foreground(neonCyan).
			Bold(true).
			PaddingTop(0).
			BorderBottom(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("239")).
			Width(width - 6).
			Render(titleText)
		content.WriteString(updatesTitle)
		content.WriteString("\n")

		// Display live updates (already sorted by minute descending - newest first)
		if len(liveUpdates) == 0 && !loading && !isPolling {
			emptyUpdates := lipgloss.NewStyle().
				Foreground(neonDim).
				Padding(0, 0).
				Render(constants.EmptyNoUpdates)
			content.WriteString(emptyUpdates)
		} else if len(liveUpdates) > 0 {
			// Events are already sorted descending by minute
			var updatesList []string
			for _, update := range liveUpdates {
				updateLine := renderStyledLiveUpdate(update)
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
		MaxHeight(height).
		Render(panelContent)

	return panel
}

// formatMatchEventForDisplay formats a match event for display in the stats view
// Uses neon styling with red/cyan theme and no emojis
// renderStyledLiveUpdate renders a live update string with appropriate colors based on symbol prefix.
// Uses minimal symbol styling: ● gradient for goals, ▪ cyan for yellow cards, ■ red for red cards,
// ↔ dim for substitutions, · dim for other events.
func renderStyledLiveUpdate(update string) string {
	if len(update) == 0 {
		return update
	}

	// Get the first rune (symbol prefix)
	runes := []rune(update)
	symbol := string(runes[0])
	rest := string(runes[1:])

	// Neon colors matching theme
	neonRed := lipgloss.Color("196")
	neonDim := lipgloss.Color("244")
	neonWhite := lipgloss.Color("255")

	switch symbol {
	case "●": // Goal - gradient on [GOAL] label, white text for rest
		return renderGoalWithGradient(update)
	case "▪": // Yellow card - yellow up to [CARD], white for rest
		neonYellow := lipgloss.Color("226") // Bright yellow
		return renderCardWithColor(update, neonYellow)
	case "■": // Red card - red up to [CARD], white for rest
		return renderCardWithColor(update, neonRed)
	case "↔": // Substitution - color coded players
		return renderSubstitutionWithColors(update)
	case "·": // Other - dim symbol and text
		symbolStyle := lipgloss.NewStyle().Foreground(neonDim)
		textStyle := lipgloss.NewStyle().Foreground(neonDim)
		return symbolStyle.Render(symbol) + textStyle.Render(rest)
	default:
		// Unknown prefix, render as-is with default style
		return lipgloss.NewStyle().Foreground(neonWhite).Render(update)
	}
}

// renderSubstitutionWithColors renders a substitution event with color-coded players.
// Cyan ← arrow = player coming IN (entering the pitch)
// Red → arrow = player going OUT (leaving the pitch)
// Format: ↔ 45' [SUB] {OUT}PlayerOut {IN}PlayerIn - Team
func renderSubstitutionWithColors(update string) string {
	neonRed := lipgloss.Color("196")
	neonCyan := lipgloss.Color("51")
	neonDim := lipgloss.Color("244")
	neonWhite := lipgloss.Color("255")

	dimStyle := lipgloss.NewStyle().Foreground(neonDim)
	whiteStyle := lipgloss.NewStyle().Foreground(neonWhite)
	outStyle := lipgloss.NewStyle().Foreground(neonRed) // Red = going OUT
	inStyle := lipgloss.NewStyle().Foreground(neonCyan) // Cyan = coming IN

	// Find the markers
	outIdx := strings.Index(update, "{OUT}")
	inIdx := strings.Index(update, "{IN}")
	teamIdx := strings.LastIndex(update, " - ")

	if outIdx == -1 || inIdx == -1 {
		// Fallback to dim rendering if markers not found
		return dimStyle.Render(update)
	}

	// Split the string into parts
	prefix := update[:outIdx]             // "↔ 45' [SUB] "
	playerOut := update[outIdx+5 : inIdx] // Player going OUT (after {OUT}, before {IN})
	playerIn := update[inIdx+4 : teamIdx] // Player coming IN (after {IN}, before " - ")
	suffix := update[teamIdx:]            // " - Team"

	// Render prefix (symbol, time, [SUB]) in dim
	result := dimStyle.Render(prefix)

	// Render player coming IN with cyan ← arrow (entering the pitch)
	result += inStyle.Render("← " + strings.TrimSpace(playerIn))
	result += whiteStyle.Render(" ")

	// Render player going OUT with red → arrow (leaving the pitch)
	result += outStyle.Render("→ " + strings.TrimSpace(playerOut))

	// Render suffix (team) in white
	result += whiteStyle.Render(suffix)

	return result
}

// renderCardWithColor renders a card event with color on symbol, time, and [CARD] label.
// The rest of the text (player, team) is rendered in white.
func renderCardWithColor(update string, color lipgloss.Color) string {
	neonWhite := lipgloss.Color("255")
	colorStyle := lipgloss.NewStyle().Foreground(color).Bold(true)
	whiteStyle := lipgloss.NewStyle().Foreground(neonWhite)

	// Find [CARD] in the string
	cardEnd := strings.Index(update, "[CARD]")
	if cardEnd == -1 {
		// No [CARD] found, color entire line
		return colorStyle.Render(update)
	}
	cardEnd += len("[CARD]")

	// Split: colored prefix (symbol + time + [CARD]) and white suffix (player + team)
	prefix := update[:cardEnd]
	suffix := update[cardEnd:]

	return colorStyle.Render(prefix) + whiteStyle.Render(suffix)
}

// renderGoalWithGradient renders a goal event with gradient on the [GOAL] label.
// The gradient matches the spinner theme (cyan → red).
func renderGoalWithGradient(update string) string {
	// Parse gradient colors
	startColor, _ := colorful.Hex(constants.GradientStartColor) // Cyan
	endColor, _ := colorful.Hex(constants.GradientEndColor)     // Red

	neonWhite := lipgloss.Color("255")
	whiteStyle := lipgloss.NewStyle().Foreground(neonWhite)

	// Find [GOAL] in the string and apply gradient to it
	goalStart := strings.Index(update, "[GOAL]")
	if goalStart == -1 {
		// No [GOAL] found, just render with gradient on first part
		return applyGradientToText(update, startColor, endColor)
	}

	goalEnd := goalStart + len("[GOAL]")

	// Build: prefix + gradient[GOAL] + suffix
	prefix := update[:goalStart]
	goalText := update[goalStart:goalEnd]
	suffix := update[goalEnd:]

	// Apply gradient to [GOAL] text character by character
	gradientGoal := applyGradientToText(goalText, startColor, endColor)

	// Render prefix (● and time) with gradient too for cohesion
	gradientPrefix := applyGradientToText(prefix, startColor, endColor)

	return gradientPrefix + gradientGoal + whiteStyle.Render(suffix)
}

// applyGradientToText applies a cyan→red gradient to text, character by character.
func applyGradientToText(text string, startColor, endColor colorful.Color) string {
	runes := []rune(text)
	if len(runes) == 0 {
		return text
	}

	var result strings.Builder
	for i, char := range runes {
		ratio := float64(i) / float64(max(len(runes)-1, 1))
		color := startColor.BlendLab(endColor, ratio)
		hexColor := color.Hex()
		charStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(hexColor)).Bold(true)
		result.WriteString(charStyle.Render(string(char)))
	}

	return result.String()
}

func formatMatchEventForDisplay(event api.MatchEvent, homeTeam, awayTeam string) string {
	// Uses package-level neon colors from neon_styles.go

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
		// Use shared card styles for consistency
		cardSymbol := CardSymbolYellow
		cardStyle := neonYellowCardStyle
		if cardType == "red" {
			cardSymbol = CardSymbolRed
			cardStyle = neonRedCardStyle
		}
		cardIndicator := cardStyle.Render(cardSymbol)
		textStyle := lipgloss.NewStyle().Foreground(neonWhite)
		eventText = lipgloss.JoinHorizontal(lipgloss.Left, cardIndicator, textStyle.Render(fmt.Sprintf(" %s - %s", teamName, playerName)))
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

// renderLargeScore renders the score in a large, prominent format using block digits.
// The score is centered within the given width.
func renderLargeScore(homeScore, awayScore int, width int) string {
	// Large block-style digits (3 lines tall)
	digits := map[int][]string{
		0: {"█▀█", "█ █", "▀▀▀"},
		1: {" █ ", " █ ", " ▀ "},
		2: {"▀▀█", "█▀▀", "▀▀▀"},
		3: {"▀▀█", " ▀█", "▀▀▀"},
		4: {"█ █", "▀▀█", "  ▀"},
		5: {"█▀▀", "▀▀█", "▀▀▀"},
		6: {"█▀▀", "█▀█", "▀▀▀"},
		7: {"▀▀█", "  █", "  ▀"},
		8: {"█▀█", "█▀█", "▀▀▀"},
		9: {"█▀█", "▀▀█", "▀▀▀"},
	}

	dash := []string{"   ", "▀▀▀", "   "}

	// Helper to get digit patterns for a number (handles multi-digit)
	getDigitPatterns := func(score int) [][]string {
		if score < 10 {
			return [][]string{digits[score]}
		}
		// Multi-digit: split into individual digits
		var patterns [][]string
		scoreStr := fmt.Sprintf("%d", score)
		for _, ch := range scoreStr {
			d := int(ch - '0')
			patterns = append(patterns, digits[d])
		}
		return patterns
	}

	homePatterns := getDigitPatterns(homeScore)
	awayPatterns := getDigitPatterns(awayScore)

	// Build 3-line score display
	var lines []string
	neonRed := lipgloss.Color("196")
	scoreStyle := lipgloss.NewStyle().Foreground(neonRed).Bold(true)

	for i := 0; i < 3; i++ {
		// Build home score line
		var homeLine string
		for j, p := range homePatterns {
			if j > 0 {
				homeLine += " " // Space between digits
			}
			homeLine += p[i]
		}

		// Build away score line
		var awayLine string
		for j, p := range awayPatterns {
			if j > 0 {
				awayLine += " " // Space between digits
			}
			awayLine += p[i]
		}

		line := homeLine + "  " + dash[i] + "  " + awayLine
		lines = append(lines, scoreStyle.Render(line))
	}

	// Join lines and center the entire block
	scoreBlock := strings.Join(lines, "\n")

	return lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center).
		Render(scoreBlock)
}
