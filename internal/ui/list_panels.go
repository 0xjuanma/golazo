package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/0xjuanma/golazo/internal/api"
	"github.com/0xjuanma/golazo/internal/constants"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
)

// RenderLiveMatchesListPanel renders the left panel using bubbletea list component.
// Note: listModel is passed by value, so SetSize must be called before this function.
// Uses Neon design with Golazo red/cyan theme.
func RenderLiveMatchesListPanel(width, height int, listModel list.Model) string {
	// Wrap list in panel with neon styling
	title := neonPanelTitleStyle.Width(width - 6).Render(constants.PanelLiveMatches)
	listView := listModel.View()

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		listView,
	)

	// Truncate inner content before applying border to preserve border rendering
	// neonPanelStyle has 2 lines for border (top + bottom), so inner height = height - 2
	innerHeight := height - 2
	if innerHeight > 0 {
		content = truncateToHeight(content, innerHeight)
	}

	panel := neonPanelStyle.
		Width(width).
		Height(height).
		Render(content)

	return panel
}

// RenderStatsListPanel renders the left panel for stats view using bubbletea list component.
// Note: listModel is passed by value, so SetSize must be called before this function.
// Uses Neon design with Golazo red/cyan theme.
// List titles are only shown when there are items. Empty lists show gray messages instead.
// For 1-day view, shows both finished and upcoming lists stacked vertically.
func RenderStatsListPanel(width, height int, finishedList list.Model, upcomingList list.Model, dateRange int) string {
	// Render date range selector with neon styling
	dateSelector := renderDateRangeSelector(width-6, dateRange)

	emptyStyle := neonEmptyStyle.Width(width - 6)

	var finishedListView string
	finishedItems := finishedList.Items()
	if len(finishedItems) == 0 {
		// No items - show empty message, no list title
		finishedListView = emptyStyle.Render(constants.EmptyNoFinishedMatches + "\n\nTry selecting a different date range (h/l keys)")
	} else {
		// Has items - show list (which includes its title)
		finishedListView = finishedList.View()
	}

	// For 1-day view, show both lists stacked vertically
	if dateRange == 1 {
		var upcomingListView string
		upcomingItems := upcomingList.Items()
		if len(upcomingItems) == 0 {
			// No upcoming matches - show empty message, no list title
			upcomingListView = emptyStyle.Render("No upcoming matches scheduled for today")
		} else {
			// Has items - show list (which includes its title)
			upcomingListView = upcomingList.View()
		}

		// Combine both lists with date selector
		content := lipgloss.JoinVertical(
			lipgloss.Left,
			dateSelector,
			"",
			finishedListView,
			"",
			upcomingListView,
		)

		// Truncate inner content before applying border to preserve border rendering
		innerHeight := height - 2
		if innerHeight > 0 {
			content = truncateToHeight(content, innerHeight)
		}

		panel := neonPanelStyle.
			Width(width).
			Height(height).
			Render(content)
		return panel
	}

	// For 3-day view, only show finished matches
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		dateSelector,
		"",
		finishedListView,
	)

	// Truncate inner content before applying border to preserve border rendering
	innerHeight := height - 2
	if innerHeight > 0 {
		content = truncateToHeight(content, innerHeight)
	}

	panel := neonPanelStyle.
		Width(width).
		Height(height).
		Render(content)

	return panel
}

// renderDateRangeSelector renders a horizontal date range selector (Today, 3d, 5d).
func renderDateRangeSelector(width int, selected int) string {
	options := []struct {
		days  int
		label string
	}{
		{1, "Today"},
		{3, "3d"},
		{5, "5d"},
	}

	items := make([]string, 0, len(options))
	for _, opt := range options {
		if opt.days == selected {
			// Selected option - neon red
			item := neonDateSelectedStyle.Render(opt.label)
			items = append(items, item)
		} else {
			// Unselected option - dim
			item := neonDateUnselectedStyle.Render(opt.label)
			items = append(items, item)
		}
	}

	// Join items with separator
	separator := "  "
	selector := strings.Join(items, separator)

	// Center the selector
	selectorStyle := lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center).
		Padding(0, 1)

	return selectorStyle.Render(selector)
}

