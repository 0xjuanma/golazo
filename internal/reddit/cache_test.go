package reddit

import (
	"testing"
	"time"
)

// TestGetNotFoundTTL exercises the TTL comparison inside Get() for not-found
// markers (cache.go: time.Since(link.FetchedAt) > NotFoundTTL). A marker just
// under NotFoundTTL must be returned (so we don't re-search), and a marker
// just past it must be reported as expired (return nil so a retry is allowed).
// This is the function where the TTL behavior actually lives — bumping the
// constant without verifying the comparison would not catch a regression.
func TestGetNotFoundTTL(t *testing.T) {
	if NotFoundTTL != time.Hour {
		t.Fatalf("NotFoundTTL = %v, want 1h", NotFoundTTL)
	}

	cache := &GoalLinkCache{
		links: make(map[string]GoalLink),
	}

	const matchID = 12345
	const minute = 42
	key := GoalLinkKey{MatchID: matchID, Minute: minute}

	cases := []struct {
		name      string
		fetchedAt time.Time
		wantFound bool
	}{
		{
			name:      "just under TTL returns marker",
			fetchedAt: time.Now().Add(-(NotFoundTTL - time.Minute)),
			wantFound: true,
		},
		{
			name:      "just past TTL returns nil (allows retry)",
			fetchedAt: time.Now().Add(-(NotFoundTTL + time.Minute)),
			wantFound: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cache.links[makeKey(key)] = GoalLink{
				MatchID:   matchID,
				Minute:    minute,
				URL:       NotFoundMarker,
				FetchedAt: tc.fetchedAt,
			}

			got := cache.Get(key)
			if tc.wantFound {
				if got == nil {
					t.Fatalf("Get returned nil for marker aged %v (TTL %v); expected marker", time.Since(tc.fetchedAt), NotFoundTTL)
				}
				if !IsNotFound(got) {
					t.Fatalf("Get returned non-marker entry: %+v", got)
				}
				return
			}
			if got != nil {
				t.Fatalf("Get returned %+v for marker aged %v (TTL %v); expected nil", got, time.Since(tc.fetchedAt), NotFoundTTL)
			}
		})
	}
}
