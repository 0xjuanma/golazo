package ui

import (
	"fmt"
	"strings"

	"github.com/0xjuanma/golazo/internal/api"
	"github.com/0xjuanma/golazo/internal/constants"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

// GoalLinksMap maps goal keys (matchID:minute) to replay URLs.
type GoalLinksMap map[string]string

const (
	minPanelHeight    = 10
	minScrollableArea = 3
	minListHeight     = 3
)

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
func RenderLiveMatchesListPanel(width, height int, listModel list.Model, upcomingMatches []MatchDisplay) string {
	contentWidth := width - 6

	title := neonPanelTitleStyle.Width(contentWidth).Render(constants.PanelLiveMatches)

	var listView string
	if len(listModel.Items()) == 0 {
		listView = neonEmptyStyle.Width(contentWidth).Render(constants.EmptyNoLiveMatches)
	} else {
		listView = listModel.View()
	}

	borderHeight := 2
	titleHeight := 2
	innerHeight := height - borderHeight - titleHeight

	var upcomingSection string
	upcomingHeight := 0
	if len(upcomingMatches) > 0 {
		maxUpcomingHeight := innerHeight / 2

		upcomingTitle := neonHeaderStyle.Render("Upcoming")

		var upcomingLines []string
		upcomingLines = append(upcomingLines, upcomingTitle)
		for _, match := range upcomingMatches {
			matchLine := renderUpcomingMatchLine(match, contentWidth)
			upcomingLines = append(upcomingLines, matchLine)
		}
		upcomingSection = strings.Join(upcomingLines, "\n")

		upcomingHeight = len(upcomingLines) + 1
		if upcomingHeight > maxUpcomingHeight {
			upcomingSection = truncateToHeight(upcomingSection, maxUpcomingHeight)
			upcomingHeight = maxUpcomingHeight
		}
	}

	availableListHeight := max(innerHeight-upcomingHeight-1, minListHeight)
	listView = truncateToHeight(listView, availableListHeight)

	var content string
	if upcomingHeight > 0 {
		content = lipgloss.JoinVertical(lipgloss.Left, title, "", listView, "", upcomingSection)
	} else {
		content = lipgloss.JoinVertical(lipgloss.Left, title, "", listView)
	}

	totalInnerHeight := height - 2
	if totalInnerHeight > 0 {
		content = truncateToHeight(content, totalInnerHeight)
	}

	return neonPanelStyle.Width(width).Height(height).Render(content)
}

func renderUpcomingMatchLine(match MatchDisplay, maxWidth int) string {
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

	maxTeamLen := (maxWidth - 15) / 2
	if len(homeTeam) > maxTeamLen {
		homeTeam = homeTeam[:maxTeamLen-1] + "…"
	}
	if len(awayTeam) > maxTeamLen {
		awayTeam = awayTeam[:maxTeamLen-1] + "…"
	}

	return fmt.Sprintf("  %s  %s vs %s",
		neonDimStyle.Render(timeStr),
		neonValueStyle.Render(homeTeam),
		neonValueStyle.Render(awayTeam))
}

// RenderStatsListPanel renders the left panel for stats view.
func RenderStatsListPanel(width, height int, finishedList list.Model, dateRange int, rightPanelFocused bool) string {
	var header string
	if rightPanelFocused {
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
		header = neonHeaderStyle.Render("Match List")
	}

	dateSelector := renderDateRangeSelector(width-6, dateRange)
	emptyStyle := neonEmptyStyle.Width(width - 6)

	var finishedListView string
	if len(finishedList.Items()) == 0 {
		finishedListView = emptyStyle.Render(constants.EmptyNoFinishedMatches + "\n\nTry selecting a different date range (h/l keys)")
	} else {
		finishedListView = finishedList.View()
	}

	content := lipgloss.JoinVertical(lipgloss.Left, header, "", dateSelector, "", finishedListView)

	innerHeight := height - 2
	if innerHeight > 0 {
		content = truncateToHeight(content, innerHeight)
	}

	var panel string
	if rightPanelFocused {
		panel = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(neonDim).
			Padding(0, 1).
			Width(width).
			Height(height).
			Render(content)
	} else {
		panel = neonPanelStyle.Width(width).Height(height).Render(content)
	}

	return panel
}

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
			items = append(items, neonDateSelectedStyle.Render(opt.label))
		} else {
			items = append(items, neonDateUnselectedStyle.Render(opt.label))
		}
	}

	selector := strings.Join(items, "  ")
	return lipgloss.NewStyle().Width(width).Align(lipgloss.Center).Padding(0, 1).Render(selector)
}

