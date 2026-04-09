package app

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/0xjuanma/golazo/internal/ui"
)

// wcSubView represents the current sub-view within the World Cup view.
type wcSubView int

const (
	wcSubViewGroups      wcSubView = iota // scrollable group list
	wcSubViewGroupDetail                  // single group expanded detail
	wcSubViewBracket                      // knockout bracket
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
	}
	return m, nil
}

// handleWCGroupsKeys handles input on the groups list.
// Enter navigates to group detail; b opens the bracket; all other keys are
// delegated to the bubbles/list component for built-in navigation and filtering.
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
		m.wcBracketScroll++
	case "k", "up":
		if m.wcBracketScroll > 0 {
			m.wcBracketScroll--
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
