package worldcup

import (
	"strings"
	"testing"
	"time"

	"github.com/0xjuanma/golazo/internal/api"
)

func TestRenderUpcoming_GroupsByDate(t *testing.T) {
	// Three matches across three distinct local dates, already sorted ascending.
	base := time.Date(2026, 6, 14, 18, 0, 0, 0, time.Local)
	m1 := base
	m2 := base.AddDate(0, 0, 1).Add(2 * time.Hour) // next day, later
	m3 := base.AddDate(0, 0, 2)                    // day after

	matches := []api.Match{
		{
			ID:        1,
			HomeTeam:  api.Team{Name: "Argentina", ShortName: "ARG"},
			AwayTeam:  api.Team{Name: "France", ShortName: "FRA"},
			MatchTime: &m1,
			Round:     "Group A",
		},
		{
			ID:        2,
			HomeTeam:  api.Team{Name: "England", ShortName: "ENG"},
			AwayTeam:  api.Team{Name: "USA", ShortName: "USA"},
			MatchTime: &m2,
		},
		{
			ID:        3,
			HomeTeam:  api.Team{Name: "Brazil", ShortName: "BRA"},
			AwayTeam:  api.Team{Name: "Spain", ShortName: "ESP"},
			MatchTime: &m3,
		},
	}

	out := RenderUpcoming(120, 40, matches, false, "", "")

	for _, expected := range []string{
		upcomingFormatDateHeader(m1),
		upcomingFormatDateHeader(m2),
		upcomingFormatDateHeader(m3),
		"ARG", "FRA", "ENG", "USA", "BRA", "ESP",
		"Group A",
		m1.Format("15:04"),
		m2.Format("15:04"),
		m3.Format("15:04"),
	} {
		if !strings.Contains(out, expected) {
			t.Errorf("expected output to contain %q, but it did not.\nOutput:\n%s", expected, out)
		}
	}

	// Each distinct date header must appear exactly once.
	for _, header := range []string{
		upcomingFormatDateHeader(m1),
		upcomingFormatDateHeader(m2),
		upcomingFormatDateHeader(m3),
	} {
		if c := strings.Count(out, header); c != 1 {
			t.Errorf("date header %q appears %d times, want 1", header, c)
		}
	}

	// Day 1 header must appear before day 2 header (ordering check).
	pos1 := strings.Index(out, upcomingFormatDateHeader(m1))
	pos2 := strings.Index(out, upcomingFormatDateHeader(m2))
	pos3 := strings.Index(out, upcomingFormatDateHeader(m3))
	if !(pos1 < pos2 && pos2 < pos3) {
		t.Errorf("date headers out of order: pos1=%d pos2=%d pos3=%d", pos1, pos2, pos3)
	}
}

func TestRenderUpcoming_SameDayMultipleMatches(t *testing.T) {
	// Two matches same local day, second later than the first.
	day := time.Date(2026, 6, 14, 0, 0, 0, 0, time.Local)
	early := day.Add(15 * time.Hour)
	late := day.Add(20 * time.Hour)

	matches := []api.Match{
		{ID: 1, HomeTeam: api.Team{ShortName: "ARG"}, AwayTeam: api.Team{ShortName: "FRA"}, MatchTime: &early},
		{ID: 2, HomeTeam: api.Team{ShortName: "ENG"}, AwayTeam: api.Team{ShortName: "USA"}, MatchTime: &late},
	}

	out := RenderUpcoming(120, 40, matches, false, "", "")

	// One date header only.
	header := upcomingFormatDateHeader(early)
	if c := strings.Count(out, header); c != 1 {
		t.Errorf("same-day header should appear once, appeared %d times", c)
	}

	// Earlier match must render before later match in the output.
	posARG := strings.Index(out, "ARG")
	posENG := strings.Index(out, "ENG")
	if posARG == -1 || posENG == -1 || posARG >= posENG {
		t.Errorf("matches not rendered in ascending kickoff order: ARG@%d ENG@%d", posARG, posENG)
	}
}

func TestRenderUpcoming_EmptyState(t *testing.T) {
	out := RenderUpcoming(80, 24, nil, false, "", "")
	if !strings.Contains(out, "No matches in the next 4 days") {
		t.Errorf("expected empty-state message, got:\n%s", out)
	}
}

func TestRenderUpcoming_Loading(t *testing.T) {
	out := RenderUpcoming(80, 24, nil, true, "", "")
	if !strings.Contains(out, "Loading upcoming matches") {
		t.Errorf("expected loading text, got:\n%s", out)
	}
}

func TestRenderUpcoming_Error(t *testing.T) {
	out := RenderUpcoming(80, 24, nil, false, "boom", "")
	if !strings.Contains(out, "boom") {
		t.Errorf("expected error text in output, got:\n%s", out)
	}
}

func TestRenderUpcoming_PrefixesTeamNamesWithFlagEmoji(t *testing.T) {
	kickoff := time.Date(2026, 6, 14, 18, 0, 0, 0, time.Local)
	matches := []api.Match{
		{
			ID:        1,
			HomeTeam:  api.Team{Name: "Argentina", ShortName: "ARG"},
			AwayTeam:  api.Team{Name: "France", ShortName: "FRA"},
			MatchTime: &kickoff,
		},
	}

	out := RenderUpcoming(120, 24, matches, false, "", "")

	argFlag := FlagEmoji("ARG")
	fraFlag := FlagEmoji("FRA")
	if argFlag == "" || fraFlag == "" {
		t.Fatalf("expected ARG/FRA to have flag emojis registered")
	}

	if !strings.Contains(out, argFlag+" ARG") {
		t.Errorf("expected home team to be prefixed with flag emoji (e.g. %q ARG), got:\n%s", argFlag, out)
	}
	if !strings.Contains(out, fraFlag+" FRA") {
		t.Errorf("expected away team to be prefixed with flag emoji (e.g. %q FRA), got:\n%s", fraFlag, out)
	}
}

func TestRenderUpcoming_FallsBackWhenFlagMissing(t *testing.T) {
	// A made-up short code with no flag mapping; output must contain the
	// short code without an "<emoji> " prefix.
	kickoff := time.Date(2026, 6, 14, 18, 0, 0, 0, time.Local)
	matches := []api.Match{
		{
			ID:        1,
			HomeTeam:  api.Team{Name: "Nowhereland", ShortName: "ZZZ"},
			AwayTeam:  api.Team{Name: "Otherland", ShortName: "QQQ"},
			MatchTime: &kickoff,
		},
	}

	out := RenderUpcoming(120, 24, matches, false, "", "")
	if !strings.Contains(out, "ZZZ") || !strings.Contains(out, "QQQ") {
		t.Errorf("expected short codes to still render when no flag, got:\n%s", out)
	}
}
