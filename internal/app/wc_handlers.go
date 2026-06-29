package app

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/0xjuanma/golazo/internal/ui"
)

// handleWorldCupKeys routes keyboard input to the active WC sub-view handler.
func (m model) handleWorldCupKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.wcLoading {
		return m, nil
	}
	switch m.wcSubView {
	case wcSubViewGroups:
		return m.handleWCGroupsKeys(msg)
	case wcSubViewGroupDetail:
		return m.handleWCGroupDetailKeys(msg)
	case wcSubViewBracket:
		return m.handleWCBracketKeys(msg)
	case wcSubViewGroupGrid:
		return m.handleWCGroupGridKeys(msg)
	case wcSubViewUpcoming:
		return m.handleWCUpcomingKeys(msg)
	}
	return m, nil
}

// handleWCGroupsKeys handles input on the groups list sub-view.
// Enter navigates to group detail; b opens the bracket; u opens upcoming;
// Esc returns to the grid (the home sub-view).
func (m model) handleWCGroupsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.wcData == nil {
		return m, nil
	}

	switch msg.String() {
	case "esc":
		m.wcSubView = wcSubViewGroupGrid
		return m, tea.ClearScreen

	case "enter":
		if item, ok := m.wcGroupsList.SelectedItem().(ui.WCGroupItem); ok {
			for i, g := range m.wcData.Groups {
				if g.Letter == item.Group.Letter {
					m.wcSelectedGroup = i
					break
				}
			}
			m.wcSubView = wcSubViewGroupDetail
			return m, tea.ClearScreen
		}
		return m, nil

	case "b":
		if len(m.wcData.KnockoutRounds) > 0 {
			m.wcSubView = wcSubViewBracket
			return m, tea.ClearScreen
		}
		m.wcLastError = "Bracket not available yet — group stage in progress"
		return m, nil

	case "u":
		m.wcSubView = wcSubViewUpcoming
		m.wcUpcomingLoading = true
		m.wcUpcomingLastError = ""
		return m, tea.Batch(tea.ClearScreen, fetchWorldCupUpcoming(m.loadCtx, m.fotmobClient))

	case "s":
		return m.openTopScorersDialog()

	default:
		var cmd tea.Cmd
		m.wcGroupsList, cmd = m.wcGroupsList.Update(msg)
		return m, cmd
	}
}

// handleWCGroupDetailKeys handles input on the group detail view.
func (m model) handleWCGroupDetailKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.wcSubView = wcSubViewGroupGrid
		return m, tea.ClearScreen
	}
	return m, nil
}

// handleWCBracketKeys handles input on the bracket view.
func (m model) handleWCBracketKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.wcSubView = wcSubViewGroupGrid
		return m, tea.ClearScreen
	case "u":
		m.wcSubView = wcSubViewUpcoming
		m.wcUpcomingLoading = true
		m.wcUpcomingLastError = ""
		return m, tea.Batch(tea.ClearScreen, fetchWorldCupUpcoming(m.loadCtx, m.fotmobClient))

	case "s":
		return m.openTopScorersDialog()
	}
	return m, nil
}

// handleWCGroupGridKeys handles input on the all-groups grid view, which is
// the home sub-view of the World Cup view. Esc on the grid is absorbed by
// the outer update flow (resets to the main menu); t opens the scrollable
// groups list (table view).
func (m model) handleWCGroupGridKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.wcData == nil {
		return m, nil
	}
	n := len(m.wcData.Groups)
	if n == 0 {
		return m, nil
	}

	// Determine column count matching RenderGroupGrid's logic
	cols := 2
	if m.width > 120 {
		cols = 4
	} else if m.width > 80 {
		cols = 3
	}

	switch msg.String() {
	case "enter":
		m.wcSelectedGroup = m.wcGridSelectedIdx
		m.wcSubView = wcSubViewGroupDetail
		return m, tea.ClearScreen

	case "t":
		m.wcSubView = wcSubViewGroups
		return m, tea.ClearScreen

	case "b":
		if len(m.wcData.KnockoutRounds) > 0 {
			m.wcSubView = wcSubViewBracket
			return m, tea.ClearScreen
		}
		m.wcLastError = "Bracket not available yet — group stage in progress"
		m.wcSubView = wcSubViewGroups
		return m, tea.ClearScreen

	case "u":
		m.wcSubView = wcSubViewUpcoming
		m.wcUpcomingLoading = true
		m.wcUpcomingLastError = ""
		return m, tea.Batch(tea.ClearScreen, fetchWorldCupUpcoming(m.loadCtx, m.fotmobClient))

	case "s":
		return m.openTopScorersDialog()

	case "right", "l":
		if m.wcGridSelectedIdx < n-1 {
			m.wcGridSelectedIdx++
		}

	case "left", "h":
		if m.wcGridSelectedIdx > 0 {
			m.wcGridSelectedIdx--
		}

	case "down", "j":
		if m.wcGridSelectedIdx+cols < n {
			m.wcGridSelectedIdx += cols
		}

	case "up", "k":
		if m.wcGridSelectedIdx-cols >= 0 {
			m.wcGridSelectedIdx -= cols
		}
	}
	return m, nil
}

// handleWCData processes the World Cup data message and populates the groups list.
func (m model) handleWCData(msg wcDataMsg) (tea.Model, tea.Cmd) {
	m.wcLoading = false
	if msg.err != nil {
		m.wcLastError = "Failed to load World Cup data"
		return m, nil
	}
	m.wcData = msg.data
	m.wcLastError = ""

	items := make([]list.Item, len(msg.data.Groups))
	for i, g := range msg.data.Groups {
		items[i] = ui.WCGroupItem{Group: g}
	}
	m.wcGroupsList.SetItems(items)
	return m, nil
}

// handleWCUpcomingKeys handles input on the upcoming-matches sub-view.
// Only Esc is meaningful — it returns to the grid (home sub-view).
func (m model) handleWCUpcomingKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.wcSubView = wcSubViewGroupGrid
		return m, tea.ClearScreen
	}
	return m, nil
}

// handleWCUpcoming processes the upcoming-matches response message.
func (m model) handleWCUpcoming(msg wcUpcomingMsg) (tea.Model, tea.Cmd) {
	m.wcUpcomingLoading = false
	if msg.err != nil {
		m.wcUpcomingLastError = "Failed to load upcoming matches"
		return m, nil
	}
	m.wcUpcoming = msg.matches
	m.wcUpcomingLastError = ""
	return m, nil
}

// handleWCTopScorers processes the top scorers response, stores the data, and opens the dialog.
func (m model) handleWCTopScorers(msg wcTopScorersMsg) (tea.Model, tea.Cmd) {
	m.wcTopScorersLoading = false
	if msg.err != nil {
		return m, nil
	}
	m.wcTopScorers = msg.scorers
	if m.dialogOverlay != nil {
		m.dialogOverlay.OpenDialog(ui.NewTopScorersDialog(msg.scorers))
	}
	return m, nil
}

// openTopScorersDialog opens the top scorers dialog, using cached data when available.
func (m model) openTopScorersDialog() (tea.Model, tea.Cmd) {
	if len(m.wcTopScorers) > 0 && m.dialogOverlay != nil {
		m.dialogOverlay.OpenDialog(ui.NewTopScorersDialog(m.wcTopScorers))
		return m, nil
	}
	m.wcTopScorersLoading = true
	return m, fetchWorldCupTopScorers(m.loadCtx, m.fotmobClient, m.wcYear)
}
