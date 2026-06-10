package app

import (
	"testing"
	"time"

	"github.com/0xjuanma/golazo/internal/api"
)

func TestSortAndDedupeWCUpcoming(t *testing.T) {
	mk := func(id int, offsetHours int) api.Match {
		t := time.Date(2026, 6, 14, 12, 0, 0, 0, time.UTC).Add(time.Duration(offsetHours) * time.Hour)
		return api.Match{ID: id, MatchTime: &t}
	}

	// Five matches across 3 days, intentionally out of order, plus one duplicate
	// of ID=2 and one match without MatchTime (must be dropped).
	in := []api.Match{
		mk(3, 48),       // day 3
		mk(1, 0),        // day 1
		mk(2, 24),       // day 2 (first occurrence)
		mk(2, 25),       // duplicate ID — must be dropped (first occurrence wins)
		mk(4, 26),       // day 2, later than #2
		{ID: 999},       // no MatchTime — must be dropped
		mk(5, 49),       // day 3, later than #3
	}

	out := sortAndDedupeWCUpcoming(in)

	if len(out) != 5 {
		t.Fatalf("expected 5 matches after dedup+drop, got %d", len(out))
	}

	wantIDs := []int{1, 2, 4, 3, 5}
	for i, m := range out {
		if m.ID != wantIDs[i] {
			t.Errorf("position %d: expected ID %d, got %d", i, wantIDs[i], m.ID)
		}
	}

	// Ensure ascending kickoff order
	for i := 1; i < len(out); i++ {
		if out[i].MatchTime.Before(*out[i-1].MatchTime) {
			t.Errorf("matches not sorted ascending at index %d", i)
		}
	}

	// Verify the kept occurrence of duplicate ID=2 is the first one (offset 24h),
	// not the second (offset 25h). Silent-break guard.
	for _, m := range out {
		if m.ID == 2 {
			expected := time.Date(2026, 6, 14, 12, 0, 0, 0, time.UTC).Add(24 * time.Hour)
			if !m.MatchTime.Equal(expected) {
				t.Errorf("dedup kept wrong occurrence: got %v, want %v", m.MatchTime, expected)
			}
		}
	}
}

func TestSortAndDedupeWCUpcoming_Empty(t *testing.T) {
	if got := sortAndDedupeWCUpcoming(nil); len(got) != 0 {
		t.Errorf("expected empty result for nil input, got %d", len(got))
	}
	if got := sortAndDedupeWCUpcoming([]api.Match{}); len(got) != 0 {
		t.Errorf("expected empty result for empty input, got %d", len(got))
	}
}
