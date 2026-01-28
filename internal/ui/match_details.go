package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/0xjuanma/golazo/internal/api"
	"github.com/0xjuanma/golazo/internal/constants"
	"github.com/0xjuanma/golazo/internal/ui/design"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/lipgloss"
)

// MatchDetailsConfig holds all parameters for rendering match details.
type MatchDetailsConfig struct {
	Width, Height int
	Details       *api.MatchDetails
	GoalLinks     GoalLinksMap

	// View-specific features
	ShowStatistics bool // Stats view only
	ShowHighlights bool // Stats view only

	// Live view state
	LiveUpdates    []string
	PollingSpinner *RandomCharSpinner
	IsPolling      bool
	Loading        bool

	// Stats view state
	Focused bool
}

// RenderMatchDetails renders match details content, returning header and scrollable content separately.
// This unified function is used by both live and stats views.
func RenderMatchDetails(cfg MatchDetailsConfig) (headerContent, scrollableContent string) {
	if cfg.Details == nil {
		return "", ""
	}

	contentWidth := cfg.Width - 6
	details := cfg.Details

	var headerLines []string
	var scrollableLines []string

	// Team names
	homeTeam := details.HomeTeam.ShortName
	if homeTeam == "" {
		homeTeam = details.HomeTeam.Name
	}
	awayTeam := details.AwayTeam.ShortName
	if awayTeam == "" {
		awayTeam = details.AwayTeam.Name
	}

	// Header with optional focus styling
	headerLines = append(headerLines, renderPanelHeader("Match Details", cfg.Focused))
	headerLines = append(headerLines, "")

	// Status and league info
	headerLines = append(headerLines, renderStatusLine(details, contentWidth))
	headerLines = append(headerLines, "")

	// Teams display
	teamsDisplay := fmt.Sprintf("%s  vs  %s",
		neonTeamStyle.Render(homeTeam),
		neonTeamStyle.Render(awayTeam))
	headerLines = append(headerLines, lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Center).Render(teamsDisplay))
	headerLines = append(headerLines, "")

	// Large score
	if details.HomeScore != nil && details.AwayScore != nil {
		headerLines = append(headerLines, renderLargeScore(*details.HomeScore, *details.AwayScore, contentWidth))
	} else {
		vsText := lipgloss.NewStyle().
			Foreground(neonDim).
			Width(contentWidth).
			Align(lipgloss.Center).
			Render("vs")
		headerLines = append(headerLines, vsText)
	}
	headerLines = append(headerLines, "")

	// Match context (detailed info)
	headerLines = append(headerLines, renderMatchContext(details, contentWidth)...)

	// Penalties (prominent section)
	if details.Penalties != nil && details.Penalties.Home != nil && details.Penalties.Away != nil {
		headerLines = append(headerLines, renderPenaltiesSection(details, contentWidth)...)
	}

	// For live matches, show live updates instead of event details
	if details.Status == api.MatchStatusLive || details.Status == api.MatchStatusNotStarted {
		liveSection := renderLiveUpdatesSection(cfg, contentWidth)
		scrollableLines = append(scrollableLines, liveSection)
	} else {
		// Finished match content
		if cfg.ShowHighlights && details.Highlight != nil && details.Highlight.URL != "" {
			scrollableLines = append(scrollableLines, "")
			highlightLink := lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Center).Render(
				Hyperlink("▶ Official Match Highlights", details.Highlight.URL),
			)
			scrollableLines = append(scrollableLines, neonValueStyle.Render(highlightLink))
		}

		// Goals section (with gradient)
		goalsSection := renderGoalsSection(cfg, contentWidth)
		if goalsSection != "" {
			scrollableLines = append(scrollableLines, goalsSection)
		}

		// Cards section
		cardsSection := renderCardsSection(cfg, contentWidth)
		if cardsSection != "" {
			scrollableLines = append(scrollableLines, cardsSection)
		}

		// Statistics section (stats view only)
		if cfg.ShowStatistics && len(details.Statistics) > 0 {
			statsSection := renderStatisticsSection(cfg, contentWidth, homeTeam, awayTeam)
			scrollableLines = append(scrollableLines, statsSection)
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, headerLines...),
		lipgloss.JoinVertical(lipgloss.Left, scrollableLines...)
}

