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

func TestHandleWCTopScorers_StoresScorers(t *testing.T) {
	m := newWCTestModel()
	m.wcTopScorersLoading = true

	scorers := []api.WCTopScorer{
		{PlayerName: "Messi", Team: "Argentina", Goals: 6},
		{PlayerName: "Mbappé", Team: "France", Goals: 4},
	}
	next, _ := m.handleWCTopScorers(wcTopScorersMsg{scorers: scorers})

	nm := next.(model)
	if nm.wcTopScorersLoading {
		t.Error("wcTopScorersLoading should be false after message handled")
	}
	if len(nm.wcTopScorers) != 2 {
		t.Errorf("wcTopScorers len = %d, want 2", len(nm.wcTopScorers))
	}
	if nm.wcTopScorers[0].Goals != 6 {
		t.Errorf("wcTopScorers[0].Goals = %d, want 6", nm.wcTopScorers[0].Goals)
	}
}

func TestHandleWCGroupsKeys_SEmitsTopScorersCmd(t *testing.T) {
	m := newWCTestModel()
	m.wcSubView = wcSubViewGroups

	_, cmd := m.handleWCGroupsKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	if cmd == nil {
		t.Error("expected non-nil tea.Cmd after pressing s (top scorers fetch)")
	}
}

func TestHandleWCGroupGridKeys_SEmitsTopScorersCmd(t *testing.T) {
	m := newWCTestModel()

	_, cmd := m.handleWCGroupGridKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	if cmd == nil {
		t.Error("expected non-nil tea.Cmd after pressing s (top scorers fetch)")
	}
}

func TestHandleWCTopScorers_ErrorClearsLoading(t *testing.T) {
	m := newWCTestModel()
	m.wcTopScorersLoading = true

	next, _ := m.handleWCTopScorers(wcTopScorersMsg{err: context.DeadlineExceeded})

	nm := next.(model)
	if nm.wcTopScorersLoading {
		t.Error("wcTopScorersLoading should be false after error")
	}
	if len(nm.wcTopScorers) != 0 {
		t.Errorf("wcTopScorers should be empty after error, got %d entries", len(nm.wcTopScorers))
	}
}

// ── bracket unavailable feedback (Phase 2) ───────────────────────────────────

// TestHandleWCGroupsKeys_BWithNoKnockoutRounds verifies that pressing 'b' from
// the groups list view sets wcLastError and keeps the sub-view unchanged when
// there are no knockout rounds yet.
func TestHandleWCGroupsKeys_BWithNoKnockoutRounds(t *testing.T) {
	m := newWCTestModel()
	m.wcSubView = wcSubViewGroups
	// KnockoutRounds is nil (zero value from newWCTestModel)

	next, _ := m.handleWCGroupsKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})

	nm := next.(model)
	if nm.wcLastError == "" {
		t.Error("wcLastError should be set when KnockoutRounds is empty")
	}
	if nm.wcSubView == wcSubViewBracket {
		t.Error("wcSubView should not switch to bracket when KnockoutRounds is empty")
	}
}

// TestHandleWCGroupsKeys_BWithKnockoutRounds verifies the inverse: when rounds
// exist, 'b' opens the bracket and leaves wcLastError empty. This pins both
// sides of the branch so a off-by-one on the guard condition fails.
func TestHandleWCGroupsKeys_BWithKnockoutRounds(t *testing.T) {
	m := newWCTestModel()
	m.wcSubView = wcSubViewGroups
	m.wcData.KnockoutRounds = []api.WCKnockoutRound{{Stage: "1/8"}}

	next, _ := m.handleWCGroupsKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})

	nm := next.(model)
	if nm.wcLastError != "" {
		t.Errorf("wcLastError = %q, want empty when rounds present", nm.wcLastError)
	}
	if nm.wcSubView != wcSubViewBracket {
		t.Errorf("wcSubView = %v, want wcSubViewBracket", nm.wcSubView)
	}
}

// TestHandleWCGroupGridKeys_BWithNoKnockoutRounds verifies that pressing 'b'
// from the grid view navigates to the groups list (where the error is rendered)
// and sets wcLastError when there are no knockout rounds.
func TestHandleWCGroupGridKeys_BWithNoKnockoutRounds(t *testing.T) {
	m := newWCTestModel()
	// KnockoutRounds is nil (zero value)

	next, _ := m.handleWCGroupGridKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})

	nm := next.(model)
	if nm.wcLastError == "" {
		t.Error("wcLastError should be set when KnockoutRounds is empty")
	}
	if nm.wcSubView != wcSubViewGroups {
		t.Errorf("wcSubView = %v, want wcSubViewGroups (so error is visible)", nm.wcSubView)
	}
}

// TestHandleWCGroupGridKeys_BWithKnockoutRounds verifies the inverse.
func TestHandleWCGroupGridKeys_BWithKnockoutRounds(t *testing.T) {
	m := newWCTestModel()
	m.wcData.KnockoutRounds = []api.WCKnockoutRound{{Stage: "1/8"}}

	next, _ := m.handleWCGroupGridKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})

	nm := next.(model)
	if nm.wcLastError != "" {
		t.Errorf("wcLastError = %q, want empty when rounds present", nm.wcLastError)
	}
	if nm.wcSubView != wcSubViewBracket {
		t.Errorf("wcSubView = %v, want wcSubViewBracket", nm.wcSubView)
	}
}
