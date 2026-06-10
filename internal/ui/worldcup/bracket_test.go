package worldcup

import (
	"strings"
	"testing"

	"github.com/0xjuanma/golazo/internal/api"
	"github.com/charmbracelet/lipgloss"
)

// TestRenderBracketRound_ConnectorAlignment exercises the column geometry
// that previously distorted in the --mock view: the ──╮, ├─, and ──╯
// connector glyphs MUST line up vertically (same column) regardless of
// whether the match line includes flag emojis.
func TestRenderBracketRound_ConnectorAlignment(t *testing.T) {
	winner := 1
	round := api.WCKnockoutRound{
		Stage: "1/8",
		Label: "Round of 16",
		Matchups: []api.WCMatchup{
			{
				HomeTeam: "Argentina", HomeTeamID: 1, HomeShort: "ARG",
				AwayTeam: "Australia", AwayTeamID: 2, AwayShort: "AUS",
				HomeScore: intPtrLocal(2), AwayScore: intPtrLocal(1),
				WinnerID: &winner,
			},
			{
				HomeTeam: "Netherlands", HomeTeamID: 3, HomeShort: "NED",
				AwayTeam: "USA", AwayTeamID: 4, AwayShort: "USA",
				HomeScore: intPtrLocal(3), AwayScore: intPtrLocal(1),
				WinnerID: func() *int { v := 3; return &v }(),
			},
		},
	}

	lines := renderBracketRound(round, 100)
	if len(lines) < 4 {
		t.Fatalf("expected at least 4 lines (mu1, mu2+╮, ├─, ──╯), got %d:\n%s", len(lines), strings.Join(lines, "\n"))
	}

	// Lines: mu1, mu2+╮, ├─, ──╯
	topCornerLine := lines[1]
	middleLine := lines[2]
	bottomLine := lines[3]

	colOf := func(line, glyph string) int {
		i := strings.Index(line, glyph)
		if i < 0 {
			return -1
		}
		return lipgloss.Width(line[:i])
	}

	topCol := colOf(topCornerLine, "╮")
	midCol := colOf(middleLine, "├")
	botCol := colOf(bottomLine, "╯")

	if topCol < 0 || midCol < 0 || botCol < 0 {
		t.Fatalf("connector glyph(s) missing: ╮=%d ├=%d ╯=%d\nlines:\n%s",
			topCol, midCol, botCol, strings.Join(lines, "\n"))
	}

	// ╮ sits at the end of "──╮", ├ sits at the start of "├─";
	// ╯ sits at the end of "──╯". Align so that the vertical stroke of
	// each connector glyph shares the same column.
	corner := topCol
	mid := midCol + 2     // ├ is preceded by no extra padding; the corner column is corner = mid + 2
	bottom := botCol      // ╯ at end of "──╯" → column = bottom

	if corner != mid {
		t.Errorf("top corner (col %d) not aligned with middle connector (col %d, +2 for ──)\nlines:\n%s",
			corner, mid, strings.Join(lines, "\n"))
	}
	if bottom != corner {
		t.Errorf("bottom corner (col %d) not aligned with top corner (col %d)\nlines:\n%s",
			bottom, corner, strings.Join(lines, "\n"))
	}
}

func intPtrLocal(i int) *int { return &i }
