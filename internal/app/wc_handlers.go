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
	}
	return m, nil
}

// handleWCGroupsKeys handles input on the groups list.
// Enter navigates to group detail; b opens the bracket; g opens the grid.
func (m model) handleWCGroupsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.wcData == nil {
		return m, nil
	}

	switch msg.String() {
	case "enter":
		if item, ok := m.wcGroupsList.SelectedItem().(ui.WCGroupItem); ok {
			for i, g := range m.wcData.Groups {
				if g.Letter == item.Group.Letter {
					m.wcSelectedGroup = i
					break
				}
			}
			m.wcSubView = wcSubViewGroupDetail
		}
		return m, nil

	case "b":
		if len(m.wcData.KnockoutRounds) > 0 {
			m.wcBracketScroll = 0
			m.wcSubView = wcSubViewBracket
		}
		return m, nil

	case "g":
		m.wcGridSelectedIdx = 0
		m.wcSubView = wcSubViewGroupGrid
		return m, nil

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
		m.wcSubView = wcSubViewGroups
	}
	return m, nil
}

// handleWCBracketKeys handles input on the bracket view.
func (m model) handleWCBracketKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.wcSubView = wcSubViewGroups
	case "j", "down":
		if m.wcBracketLines > 0 && m.wcBracketScroll < m.wcBracketLines-1 {
			m.wcBracketScroll++
		}
	case "k", "up":
		if m.wcBracketScroll > 0 {
			m.wcBracketScroll--
		}
	}
	return m, nil
}

// handleWCGroupGridKeys handles input on the all-groups grid view.
func (m model) handleWCGroupGridKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.wcData == nil {
		return m, nil
	}
	n := len(m.wcData.Groups)
	if n == 0 {
		if msg.String() == "esc" {
			m.wcSubView = wcSubViewGroups
		}
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
	case "esc":
		m.wcSubView = wcSubViewGroups

	case "enter":
		m.wcSelectedGroup = m.wcGridSelectedIdx
		m.wcSubView = wcSubViewGroupDetail

	case "b":
		if len(m.wcData.KnockoutRounds) > 0 {
			m.wcBracketScroll = 0
			m.wcSubView = wcSubViewBracket
		}

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
	m.wcBracketLines = msg.data.BracketLineCount()

	items := make([]list.Item, len(msg.data.Groups))
	for i, g := range msg.data.Groups {
		items[i] = ui.WCGroupItem{Group: g}
	}
	m.wcGroupsList.SetItems(items)
	return m, nil
}
