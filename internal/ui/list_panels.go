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
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

// GoalLinksMap maps goal keys (matchID:minute) to replay URLs.
// Used to enhance goal event display with video replay links.
type GoalLinksMap map[string]string

// MakeGoalLinkKey creates a key for the goal links map.
func MakeGoalLinkKey(matchID, minute int) string {
	return fmt.Sprintf("%d:%d", matchID, minute)
}

// GetReplayURL returns the replay URL for a goal if available.
func (g GoalLinksMap) GetReplayURL(matchID, minute int) string {
	if g == nil {
		return ""
	}
	return g[MakeGoalLinkKey(matchID, minute)]
}

// RenderLiveMatchesListPanel renders the left panel using bubbletea list component.
// Note: listModel is passed by value, so SetSize must be called before this function.
// Uses Neon design with Golazo red/cyan theme.
// upcomingMatches are displayed at the bottom of the panel (fixed, not scrollable).
func RenderLiveMatchesListPanel(width, height int, listModel list.Model, upcomingMatches []MatchDisplay) string {
	contentWidth := width - 6 // Account for border and padding

	// Wrap list in panel with neon styling
	title := neonPanelTitleStyle.Width(contentWidth).Render(constants.PanelLiveMatches)

	// Check if list is empty - show custom message instead of list view to avoid duplicate "no items"
	var listView string
	if len(listModel.Items()) == 0 {
		listView = neonEmptyStyle.Width(contentWidth).Render(constants.EmptyNoLiveMatches)
	} else {
		listView = listModel.View()
	}

	// Calculate available inner height (minus borders)
	borderHeight := 2
	titleHeight := 2
	innerHeight := height - borderHeight - titleHeight

	// Build upcoming section if there are upcoming matches
	var upcomingSection string
	upcomingHeight := 0
	if len(upcomingMatches) > 0 {
		// Split 50-50 between live matches and upcoming
		maxUpcomingHeight := innerHeight / 2

		// Render upcoming section header
		upcomingTitle := neonHeaderStyle.Render("Upcoming")

		// Render upcoming matches as simple text (not selectable)
		var upcomingLines []string
		upcomingLines = append(upcomingLines, upcomingTitle)
		for _, match := range upcomingMatches {
			matchLine := renderUpcomingMatchLine(match, contentWidth)
			upcomingLines = append(upcomingLines, matchLine)
		}
		upcomingSection = strings.Join(upcomingLines, "\n")

		// Truncate upcoming section if it exceeds max height
		upcomingHeight = len(upcomingLines) + 1 // +1 for separator
		if upcomingHeight > maxUpcomingHeight {
			upcomingSection = truncateToHeight(upcomingSection, maxUpcomingHeight)
			upcomingHeight = maxUpcomingHeight
		}
	}

	// Calculate available height for the live list
	availableListHeight := max(
		// -1 for separator
		innerHeight-upcomingHeight-1,
		// Minimum height for list
		3)

	// Truncate list view to fit
	listView = truncateToHeight(listView, availableListHeight)

	// Build content
	var content string
	if upcomingHeight > 0 {
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			title,
			"",
			listView,
			"",
			upcomingSection,
		)
	} else {
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			title,
			"",
			listView,
		)
	}

	// Truncate inner content before applying border to preserve border rendering
	totalInnerHeight := height - 2
	if totalInnerHeight > 0 {
		content = truncateToHeight(content, totalInnerHeight)
	}

	panel := neonPanelStyle.
		Width(width).
		Height(height).
		Render(content)

	return panel
}

// renderUpcomingMatchLine renders a single upcoming match as a simple text line.
func renderUpcomingMatchLine(match MatchDisplay, maxWidth int) string {
	// Format: "  HH:MM  Team A vs Team B"
	var timeStr string
	if match.MatchTime != nil {
		timeStr = match.MatchTime.Local().Format("15:04")
	} else {
		timeStr = "--:--"
	}

	homeTeam := match.HomeTeam.ShortName
	if homeTeam == "" {
		homeTeam = match.HomeTeam.Name
	}
	awayTeam := match.AwayTeam.ShortName
	if awayTeam == "" {
		awayTeam = match.AwayTeam.Name
	}

	// Truncate team names if too long
	maxTeamLen := (maxWidth - 15) / 2 // 15 = time(5) + " vs "(4) + padding(6)
	if len(homeTeam) > maxTeamLen {
		homeTeam = homeTeam[:maxTeamLen-1] + "…"
	}
	if len(awayTeam) > maxTeamLen {
		awayTeam = awayTeam[:maxTeamLen-1] + "…"
	}

	timeStyle := neonDimStyle
	teamStyle := neonValueStyle

	return fmt.Sprintf("  %s  %s vs %s",
		timeStyle.Render(timeStr),
		teamStyle.Render(homeTeam),
		teamStyle.Render(awayTeam))
}