// RenderMultiPanelViewWithList renders the live matches view with list component.
func RenderMultiPanelViewWithList(width, height int, listModel list.Model, details *api.MatchDetails, liveUpdates []string, sp spinner.Model, loading bool, randomSpinner *RandomCharSpinner, viewLoading bool, leaguesLoaded int, totalLeagues int, pollingSpinner *RandomCharSpinner, isPolling bool, upcomingMatches []MatchDisplay, goalLinks GoalLinksMap, bannerType constants.StatusBannerType) string {
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 24
	}

	spinnerHeight := 3
	availableHeight := max(height-spinnerHeight, minPanelHeight)

	spinnerStyle := lipgloss.NewStyle().
		Width(width).
		Height(spinnerHeight).
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center)

	var spinnerArea string
	if viewLoading && randomSpinner != nil {
		spinnerView := randomSpinner.View()
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
		spinnerArea = spinnerStyle.Render("")
	}

	leftWidth := max(width*35/100, 25)
	rightWidth := width - leftWidth - 1
	if rightWidth < 35 {
		rightWidth = 35
		leftWidth = width - rightWidth - 1
	}

	panelHeight := availableHeight - 2

	leftPanel := RenderLiveMatchesListPanel(leftWidth, panelHeight, listModel, upcomingMatches)
	rightPanel := renderMatchDetailsPanelWithPolling(rightWidth, panelHeight, details, liveUpdates, sp, loading, pollingSpinner, isPolling, goalLinks)

	separatorStyle := neonSeparatorStyle.Height(panelHeight)
	separator := separatorStyle.Render("┃")

	panels := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, separator, rightPanel)
	statusBanner := renderStatusBanner(bannerType, width)

	return lipgloss.JoinVertical(lipgloss.Left, spinnerArea, statusBanner, panels)
}

// RenderStatsViewWithList renders the stats view with list component.
func RenderStatsViewWithList(width, height int, finishedList list.Model, details *api.MatchDetails, randomSpinner *RandomCharSpinner, viewLoading bool, dateRange int, daysLoaded int, totalDays int, goalLinks GoalLinksMap, bannerType constants.StatusBannerType, detailsViewport *viewport.Model, rightPanelFocused bool, scrollOffset int) string {
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 24
	}

	spinnerHeight := 3
	availableHeight := max(height-spinnerHeight, minPanelHeight)

	spinnerStyle := lipgloss.NewStyle().
		Width(width).
		Height(spinnerHeight).
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center)

	var spinnerArea string
	if viewLoading && randomSpinner != nil {
		spinnerView := randomSpinner.View()
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
		spinnerArea = spinnerStyle.Render("")
	}

	leftWidth := max(width*35/100, 25)
	rightWidth := width - leftWidth - 1
	if rightWidth < 35 {
		rightWidth = 35
		leftWidth = width - rightWidth - 1
	}

	panelHeight := availableHeight - 2

	leftPanel := RenderStatsListPanel(leftWidth, panelHeight, finishedList, dateRange, rightPanelFocused)
	headerContent, scrollableContent := renderStatsMatchDetailsPanel(rightWidth, panelHeight, details, goalLinks, rightPanelFocused)

	var rightPanel string
	scrollableLines := strings.Split(scrollableContent, "\n")
	headerHeight := strings.Count(headerContent, "\n") + 1
	availableHeight = max(panelHeight-headerHeight, minScrollableArea)

	visibleLines := scrollableLines
	if rightPanelFocused && len(scrollableLines) > availableHeight {
		start := scrollOffset
		end := min(start+availableHeight, len(scrollableLines))
		if start < len(scrollableLines) && start >= 0 {
			visibleLines = scrollableLines[start:end]
		} else if start >= len(scrollableLines) {
			visibleLines = []string{}
		}
		for len(visibleLines) < availableHeight && len(visibleLines) < len(scrollableLines) {
			visibleLines = append(visibleLines, "")
		}
	} else {
		if len(scrollableLines) > availableHeight {
			visibleLines = scrollableLines[:availableHeight]
		}
	}

	visibleContent := strings.Join(visibleLines, "\n")
	rightPanel = lipgloss.JoinVertical(lipgloss.Left, headerContent, visibleContent)

	if rightPanelFocused {
		rightPanel = lipgloss.NewStyle().
			BorderTop(true).
			BorderBottom(true).
			BorderForeground(neonCyan).
			Padding(0, 1).
			Width(rightWidth).
			MaxHeight(panelHeight).
			Render(rightPanel)
	} else {
		rightPanel = lipgloss.NewStyle().
			BorderTop(true).
			BorderBottom(true).
			BorderForeground(neonDim).
			Padding(0, 1).
			Width(rightWidth).
			MaxHeight(panelHeight).
			Render(rightPanel)
	}

	separatorStyle := neonSeparatorStyle.Height(panelHeight)
	separator := separatorStyle.Render("┃")

	panels := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, separator, rightPanel)
	statusBanner := renderStatusBanner(bannerType, width)

	return lipgloss.JoinVertical(lipgloss.Left, spinnerArea, statusBanner, panels)
}

// renderStatsMatchDetailsPanel renders match details using unified rendering.
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

		return "", emptyPanel
	}

	cfg := MatchDetailsConfig{
		Width:          width,
		Height:         height,
		Details:        details,
		GoalLinks:      goalLinks,
		ShowStatistics: true,
		ShowHighlights: true,
		Focused:        focused,
	}

	return RenderMatchDetails(cfg)
}

// RenderMatchDetailsPanel is an exported version for debug scripts.
func RenderMatchDetailsPanel(width, height int, details *api.MatchDetails) string {
	header, scrollable := renderStatsMatchDetailsPanel(width, height, details, nil, false)
	content := lipgloss.JoinVertical(lipgloss.Left, header, scrollable)
	return neonPanelCyanStyle.
		Width(width).
		Height(height).
		MaxHeight(height).
		Render(content)
}

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