// RenderMultiPanelViewWithList renders the live matches view with list component.
// leaguesLoaded and totalLeagues show loading progress during progressive loading.
// pollingSpinner and isPolling control the small polling indicator in the right panel.
func RenderMultiPanelViewWithList(width, height int, listModel list.Model, details *api.MatchDetails, liveUpdates []string, sp spinner.Model, loading bool, randomSpinner *RandomCharSpinner, viewLoading bool, leaguesLoaded int, totalLeagues int, pollingSpinner *RandomCharSpinner, isPolling bool) string {
	// Handle edge case: if width/height not set, use defaults
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 24
	}

	// Reserve 3 lines at top for spinner (always reserve to prevent layout shift)
	spinnerHeight := 3
	availableHeight := height - spinnerHeight
	if availableHeight < 10 {
		availableHeight = 10 // Minimum height for panels
	}

	// Render spinner centered in reserved space
	// ALWAYS use styled approach with explicit height to prevent layout shifts
	spinnerStyle := lipgloss.NewStyle().
		Width(width).
		Height(spinnerHeight).
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center)

	var spinnerArea string
	if viewLoading && randomSpinner != nil {
		spinnerView := randomSpinner.View()
		// Add progress indicator during progressive loading (batches of 4 leagues)
		var progressText string
		if totalLeagues > 0 && leaguesLoaded < totalLeagues {
			progressText = fmt.Sprintf("  Scanning batch %d/%d...", leaguesLoaded+1, totalLeagues)
		}
		if spinnerView != "" {
			spinnerArea = spinnerStyle.Render(spinnerView + progressText)
		} else {
			spinnerArea = spinnerStyle.Render("Loading..." + progressText)
		}
	} else {
		// Reserve space with empty styled box - explicit height prevents layout shifts
		spinnerArea = spinnerStyle.Render("")
	}

	// Calculate panel dimensions
	leftWidth := width * 35 / 100
	if leftWidth < 25 {
		leftWidth = 25
	}
	rightWidth := width - leftWidth - 1 // -1 for separator
	if rightWidth < 35 {
		rightWidth = 35
		leftWidth = width - rightWidth - 1
	}

	// Use panelHeight similar to stats view to ensure proper spacing
	panelHeight := availableHeight - 2

	// Render left panel (matches list) - shifted down
	leftPanel := RenderLiveMatchesListPanel(leftWidth, panelHeight, listModel)

	// Render right panel (match details with live updates) - shifted down
	rightPanel := renderMatchDetailsPanelWithPolling(rightWidth, panelHeight, details, liveUpdates, sp, loading, pollingSpinner, isPolling)

	// Create separator with neon red accent
	separatorStyle := neonSeparatorStyle.Height(panelHeight)
	separator := separatorStyle.Render("┃")

	// Combine panels
	panels := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftPanel,
		separator,
		rightPanel,
	)

	// Combine spinner area and panels - this shifts panels down
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		spinnerArea,
		panels,
	)

	return content
}

