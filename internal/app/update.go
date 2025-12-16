package app

import (
	"github.com/0xjuanma/golazo/internal/api"
	"github.com/0xjuanma/golazo/internal/ui"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// Update handles all incoming messages and updates the model accordingly.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleWindowSize(msg)

	case spinner.TickMsg:
		return m.handleSpinnerTick(msg)

	case liveUpdateMsg:
		return m.handleLiveUpdate(msg)

	case matchDetailsMsg:
		return m.handleMatchDetails(msg)

	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case liveMatchesMsg:
		return m.handleLiveMatches(msg)

	case finishedMatchesMsg:
		return m.handleFinishedMatches(msg)

	case upcomingMatchesMsg:
		return m.handleUpcomingMatches(msg)

	case ui.TickMsg:
		return m.handleRandomSpinnerTick(msg)

	case mainViewCheckMsg:
		return m.handleMainViewCheck(msg)

	default:
		// Fallback handler for ui.TickMsg type assertion
		if _, ok := msg.(ui.TickMsg); ok {
			return m.handleRandomSpinnerTick(msg.(ui.TickMsg))
		}
	}

	return m, tea.Batch(cmds...)
}

// handleWindowSize updates list sizes when window dimensions change.
func (m model) handleWindowSize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height

	const (
		frameH        = 2
		frameV        = 2
		titleHeight   = 3
		spinnerHeight = 3
	)

	switch m.currentView {
	case viewLiveMatches:
		leftWidth := max(m.width*35/100, 25)
		availableWidth := leftWidth - frameH*2
		availableHeight := m.height - frameV*2 - titleHeight - spinnerHeight
		if availableWidth > 0 && availableHeight > 0 {
			m.liveMatchesList.SetSize(availableWidth, availableHeight)
		}

	case viewStats:
		leftWidth := max(m.width*40/100, 30)
		availableWidth := leftWidth - frameH*2
		availableHeight := m.height - frameV*2 - titleHeight - spinnerHeight
		if availableWidth > 0 && availableHeight > 0 {
			if m.statsDateRange == 1 {
				finishedHeight := availableHeight * 60 / 100
				upcomingHeight := availableHeight - finishedHeight
				m.statsMatchesList.SetSize(availableWidth, finishedHeight)
				m.upcomingMatchesList.SetSize(availableWidth, upcomingHeight)
			} else {
				m.statsMatchesList.SetSize(availableWidth, availableHeight)
				m.upcomingMatchesList.SetSize(availableWidth, 0)
			}
		}
	}

	return m, nil
}

