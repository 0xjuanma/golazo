package app

import (
	"github.com/0xjuanma/golazo/internal/api"
	"github.com/0xjuanma/golazo/internal/fotmob"
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
		if m.selected < 2 && !m.mainViewLoading { // 3 menu items: 0, 1, 2
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

		// Handle Settings view separately (no API calls needed)
		if m.selected == 2 {
			m.settingsState = ui.NewSettingsState()
			m.currentView = viewSettings
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
		m.lastHomeScore = 0
		m.lastAwayScore = 0
		m.polling = false
		m.upcomingMatchesList.SetItems([]list.Item{})
		m.matchDetailsCache = make(map[int]*api.MatchDetails)

		// Start API calls immediately while showing main view spinner
		cmds := []tea.Cmd{
			m.spinner.Tick,
			performMainViewCheck(m.selected),
		}

		switch m.selected {
		case 0: // Stats view - fetch data progressively (day by day)
			m.statsViewLoading = true
			m.loading = true
			m.statsData = nil                          // Clear cached data to force fresh fetch
			m.statsDaysLoaded = 0                      // Reset progress
			m.statsTotalDays = fotmob.StatsDataDays    // Set total days to load
			m.statsMatchesList.SetItems([]list.Item{}) // Clear list
			cmds = append(cmds, ui.SpinnerTick())
			// Start fetching day 0 (today) first - results shown immediately when it completes
			cmds = append(cmds, fetchStatsDayData(m.fotmobClient, m.useMockData, 0, fotmob.StatsDataDays))
		case 1: // Live Matches view - preload live matches progressively (parallel batches)
			m.liveViewLoading = true
			m.loading = true
			m.liveBatchesLoaded = 0
			totalLeagues := fotmob.TotalLeagues()
			m.liveTotalBatches = (totalLeagues + LiveBatchSize - 1) / LiveBatchSize // Ceiling division
			m.liveMatchesBuffer = nil                                               // Clear buffer
			m.liveMatchesList.SetItems([]list.Item{})
			cmds = append(cmds, ui.SpinnerTick())
			// Start fetching batch 0 (4 leagues in parallel) - results shown when batch completes
			cmds = append(cmds, fetchLiveBatchData(m.fotmobClient, m.useMockData, 0))
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
// Uses client-side filtering from cached data - no new API calls needed!
func (m model) handleStatsViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "l", "right":
		// Cycle date range forward: 1 -> 3 -> 5 -> 1
		switch m.statsDateRange {
		case 1:
			m.statsDateRange = 3
		case 3:
			m.statsDateRange = 5
		default:
			m.statsDateRange = 1
		}
	case "h", "left":
		// Cycle date range backward: 1 -> 5 -> 3 -> 1
		switch m.statsDateRange {
		case 1:
			m.statsDateRange = 5
		case 5:
			m.statsDateRange = 3
		default:
			m.statsDateRange = 1
		}
	default:
		return m, nil
	}

	// If we have cached stats data, just filter client-side (instant!)
	if m.statsData != nil {
		m.matchDetails = nil
		m.matchDetailsCache = make(map[int]*api.MatchDetails)
		m.applyStatsDateFilter()
		m.selected = 0

		// Load details for first match if available
		if len(m.matches) > 0 {
			m.statsMatchesList.Select(0)
			return m.loadStatsMatchDetails(m.matches[0].ID)
		}
		return m, nil
	}

	// No cached data - need to fetch (shouldn't happen normally)
	m.statsViewLoading = true
	m.loading = true
	m.statsDaysLoaded = 0
	m.statsTotalDays = fotmob.StatsDataDays
	return m, tea.Batch(m.spinner.Tick, ui.SpinnerTick(), fetchStatsDayData(m.fotmobClient, m.useMockData, 0, fotmob.StatsDataDays))
}

// loadMatchDetails loads match details for the live matches view.
// Resets live updates and event history before fetching new details.
func (m model) loadMatchDetails(matchID int) (tea.Model, tea.Cmd) {
	m.liveUpdates = nil
	m.lastEvents = nil
	m.lastHomeScore = 0
	m.lastAwayScore = 0
	m.loading = true
	m.liveViewLoading = true
	m.polling = false // Reset polling state - this is a new match load, not a poll refresh
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

// handleSettingsViewKeys processes keyboard input for the settings view.
// Follows the same pattern as handleStatsSelection for consistent behavior.
func (m model) handleSettingsViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.settingsState == nil {
		return m, nil
	}

	// Check if list is filtering - if so, let list handle ALL keys
	isFiltering := m.settingsState.List.FilterState() == list.Filtering

	// Only handle custom keys when NOT filtering
	if !isFiltering {
		switch msg.String() {
		case " ": // Space to toggle selection
			m.settingsState.Toggle()
			return m, nil
		case "enter":
			// Save settings and return to main menu
			_ = m.settingsState.Save() // Best-effort save
			m.settingsState = nil
			m.currentView = viewMain
			m.selected = 0
			return m, nil
		}
	}

	// Delegate to list component for navigation, filtering, etc.
	var listCmd tea.Cmd
	m.settingsState.List, listCmd = m.settingsState.List.Update(msg)
	return m, listCmd
}