// RenderStatsViewWithList renders the stats view with list component.
// Rebuilt to match live view structure exactly: spinner at top, left panel (matches), right panel (details).
// daysLoaded and totalDays show loading progress during progressive loading.
func RenderStatsViewWithList(width, height int, finishedList list.Model, upcomingList list.Model, details *api.MatchDetails, randomSpinner *RandomCharSpinner, viewLoading bool, dateRange int, daysLoaded int, totalDays int) string {
	// Handle edge case: if width/height not set, use defaults
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 24
	}

	// Reserve 3 lines at top for spinner (always reserve to prevent layout shift)
	// Match live view exactly
	spinnerHeight := 3
	availableHeight := height - spinnerHeight
	if availableHeight < 10 {
		availableHeight = 10 // Minimum height for panels
	}

	// Render spinner centered in reserved space - match live view exactly
	// ALWAYS use styled approach with explicit height to prevent layout shifts
	spinnerStyle := lipgloss.NewStyle().
		Width(width).
		Height(spinnerHeight).
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center)

	var spinnerArea string
	if viewLoading && randomSpinner != nil {
		spinnerView := randomSpinner.View()
		// Add progress indicator during progressive loading
		var progressText string
		if totalDays > 0 && daysLoaded < totalDays {
			progressText = fmt.Sprintf("  Loading day %d/%d...", daysLoaded+1, totalDays)
		}
		if spinnerView != "" {
			spinnerArea = spinnerStyle.Render(spinnerView + progressText)
		} else {
			spinnerArea = spinnerStyle.Render("Loading..." + progressText)
		}
	} else {
		// Reserve space with empty styled box - explicit height prevents layout shifts
		spinnerArea = spinnerStyle.Render("")
	}

	// Calculate panel dimensions - match live view exactly (35% left, 65% right)
	leftWidth := width * 35 / 100
	if leftWidth < 25 {
		leftWidth = 25
	}
	rightWidth := width - leftWidth - 1 // -1 for separator
	if rightWidth < 35 {
		rightWidth = 35
		leftWidth = width - rightWidth - 1
	}

	// Use panelHeight similar to live view to ensure proper spacing
	panelHeight := availableHeight - 2

	// Render left panel (finished matches list) - match live view structure
	// For 1-day view, combine finished and upcoming lists vertically
	leftPanel := RenderStatsListPanel(leftWidth, panelHeight, finishedList, upcomingList, dateRange)

	// Render right panel (match details) - use dedicated stats panel renderer
	rightPanel := renderStatsMatchDetailsPanel(rightWidth, panelHeight, details)

	// Create separator with neon red accent
	separatorStyle := neonSeparatorStyle.Height(panelHeight)
	separator := separatorStyle.Render("┃")

	// Combine panels
	panels := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftPanel,
		separator,
		rightPanel,
	)

	// Combine spinner area and panels - this shifts panels down
	// Match live view exactly - use lipgloss.Left
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		spinnerArea,
		panels,
	)

	return content
}

