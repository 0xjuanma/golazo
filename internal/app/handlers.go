package app

import (
	"github.com/0xjuanma/golazo/internal/api"
	"github.com/0xjuanma/golazo/internal/ui"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// handleMainViewKeys processes keyboard input for the main menu view.
// Handles navigation (up/down) and selection (enter) to switch between views.
// On selection, immediately starts API preloading while showing spinner for 2 seconds.
func (m model) handleMainViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		if m.selected < 1 && !m.mainViewLoading {
			m.selected++
		}
	case "k", "up":
		if m.selected > 0 && !m.mainViewLoading {
			m.selected--
		}
	case "enter":
		if m.mainViewLoading {
			return m, nil
		}
		m.mainViewLoading = true
		m.pendingSelection = m.selected

		// Clear previous view state
		m.matches = nil
		m.upcomingMatches = nil
		m.matchDetails = nil
		m.liveUpdates = nil
		m.lastEvents = nil
		m.polling = false
		m.upcomingMatchesList.SetItems([]list.Item{})
		m.matchDetailsCache = make(map[int]*api.MatchDetails)

		// Start API calls immediately while showing main view spinner
		cmds := []tea.Cmd{
			m.spinner.Tick,
			performMainViewCheck(m.selected),
		}

		switch m.selected {
		case 0: // Stats view - preload finished matches
			m.statsViewLoading = true
			m.loading = true
			cmds = append(cmds, ui.SpinnerTick())
			cmds = append(cmds, fetchFinishedMatchesFotmob(m.fotmobClient, m.useMockData, m.statsDateRange))
			if m.statsDateRange == 1 {
				cmds = append(cmds, fetchUpcomingMatchesFotmob(m.fotmobClient, m.useMockData))
			}
		case 1: // Live Matches view - preload live matches
			m.liveViewLoading = true
			m.loading = true
			cmds = append(cmds, ui.SpinnerTick())
			cmds = append(cmds, fetchLiveMatches(m.fotmobClient, m.useMockData))
		}

		return m, tea.Batch(cmds...)
	}
	return m, nil
}

// handleLiveMatchesKeys processes keyboard input for the live matches view.
// Handles navigation between matches and loading match details on selection.
// Note: Currently unused as list component handles navigation directly.
func (m model) handleLiveMatchesKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		if m.selected < len(m.matches)-1 {
			m.selected++
			if m.selected < len(m.matches) {
				return m.loadMatchDetails(m.matches[m.selected].ID)
			}
		}
	case "k", "up":
		if m.selected > 0 {
			m.selected--
			if m.selected >= 0 && m.selected < len(m.matches) {
				return m.loadMatchDetails(m.matches[m.selected].ID)
			}
		}
	}
	return m, nil
}

// handleStatsViewKeys processes keyboard input for the stats view.
// Handles date range navigation (left/right) to change the time period.
func (m model) handleStatsViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "h", "left", "l", "right":
		// Cycle date range: 1 -> 3 -> 1
		if m.statsDateRange == 1 {
			m.statsDateRange = 3
		} else {
			m.statsDateRange = 1
		}

		// Reset state for new date range
		m.statsViewLoading = true
		m.loading = true
		m.upcomingMatches = nil
		m.upcomingMatchesList.SetItems([]list.Item{})
		m.matchDetailsCache = make(map[int]*api.MatchDetails)
		m.matchDetails = nil

		cmds := []tea.Cmd{
			m.spinner.Tick,
			ui.SpinnerTick(),
			fetchFinishedMatchesFotmob(m.fotmobClient, m.useMockData, m.statsDateRange),
		}

		// Fetch upcoming matches only for 1-day view
		if m.statsDateRange == 1 {
			cmds = append(cmds, fetchUpcomingMatchesFotmob(m.fotmobClient, m.useMockData))
		}

		return m, tea.Batch(cmds...)
	}
	return m, nil
}

// loadMatchDetails loads match details for the live matches view.
// Resets live updates and event history before fetching new details.
func (m model) loadMatchDetails(matchID int) (tea.Model, tea.Cmd) {
	m.liveUpdates = nil
	m.lastEvents = nil
	m.loading = true
	m.liveViewLoading = true
	return m, tea.Batch(m.spinner.Tick, ui.SpinnerTick(), fetchMatchDetails(m.fotmobClient, matchID, m.useMockData))
}

// loadStatsMatchDetails loads match details for the stats view.
// Checks cache first to avoid redundant API calls.
func (m model) loadStatsMatchDetails(matchID int) (tea.Model, tea.Cmd) {
	// Return cached details if available
	if cached, ok := m.matchDetailsCache[matchID]; ok {
		m.matchDetails = cached
		return m, nil
	}

	// Fetch from API
	m.loading = true
	m.statsViewLoading = true
	return m, tea.Batch(m.spinner.Tick, ui.SpinnerTick(), fetchStatsMatchDetailsFotmob(m.fotmobClient, matchID, m.useMockData))
}
