package app

import "github.com/0xjuanma/golazo/internal/ui"

// View renders the current application state.
func (m model) View() string {
	switch m.currentView {
	case viewMain:
		return ui.RenderMainMenu(m.width, m.height, m.selected, m.spinner, m.randomSpinner, m.mainViewLoading)

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
			m.detailsViewport,
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
			m.detailsViewport,
		)

	case viewSettings:
		return ui.RenderSettingsView(m.width, m.height, m.settingsState)

	default:
		return ui.RenderMainMenu(m.width, m.height, m.selected, m.spinner, m.randomSpinner, m.mainViewLoading)
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
		frameH        = 2
		frameV        = 2
		titleHeight   = 3
		spinnerHeight = 3
	)

	leftWidth := max(m.width*40/100, 30)
	availableWidth := leftWidth - frameH*2
	availableHeight := m.height - frameV*2 - titleHeight - spinnerHeight

	if availableWidth > 0 && availableHeight > 0 {
		// Upcoming matches are now shown in Live view, so give full height to finished list
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