// renderStatsMatchDetailsPanel renders the right panel for stats view with match details.
// Uses Neon design with Golazo red/cyan theme.
// Displays expanded match information including statistics, lineups, and more.
func renderStatsMatchDetailsPanel(width, height int, details *api.MatchDetails) string {
	if details == nil {
		emptyMessage := neonDimStyle.
			Align(lipgloss.Center).
			Width(width - 6).
			PaddingTop(height / 4).
			Render("Select a match to view details")

		return neonPanelCyanStyle.
			Width(width).
			Height(height).
			MaxHeight(height).
			Render(emptyMessage)
	}

	contentWidth := width - 6 // Account for border padding
	var lines []string

	// Team names
	homeTeam := details.HomeTeam.ShortName
	if homeTeam == "" {
		homeTeam = details.HomeTeam.Name
	}
	awayTeam := details.AwayTeam.ShortName
	if awayTeam == "" {
		awayTeam = details.AwayTeam.Name
	}

	// ═══════════════════════════════════════════════
	// MATCH HEADER
	// ═══════════════════════════════════════════════
	lines = append(lines, neonHeaderStyle.Render("Match Info"))
	lines = append(lines, "")

	// Line 1: Team A vs Team B (centered)
	teamsDisplay := fmt.Sprintf("%s  vs  %s",
		neonTeamStyle.Render(homeTeam),
		neonTeamStyle.Render(awayTeam))
	lines = append(lines, lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Center).Render(teamsDisplay))
	lines = append(lines, "")

	// Line 2: Large score (like live view)
	if details.HomeScore != nil && details.AwayScore != nil {
		largeScore := renderLargeScore(*details.HomeScore, *details.AwayScore, contentWidth)
		lines = append(lines, largeScore)
	} else {
		vsText := lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")).
			Width(contentWidth).
			Align(lipgloss.Center).
			Render("vs")
		lines = append(lines, vsText)
	}
	lines = append(lines, "")

	// Match context row
	if details.League.Name != "" {
		lines = append(lines, neonLabelStyle.Render("League:      ")+neonValueStyle.Render(details.League.Name))
	}
	if details.Venue != "" {
		lines = append(lines, neonLabelStyle.Render("Venue:       ")+neonValueStyle.Render(truncateString(details.Venue, contentWidth-14)))
	}
	if details.MatchTime != nil {
		lines = append(lines, neonLabelStyle.Render("Date:        ")+neonValueStyle.Render(details.MatchTime.Format("02 Jan 2006, 15:04")))
	}
	if details.Referee != "" {
		lines = append(lines, neonLabelStyle.Render("Referee:     ")+neonValueStyle.Render(details.Referee))
	}
	if details.Attendance > 0 {
		lines = append(lines, neonLabelStyle.Render("Attendance:  ")+neonValueStyle.Render(formatNumber(details.Attendance)))
	}

	// ═══════════════════════════════════════════════
	// GOALS TIMELINE
	// ═══════════════════════════════════════════════
	var homeGoals, awayGoals []api.MatchEvent
	for _, event := range details.Events {
		if event.Type == "goal" {
			if event.Team.ID == details.HomeTeam.ID {
				homeGoals = append(homeGoals, event)
			} else {
				awayGoals = append(awayGoals, event)
			}
		}
	}

	if len(homeGoals) > 0 || len(awayGoals) > 0 {
		lines = append(lines, "")
		lines = append(lines, neonHeaderStyle.Render("Goals"))

		if len(homeGoals) > 0 {
			lines = append(lines, neonTeamStyle.Render(homeTeam))
			for _, g := range homeGoals {
				goalLine := renderGoalLine(g, contentWidth-2)
				lines = append(lines, "  "+goalLine)
			}
		}

		if len(awayGoals) > 0 {
			lines = append(lines, neonTeamStyle.Render(awayTeam))
			for _, g := range awayGoals {
				goalLine := renderGoalLine(g, contentWidth-2)
				lines = append(lines, "  "+goalLine)
			}
		}
	}

	// ═══════════════════════════════════════════════
	// CARDS - Detailed list with player, minute, team
	// ═══════════════════════════════════════════════
	var cardEvents []api.MatchEvent
	for _, event := range details.Events {
		if event.Type == "card" {
			cardEvents = append(cardEvents, event)
		}
	}

	if len(cardEvents) > 0 {
		lines = append(lines, "")
		lines = append(lines, neonHeaderStyle.Render("Cards"))

		for _, card := range cardEvents {
			player := "Unknown"
			if card.Player != nil {
				player = *card.Player
			}
			teamName := card.Team.ShortName
			if teamName == "" {
				teamName = card.Team.Name
			}

			// Determine card type and apply appropriate color (using shared styles)
			cardSymbol := CardSymbolYellow
			cardStyle := neonYellowCardStyle
			if card.EventType != nil && (*card.EventType == "red" || *card.EventType == "redcard" || *card.EventType == "secondyellow") {
				cardSymbol = CardSymbolRed
				cardStyle = neonRedCardStyle
			}

			// Format: ▪ 28' PlayerName (Team)
			cardLine := fmt.Sprintf("  %s %s %s (%s)",
				cardStyle.Render(cardSymbol),
				neonScoreStyle.Render(fmt.Sprintf("%d'", card.Minute)),
				neonValueStyle.Render(player),
				neonDimStyle.Render(teamName))
			lines = append(lines, cardLine)
		}
	}

	// ═══════════════════════════════════════════════
	// MATCH STATISTICS (Visual Progress Bars)
	// ═══════════════════════════════════════════════
	if len(details.Statistics) > 0 {
		lines = append(lines, "")
		lines = append(lines, neonHeaderStyle.Render("Statistics"))

		// Only show these 5 specific stats
		wantedStats := []struct {
			patterns   []string
			label      string
			isProgress bool // true = show as progress bar
		}{
			{[]string{"possession", "ball possession", "ballpossesion"}, "Possession", true},
			{[]string{"total_shots", "total shots"}, "Total Shots", false},
			{[]string{"shots_on_target", "on target", "shotsontarget"}, "Shots on Target", false},
			{[]string{"accurate_passes", "accurate passes"}, "Accurate Passes", false},
			{[]string{"fouls", "fouls committed"}, "Fouls", false},
		}

		// Style for centering stat blocks
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
					// Add spacing before each stat
					lines = append(lines, "")

					if wanted.isProgress {
						// Render as visual progress bar (centered)
						statLine := renderStatProgressBar(wanted.label, stat.HomeValue, stat.AwayValue, contentWidth, homeTeam, awayTeam)
						lines = append(lines, centerStyle.Render(statLine))
					} else {
						// Render as comparison bar (centered)
						statLine := renderStatComparison(wanted.label, stat.HomeValue, stat.AwayValue, contentWidth)
						lines = append(lines, centerStyle.Render(statLine))
					}
					break
				}
			}
		}
	}

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)

	return neonPanelCyanStyle.
		Width(width).
		Height(height).
		MaxHeight(height).
		Render(content)
}

