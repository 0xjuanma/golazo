package ui

import (
	"fmt"
	"strings"

	"github.com/0xjuanma/golazo/internal/api"
	"github.com/0xjuanma/golazo/internal/constants"
	"github.com/charmbracelet/bubbles/list"
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

	panel := neonPanelStyle.
		Width(width).
		Height(height).
		Render(content)

	return panel
}

// renderDateRangeSelector renders a horizontal date range selector (Today, 3d).
func renderDateRangeSelector(width int, selected int) string {
	options := []struct {
		days  int
		label string
	}{
		{1, "Today"},
		{3, "3d"},
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
func RenderMultiPanelViewWithList(width, height int, listModel list.Model, details *api.MatchDetails, liveUpdates []string, sp spinner.Model, loading bool, randomSpinner *RandomCharSpinner, viewLoading bool) string {
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
	var spinnerArea string
	if viewLoading && randomSpinner != nil {
		spinnerView := randomSpinner.View()
		if spinnerView != "" {
			// Center the spinner horizontally using style with width and alignment
			spinnerStyle := lipgloss.NewStyle().
				Width(width).
				Height(spinnerHeight).
				Align(lipgloss.Center).
				AlignVertical(lipgloss.Center)
			spinnerArea = spinnerStyle.Render(spinnerView)
		} else {
			// Fallback if spinner view is empty
			spinnerStyle := lipgloss.NewStyle().
				Width(width).
				Height(spinnerHeight).
				Align(lipgloss.Center).
				AlignVertical(lipgloss.Center)
			spinnerArea = spinnerStyle.Render("Loading...")
		}
	} else {
		// Reserve space with empty lines - ensure it takes up exactly spinnerHeight lines
		spinnerArea = strings.Repeat("\n", spinnerHeight)
	}

	// Calculate panel dimensions
	leftWidth := width * 35 / 100
	if leftWidth < 25 {
		leftWidth = 25
	}
	rightWidth := width - leftWidth - 1
	if rightWidth < 35 {
		rightWidth = 35
		leftWidth = width - rightWidth - 1
	}

	// Use panelHeight similar to stats view to ensure proper spacing
	panelHeight := availableHeight - 2

	// Render left panel (matches list) - shifted down
	leftPanel := RenderLiveMatchesListPanel(leftWidth, panelHeight, listModel)

	// Render right panel (match details with live updates) - shifted down
	rightPanel := renderMatchDetailsPanel(rightWidth, panelHeight, details, liveUpdates, sp, loading)

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
func RenderStatsViewWithList(width, height int, finishedList list.Model, upcomingList list.Model, details *api.MatchDetails, randomSpinner *RandomCharSpinner, viewLoading bool, dateRange int) string {
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
	var spinnerArea string
	if viewLoading && randomSpinner != nil {
		spinnerView := randomSpinner.View()
		if spinnerView != "" {
			// Center the spinner horizontally using style with width and alignment
			spinnerStyle := lipgloss.NewStyle().
				Width(width).
				Height(spinnerHeight).
				Align(lipgloss.Center).
				AlignVertical(lipgloss.Center)
			spinnerArea = spinnerStyle.Render(spinnerView)
		} else {
			// Fallback if spinner view is empty
			spinnerStyle := lipgloss.NewStyle().
				Width(width).
				Height(spinnerHeight).
				Align(lipgloss.Center).
				AlignVertical(lipgloss.Center)
			spinnerArea = spinnerStyle.Render("Loading...")
		}
	} else {
		// Reserve space with empty lines - ensure it takes up exactly spinnerHeight lines
		spinnerArea = strings.Repeat("\n", spinnerHeight)
	}

	// Calculate panel dimensions - match live view exactly (35% left, 65% right)
	leftWidth := width * 35 / 100
	if leftWidth < 25 {
		leftWidth = 25
	}
	rightWidth := width - leftWidth - 1
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

	// Header: Match Info
	lines = append(lines, neonHeaderStyle.Render("Match Info"))
	lines = append(lines, "")

	// Score line - centered
	var scoreDisplay string
	if details.HomeScore != nil && details.AwayScore != nil {
		scoreDisplay = fmt.Sprintf("%s  %d - %d  %s",
			neonTeamStyle.Render(homeTeam),
			*details.HomeScore,
			*details.AwayScore,
			neonTeamStyle.Render(awayTeam))
	} else {
		scoreDisplay = fmt.Sprintf("%s  vs  %s",
			neonTeamStyle.Render(homeTeam),
			neonTeamStyle.Render(awayTeam))
	}
	lines = append(lines, lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Center).Render(scoreDisplay))
	lines = append(lines, "")

	// Status
	var statusStr string
	switch details.Status {
	case api.MatchStatusFinished:
		statusStr = neonFinishedStyle.Render("FT")
	case api.MatchStatusLive:
		if details.LiveTime != nil {
			statusStr = neonLiveStyle.Render(*details.LiveTime)
		} else {
			statusStr = neonLiveStyle.Render("LIVE")
		}
	default:
		statusStr = neonDimStyle.Render(string(details.Status))
	}
	lines = append(lines, neonLabelStyle.Render("Status:      ")+statusStr)

	// League
	if details.League.Name != "" {
		lines = append(lines, neonLabelStyle.Render("League:      ")+neonValueStyle.Render(details.League.Name))
	}

	// Venue
	if details.Venue != "" {
		lines = append(lines, neonLabelStyle.Render("Venue:       ")+neonValueStyle.Render(details.Venue))
	}

	// Date
	if details.MatchTime != nil {
		lines = append(lines, neonLabelStyle.Render("Date:        ")+neonValueStyle.Render(details.MatchTime.Format("02 Jan 2006")))
	}

	// Half-time score
	if details.HalfTimeScore != nil && details.HalfTimeScore.Home != nil && details.HalfTimeScore.Away != nil {
		htStr := fmt.Sprintf("%d - %d", *details.HalfTimeScore.Home, *details.HalfTimeScore.Away)
		lines = append(lines, neonLabelStyle.Render("Half-Time:   ")+neonValueStyle.Render(htStr))
	}

	// Goals section
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
				player := "Unknown"
				if g.Player != nil {
					player = *g.Player
				}
				lines = append(lines, fmt.Sprintf("  %s %s", neonScoreStyle.Render(fmt.Sprintf("%d'", g.Minute)), neonValueStyle.Render(player)))
			}
		}

		if len(awayGoals) > 0 {
			lines = append(lines, neonTeamStyle.Render(awayTeam))
			for _, g := range awayGoals {
				player := "Unknown"
				if g.Player != nil {
					player = *g.Player
				}
				lines = append(lines, fmt.Sprintf("  %s %s", neonScoreStyle.Render(fmt.Sprintf("%d'", g.Minute)), neonValueStyle.Render(player)))
			}
		}
	}

	// Cards section
	var yellowCards, redCards int
	for _, event := range details.Events {
		if event.Type == "yellowCard" {
			yellowCards++
		} else if event.Type == "redCard" {
			redCards++
		}
	}

	if yellowCards > 0 || redCards > 0 {
		lines = append(lines, "")
		lines = append(lines, neonHeaderStyle.Render("Cards"))
		if yellowCards > 0 {
			lines = append(lines, fmt.Sprintf("  %s %s", neonTeamStyle.Render("Yellow:"), neonValueStyle.Render(fmt.Sprintf("%d", yellowCards))))
		}
		if redCards > 0 {
			lines = append(lines, fmt.Sprintf("  %s %s", neonLiveStyle.Render("Red:"), neonValueStyle.Render(fmt.Sprintf("%d", redCards))))
		}
	}

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)

	return neonPanelCyanStyle.
		Width(width).
		Height(height).
		Render(content)
}
