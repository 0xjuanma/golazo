package worldcup

import (
	"strings"
	"testing"

	"github.com/0xjuanma/golazo/internal/api"
	"github.com/charmbracelet/lipgloss"
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
	cell := renderGroupGridCell(sampleGroup(), 30, false)

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

// TestRenderGroupGridCell_WidthInvariant locks in the fix for #158: every
// team row inside a grid cell must occupy the same visual width regardless
// of whether the label uses a regional-indicator flag, a tag-sequence
// subdivision flag, or a placeholder. Without the width pin in
// renderGroupGridCell, the v0.27.0 flag/name-override backfill caused rows
// to drift in terminals whose width metric disagrees with lipgloss.
func TestRenderGroupGridCell_WidthInvariant(t *testing.T) {
	mixed := api.WCGroup{
		ID: 1, Letter: "M", Name: "Group M",
		Teams: []api.LeagueTableEntry{
			// Regional-indicator pair
			{Position: 1, Team: api.Team{Name: "Mexico", ShortName: "MEX"}, Points: 9},
			// Tag-sequence subdivision flag
			{Position: 2, Team: api.Team{Name: "England", ShortName: "ENG"}, Points: 6},
			// Override-fallback (ambiguous SOU → KOR)
			{Position: 3, Team: api.Team{Name: "South Korea", ShortName: "SOU"}, Points: 4},
			// No-flag placeholder
			{Position: 4, Team: api.Team{Name: "Nowhereland", ShortName: "ZZZ"}, Points: 0},
		},
	}

	cell := renderGroupGridCell(mixed, 30, false)
	rows := strings.Split(cell, "\n")
	if len(rows) < 5 {
		t.Fatalf("expected header + 4 team rows, got %d lines:\n%s", len(rows), cell)
	}

	teamRows := rows[1:]
	wantW := lipgloss.Width(teamRows[0])
	for i, row := range teamRows {
		if w := lipgloss.Width(row); w != wantW {
			t.Errorf("row %d width mismatch: got %d, want %d\nrow=%q\nfull cell:\n%s",
				i, w, wantW, row, cell)
		}
	}
}

// TestPadToHeight covers the fix-screen-leak helper from #158: every WC
// render now passes through padToHeight so the returned frame matches the
// terminal height exactly. Bubbletea's diffing model can leave residue from
// a previous, taller view at the bottom of a shorter frame; pinning frame
// height removes that whole class of bug.
func TestPadToHeight(t *testing.T) {
	tests := []struct {
		name      string
		in        string
		height    int
		wantLines int
	}{
		{"pads short to height", "a\nb", 5, 5},
		{"truncates tall to height", "a\nb\nc\nd\ne\nf", 3, 3},
		{"matches exact height", "a\nb\nc", 3, 3},
		{"zero height returns input", "a\nb", 0, 2},
		{"negative height returns input", "a\nb", -1, 2},
		{"empty input pads to height", "", 4, 4},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := padToHeight(tc.in, tc.height)
			gotLines := strings.Count(got, "\n") + 1
			if gotLines != tc.wantLines {
				t.Errorf("padToHeight(%q, %d) produced %d lines, want %d",
					tc.in, tc.height, gotLines, tc.wantLines)
			}
		})
	}
}

func TestRenderGroupGrid_FrameHeightInvariant(t *testing.T) {
	mkGroup := func(letter string) api.WCGroup {
		return api.WCGroup{
			ID: int(letter[0]), Letter: letter, Name: "Group " + letter,
			Teams: []api.LeagueTableEntry{
				{Position: 1, Team: api.Team{Name: "Team", ShortName: "TM" + letter}, Points: 0},
				{Position: 2, Team: api.Team{Name: "Team", ShortName: "TM" + letter}, Points: 0},
				{Position: 3, Team: api.Team{Name: "Team", ShortName: "TM" + letter}, Points: 0},
				{Position: 4, Team: api.Team{Name: "Team", ShortName: "TM" + letter}, Points: 0},
			},
		}
	}
	wc := &api.WorldCupData{Name: "Test", Groups: []api.WCGroup{mkGroup("A"), mkGroup("B"), mkGroup("C"), mkGroup("D")}}

	const wantH = 40
	out := RenderGroupGrid(120, wantH, wc, 0, "")
	got := strings.Count(out, "\n") + 1
	if got != wantH {
		t.Errorf("RenderGroupGrid produced %d lines, want %d", got, wantH)
	}

	out2 := RenderGroupDetail(120, wantH, wc, 0, "")
	got2 := strings.Count(out2, "\n") + 1
	if got2 != wantH {
		t.Errorf("RenderGroupDetail produced %d lines, want %d", got2, wantH)
	}
}