// RenderMatchDetailsPanel is an exported version of renderStatsMatchDetailsPanel
// for use by debug scripts. Renders match details in the Golazo stats view style.
func RenderMatchDetailsPanel(width, height int, details *api.MatchDetails) string {
	return renderStatsMatchDetailsPanel(width, height, details)
}

// renderGoalLine renders a single goal with scorer, minute, and assist
func renderGoalLine(g api.MatchEvent, maxWidth int) string {
	player := "Unknown"
	if g.Player != nil {
		player = *g.Player
	}

	minuteStr := neonScoreStyle.Render(fmt.Sprintf("%d'", g.Minute))
	playerStr := neonValueStyle.Render(truncateString(player, maxWidth-10))

	line := fmt.Sprintf("%s %s", minuteStr, playerStr)

	// Add assist if available
	if g.Assist != nil && *g.Assist != "" {
		line += neonDimStyle.Render(fmt.Sprintf(" (%s)", truncateString(*g.Assist, 15)))
	}

	return line
}

// Fixed bar width for consistent UI
const statBarWidth = 20

// renderStatProgressBar renders a stat as a visual progress bar using bubbles progress component
// Uses gradient fill from cyan to red for the Golazo theme
// Fixed width of 20 squares for consistent UI
// Renders label on first line, bar on second line (both centered)
func renderStatProgressBar(label, homeVal, awayVal string, maxWidth int, homeTeam, awayTeam string) string {
	// Parse percentage values (e.g., "59" or "59%")
	homePercent := parsePercent(homeVal)
	awayPercent := parsePercent(awayVal)

	// Normalize if they don't add up to 100
	total := homePercent + awayPercent
	if total > 0 && total != 100 {
		homePercent = (homePercent * 100) / total
		awayPercent = 100 - homePercent
	}

	// Create bubbles progress bar with gradient (cyan -> red for Golazo theme)
	prog := progress.New(
		progress.WithScaledGradient("#00FFFF", "#FF0055"), // Cyan to Red gradient
		progress.WithWidth(statBarWidth),
		progress.WithoutPercentage(),
	)

	// Render the progress bar at home team's percentage
	progressView := prog.ViewAs(float64(homePercent) / 100.0)

	// Format values
	homeValStyled := neonValueStyle.Render(fmt.Sprintf("%3d%%", homePercent))
	awayValStyled := neonDimStyle.Render(fmt.Sprintf("%3d%%", awayPercent))

	// Line 1: Label (centered via parent, no width constraint)
	labelStyle := lipgloss.NewStyle().Foreground(neonDim)
	labelLine := labelStyle.Render(label)

	// Line 2: Bar with values
	barLine := fmt.Sprintf("%s %s %s", homeValStyled, progressView, awayValStyled)

	return labelLine + "\n" + barLine
}

