package fotmob

import (
	"testing"
	"time"
)

// edtZone returns a fixed EDT zone for deterministic local-day tests.
func edtZone() *time.Location {
	return time.FixedZone("EDT", -4*3600)
}

func TestClassifyLeagueMatches_LiveAcrossUTCBoundary(t *testing.T) {
	// Saudi Arabia vs Uruguay: kickoff 2026-06-15T22:00Z, user clock is
	// 2026-06-16 00:30 UTC (= 2026-06-15 20:30 EDT). UTC "today" has rolled
	// to June 16, but the match is still in progress on UTC June 15. The
	// status-only filter must keep it.
	matches := []fotmobMatch{
		{
			ID: "1",
			Status: status{
				UTCTime:  "2026-06-15T22:00:00.000Z",
				Started:  boolPtr(true),
				Finished: boolPtr(false),
			},
			Home: team{ID: "100", Name: "Saudi Arabia", ShortName: "KSA"},
			Away: team{ID: "200", Name: "Uruguay", ShortName: "URU"},
		},
	}
	leagueInfo := league{ID: 77, Name: "FIFA World Cup"}

	now := time.Date(2026, 6, 15, 20, 30, 0, 0, edtZone())
	live, upcoming := classifyLeagueMatches(matches, leagueInfo, now)

	if len(live) != 1 {
		t.Fatalf("live len = %d, want 1", len(live))
	}
	if len(upcoming) != 0 {
		t.Fatalf("upcoming len = %d, want 0", len(upcoming))
	}
	if live[0].HomeTeam.Name != "Saudi Arabia" {
		t.Errorf("home = %q, want Saudi Arabia", live[0].HomeTeam.Name)
	}
	if live[0].League.ID != 77 {
		t.Errorf("league ID = %d, want 77 (filled from leagueInfo)", live[0].League.ID)
	}
}

func TestClassifyLeagueMatches_UpcomingGatedByLocalToday(t *testing.T) {
	// Two not-started matches:
	//   - kickoff 2026-06-16T00:30Z (= 2026-06-15 20:30 EDT) → local-today (keep)
	//   - kickoff 2026-06-16T20:00Z (= 2026-06-16 16:00 EDT) → tomorrow local (drop)
	matches := []fotmobMatch{
		{
			ID: "1",
			Status: status{
				UTCTime:  "2026-06-16T00:30:00.000Z",
				Started:  boolPtr(false),
				Finished: boolPtr(false),
			},
		},
		{
			ID: "2",
			Status: status{
				UTCTime:  "2026-06-16T20:00:00.000Z",
				Started:  boolPtr(false),
				Finished: boolPtr(false),
			},
		},
	}

	// User clock: 2026-06-15 18:00 EDT.
	now := time.Date(2026, 6, 15, 18, 0, 0, 0, edtZone())
	live, upcoming := classifyLeagueMatches(matches, league{ID: 77}, now)

	if len(live) != 0 {
		t.Fatalf("live len = %d, want 0", len(live))
	}
	if len(upcoming) != 1 {
		t.Fatalf("upcoming len = %d, want 1", len(upcoming))
	}
	if upcoming[0].ID != 1 {
		t.Errorf("upcoming ID = %d, want 1", upcoming[0].ID)
	}
}

func TestClassifyLeagueMatches_CancelledExcluded(t *testing.T) {
	matches := []fotmobMatch{
		{
			ID: "1",
			Status: status{
				UTCTime:   "2026-06-15T22:00:00.000Z",
				Started:   boolPtr(true),
				Finished:  boolPtr(false),
				Cancelled: boolPtr(true),
			},
		},
	}
	now := time.Date(2026, 6, 15, 20, 30, 0, 0, edtZone())
	live, upcoming := classifyLeagueMatches(matches, league{ID: 77}, now)

	if len(live) != 0 {
		t.Errorf("live len = %d, want 0 (cancelled match)", len(live))
	}
	if len(upcoming) != 0 {
		t.Errorf("upcoming len = %d, want 0 (cancelled match)", len(upcoming))
	}
}

func TestClassifyLeagueMatches_FinishedExcluded(t *testing.T) {
	matches := []fotmobMatch{
		{
			ID: "1",
			Status: status{
				UTCTime:  "2026-06-15T18:00:00.000Z",
				Started:  boolPtr(true),
				Finished: boolPtr(true),
			},
		},
	}
	now := time.Date(2026, 6, 15, 20, 30, 0, 0, edtZone())
	live, upcoming := classifyLeagueMatches(matches, league{ID: 77}, now)

	if len(live) != 0 {
		t.Errorf("live len = %d, want 0 (finished match)", len(live))
	}
	if len(upcoming) != 0 {
		t.Errorf("upcoming len = %d, want 0", len(upcoming))
	}
}

func TestClassifyLeagueMatches_MissingUTCTimeSkipped(t *testing.T) {
	matches := []fotmobMatch{
		{
			ID: "1",
			Status: status{
				Started:  boolPtr(true),
				Finished: boolPtr(false),
			},
		},
	}
	now := time.Date(2026, 6, 15, 20, 30, 0, 0, edtZone())
	live, upcoming := classifyLeagueMatches(matches, league{ID: 77}, now)
	if len(live) != 0 || len(upcoming) != 0 {
		t.Errorf("live=%d upcoming=%d, want both 0 (empty utcTime should be skipped)", len(live), len(upcoming))
	}
}