// RenderStatsListPanel renders the left panel for stats view using bubbletea list component.
// Note: listModel is passed by value, so SetSize must be called before this function.
// Uses Neon design with Golazo red/cyan theme.
// List titles are only shown when there are items. Empty lists show gray messages instead.
// Upcoming matches are now shown in the Live view instead.
func RenderStatsListPanel(width, height int, finishedList list.Model, dateRange int, rightPanelFocused bool) string {
	// Add header with focus state (color-based, not text-based)
	var header string
	if rightPanelFocused {
		// Dimmed header when right panel is focused
		header = lipgloss.NewStyle().
			Foreground(neonDim).
			Bold(true).
			PaddingBottom(0).
			BorderBottom(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(neonDim).
			MarginBottom(0).
			Render("Match List")
	} else {
		// Bright header when left panel is focused
		header = neonHeaderStyle.Render("Match List")
	}

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

	// Show finished matches with header, date selector
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		dateSelector,
		"",
		finishedListView,
	)

	// Truncate inner content before applying border to preserve border rendering
	innerHeight := height - 2
	if innerHeight > 0 {
		content = truncateToHeight(content, innerHeight)
	}

	// Apply panel styling based on focus state
	var panel string
	if rightPanelFocused {
		// Left panel unfocused - dim red border
		panel = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(neonDim).
			Padding(0, 1).
			Width(width).
			Height(height).
			Render(content)
	} else {
		// Left panel focused - bright red border
		panel = neonPanelStyle.
			Width(width).
			Height(height).
			Render(content)
	}

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
// upcomingMatches are displayed at the bottom of the left panel (fixed, not scrollable).
func RenderMultiPanelViewWithList(width, height int, listModel list.Model, details *api.MatchDetails, liveUpdates []string, sp spinner.Model, loading bool, randomSpinner *RandomCharSpinner, viewLoading bool, leaguesLoaded int, totalLeagues int, pollingSpinner *RandomCharSpinner, isPolling bool, upcomingMatches []MatchDisplay, goalLinks GoalLinksMap, bannerType constants.StatusBannerType) string {
	// Handle edge case: if width/height not set, use defaults
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 24
	}

	// Reserve 3 lines at top for spinner (always reserve to prevent layout shift)
	spinnerHeight := 3
	availableHeight := max(height-spinnerHeight,
		// Minimum height for panels
		10)

	// Render spinner centered in reserved space
	// ALWAYS use styled approach with explicit height to prevent layout shifts
	spinnerStyle := lipgloss.NewStyle().
		Width(width).
		Height(spinnerHeight).
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center)

	// Debug indicator moved to separate status line

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
	leftWidth := max(width*35/100, 25)
	rightWidth := width - leftWidth - 1 // -1 for separator
	if rightWidth < 35 {
		rightWidth = 35
		leftWidth = width - rightWidth - 1
	}

	// Use panelHeight similar to stats view to ensure proper spacing
	panelHeight := availableHeight - 2

	// Render left panel (matches list) - shifted down
	// Upcoming matches are displayed at the bottom of the left panel
	leftPanel := RenderLiveMatchesListPanel(leftWidth, panelHeight, listModel, upcomingMatches)

	// Render right panel (match details with live updates) - shifted down
	rightPanel := renderMatchDetailsPanelWithPolling(rightWidth, panelHeight, details, liveUpdates, sp, loading, pollingSpinner, isPolling, goalLinks)

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

	// Add status banner between spinner and panels (more stable than spinner area)
	statusBanner := renderStatusBanner(bannerType, width)

	// Combine spinner area, status banner, and panels
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		spinnerArea,
		statusBanner,
		panels,
	)

	return content
}

