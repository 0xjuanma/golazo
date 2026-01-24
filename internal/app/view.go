package app

import (
	"fmt"

	"github.com/0xjuanma/golazo/internal/reddit"
	"github.com/0xjuanma/golazo/internal/ui"
)

// View renders the current application state.
func (m model) View() string {
	// DEBUG: Log that view is being called
	m.debugLog(fmt.Sprintf("VIEW: View() called, currentView=%v, width=%d, height=%d, matchDetails=%v", m.currentView, m.width, m.height, m.matchDetails != nil))
	if m.matchDetails != nil {
		m.debugLog(fmt.Sprintf("VIEW: matchDetails ID=%d, Status=%s, Highlights=%v", m.matchDetails.ID, m.matchDetails.Status, m.matchDetails.Highlight != nil))
	}

	switch m.currentView {
	case viewMain:
		return ui.RenderMainMenu(m.width, m.height, m.selected, m.spinner, m.randomSpinner, m.mainViewLoading, m.getStatusBannerType())

	case viewLiveMatches:
		m.ensureLiveListSize()
		return ui.RenderMultiPanelViewWithList(
			m.width, m.height,
			m.liveMatchesList,
			m.matchDetails,
			m.liveUpdates,
			m.spinner,
			m.loading,
			m.randomSpinner,
			m.liveViewLoading,
			m.liveBatchesLoaded,
			m.liveTotalBatches,
			m.pollingSpinner,
			m.polling,
			m.liveUpcomingMatches,
			m.buildGoalLinksMap(),
			m.getStatusBannerType(),
		)

	case viewStats:
		m.ensureStatsListSize()
		spinner := m.ensureStatsSpinner()
		return ui.RenderStatsViewWithList(
			m.width, m.height,
			m.statsMatchesList,
			m.matchDetails,
			spinner,
			m.statsViewLoading,
			m.statsDateRange,
			m.statsDaysLoaded,
			m.statsTotalDays,
			m.buildGoalLinksMap(),
			m.getStatusBannerType(),
			&m.statsDetailsViewport,
			m.statsRightPanelFocused,
			m.statsScrollOffset,
		)

	case viewSettings:
		return ui.RenderSettingsView(m.width, m.height, m.settingsState, m.getStatusBannerType())

	default:
		return ui.RenderMainMenu(m.width, m.height, m.selected, m.spinner, m.randomSpinner, m.mainViewLoading, m.getStatusBannerType())
	}
}

// ensureLiveListSize ensures list dimensions are set before rendering.
func (m *model) ensureLiveListSize() {
	if m.width <= 0 || m.height <= 0 {
		return
	}

	const (
		frameH        = 2
		frameV        = 2
		titleHeight   = 3
		spinnerHeight = 3
	)

	leftWidth := max(m.width*35/100, 25)
	availableWidth := leftWidth - frameH*2
	availableHeight := m.height - frameV*2 - titleHeight - spinnerHeight

	if availableWidth > 0 && availableHeight > 0 {
		m.liveMatchesList.SetSize(availableWidth, availableHeight)
	}
}

// ensureStatsListSize ensures stats list dimensions are set before rendering.
func (m *model) ensureStatsListSize() {
	if m.width <= 0 || m.height <= 0 {
		return
	}

	const (
		frameH         = 2
		frameV         = 2
		titleHeight    = 3
		spinnerHeight  = 3
		headerHeight   = 2 // "Match List" header + spacing
		selectorHeight = 2 // Date selector + spacing
	)

	leftWidth := max(m.width*40/100, 30)
	availableWidth := leftWidth - frameH*2
	availableHeight := m.height - frameV*2 - titleHeight - spinnerHeight - headerHeight - selectorHeight

	if availableWidth > 0 && availableHeight > 0 {
		m.statsMatchesList.SetSize(availableWidth, availableHeight)
	}
}

// ensureStatsSpinner ensures stats spinner is initialized.
func (m *model) ensureStatsSpinner() *ui.RandomCharSpinner {
	if m.statsViewSpinner == nil {
		m.statsViewSpinner = ui.NewRandomCharSpinner()
		m.statsViewSpinner.SetWidth(30)
	}
	return m.statsViewSpinner
}

// buildGoalLinksMap converts the model's goal links to a UI-friendly map.
// Also triggers fetching for any goals that exist in match details but are not cached.
func (m *model) buildGoalLinksMap() ui.GoalLinksMap {
	if len(m.goalLinks) == 0 {
		return nil
	}

	result := make(ui.GoalLinksMap)
	for key, link := range m.goalLinks {
		// Filter out "__NOT_FOUND__" and invalid URLs using helper function
		if link != nil && ui.IsValidReplayURL(link.URL) {
			uiKey := ui.MakeGoalLinkKey(key.MatchID, key.Minute)
			result[uiKey] = link.URL
		}
	}
	return result
}

// Ensure reddit.GoalLinkKey is used (avoid unused import)
var _ reddit.GoalLinkKey
