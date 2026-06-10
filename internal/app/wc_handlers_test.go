package app

import (
	"context"
	"testing"

	"github.com/0xjuanma/golazo/internal/api"
	tea "github.com/charmbracelet/bubbletea"
)

// newWCTestModel builds the minimal model needed to exercise the WC handlers
// in isolation. Live/stats fields are left at their zero value because the
// World Cup handlers never read them.
func newWCTestModel() model {
	return model{
		loadCtx:   context.Background(),
		wcSubView: wcSubViewGroupGrid,
		wcData: &api.WorldCupData{
			Groups: []api.WCGroup{{Letter: "A"}, {Letter: "B"}},
		},
	}
}

func TestHandleWCGroupGridKeys_UTransitionsToUpcoming(t *testing.T) {
	m := newWCTestModel()

	next, cmd := m.handleWCGroupGridKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})

	nm := next.(model)
	if nm.wcSubView != wcSubViewUpcoming {
		t.Errorf("wcSubView = %v, want wcSubViewUpcoming", nm.wcSubView)
	}
	if !nm.wcUpcomingLoading {
		t.Error("wcUpcomingLoading = false, want true after pressing u")
	}
	if cmd == nil {
		t.Error("expected non-nil tea.Cmd to dispatch fetchWorldCupUpcoming")
	}
}

func TestHandleWCGroupGridKeys_TTransitionsToGroupsList(t *testing.T) {
	m := newWCTestModel()

	next, _ := m.handleWCGroupGridKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})

	nm := next.(model)
	if nm.wcSubView != wcSubViewGroups {
		t.Errorf("wcSubView = %v, want wcSubViewGroups", nm.wcSubView)
	}
}

func TestHandleWCGroupsKeys_EscReturnsToGrid(t *testing.T) {
	m := newWCTestModel()
	m.wcSubView = wcSubViewGroups

	next, _ := m.handleWCGroupsKeys(tea.KeyMsg{Type: tea.KeyEsc})

	nm := next.(model)
	if nm.wcSubView != wcSubViewGroupGrid {
		t.Errorf("wcSubView = %v, want wcSubViewGroupGrid", nm.wcSubView)
	}
}

func TestHandleWCUpcomingKeys_EscReturnsToGrid(t *testing.T) {
	m := newWCTestModel()
	m.wcSubView = wcSubViewUpcoming

	next, _ := m.handleWCUpcomingKeys(tea.KeyMsg{Type: tea.KeyEsc})

	nm := next.(model)
	if nm.wcSubView != wcSubViewGroupGrid {
		t.Errorf("wcSubView = %v, want wcSubViewGroupGrid", nm.wcSubView)
	}
}

func TestHandleWCUpcoming_StoresMatches(t *testing.T) {
	m := newWCTestModel()
	m.wcUpcomingLoading = true

	matches := []api.Match{{ID: 1}, {ID: 2}}
	next, _ := m.handleWCUpcoming(wcUpcomingMsg{matches: matches})

	nm := next.(model)
	if nm.wcUpcomingLoading {
		t.Error("wcUpcomingLoading should be false after message handled")
	}
	if len(nm.wcUpcoming) != 2 {
		t.Errorf("wcUpcoming len = %d, want 2", len(nm.wcUpcoming))
	}
	if nm.wcUpcomingLastError != "" {
		t.Errorf("wcUpcomingLastError = %q, want empty", nm.wcUpcomingLastError)
	}
}

func TestHandleWCUpcoming_StoresError(t *testing.T) {
	m := newWCTestModel()
	m.wcUpcomingLoading = true

	next, _ := m.handleWCUpcoming(wcUpcomingMsg{err: context.DeadlineExceeded})

	nm := next.(model)
	if nm.wcUpcomingLoading {
		t.Error("wcUpcomingLoading should be false after error message handled")
	}
	if nm.wcUpcomingLastError == "" {
		t.Error("wcUpcomingLastError should be set after error message")
	}
}