// RenderStatsViewWithList renders the stats view with list component.
// Rebuilt to match live view structure exactly: spinner at top, left panel (matches), right panel (details).
// daysLoaded and totalDays show loading progress during progressive loading.
// Note: Upcoming matches are now shown in the Live view instead.
func RenderStatsViewWithList(width, height int, finishedList list.Model, details *api.MatchDetails, randomSpinner *RandomCharSpinner, viewLoading bool, dateRange int, daysLoaded int, totalDays int, goalLinks GoalLinksMap, bannerType constants.StatusBannerType, detailsViewport *viewport.Model, rightPanelFocused bool, scrollOffset int) string {
	// Handle edge case: if width/height not set, use defaults
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 24
	}

	// Reserve 3 lines at top for spinner (always reserve to prevent layout shift)
	spinnerHeight := 3
	availableHeight := max(height-spinnerHeight,
		// Minimum height for panels
		10)

	// Render spinner centered in reserved space - match live view exactly
	// ALWAYS use styled approach with explicit height to prevent layout shifts
	spinnerStyle := lipgloss.NewStyle().
		Width(width).
		Height(spinnerHeight).
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center)

	// Debug indicator moved to separate status line

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
	leftWidth := max(width*35/100, 25)
	rightWidth := width - leftWidth - 1 // -1 for separator
	if rightWidth < 35 {
		rightWidth = 35
		leftWidth = width - rightWidth - 1
	}

	// Use panelHeight similar to live view to ensure proper spacing
	panelHeight := availableHeight - 2

	// Render left panel (finished matches list) - match live view structure
	leftPanel := RenderStatsListPanel(leftWidth, panelHeight, finishedList, dateRange, rightPanelFocused)

	// Render right panel (match details) - split into fixed header and scrollable content
	headerContent, scrollableContent := renderStatsMatchDetailsPanel(rightWidth, panelHeight, details, goalLinks, rightPanelFocused)

	var rightPanel string

	// Manual scrolling: split content into lines and show visible portion
	scrollableLines := strings.Split(scrollableContent, "\n")

	// Calculate available height for scrolling (reserve space for header)
	headerHeight := strings.Count(headerContent, "\n") + 1
	availableHeight = max(panelHeight-headerHeight,
		// Minimum scrollable area
		3)

	// Apply manual scroll offset when focused, otherwise show beginning of content
	visibleLines := scrollableLines
	if rightPanelFocused && len(scrollableLines) > availableHeight {
		// Show only visible portion based on scroll offset
		start := scrollOffset
		end := min(start+availableHeight, len(scrollableLines))
		if start < len(scrollableLines) && start >= 0 {
			visibleLines = scrollableLines[start:end]
		} else if start >= len(scrollableLines) {
			visibleLines = []string{}
		}
		// When focused, ensure we have enough lines to fill the available height
		for len(visibleLines) < availableHeight && len(visibleLines) < len(scrollableLines) {
			visibleLines = append(visibleLines, "")
		}
	} else {
		// When not focused, show the beginning of content (same as scroll offset 0)
		// This ensures consistent display when toggling focus
		if len(scrollableLines) > availableHeight {
			visibleLines = scrollableLines[:availableHeight]
		}
		// Reset scroll offset when not focused to maintain consistency
		if scrollOffset != 0 {
			// Note: This is just for display consistency, the actual offset reset happens in the tab handler
		}
	}

	// Combine visible lines back into content
	visibleContent := strings.Join(visibleLines, "\n")

	// Always include header
	rightPanel = lipgloss.JoinVertical(lipgloss.Left, headerContent, visibleContent)

	// Apply panel styling based on focus state (only top/bottom borders for right panel)
	if rightPanelFocused {
		// Focused right panel - bright neon cyan top/bottom borders
		rightPanel = lipgloss.NewStyle().
			BorderTop(true).
			BorderBottom(true).
			BorderForeground(neonCyan).
			Padding(0, 1).
			Width(rightWidth).
			MaxHeight(panelHeight).
			Render(rightPanel)
	} else {
		// Unfocused right panel - dim top/bottom borders
		rightPanel = lipgloss.NewStyle().
			BorderTop(true).
			BorderBottom(true).
			BorderForeground(neonDim).
			Padding(0, 1).
			Width(rightWidth).
			MaxHeight(panelHeight).
			Render(rightPanel)
	}

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

	// Add status banner between spinner and panels (more stable than spinner area)
	statusBanner := renderStatusBanner(bannerType, width)

	// Combine all elements (panels already include headers with focus styling)
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		spinnerArea,
		statusBanner,
		panels,
	)

	return content
}