// renderStatComparison renders a stat as a visual comparison (for counts like shots, fouls)
// Fixed width of 20 squares total (10 per side) for consistent UI
// Renders label on first line, bar on second line (both centered)
func renderStatComparison(label, homeVal, awayVal string, maxWidth int) string {
	// Parse numeric values
	homeNum := parseNumber(homeVal)
	awayNum := parseNumber(awayVal)

	// Determine who has more (for highlighting)
	homeStyle := neonValueStyle
	awayStyle := neonValueStyle
	if homeNum > awayNum {
		homeStyle = lipgloss.NewStyle().Foreground(neonCyan).Bold(true)
	} else if awayNum > homeNum {
		awayStyle = lipgloss.NewStyle().Foreground(neonCyan).Bold(true)
	}

	// Fixed bar width: 10 squares per side = 20 total
	halfBar := statBarWidth / 2

	// Visual bar comparison - proportional to max value
	maxVal := max(homeNum, awayNum)
	if maxVal == 0 {
		maxVal = 1
	}

	// Home bar (right-aligned, grows left)
	homeFilled := (homeNum * halfBar) / maxVal
	if homeFilled > halfBar {
		homeFilled = halfBar
	}
	homeEmpty := halfBar - homeFilled
	homeBar := strings.Repeat(" ", homeEmpty) + strings.Repeat("▪", homeFilled)
	homeBarStyled := lipgloss.NewStyle().Foreground(neonCyan).Render(homeBar)

	// Away bar (left-aligned, grows right)
	awayFilled := (awayNum * halfBar) / maxVal
	if awayFilled > halfBar {
		awayFilled = halfBar
	}
	awayEmpty := halfBar - awayFilled
	awayBar := strings.Repeat("▪", awayFilled) + strings.Repeat(" ", awayEmpty)
	awayBarStyled := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(awayBar)

	// Line 1: Label (centered via parent, no width constraint)
	labelStyle := lipgloss.NewStyle().Foreground(neonDim)
	labelLine := labelStyle.Render(label)

	// Line 2: Bar with values
	barLine := fmt.Sprintf("%s %s %s %s",
		homeStyle.Render(fmt.Sprintf("%10s", homeVal)),
		homeBarStyled,
		awayBarStyled,
		awayStyle.Render(fmt.Sprintf("%-10s", awayVal)))

	return labelLine + "\n" + barLine
}

// parsePercent extracts a percentage value from a string like "59" or "59%"
func parsePercent(s string) int {
	s = strings.TrimSuffix(s, "%")
	s = strings.TrimSpace(s)
	val, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return val
}

// parseNumber extracts a numeric value from a string, handling formats like "476 (89%)"
func parseNumber(s string) int {
	// Handle formats like "476 (89%)" - extract first number
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

// truncateString truncates a string to maxLen, adding "..." if truncated
func truncateString(s string, maxLen int) string {
	if maxLen <= 3 {
		return s
	}
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// formatNumber formats a number with thousand separators
func formatNumber(n int) string {
	s := fmt.Sprintf("%d", n)
	if n < 1000 {
		return s
	}

	// Insert commas from right to left
	result := ""
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result += ","
		}
		result += string(c)
	}
	return result
}

// min returns the smaller of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// truncateToHeight truncates content to fit within maxLines.
// This is used to truncate inner content before applying bordered styles,
// ensuring borders are always rendered completely.
func truncateToHeight(content string, maxLines int) string {
	if maxLines <= 0 {
		return content
	}

	lines := strings.Split(content, "\n")
	if len(lines) <= maxLines {
		return content
	}

	return strings.Join(lines[:maxLines], "\n")
}
