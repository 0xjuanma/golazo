package app

import (
	"strings"
	"time"

	"github.com/0xjuanma/golazo/internal/api"
	"github.com/0xjuanma/golazo/internal/fotmob"
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

	case liveRefreshMsg:
		return m.handleLiveRefresh(msg)

	case liveBatchDataMsg:
		return m.handleLiveBatchData(msg)

	case statsDataMsg:
		return m.handleStatsData(msg)

	case statsDayDataMsg:
		return m.handleStatsDayData(msg)

	case ui.TickMsg:
		return m.handleRandomSpinnerTick(msg)

	case mainViewCheckMsg:
		return m.handleMainViewCheck(msg)

	case pollTickMsg:
		return m.handlePollTick(msg)

	case pollDisplayCompleteMsg:
		return m.handlePollDisplayComplete()

	case list.FilterMatchesMsg:
		// Route filter matches message to the appropriate list based on current view
		return m.handleFilterMatches(msg)

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

	// Update viewport size for match details panel
	m.updateDetailsViewportSize()

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
			// Upcoming matches are now shown in Live view, so give full height to finished list
			m.statsMatchesList.SetSize(availableWidth, availableHeight)
		}

	case viewSettings:
		// Settings list size is handled in RenderSettingsView
		// but we update it here too for consistency
		if m.settingsState != nil {
			listHeight := m.height - 11 // Account for title, info, help
			if listHeight < 5 {
				listHeight = 5
			}
			m.settingsState.List.SetSize(48, listHeight)
		}
	}

	return m, nil
}

// updateDetailsViewportSize calculates and sets the viewport dimensions for the match details panel.
func (m *model) updateDetailsViewportSize() {
	if m.width <= 0 || m.height <= 0 {
		return
	}

	// Calculate right panel dimensions (matches the render functions)
	spinnerHeight := 3
	availableHeight := m.height - spinnerHeight
	if availableHeight < 10 {
		availableHeight = 10
	}

	leftWidth := m.width * 35 / 100
	if leftWidth < 25 {
		leftWidth = 25
	}
	rightWidth := m.width - leftWidth - 1
	if rightWidth < 35 {
		rightWidth = 35
	}

	panelHeight := availableHeight - 2

	// Header takes approximately 8 lines (status, teams, score block)
	// Panel border/padding takes 4 lines, title takes 2 lines
	headerHeight := 10
	viewportHeight := panelHeight - headerHeight - 4
	if viewportHeight < 3 {
		viewportHeight = 3
	}

	viewportWidth := rightWidth - 6 // Account for panel padding
	if viewportWidth < 10 {
		viewportWidth = 10
	}

	m.detailsViewport.Width = viewportWidth
	m.detailsViewport.Height = viewportHeight
}