// renderStatsMatchDetailsPanel renders the right panel for stats view with match details.
// Uses Neon design with Golazo red/cyan theme.
// Returns fixed header and scrollable content separately for viewport scrolling.
// Displays expanded match information including statistics, lineups, and more.
func renderStatsMatchDetailsPanel(width, height int, details *api.MatchDetails, goalLinks GoalLinksMap, focused bool) (string, string) {
	if details == nil {
		emptyMessage := neonDimStyle.
			Align(lipgloss.Center).
			Width(width - 6).
			PaddingTop(height / 4).
			Render("Select a match to view details")

		emptyPanel := neonPanelCyanStyle.
			Width(width).
			Height(height).
			MaxHeight(height).
			Render(emptyMessage)

		return "", emptyPanel // No header, all content is empty message
	}

	contentWidth := width - 6 // Account for border padding
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

	// ═══════════════════════════════════════════════
	// MATCH HEADER (FIXED - always visible)
	// ═══════════════════════════════════════════════
	var headerText string
	if focused {
		// Bright cyan header when focused
		headerText = lipgloss.NewStyle().
			Foreground(neonCyan).
			Bold(true).
			PaddingBottom(0).
			BorderBottom(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(neonCyan).
			MarginBottom(0).
			Render("Match Details")
	} else {
		// Dim header when not focused
		headerText = lipgloss.NewStyle().
			Foreground(neonDim).
			Bold(true).
			PaddingBottom(0).
			BorderBottom(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(neonDim).
			MarginBottom(0).
			Render("Match Details")
	}
	headerLines = append(headerLines, headerText)
	headerLines = append(headerLines, "")

	// Line 1: Team A vs Team B (centered)
	teamsDisplay := fmt.Sprintf("%s  vs  %s",
		neonTeamStyle.Render(homeTeam),
		neonTeamStyle.Render(awayTeam))
	headerLines = append(headerLines, lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Center).Render(teamsDisplay))
	headerLines = append(headerLines, "")

	// Line 2: Large score (like live view)
	if details.HomeScore != nil && details.AwayScore != nil {
		largeScore := renderLargeScore(*details.HomeScore, *details.AwayScore, contentWidth)
		headerLines = append(headerLines, largeScore)
	} else {
		vsText := lipgloss.NewStyle().
			Foreground(neonDim).
			Width(contentWidth).
			Align(lipgloss.Center).
			Render("vs")
		headerLines = append(headerLines, vsText)
	}
	headerLines = append(headerLines, "")

	// Match context row
	if details.League.Name != "" {
		headerLines = append(headerLines, neonLabelStyle.Render("League:      ")+neonValueStyle.Render(details.League.Name))
	}
	if details.Venue != "" {
		headerLines = append(headerLines, neonLabelStyle.Render("Venue:       ")+neonValueStyle.Render(truncateString(details.Venue, contentWidth-14)))
	}
	if details.MatchTime != nil {
		headerLines = append(headerLines, neonLabelStyle.Render("Date:        ")+neonValueStyle.Render(details.MatchTime.Format("02 Jan 2006, 15:04")+" UTC"))
	}
	if details.Referee != "" {
		headerLines = append(headerLines, neonLabelStyle.Render("Referee:     ")+neonValueStyle.Render(details.Referee))
	}
	if details.Attendance > 0 {
		headerLines = append(headerLines, neonLabelStyle.Render("Attendance:  ")+neonValueStyle.Render(formatNumber(details.Attendance)))
	}

	// ═══════════════════════════════════════════════
	// GOALS TIMELINE (chronological with home/away alignment)
	// ═══════════════════════════════════════════════
	var goals []api.MatchEvent
	for _, event := range details.Events {
		if event.Type == "goal" {
			goals = append(goals, event)
		}
	}

	if len(goals) > 0 {
		scrollableLines = append(scrollableLines, "")
		scrollableLines = append(scrollableLines, neonHeaderStyle.Render("Goals"))

		for _, g := range goals {
			isHome := g.Team.ID == details.HomeTeam.ID
			// Build goal content with symbol+type adjacent to center time
			player := "Unknown"
			if g.Player != nil {
				player = *g.Player
			}
			playerDetails := neonValueStyle.Render(player)

			// Check for replay link and create indicator
			replayIndicator := getReplayIndicator(details, goalLinks, g.Minute)

			goalContent := buildEventContent(playerDetails, replayIndicator, "●", neonScoreStyle.Render("GOAL"), isHome)
			minuteStr := g.DisplayMinute
			if minuteStr == "" {
				minuteStr = fmt.Sprintf("%d'", g.Minute) // Fallback
			}
			goalLine := renderCenterAlignedEvent(minuteStr, goalContent, isHome, contentWidth)
			scrollableLines = append(scrollableLines, goalLine)
		}
	}

	// ═══════════════════════════════════════════════
	// CARDS - Detailed list with player, minute (aligned by team)
	// ═══════════════════════════════════════════════
	var cardEvents []api.MatchEvent
	for _, event := range details.Events {
		if event.Type == "card" {
			cardEvents = append(cardEvents, event)
		}
	}

	if len(cardEvents) > 0 {
		scrollableLines = append(scrollableLines, "")
		scrollableLines = append(scrollableLines, neonHeaderStyle.Render("Cards"))

		for _, card := range cardEvents {
			player := "Unknown"
			if card.Player != nil {
				player = *card.Player
			}
			isHome := card.Team.ID == details.HomeTeam.ID

			// Determine card type and apply appropriate color (using shared styles)
			cardSymbol := CardSymbolYellow
			cardStyle := neonYellowCardStyle
			if card.EventType != nil && (*card.EventType == "red" || *card.EventType == "redcard" || *card.EventType == "secondyellow") {
				cardSymbol = CardSymbolRed
				cardStyle = neonRedCardStyle
			}

			// Build card content with symbol+type adjacent to center time
			playerDetails := neonValueStyle.Render(player)
			cardContent := buildEventContent(playerDetails, "", cardSymbol, cardStyle.Render("CARD"), isHome)
			minuteStr := card.DisplayMinute
			if minuteStr == "" {
				minuteStr = fmt.Sprintf("%d'", card.Minute) // Fallback
			}
			cardLine := renderCenterAlignedEvent(minuteStr, cardContent, isHome, contentWidth)
			scrollableLines = append(scrollableLines, cardLine)
		}
	}

	// ═══════════════════════════════════════════════
	// MATCH STATISTICS (Visual Progress Bars)
	// ═══════════════════════════════════════════════
	if len(details.Statistics) > 0 {
		scrollableLines = append(scrollableLines, "")
		scrollableLines = append(scrollableLines, neonHeaderStyle.Render("Statistics"))

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
					scrollableLines = append(scrollableLines, "")

					if wanted.isProgress {
						// Render as visual progress bar (centered)
						statLine := renderStatProgressBar(wanted.label, stat.HomeValue, stat.AwayValue, contentWidth, homeTeam, awayTeam)
						scrollableLines = append(scrollableLines, centerStyle.Render(statLine))
					} else {
						// Render as comparison bar (centered)
						statLine := renderStatComparison(wanted.label, stat.HomeValue, stat.AwayValue, contentWidth)
						scrollableLines = append(scrollableLines, centerStyle.Render(statLine))
					}
					break
				}
			}
		}
	}

	// Combine header and scrollable content
	headerContent := lipgloss.JoinVertical(lipgloss.Left, headerLines...)
	scrollableContent := lipgloss.JoinVertical(lipgloss.Left, scrollableLines...)

	return headerContent, scrollableContent
}

// RenderMatchDetailsPanel is an exported version of renderStatsMatchDetailsPanel
// for use by debug scripts. Renders match details in the Golazo stats view style.
func RenderMatchDetailsPanel(width, height int, details *api.MatchDetails) string {
	header, scrollable := renderStatsMatchDetailsPanel(width, height, details, nil, false)
	content := lipgloss.JoinVertical(lipgloss.Left, header, scrollable)
	return neonPanelCyanStyle.
		Width(width).
		Height(height).
		MaxHeight(height).
		Render(content)
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
	homeFilled := min((homeNum*halfBar)/maxVal, halfBar)
	homeEmpty := halfBar - homeFilled
	homeBar := strings.Repeat(" ", homeEmpty) + strings.Repeat("▪", homeFilled)
	homeBarStyled := lipgloss.NewStyle().Foreground(neonCyan).Render(homeBar)

	// Away bar (left-aligned, grows right)
	awayFilled := min((awayNum*halfBar)/maxVal, halfBar)
	awayEmpty := halfBar - awayFilled
	awayBar := strings.Repeat("▪", awayFilled) + strings.Repeat(" ", awayEmpty)
	awayBarStyled := lipgloss.NewStyle().Foreground(neonGray).Render(awayBar)

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
