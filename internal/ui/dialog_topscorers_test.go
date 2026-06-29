package ui

import (
	"strings"
	"testing"

	"github.com/0xjuanma/golazo/internal/api"
	tea "github.com/charmbracelet/bubbletea"
)

func stubTopScorers() []api.WCTopScorer {
	return []api.WCTopScorer{
		{PlayerName: "Lionel Messi", Team: "Argentina", Goals: 6},
		{PlayerName: "Kylian Mbappé", Team: "France", Goals: 4},
		{PlayerName: "Olivier Giroud", Team: "France", Goals: 3},
	}
}

func TestTopScorersDialog_ID(t *testing.T) {
	d := NewTopScorersDialog(nil)
	if d.ID() != TopScorersDialogID {
		t.Errorf("ID() = %q, want %q", d.ID(), TopScorersDialogID)
	}
}

func TestTopScorersDialog_ViewContainsPlayerName(t *testing.T) {
	d := NewTopScorersDialog(stubTopScorers())
	out := d.View(120, 40)
	if !strings.Contains(out, "Lionel Messi") {
		t.Error("View() output missing player name 'Lionel Messi'")
	}
	if !strings.Contains(out, "Argentina") {
		t.Error("View() output missing team 'Argentina'")
	}
	if !strings.Contains(out, "6") {
		t.Error("View() output missing goal count '6'")
	}
}

func TestTopScorersDialog_ViewContainsTitle(t *testing.T) {
	d := NewTopScorersDialog(stubTopScorers())
	out := d.View(120, 40)
	if !strings.Contains(out, "Top Scorers") {
		t.Error("View() output missing title 'Top Scorers'")
	}
}

func TestTopScorersDialog_CloseOnEsc(t *testing.T) {
	d := NewTopScorersDialog(stubTopScorers())
	_, action := d.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if _, ok := action.(DialogActionClose); !ok {
		t.Errorf("Update(esc) action = %T, want DialogActionClose", action)
	}
}

func TestTopScorersDialog_CloseOnS(t *testing.T) {
	d := NewTopScorersDialog(stubTopScorers())
	_, action := d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	if _, ok := action.(DialogActionClose); !ok {
		t.Errorf("Update('s') action = %T, want DialogActionClose", action)
	}
}

func TestTopScorersDialog_ScrollDown(t *testing.T) {
	scorers := make([]api.WCTopScorer, 20)
	for i := range scorers {
		scorers[i] = api.WCTopScorer{PlayerName: "Player", Team: "Team", Goals: 20 - i}
	}
	d := NewTopScorersDialog(scorers)
	d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if d.scrollIndex != 1 {
		t.Errorf("scrollIndex = %d after j, want 1", d.scrollIndex)
	}
}

func TestTopScorersDialog_EmptyScorers(t *testing.T) {
	d := NewTopScorersDialog(nil)
	out := d.View(120, 40)
	if !strings.Contains(out, "No scorer data") {
		t.Error("View() with nil scorers should show empty state message")
	}
}