// handleSpinnerTick updates the standard spinner animation.
func (m model) handleSpinnerTick(msg spinner.TickMsg) (tea.Model, tea.Cmd) {
	if m.loading || m.mainViewLoading {
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

// handleLiveUpdate processes live match update messages.
func (m model) handleLiveUpdate(msg liveUpdateMsg) (tea.Model, tea.Cmd) {
	if msg.update != "" {
		m.liveUpdates = append(m.liveUpdates, msg.update)
	}

	// Continue polling if match is live
	if m.polling && m.matchDetails != nil && m.matchDetails.Status == api.MatchStatusLive {
		return m, pollMatchDetails(m.fotmobClient, m.parser, m.matchDetails.ID, m.lastEvents, m.useMockData)
	}

	m.loading = false
	m.polling = false
	return m, nil
}

// handleMatchDetails processes match details response messages.
func (m model) handleMatchDetails(msg matchDetailsMsg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	if msg.details == nil {
		// Clear match details when API call fails so we don't show stale data
		m.matchDetails = nil
		m.loading = false
		m.liveViewLoading = false
		m.statsViewLoading = false
		return m, nil
	}

	m.matchDetails = msg.details

	// Cache for stats view
	if m.currentView == viewStats {
		m.matchDetailsCache[msg.details.ID] = msg.details
		m.loading = false
		m.statsViewLoading = false
		return m, nil
	}

	// Handle live matches view
	if m.currentView == viewLiveMatches {
		m.liveViewLoading = false

		// Parse events for live updates
		var eventsToParse []api.MatchEvent
		if len(m.lastEvents) == 0 {
			eventsToParse = msg.details.Events
		} else {
			eventsToParse = m.parser.NewEvents(m.lastEvents, msg.details.Events)
		}

		if len(eventsToParse) > 0 {
			updates := m.parser.ParseEvents(eventsToParse, msg.details.HomeTeam, msg.details.AwayTeam)
			m.liveUpdates = append(m.liveUpdates, updates...)
		}
		m.lastEvents = msg.details.Events

		// Continue polling if match is live
		if msg.details.Status == api.MatchStatusLive {
			m.polling = true
			m.loading = true
			cmds = append(cmds, pollMatchDetails(m.fotmobClient, m.parser, msg.details.ID, m.lastEvents, m.useMockData))
		} else {
			m.loading = false
			m.polling = false
		}
		return m, tea.Batch(cmds...)
	}

	// Default: turn off all loading states
	m.loading = false
	m.liveViewLoading = false
	m.statsViewLoading = false
	return m, nil
}

// handleKeyPress routes key events to view-specific handlers.
func (m model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "esc":
		if m.currentView != viewMain {
			return m.resetToMainView()
		}
	}

	// View-specific key handling
	switch m.currentView {
	case viewMain:
		return m.handleMainViewKeys(msg)
	case viewLiveMatches:
		return m.handleLiveMatchesSelection(msg)
	case viewStats:
		return m.handleStatsSelection(msg)
	}

	return m, nil
}

// resetToMainView clears state and returns to main menu.
func (m model) resetToMainView() (tea.Model, tea.Cmd) {
	m.currentView = viewMain
	m.selected = 0
	m.matchDetails = nil
	m.matchDetailsCache = make(map[int]*api.MatchDetails)
	m.liveUpdates = nil
	m.lastEvents = nil
	m.loading = false
	m.polling = false
	m.matches = nil
	m.upcomingMatches = nil
	return m, nil
}

// handleLiveMatchesSelection handles list navigation in live matches view.
func (m model) handleLiveMatchesSelection(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var listCmd tea.Cmd
	m.liveMatchesList, listCmd = m.liveMatchesList.Update(msg)

	if selectedItem := m.liveMatchesList.SelectedItem(); selectedItem != nil {
		if item, ok := selectedItem.(ui.MatchListItem); ok {
			for i, match := range m.matches {
				if match.ID == item.Match.ID && i != m.selected {
					m.selected = i
					return m.loadMatchDetails(m.matches[m.selected].ID)
				}
			}
		}
	}

	return m, listCmd
}

// handleStatsSelection handles list navigation and date range changes in stats view.
func (m model) handleStatsSelection(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle date range navigation
	if msg.String() == "h" || msg.String() == "left" || msg.String() == "l" || msg.String() == "right" {
		return m.handleStatsViewKeys(msg)
	}

	// Handle list navigation
	var listCmd tea.Cmd
	m.statsMatchesList, listCmd = m.statsMatchesList.Update(msg)

	if selectedItem := m.statsMatchesList.SelectedItem(); selectedItem != nil {
		if item, ok := selectedItem.(ui.MatchListItem); ok {
			for i, match := range m.matches {
				if match.ID == item.Match.ID && i != m.selected {
					m.selected = i
					return m.loadStatsMatchDetails(m.matches[m.selected].ID)
				}
			}
		}
	}

	return m, listCmd
}

// handleLiveMatches processes live matches API response.
func (m model) handleLiveMatches(msg liveMatchesMsg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	if len(msg.matches) == 0 {
		m.liveViewLoading = false
		m.loading = false
		return m, nil
	}

	// Convert to display format
	displayMatches := make([]ui.MatchDisplay, 0, len(msg.matches))
	for _, match := range msg.matches {
		displayMatches = append(displayMatches, ui.MatchDisplay{Match: match})
	}

	m.matches = displayMatches
	m.selected = 0
	m.loading = false
	cmds = append(cmds, m.randomSpinner.Init())

	// Update list
	m.liveMatchesList.SetItems(ui.ToMatchListItems(displayMatches))
	m.updateLiveListSize()

	if len(displayMatches) > 0 {
		m.liveMatchesList.Select(0)
		updatedModel, loadCmd := m.loadMatchDetails(m.matches[0].ID)
		if updatedM, ok := updatedModel.(model); ok {
			m = updatedM
		}
		cmds = append(cmds, loadCmd)
		return m, tea.Batch(cmds...)
	}

	m.liveViewLoading = false
	return m, tea.Batch(cmds...)
}

// updateLiveListSize sets the live list dimensions based on window size.
func (m *model) updateLiveListSize() {
	const spinnerHeight = 3
	leftWidth := max(m.width*35/100, 25)
	if m.width == 0 {
		leftWidth = 40
	}

	frameWidth := 4
	frameHeight := 6
	titleHeight := 3
	availableWidth := leftWidth - frameWidth
	availableHeight := m.height - frameHeight - titleHeight - spinnerHeight
	if m.height == 0 {
		availableHeight = 20
	}

	if availableWidth > 0 && availableHeight > 0 {
		m.liveMatchesList.SetSize(availableWidth, availableHeight)
	}
}

// handleFinishedMatches processes finished matches API response.
func (m model) handleFinishedMatches(msg finishedMatchesMsg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	if len(msg.matches) == 0 {
		if m.statsDateRange == 1 {
			return m, nil // Wait for upcoming matches
		}
		m.statsViewLoading = false
		m.loading = false
		return m, nil
	}

	// Deduplicate matches by ID
	seen := make(map[int]bool)
	var uniqueMatches []api.Match
	for _, match := range msg.matches {
		if !seen[match.ID] {
			seen[match.ID] = true
			uniqueMatches = append(uniqueMatches, match)
		}
	}

	// Convert to display format
	displayMatches := make([]ui.MatchDisplay, 0, len(uniqueMatches))
	for _, match := range uniqueMatches {
		displayMatches = append(displayMatches, ui.MatchDisplay{Match: match})
	}

	m.matches = displayMatches
	m.selected = 0
	m.loading = false
	cmds = append(cmds, m.statsViewSpinner.Init())

	// Update list
	m.statsMatchesList.SetItems(ui.ToMatchListItems(displayMatches))
	if len(displayMatches) > 0 {
		m.statsMatchesList.Select(0)
		updatedModel, loadCmd := m.loadStatsMatchDetails(m.matches[0].ID)
		if updatedM, ok := updatedModel.(model); ok {
			m = updatedM
		}
		cmds = append(cmds, loadCmd)
		return m, tea.Batch(cmds...)
	}

	m.statsViewLoading = false
	return m, tea.Batch(cmds...)
}

// handleUpcomingMatches processes upcoming matches API response.
func (m model) handleUpcomingMatches(msg upcomingMatchesMsg) (tea.Model, tea.Cmd) {
	displayMatches := make([]ui.MatchDisplay, 0, len(msg.matches))
	for _, match := range msg.matches {
		displayMatches = append(displayMatches, ui.MatchDisplay{Match: match})
	}

	m.upcomingMatches = displayMatches
	m.upcomingMatchesList.SetItems(ui.ToMatchListItems(displayMatches))

	if m.matchDetails != nil {
		m.statsViewLoading = false
	}

	return m, nil
}

// handleRandomSpinnerTick updates random spinner animations.
func (m model) handleRandomSpinnerTick(msg ui.TickMsg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	if m.mainViewLoading || m.liveViewLoading {
		updatedModel, cmd := m.randomSpinner.Update(msg)
		if s, ok := updatedModel.(*ui.RandomCharSpinner); ok {
			m.randomSpinner = s
		}
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	if m.statsViewLoading {
		updatedModel, cmd := m.statsViewSpinner.Update(msg)
		if s, ok := updatedModel.(*ui.RandomCharSpinner); ok {
			m.statsViewSpinner = s
		}
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

// handleMainViewCheck processes main view check completion and navigates to selected view.
func (m model) handleMainViewCheck(msg mainViewCheckMsg) (tea.Model, tea.Cmd) {
	m.mainViewLoading = false

	// Clear previous view state
	m.matches = nil
	m.upcomingMatches = nil
	m.matchDetails = nil
	m.liveUpdates = nil
	m.lastEvents = nil
	m.polling = false
	m.selected = 0
	m.upcomingMatchesList.SetItems([]list.Item{})

	switch msg.selection {
	case 0: // Stats view
		m.statsViewLoading = true
		m.currentView = viewStats
		m.loading = true
		m.matchDetailsCache = make(map[int]*api.MatchDetails)

		cmds := []tea.Cmd{
			m.spinner.Tick,
			m.statsViewSpinner.Init(),
			fetchFinishedMatchesFotmob(m.fotmobClient, m.useMockData, m.statsDateRange),
		}

		if m.statsDateRange == 1 {
			cmds = append(cmds, fetchUpcomingMatchesFotmob(m.fotmobClient, m.useMockData))
		}

		return m, tea.Batch(cmds...)

	case 1: // Live Matches view
		m.currentView = viewLiveMatches
		m.loading = true
		m.liveViewLoading = true
		return m, tea.Batch(m.spinner.Tick, m.randomSpinner.Init(), fetchLiveMatches(m.fotmobClient, m.useMockData))
	}

	return m, nil
}

// max returns the larger of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

