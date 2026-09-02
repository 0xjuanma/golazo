package ui

import (
	"strings"
	"testing"

	"github.com/0xjuanma/golazo/internal/api"
	"github.com/0xjuanma/golazo/internal/constants"
	"github.com/charmbracelet/lipgloss"
)

func TestRenderFIFARanking(t *testing.T) {
	ranking := &api.FIFARanking{
		Gender: "women",
		Period: api.FIFARankingPeriod{Name: "21.04.2026"},
		Teams: []api.FIFARankingEntry{
			{Rank: 1, Name: "Spain", TotalPoints: 2083.12, PointsDiff: 12.5},
		},
	}
	out := RenderFIFARanking(100, 30, ranking, "women", false, "", 0, 0, constants.StatusBannerNone)
	for _, want := range []string{"FIFA World Ranking", "Men", "Women", "Released 21.04.2026", "Spain", "2083.12", "+12.50"} {
		if !strings.Contains(out, want) {
			t.Errorf("output does not contain %q", want)
		}
	}
}

func TestRenderFIFARankingError(t *testing.T) {
	out := RenderFIFARanking(80, 24, nil, "men", false, constants.ErrorLoadFailed, 0, 0, constants.StatusBannerNone)
	if !strings.Contains(out, constants.ErrorLoadFailed) || !strings.Contains(out, constants.ErrorRetryHint) {
		t.Fatal("error state and retry hint should be visible")
	}
}

func TestRenderFIFARankingNarrowTerminalKeepsHeaderAndFirstTeam(t *testing.T) {
	teams := make([]api.FIFARankingEntry, 20)
	for i := range teams {
		teams[i] = api.FIFARankingEntry{
			Rank: i + 1, Name: "Country with a long name",
			TotalPoints: 2000 - float64(i), PointsDiff: float64(i),
		}
	}
	teams[0].Name = "Spain"
	teams[12].Name = "Croatia"
	ranking := &api.FIFARanking{
		Gender: "men",
		Period: api.FIFARankingPeriod{Name: "20.07.2026"},
		Teams:  teams,
	}

	out := RenderFIFARanking(46, 24, ranking, "men", false, "", 0, 0, constants.StatusBannerNone)
	for _, want := range []string{"FIFA World Ranking", "Men", "Women", "Rank", "Country", "Points", "Spain"} {
		if !strings.Contains(out, want) {
			t.Errorf("narrow output does not contain %q", want)
		}
	}
	if strings.Index(out, "Croatia") >= 0 && strings.Index(out, "Croatia") < strings.Index(out, "Spain") {
		t.Fatal("ranking was clipped from the top")
	}
}

func TestRenderFIFARankingColumnHeadersShareOneLine(t *testing.T) {
	ranking := &api.FIFARanking{
		Gender: "men",
		Period: api.FIFARankingPeriod{Name: "20.07.2026"},
		Teams:  []api.FIFARankingEntry{{Rank: 1, Name: "Spain", TotalPoints: 1996, PointsDiff: 121}},
	}
	out := RenderFIFARanking(100, 30, ranking, "men", false, "", 0, 0, constants.StatusBannerNone)
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "Rank ") {
			for _, header := range []string{"Country", "Points", "+/-"} {
				if !strings.Contains(line, header) {
					t.Fatalf("header %q is not on the Rank header line", header)
				}
			}
			return
		}
	}
	t.Fatal("Rank header line not found")
}

func TestFormatRankingRowNeverExceedsWidth(t *testing.T) {
	for _, width := range []int{18, 28, 42, 70} {
		row := formatRankingRow(width, "211", "Saint Vincent and The Grenadines", "1234.56", "+12.34")
		if got := lipgloss.Width(row); got > width {
			t.Errorf("formatRankingRow(%d) width = %d", width, got)
		}
	}
}