// updateDetailsViewportContent updates the viewport content based on current match details.
func (m *model) updateDetailsViewportContent() {
	if m.matchDetails == nil {
		m.detailsViewport.SetContent("")
		return
	}

	// Calculate right panel width for content rendering
	leftWidth := m.width * 35 / 100
	if leftWidth < 25 {
		leftWidth = 25
	}
	rightWidth := m.width - leftWidth - 1
	if rightWidth < 35 {
		rightWidth = 35
	}

	var content string
	if m.currentView == viewStats || m.pendingSelection == 0 {
		// Stats view - use stats scrollable content
		content = ui.RenderStatsMatchDetailsScrollableContent(rightWidth, m.matchDetails)
	} else {
		// Live view - use live scrollable content
		content = ui.RenderMatchDetailsScrollableContent(rightWidth, m.matchDetails, m.liveUpdates, m.pollingSpinner, m.polling, m.loading)
	}

	m.detailsViewport.SetContent(content)
	m.detailsViewport.GotoTop()
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
		return m, schedulePollTick(m.matchDetails.ID)
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
		m.detailsViewport.SetContent("")
		m.loading = false
		m.liveViewLoading = false
		m.statsViewLoading = false
		return m, nil
	}

	m.matchDetails = msg.details

	// Cache for stats view (including during preload)
	if m.currentView == viewStats || m.pendingSelection == 0 {
		m.matchDetailsCache[msg.details.ID] = msg.details
		m.updateDetailsViewportSize()
		m.updateDetailsViewportContent()
		m.loading = false
		m.statsViewLoading = false
		return m, nil
	}

	// Handle live matches view (including during preload)
	if m.currentView == viewLiveMatches || m.pendingSelection == 1 {
		m.liveViewLoading = false

		// Get current scores
		homeScore := 0
		awayScore := 0
		if msg.details.HomeScore != nil {
			homeScore = *msg.details.HomeScore
		}
		if msg.details.AwayScore != nil {
			awayScore = *msg.details.AwayScore
		}

		// Detect new goals during poll refresh (not initial load)
		// Only notify when: polling is active AND we have previous score data
		hasScoreData := m.lastHomeScore > 0 || m.lastAwayScore > 0 || len(m.lastEvents) > 0
		if m.polling && hasScoreData {
			m.notifyNewGoals(msg.details)
		}

		// Update tracked scores for next comparison
		m.lastHomeScore = homeScore
		m.lastAwayScore = awayScore

		// Parse ALL events to rebuild the live updates list
		// This ensures proper ordering (descending by minute) and uniqueness
		m.liveUpdates = m.parser.ParseEvents(msg.details.Events, msg.details.HomeTeam, msg.details.AwayTeam)
		m.lastEvents = msg.details.Events

		// Update viewport with new content
		m.updateDetailsViewportSize()
		m.updateDetailsViewportContent()

		// Continue polling if match is live
		if msg.details.Status == api.MatchStatusLive {
			// For initial load, clear loading state
			// For poll refresh, loading is cleared by 1s timer (pollDisplayCompleteMsg)
			if !m.polling {
				m.loading = false
			}
			// Note: if m.polling is true, m.loading stays true until the 1s timer fires

			m.polling = true
			// Schedule next poll tick (90 seconds from now)
			cmds = append(cmds, schedulePollTick(msg.details.ID))
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
		// Check if any list is in filtering mode - if so, let the list handle Esc
		// to cancel the filter instead of navigating back
		isFiltering := false
		switch m.currentView {
		case viewLiveMatches:
			isFiltering = m.liveMatchesList.FilterState() == list.Filtering ||
				m.liveMatchesList.FilterState() == list.FilterApplied
		case viewStats:
			isFiltering = m.statsMatchesList.FilterState() == list.Filtering ||
				m.statsMatchesList.FilterState() == list.FilterApplied
		case viewSettings:
			if m.settingsState != nil {
				isFiltering = m.settingsState.List.FilterState() == list.Filtering ||
					m.settingsState.List.FilterState() == list.FilterApplied
			}
		}

		if isFiltering {
			// Let the view-specific handler pass Esc to the list to cancel filter
			break
		}

		if m.currentView != viewMain {
			return m.resetToMainView()
		}
	case "pgup", "pgdown":
		// Handle PageUp/PageDown for viewport scrolling in live and stats views
		if m.currentView == viewLiveMatches || m.currentView == viewStats {
			var cmd tea.Cmd
			m.detailsViewport, cmd = m.detailsViewport.Update(msg)
			return m, cmd
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
	case viewSettings:
		return m.handleSettingsViewKeys(msg)
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
	m.lastHomeScore = 0
	m.lastAwayScore = 0
	m.loading = false
	m.polling = false
	m.matches = nil
	m.upcomingMatches = nil
	return m, nil
}

// handleLiveMatchesSelection handles list navigation in live matches view.
func (m model) handleLiveMatchesSelection(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Capture selected item BEFORE Update (critical for filter mode - selection changes after filter clears)
	var preUpdateMatchID int
	if preItem := m.liveMatchesList.SelectedItem(); preItem != nil {
		if item, ok := preItem.(ui.MatchListItem); ok {
			preUpdateMatchID = item.Match.ID
		}
	}

	var listCmd tea.Cmd
	m.liveMatchesList, listCmd = m.liveMatchesList.Update(msg)

	// Get currently displayed match ID
	currentMatchID := 0
	if m.matchDetails != nil {
		currentMatchID = m.matchDetails.ID
	}

	// Check post-update selection
	var postUpdateMatchID int
	if postItem := m.liveMatchesList.SelectedItem(); postItem != nil {
		if item, ok := postItem.(ui.MatchListItem); ok {
			postUpdateMatchID = item.Match.ID
		}
	}

	// Use pre-update selection if it was valid and different from current
	// This handles the filter case where Enter clears the filter
	targetMatchID := postUpdateMatchID
	if msg.String() == "enter" && preUpdateMatchID != 0 {
		targetMatchID = preUpdateMatchID
	}

	// Load match details if selection changed
	if targetMatchID != 0 && targetMatchID != currentMatchID {
		for i, match := range m.matches {
			if match.ID == targetMatchID {
				m.selected = i
				break
			}
		}
		return m.loadMatchDetails(targetMatchID)
	}

	return m, listCmd
}

// handleStatsSelection handles list navigation and date range changes in stats view.
func (m model) handleStatsSelection(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Check if list is in filtering mode - if so, let list handle ALL keys
	isFiltering := m.statsMatchesList.FilterState() == list.Filtering

	// Only handle date range navigation when NOT filtering
	if !isFiltering {
		if msg.String() == "h" || msg.String() == "left" || msg.String() == "l" || msg.String() == "right" {
			return m.handleStatsViewKeys(msg)
		}
	}

	// Capture selected item BEFORE Update (critical for filter mode - selection changes after filter clears)
	var preUpdateMatchID int
	if preItem := m.statsMatchesList.SelectedItem(); preItem != nil {
		if item, ok := preItem.(ui.MatchListItem); ok {
			preUpdateMatchID = item.Match.ID
		}
	}

	// Handle list navigation
	var listCmd tea.Cmd
	m.statsMatchesList, listCmd = m.statsMatchesList.Update(msg)

	// Get currently displayed match ID
	currentMatchID := 0
	if m.matchDetails != nil {
		currentMatchID = m.matchDetails.ID
	}

	// Check post-update selection
	var postUpdateMatchID int
	if postItem := m.statsMatchesList.SelectedItem(); postItem != nil {
		if item, ok := postItem.(ui.MatchListItem); ok {
			postUpdateMatchID = item.Match.ID
		}
	}

	// Use pre-update selection if it was valid and different from current
	// This handles the filter case where Enter clears the filter
	targetMatchID := postUpdateMatchID
	if msg.String() == "enter" && preUpdateMatchID != 0 {
		targetMatchID = preUpdateMatchID
	}

	// Load match details if selection changed
	if targetMatchID != 0 && targetMatchID != currentMatchID {
		for i, match := range m.matches {
			if match.ID == targetMatchID {
				m.selected = i
				break
			}
		}
		return m.loadStatsMatchDetails(targetMatchID)
	}

	return m, listCmd
}

// handleLiveMatches processes live matches API response.
func (m model) handleLiveMatches(msg liveMatchesMsg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Schedule the next refresh (5-min timer)
	cmds = append(cmds, scheduleLiveRefresh(m.fotmobClient, m.useMockData))

	if len(msg.matches) == 0 {
		m.liveViewLoading = false
		m.loading = false
		return m, tea.Batch(cmds...)
	}

	// Convert to display format
	displayMatches := make([]ui.MatchDisplay, 0, len(msg.matches))
	for _, match := range msg.matches {
		displayMatches = append(displayMatches, ui.MatchDisplay{Match: match})
	}

	m.matches = displayMatches
	m.selected = 0
	m.loading = false
	cmds = append(cmds, ui.SpinnerTick())

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

// handleLiveRefresh processes periodic live matches refresh (every 5 min).
// Only updates if still in the live view.
func (m model) handleLiveRefresh(msg liveRefreshMsg) (tea.Model, tea.Cmd) {
	// Ignore refresh if not in live view (user navigated away)
	if m.currentView != viewLiveMatches {
		return m, nil
	}

	var cmds []tea.Cmd

	// Schedule the next refresh
	cmds = append(cmds, scheduleLiveRefresh(m.fotmobClient, m.useMockData))

	if len(msg.matches) == 0 {
		// No live matches - clear list but keep view
		m.matches = nil
		m.liveMatchesList.SetItems(nil)
		return m, tea.Batch(cmds...)
	}

	// Convert to display format
	displayMatches := make([]ui.MatchDisplay, 0, len(msg.matches))
	for _, match := range msg.matches {
		displayMatches = append(displayMatches, ui.MatchDisplay{Match: match})
	}

	// Preserve current selection if possible
	currentMatchID := 0
	if m.selected >= 0 && m.selected < len(m.matches) {
		currentMatchID = m.matches[m.selected].ID
	}

	m.matches = displayMatches
	m.liveMatchesList.SetItems(ui.ToMatchListItems(displayMatches))
	m.updateLiveListSize()

	// Try to restore previous selection
	newSelected := 0
	for i, match := range displayMatches {
		if match.ID == currentMatchID {
			newSelected = i
			break
		}
	}
	m.selected = newSelected
	m.liveMatchesList.Select(newSelected)

	return m, tea.Batch(cmds...)
}

// handleLiveBatchData processes parallel batch loading - multiple leagues at once.
// Results are shown after each batch completes, giving progressive updates while being fast.
func (m model) handleLiveBatchData(msg liveBatchDataMsg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Accumulate live matches from this batch
	if len(msg.matches) > 0 {
		m.liveMatchesBuffer = append(m.liveMatchesBuffer, msg.matches...)
	}

	// Track progress
	m.liveBatchesLoaded++

	// Update UI immediately with current data
	if len(m.liveMatchesBuffer) > 0 {
		displayMatches := make([]ui.MatchDisplay, 0, len(m.liveMatchesBuffer))
		for _, match := range m.liveMatchesBuffer {
			displayMatches = append(displayMatches, ui.MatchDisplay{Match: match})
		}
		m.matches = displayMatches
		m.liveMatchesList.SetItems(ui.ToMatchListItems(displayMatches))
		m.updateLiveListSize()

		// On first batch with matches, select first match and load details
		if msg.batchIndex == 0 || (len(msg.matches) > 0 && m.matchDetails == nil && len(m.matches) > 0) {
			if m.selected == 0 && m.matchDetails == nil && len(m.matches) > 0 {
				m.liveMatchesList.Select(0)
				updatedModel, loadCmd := m.loadMatchDetails(m.matches[0].ID)
				if updatedM, ok := updatedModel.(model); ok {
					m = updatedM
				}
				cmds = append(cmds, loadCmd)
			}
		}
	}

	// If last batch, finalize loading
	if msg.isLast {
		m.liveViewLoading = false
		m.loading = false

		// Cache the final result
		if m.fotmobClient != nil && len(m.liveMatchesBuffer) > 0 {
			m.fotmobClient.Cache().SetLiveMatches(m.liveMatchesBuffer)
		}

		// Schedule periodic refresh
		cmds = append(cmds, scheduleLiveRefresh(m.fotmobClient, m.useMockData))

		return m, tea.Batch(cmds...)
	}

	// Otherwise, fetch next batch
	nextBatchIndex := msg.batchIndex + 1
	cmds = append(cmds, fetchLiveBatchData(m.fotmobClient, m.useMockData, nextBatchIndex))

	// Keep spinner running
	cmds = append(cmds, ui.SpinnerTick())

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

// handleStatsData processes the unified stats data API response.
// This is the main handler for stats view - always receives 3 days of data,
// then filters client-side based on the selected date range.
func (m model) handleStatsData(msg statsDataMsg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	if msg.data == nil {
		m.statsViewLoading = false
		m.loading = false
		return m, nil
	}

	// Store the full stats data for client-side filtering
	m.statsData = msg.data

	// Apply the current date range filter
	m.applyStatsDateFilter()

	m.selected = 0
	m.loading = false

	// If we have matches, load details for the first one
	if len(m.matches) > 0 {
		m.statsMatchesList.Select(0)
		updatedModel, loadCmd := m.loadStatsMatchDetails(m.matches[0].ID)
		if updatedM, ok := updatedModel.(model); ok {
			m = updatedM
		}
		cmds = append(cmds, loadCmd)
		return m, tea.Batch(cmds...)
	}

	// No matches - stop spinner
	m.statsViewLoading = false
	return m, nil
}

// handleStatsDayData processes progressive loading - one day's data at a time.
// Results are shown immediately as each day completes, giving instant feedback.
func (m model) handleStatsDayData(msg statsDayDataMsg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Initialize statsData if nil (first day)
	if m.statsData == nil {
		m.statsData = &fotmob.StatsData{
			AllFinished:   []api.Match{},
			TodayFinished: []api.Match{},
			TodayUpcoming: []api.Match{},
		}
	}

	// Accumulate finished matches (prepend older matches)
	if len(msg.finished) > 0 {
		m.statsData.AllFinished = append(m.statsData.AllFinished, msg.finished...)

		// Track today's finished separately
		if msg.isToday {
			m.statsData.TodayFinished = append(m.statsData.TodayFinished, msg.finished...)
		}
	}

	// Add upcoming matches (only from today), deduplicated by match ID
	if msg.isToday && len(msg.upcoming) > 0 {
		// Build a set of existing IDs to avoid duplicates
		existingIDs := make(map[int]bool)
		for _, match := range m.statsData.TodayUpcoming {
			existingIDs[match.ID] = true
		}

		// Only add matches that aren't already in the list
		for _, match := range msg.upcoming {
			if !existingIDs[match.ID] {
				m.statsData.TodayUpcoming = append(m.statsData.TodayUpcoming, match)
				existingIDs[match.ID] = true
			}
		}

		// Populate liveUpcomingMatches for the live view
		upcomingDisplay := make([]ui.MatchDisplay, 0, len(m.statsData.TodayUpcoming))
		for _, match := range m.statsData.TodayUpcoming {
			upcomingDisplay = append(upcomingDisplay, ui.MatchDisplay{Match: match})
		}
		m.liveUpcomingMatches = upcomingDisplay
	}

	// Track progress
	m.statsDaysLoaded++

	// Apply filter and update UI immediately with current data
	m.applyStatsDateFilter()

	// On first day with matches, select first match and load details
	firstDayWithMatches := msg.dayIndex == 0 && len(m.matches) > 0 && m.matchDetails == nil
	if firstDayWithMatches {
		m.selected = 0
		m.statsMatchesList.Select(0)
		updatedModel, loadCmd := m.loadStatsMatchDetails(m.matches[0].ID)
		if updatedM, ok := updatedModel.(model); ok {
			m = updatedM
		}
		cmds = append(cmds, loadCmd)
	}

	// If last day, stop loading
	if msg.isLast {
		m.statsViewLoading = false
		m.loading = false
		return m, tea.Batch(cmds...)
	}

	// Otherwise, fetch next day
	nextDayIndex := msg.dayIndex + 1
	cmds = append(cmds, fetchStatsDayData(m.fotmobClient, m.useMockData, nextDayIndex, m.statsTotalDays))

	// Keep spinner running
	cmds = append(cmds, ui.SpinnerTick())

	return m, tea.Batch(cmds...)
}

// applyStatsDateFilter applies the current date range filter to the cached stats data.
// This enables instant switching between Today/3d/5d views without new API calls.
// All filtering is done client-side from the cached 5-day data based on match MatchTime.
func (m *model) applyStatsDateFilter() {
	if m.statsData == nil {
		return
	}

	// Filter all views from AllFinished based on match's actual MatchTime date
	var finishedMatches []api.Match
	switch m.statsDateRange {
	case 1:
		// Today only - filter by match date
		finishedMatches = filterMatchesByDays(m.statsData.AllFinished, 1)
	case 3:
		// Last 3 days - filter by match date
		finishedMatches = filterMatchesByDays(m.statsData.AllFinished, 3)
	default:
		// 5 days - use all data
		finishedMatches = m.statsData.AllFinished
	}

	// Convert to display format
	displayMatches := make([]ui.MatchDisplay, 0, len(finishedMatches))
	for _, match := range finishedMatches {
		displayMatches = append(displayMatches, ui.MatchDisplay{Match: match})
	}
	m.matches = displayMatches
	m.statsMatchesList.SetItems(ui.ToMatchListItems(displayMatches))
	// Note: Upcoming matches are now shown in the Live view instead
}

// filterMatchesByDays filters matches to only include those from the last N days.
// Uses LOCAL time for date comparison so "today" matches user's actual timezone.
func filterMatchesByDays(matches []api.Match, days int) []api.Match {
	if days <= 0 {
		return matches
	}

	// Use local time so "today" matches the user's actual day
	now := time.Now().Local()
	cutoff := now.AddDate(0, 0, -(days - 1)) // Include today as day 1
	cutoffDate := cutoff.Format("2006-01-02")

	var filtered []api.Match
	for _, match := range matches {
		if match.MatchTime != nil {
			// Compare in local time
			matchDate := match.MatchTime.Local().Format("2006-01-02")
			if matchDate >= cutoffDate {
				filtered = append(filtered, match)
			}
		}
	}
	return filtered
}

// handleRandomSpinnerTick updates all active spinner animations.
// Uses a SINGLE tick chain - all spinners share the same tick rate.
func (m model) handleRandomSpinnerTick(msg ui.TickMsg) (tea.Model, tea.Cmd) {
	// Check if any spinner needs to be animated
	needsTick := m.mainViewLoading || m.liveViewLoading || m.statsViewLoading || m.polling

	if !needsTick {
		// No spinners active - don't continue the tick chain
		return m, nil
	}

	// Update the appropriate spinner(s) based on current state
	if m.mainViewLoading {
		m.randomSpinner.Tick()
	}

	if m.liveViewLoading && m.currentView == viewLiveMatches {
		m.randomSpinner.Tick()
	}

	if m.statsViewLoading {
		m.statsViewSpinner.Tick()
	}

	// Update polling spinner when polling is active
	if m.polling && m.pollingSpinner != nil {
		m.pollingSpinner.Tick()
	}

	// Return ONE tick command to continue the animation chain
	return m, ui.SpinnerTick()
}

// handleMainViewCheck processes main view check completion and navigates to selected view.
func (m model) handleMainViewCheck(msg mainViewCheckMsg) (tea.Model, tea.Cmd) {
	m.mainViewLoading = false
	m.pendingSelection = -1

	var cmds []tea.Cmd

	// Just switch to the target view - API calls already started during selection
	switch msg.selection {
	case 0: // Stats view
		m.currentView = viewStats
		m.selected = 0

		// If matches already loaded, ensure first match is selected
		if len(m.matches) > 0 {
			m.statsMatchesList.Select(0)

			// Load details from cache if available, otherwise start fetch
			if cached, ok := m.matchDetailsCache[m.matches[0].ID]; ok {
				m.matchDetails = cached
			} else if m.matchDetails == nil {
				// Details not loaded yet, start loading
				updatedModel, loadCmd := m.loadStatsMatchDetails(m.matches[0].ID)
				if updatedM, ok := updatedModel.(model); ok {
					m = updatedM
				}
				cmds = append(cmds, loadCmd)
			}
		}

		// Keep spinners running if still loading
		if m.statsViewLoading {
			cmds = append(cmds, m.spinner.Tick, ui.SpinnerTick())
		}

		return m, tea.Batch(cmds...)

	case 1: // Live Matches view
		m.currentView = viewLiveMatches
		m.selected = 0

		// If matches already loaded, ensure first match is selected
		if len(m.matches) > 0 {
			m.liveMatchesList.Select(0)
		}

		// Keep spinners running if still loading
		if m.liveViewLoading {
			cmds = append(cmds, m.spinner.Tick, ui.SpinnerTick())
		}

		return m, tea.Batch(cmds...)
	}

	return m, nil
}

// handlePollTick handles the 90-second poll tick.
// Shows "Updating..." spinner for 1s as visual feedback, then fetches data.
func (m model) handlePollTick(msg pollTickMsg) (tea.Model, tea.Cmd) {
	// Only process if we're still in live view and polling is active
	if m.currentView != viewLiveMatches || !m.polling {
		return m, nil
	}

	// Verify the poll is for the currently selected match
	if m.matchDetails == nil || m.matchDetails.ID != msg.matchID {
		return m, nil
	}

	// Set loading state to show "Updating..." spinner
	m.loading = true

	// Start the actual API call, spinner animation, and 1s display timer
	return m, tea.Batch(
		fetchPollMatchDetails(m.fotmobClient, msg.matchID, m.useMockData),
		ui.SpinnerTick(),
		schedulePollSpinnerHide(), // Hide spinner after 0.5 seconds
	)
}

// handlePollDisplayComplete hides the spinner after 1s display time.
func (m model) handlePollDisplayComplete() (tea.Model, tea.Cmd) {
	// Hide spinner - the 1s visual feedback is complete
	m.loading = false
	return m, nil
}

// handleFilterMatches routes filter matches messages to the appropriate list.
// This is required for the bubbles list filter to work - it fires async matching
// and sends results via FilterMatchesMsg which must be routed back to the list.
func (m model) handleFilterMatches(msg list.FilterMatchesMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch m.currentView {
	case viewLiveMatches:
		m.liveMatchesList, cmd = m.liveMatchesList.Update(msg)
	case viewStats:
		m.statsMatchesList, cmd = m.statsMatchesList.Update(msg)
		// Also update upcoming list in case it's being filtered
		var upCmd tea.Cmd
		m.upcomingMatchesList, upCmd = m.upcomingMatchesList.Update(msg)
		if upCmd != nil {
			cmd = tea.Batch(cmd, upCmd)
		}
	case viewSettings:
		if m.settingsState != nil {
			m.settingsState.List, cmd = m.settingsState.List.Update(msg)
		}
	}

	return m, cmd
}

// notifyNewGoals sends desktop notifications when a goal is scored.
// Uses score-based detection (more reliable than event ID comparison).
// Only called during poll refreshes when we have previous score data.
func (m *model) notifyNewGoals(details *api.MatchDetails) {
	if m.notifier == nil || details == nil {
		return
	}

	// Get current scores
	homeScore := 0
	awayScore := 0
	if details.HomeScore != nil {
		homeScore = *details.HomeScore
	}
	if details.AwayScore != nil {
		awayScore = *details.AwayScore
	}

	// Check if score increased (goal scored)
	homeGoalScored := homeScore > m.lastHomeScore
	awayGoalScored := awayScore > m.lastAwayScore

	if !homeGoalScored && !awayGoalScored {
		return
	}

	// Find the most recent goal event to get player details
	var goalEvent *api.MatchEvent
	for i := len(details.Events) - 1; i >= 0; i-- {
		event := details.Events[i]
		if strings.ToLower(event.Type) == "goal" {
			// Check if this goal matches the team that scored
			if homeGoalScored && event.Team.ID == details.HomeTeam.ID {
				goalEvent = &event
				break
			}
			if awayGoalScored && event.Team.ID == details.AwayTeam.ID {
				goalEvent = &event
				break
			}
		}
	}

	if goalEvent != nil {
		// Send notification - errors are silently ignored to not disrupt the app
		_ = m.notifier.Goal(*goalEvent, details.HomeTeam, details.AwayTeam, homeScore, awayScore)
	}
}

// max returns the larger of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