func renderPanelHeader(title string, focused bool) string {
	if focused {
		return lipgloss.NewStyle().
			Foreground(neonCyan).
			Bold(true).
			PaddingBottom(0).
			BorderBottom(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(neonCyan).
			MarginBottom(0).
			Render(title)
	}
	return lipgloss.NewStyle().
		Foreground(neonDim).
		Bold(true).
		PaddingBottom(0).
		BorderBottom(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(neonDim).
		MarginBottom(0).
		Render(title)
}

func renderStatusLine(details *api.MatchDetails, contentWidth int) string {
	infoStyle := lipgloss.NewStyle().Foreground(neonDim)
	var statusText string
	switch details.Status {
	case api.MatchStatusLive:
		liveTime := constants.StatusLive
		if details.LiveTime != nil {
			liveTime = *details.LiveTime
		}
		statusText = lipgloss.NewStyle().Foreground(neonRed).Bold(true).Render(liveTime)
	case api.MatchStatusFinished:
		statusText = lipgloss.NewStyle().Foreground(neonCyan).Render(constants.StatusFinished)
	default:
		statusText = infoStyle.Render(constants.StatusNotStartedShort)
	}

	leagueText := infoStyle.Italic(true).Render(details.League.Name)
	return lipgloss.NewStyle().
		Width(contentWidth).
		Align(lipgloss.Center).
		Render(statusText + " • " + leagueText)
}

func renderMatchContext(details *api.MatchDetails, contentWidth int) []string {
	var lines []string

	if details.League.Name != "" {
		lines = append(lines, neonLabelStyle.Render("League:      ")+neonValueStyle.Render(details.League.Name))
	}
	if details.Venue != "" {
		lines = append(lines, neonLabelStyle.Render("Venue:       ")+neonValueStyle.Render(truncateString(details.Venue, contentWidth-14)))
	}
	if details.MatchTime != nil {
		lines = append(lines, neonLabelStyle.Render("Date:        ")+neonValueStyle.Render(details.MatchTime.Format("02 Jan 2006, 15:04")+" UTC"))
	}
	if details.Referee != "" {
		lines = append(lines, neonLabelStyle.Render("Referee:     ")+neonValueStyle.Render(details.Referee))
	}
	if details.Attendance > 0 {
		lines = append(lines, neonLabelStyle.Render("Attendance:  ")+neonValueStyle.Render(formatNumber(details.Attendance)))
	}

	// Half-time score
	if details.HalfTimeScore != nil && details.HalfTimeScore.Home != nil && details.HalfTimeScore.Away != nil {
		htText := fmt.Sprintf("HT: %d - %d", *details.HalfTimeScore.Home, *details.HalfTimeScore.Away)
		lines = append(lines, neonLabelStyle.Render("Half-time:   ")+neonValueStyle.Render(htText))
	}

	// Extra time
	if details.ExtraTime {
		lines = append(lines, neonLabelStyle.Render("Duration:    ")+neonValueStyle.Render("After Extra Time"))
	}

	return lines
}

func renderPenaltiesSection(details *api.MatchDetails, contentWidth int) []string {
	var lines []string
	lines = append(lines, "")

	penaltyHeader := lipgloss.NewStyle().
		Foreground(neonRed).
		Bold(true).
		Width(contentWidth).
		Align(lipgloss.Center).
		Render("PENALTIES")
	lines = append(lines, penaltyHeader)

	penaltyScoreText := fmt.Sprintf("%d - %d", *details.Penalties.Home, *details.Penalties.Away)
	penaltyScore := lipgloss.NewStyle().
		Foreground(neonCyan).
		Bold(true).
		Width(contentWidth).
		Align(lipgloss.Center).
		Render(penaltyScoreText)
	lines = append(lines, penaltyScore)
	lines = append(lines, "")

	return lines
}

func renderGoalsSection(cfg MatchDetailsConfig, contentWidth int) string {
	details := cfg.Details
	var goals []api.MatchEvent
	for _, event := range details.Events {
		if event.Type == "goal" {
			goals = append(goals, event)
		}
	}

	if len(goals) == 0 {
		return ""
	}

	var lines []string
	lines = append(lines, "")
	lines = append(lines, neonHeaderStyle.Render("Goals"))

	for _, goal := range goals {
		player := "Unknown"
		if goal.Player != nil {
			player = *goal.Player
		}
		isHome := goal.Team.ID == details.HomeTeam.ID

		playerDetails := neonValueStyle.Render(player)
		replayIndicator := getReplayIndicator(details, cfg.GoalLinks, goal.Minute)

		// Use gradient for GOAL label
		styledGoal := design.ApplyGradientToText("GOAL")
		goalContent := buildEventContent(playerDetails, replayIndicator, "●", styledGoal, isHome)

		minuteStr := goal.DisplayMinute
		if minuteStr == "" {
			minuteStr = fmt.Sprintf("%d'", goal.Minute)
		}
		lines = append(lines, renderCenterAlignedEvent(minuteStr, goalContent, isHome, contentWidth))
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func renderCardsSection(cfg MatchDetailsConfig, contentWidth int) string {
	details := cfg.Details
	var cardEvents []api.MatchEvent
	for _, event := range details.Events {
		if event.Type == "card" {
			cardEvents = append(cardEvents, event)
		}
	}

	if len(cardEvents) == 0 {
		return ""
	}

	var lines []string
	lines = append(lines, "")
	lines = append(lines, neonHeaderStyle.Render("Cards"))

	for _, card := range cardEvents {
		player := "Unknown"
		if card.Player != nil {
			player = *card.Player
		}
		isHome := card.Team.ID == details.HomeTeam.ID

		cardSymbol := CardSymbolYellow
		cardStyle := neonYellowCardStyle
		if card.EventType != nil && (*card.EventType == "red" || *card.EventType == "redcard" || *card.EventType == "secondyellow") {
			cardSymbol = CardSymbolRed
			cardStyle = neonRedCardStyle
		}

		playerDetails := neonValueStyle.Render(player)
		cardContent := buildEventContent(playerDetails, "", cardSymbol, cardStyle.Render("CARD"), isHome)

		minuteStr := card.DisplayMinute
		if minuteStr == "" {
			minuteStr = fmt.Sprintf("%d'", card.Minute)
		}
		lines = append(lines, renderCenterAlignedEvent(minuteStr, cardContent, isHome, contentWidth))
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func renderStatisticsSection(cfg MatchDetailsConfig, contentWidth int, homeTeam, awayTeam string) string {
	details := cfg.Details
	var lines []string
	lines = append(lines, "")
	lines = append(lines, neonHeaderStyle.Render("Statistics"))

	wantedStats := []struct {
		patterns   []string
		label      string
		isProgress bool
	}{
		{[]string{"possession", "ball possession", "ballpossesion"}, "Possession", true},
		{[]string{"total_shots", "total shots"}, "Total Shots", false},
		{[]string{"shots_on_target", "on target", "shotsontarget"}, "Shots on Target", false},
		{[]string{"accurate_passes", "accurate passes"}, "Accurate Passes", false},
		{[]string{"fouls", "fouls committed"}, "Fouls", false},
	}

	centerStyle := lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Center)

	for _, wanted := range wantedStats {
		for _, stat := range details.Statistics {
			keyLower := strings.ToLower(stat.Key)
			labelLower := strings.ToLower(stat.Label)

			matched := false
			for _, pattern := range wanted.patterns {
				if strings.Contains(keyLower, pattern) || strings.Contains(labelLower, pattern) {
					matched = true
					break
				}
			}

			if matched {
				lines = append(lines, "")
				if wanted.isProgress {
					statLine := renderStatProgressBar(wanted.label, stat.HomeValue, stat.AwayValue, contentWidth, homeTeam, awayTeam)
					lines = append(lines, centerStyle.Render(statLine))
				} else {
					statLine := renderStatComparison(wanted.label, stat.HomeValue, stat.AwayValue, contentWidth)
					lines = append(lines, centerStyle.Render(statLine))
				}
				break
			}
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func renderLiveUpdatesSection(cfg MatchDetailsConfig, contentWidth int) string {
	var lines []string

	var titleText string
	if cfg.IsPolling && cfg.Loading && cfg.PollingSpinner != nil {
		pollingView := cfg.PollingSpinner.View()
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
		BorderForeground(neonDarkDim).
		Width(cfg.Width - 6).
		Render(titleText)
	lines = append(lines, updatesTitle)

	if len(cfg.LiveUpdates) == 0 && !cfg.Loading && !cfg.IsPolling {
		emptyUpdates := lipgloss.NewStyle().
			Foreground(neonDim).
			Padding(0, 0).
			Render(constants.EmptyNoUpdates)
		lines = append(lines, emptyUpdates)
	} else if len(cfg.LiveUpdates) > 0 {
		for _, update := range cfg.LiveUpdates {
			updateLine := renderStyledLiveUpdate(update, contentWidth, cfg.Details, cfg.GoalLinks)
			lines = append(lines, updateLine)
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// Statistics rendering functions

const statBarWidth = 20

func renderStatProgressBar(label, homeVal, awayVal string, maxWidth int, homeTeam, awayTeam string) string {
	homePercent := parsePercent(homeVal)
	awayPercent := parsePercent(awayVal)

	total := homePercent + awayPercent
	if total > 0 && total != 100 {
		homePercent = (homePercent * 100) / total
		awayPercent = 100 - homePercent
	}

	prog := progress.New(
		progress.WithScaledGradient("#00FFFF", "#FF0055"),
		progress.WithWidth(statBarWidth),
		progress.WithoutPercentage(),
	)

	progressView := prog.ViewAs(float64(homePercent) / 100.0)

	homeValStyled := neonValueStyle.Render(fmt.Sprintf("%3d%%", homePercent))
	awayValStyled := neonDimStyle.Render(fmt.Sprintf("%3d%%", awayPercent))

	labelStyle := lipgloss.NewStyle().Foreground(neonDim)
	labelLine := labelStyle.Render(label)
	barLine := fmt.Sprintf("%s %s %s", homeValStyled, progressView, awayValStyled)

	return labelLine + "\n" + barLine
}

func renderStatComparison(label, homeVal, awayVal string, maxWidth int) string {
	homeNum := parseNumber(homeVal)
	awayNum := parseNumber(awayVal)

	homeStyle := neonValueStyle
	awayStyle := neonValueStyle
	if homeNum > awayNum {
		homeStyle = lipgloss.NewStyle().Foreground(neonCyan).Bold(true)
	} else if awayNum > homeNum {
		awayStyle = lipgloss.NewStyle().Foreground(neonCyan).Bold(true)
	}

	halfBar := statBarWidth / 2
	maxVal := max(homeNum, awayNum)
	if maxVal == 0 {
		maxVal = 1
	}

	homeFilled := min((homeNum*halfBar)/maxVal, halfBar)
	homeEmpty := halfBar - homeFilled
	homeBar := strings.Repeat(" ", homeEmpty) + strings.Repeat("▪", homeFilled)
	homeBarStyled := lipgloss.NewStyle().Foreground(neonCyan).Render(homeBar)

	awayFilled := min((awayNum*halfBar)/maxVal, halfBar)
	awayEmpty := halfBar - awayFilled
	awayBar := strings.Repeat("▪", awayFilled) + strings.Repeat(" ", awayEmpty)
	awayBarStyled := lipgloss.NewStyle().Foreground(neonGray).Render(awayBar)

	labelStyle := lipgloss.NewStyle().Foreground(neonDim)
	labelLine := labelStyle.Render(label)
	barLine := fmt.Sprintf("%s %s %s %s",
		homeStyle.Render(fmt.Sprintf("%10s", homeVal)),
		homeBarStyled,
		awayBarStyled,
		awayStyle.Render(fmt.Sprintf("%-10s", awayVal)))

	return labelLine + "\n" + barLine
}

func parsePercent(s string) int {
	s = strings.TrimSuffix(s, "%")
	s = strings.TrimSpace(s)
	val, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return val
}

func parseNumber(s string) int {
	s = strings.TrimSpace(s)
	if idx := strings.Index(s, " "); idx > 0 {
		s = s[:idx]
	}
	if idx := strings.Index(s, "("); idx > 0 {
		s = s[:idx]
	}
	s = strings.TrimSpace(s)

	val, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return val
}

func truncateString(s string, maxLen int) string {
	if maxLen <= 3 {
		return s
	}
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func formatNumber(n int) string {
	s := fmt.Sprintf("%d", n)
	if n < 1000 {
		return s
	}

	var result strings.Builder
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result.WriteString(",")
		}
		result.WriteString(string(c))
	}
	return result.String()
}