// TestRenderGroupGridCell_FixedHeight locks in the fix for the row-4 gap
// reported on macOS Terminal under #158: every cell must render at exactly
// 1 + gridCellTeamRows lines regardless of how many teams FotMob ships.
// Without this, JoinHorizontal pads shorter cells with whitespace that
// surfaces as vertical drift in the bordered grid layout.
func TestRenderGroupGridCell_FixedHeight(t *testing.T) {
	short := api.WCGroup{
		ID: 1, Letter: "S", Name: "Group S",
		Teams: []api.LeagueTableEntry{
			{Position: 1, Team: api.Team{Name: "Mexico", ShortName: "MEX"}, Points: 3},
			{Position: 2, Team: api.Team{Name: "Brazil", ShortName: "BRA"}, Points: 1},
		},
	}
	long := api.WCGroup{
		ID: 2, Letter: "L", Name: "Group L",
		Teams: []api.LeagueTableEntry{
			{Position: 1, Team: api.Team{Name: "Mexico", ShortName: "MEX"}, Points: 9},
			{Position: 2, Team: api.Team{Name: "England", ShortName: "ENG"}, Points: 6},
			{Position: 3, Team: api.Team{Name: "South Korea", ShortName: "SOU"}, Points: 4},
			{Position: 4, Team: api.Team{Name: "Nowhereland", ShortName: "ZZZ"}, Points: 1},
			{Position: 5, Team: api.Team{Name: "Extra", ShortName: "EXT"}, Points: 0},
			{Position: 6, Team: api.Team{Name: "More", ShortName: "MOR"}, Points: 0},
		},
	}

	const wantLines = 1 + gridCellTeamRows
	for _, tc := range []struct {
		name string
		g    api.WCGroup
	}{{"short", short}, {"long", long}} {
		t.Run(tc.name, func(t *testing.T) {
			cell := renderGroupGridCell(tc.g, 30, false)
			got := strings.Count(cell, "\n") + 1
			if got != wantLines {
				t.Errorf("renderGroupGridCell(%s) produced %d lines, want %d\ncell:\n%s",
					tc.name, got, wantLines, cell)
			}
		})
	}
}

// TestRenderGroupGrid_RowHeightInvariant covers the row-4 gap reported on
// macOS Terminal under #158: every visual row of cells must occupy the same
// number of lines, even when groups in different rows carry different team
// counts. This is the property that prevents the bordered grid from drifting
// vertically when FotMob ships unusual qualifier shapes.
func TestRenderGroupGrid_RowHeightInvariant(t *testing.T) {
	mkGroup := func(letter string, n int) api.WCGroup {
		teams := make([]api.LeagueTableEntry, n)
		for i := range teams {
			teams[i] = api.LeagueTableEntry{
				Position: i + 1,
				Team:     api.Team{Name: "Team" + letter, ShortName: "TM" + letter},
				Points:   0,
			}
		}
		return api.WCGroup{ID: 100 + int(letter[0]), Letter: letter, Name: "Group " + letter, Teams: teams}
	}

	letters := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L"}
	groups := make([]api.WCGroup, len(letters))
	for i, l := range letters {
		// Simulate FotMob shipping an extra placeholder team in the last
		// row of groups, which is the pattern that triggered the report.
		n := 4
		if i >= 9 {
			n = 5
		}
		groups[i] = mkGroup(l, n)
	}
	wc := &api.WorldCupData{Name: "Test WC", Groups: groups}

	out := RenderGroupGrid(120, 200, wc, 0, "")

	// Find the rendered group titles for the first cell of each visual
	// row (A, D, G, J). With the borderless layout each cell starts on
	// its own line; verify those start lines are evenly spaced so the row
	// height is uniform regardless of differing per-group team counts.
	lines := strings.Split(out, "\n")
	var titleLineIdx []int
	for _, key := range []string{"Group A", "Group D", "Group G", "Group J"} {
		for i, line := range lines {
			if strings.Contains(line, key) {
				titleLineIdx = append(titleLineIdx, i)
				break
			}
		}
	}
	if len(titleLineIdx) != 4 {
		t.Fatalf("expected 4 row-start title lines, found %d at %v\noutput:\n%s",
			len(titleLineIdx), titleLineIdx, out)
	}
	gap := titleLineIdx[1] - titleLineIdx[0]
	for i := 2; i < len(titleLineIdx); i++ {
		if titleLineIdx[i]-titleLineIdx[i-1] != gap {
			t.Errorf("grid row spacing not uniform: gaps = %v (titles at %v)",
				diffs(titleLineIdx), titleLineIdx)
		}
	}
}

func diffs(in []int) []int {
	out := make([]int, 0, len(in)-1)
	for i := 1; i < len(in); i++ {
		out = append(out, in[i]-in[i-1])
	}
	return out
}
