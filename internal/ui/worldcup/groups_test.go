package worldcup

import (
	"strings"
	"testing"

	"github.com/0xjuanma/golazo/internal/api"
)

func sampleGroup() api.WCGroup {
	return api.WCGroup{
		ID: 868712, Letter: "A", Name: "Group A",
		Teams: []api.LeagueTableEntry{
			{Position: 1, Team: api.Team{Name: "Argentina", ShortName: "ARG"}, Points: 7},
			{Position: 2, Team: api.Team{Name: "France", ShortName: "FRA"}, Points: 6},
			{Position: 3, Team: api.Team{Name: "Brazil", ShortName: "BRA"}, Points: 4},
			{Position: 4, Team: api.Team{Name: "Nowhereland", ShortName: "ZZZ"}, Points: 0},
		},
	}
}

func TestRenderGroupGridCell_EmojiAdjacentToName(t *testing.T) {
	cell := renderGroupGridCell(sampleGroup(), 30)

	argFlag := FlagEmoji("ARG")
	fraFlag := FlagEmoji("FRA")
	braFlag := FlagEmoji("BRA")
	if argFlag == "" || fraFlag == "" || braFlag == "" {
		t.Fatalf("expected ARG/FRA/BRA to have flag emojis registered")
	}

	for _, expect := range []string{
		argFlag + " ARG",
		fraFlag + " FRA",
		braFlag + " BRA",
	} {
		if !strings.Contains(cell, expect) {
			t.Errorf("expected grid cell to contain literal %q, got:\n%s", expect, cell)
		}
	}

	// Unmapped short still appears as the bare code.
	if !strings.Contains(cell, "ZZZ") {
		t.Errorf("expected unmapped short code ZZZ in cell, got:\n%s", cell)
	}
}

func TestRenderGroupStandingsTable_EmojiAdjacentToName(t *testing.T) {
	table := renderGroupStandingsTable(sampleGroup(), 80)

	argFlag := FlagEmoji("ARG")
	fraFlag := FlagEmoji("FRA")
	braFlag := FlagEmoji("BRA")

	// Standings table now renders the consistent "<flag> CODE" label used by
	// every World Cup view.
	for _, expect := range []string{
		argFlag + " ARG",
		fraFlag + " FRA",
		braFlag + " BRA",
	} {
		if !strings.Contains(table, expect) {
			t.Errorf("expected standings table to contain literal %q, got:\n%s", expect, table)
		}
	}

	// Unmapped short code still appears as the bare code (no flag).
	if !strings.Contains(table, "ZZZ") {
		t.Errorf("expected unmapped short code ZZZ in table, got:\n%s", table)
	}
}
