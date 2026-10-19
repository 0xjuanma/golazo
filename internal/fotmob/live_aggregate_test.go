package fotmob

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/0xjuanma/golazo/internal/ratelimit"
)

// liveAggregateTestClient builds a Client whose league-page HTTP responses are
// produced entirely in-process by respond. Nothing here can reach fotmob.com:
// the client's transport is replaced, and the transport rejects any request
// whose path is not /leagues/{id}.
//
// The empty-results cache is intentionally left nil (as in other package
// tests) so a league fetch failure is never masked by a persisted entry.
func liveAggregateTestClient(t *testing.T, respond func(leagueID int) (*http.Response, error)) *Client {
	t.Helper()

	transport := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		leagueID, err := leagueIDFromPath(req.URL.Path)
		if err != nil {
			return nil, err
		}
		return respond(leagueID)
	})

	return &Client{
		httpClient:    &http.Client{Transport: transport, Timeout: 5 * time.Second},
		baseURL:       baseURL,
		rateLimiter:   ratelimit.New(0), // no delay in tests
		cache:         NewResponseCache(DefaultCacheConfig()),
		pageURLs:      make(map[int]string, 10),
		maxConcurrent: make(chan struct{}, 10),
	}
}

// leagueIDFromPath extracts {id} from /leagues/{id}.
func leagueIDFromPath(path string) (int, error) {
	const prefix = "/leagues/"
	if !strings.HasPrefix(path, prefix) {
		return 0, fmt.Errorf("unexpected path %q", path)
	}
	raw := strings.TrimPrefix(path, prefix)
	if idx := strings.IndexAny(raw, "/?#"); idx != -1 {
		raw = raw[:idx]
	}
	id, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("parse league id from path %q: %w", path, err)
	}
	return id, nil
}

// liveLeaguePageHTML renders a league page whose __NEXT_DATA__ payload matches
// the shape fetchLeaguePage expects.
func liveLeaguePageHTML(t *testing.T, leagueID int, leagueName string, matches []fotmobMatch) string {
	t.Helper()

	if matches == nil {
		matches = []fotmobMatch{}
	}
	pageProps := map[string]any{
		"details": map[string]any{
			"id":      leagueID,
			"name":    leagueName,
			"country": "England",
		},
		"fixtures": map[string]any{"allMatches": matches},
	}
	payload, err := json.Marshal(map[string]any{"props": map[string]any{"pageProps": pageProps}})
	if err != nil {
		t.Fatalf("marshal league page fixture: %v", err)
	}
	return `<html><body><script id="__NEXT_DATA__" type="application/json">` + string(payload) + `</script></body></html>`
}

// liveMatchFixture returns a FotMob match payload that toAPIMatch classifies as
// live (started, not finished, not cancelled).
func liveMatchFixture(id int, home, away string) fotmobMatch {
	return fotmobMatch{
		ID:   strconv.Itoa(id),
		Home: team{ID: "1", Name: home, ShortName: home},
		Away: team{ID: "2", Name: away, ShortName: away},
		Status: status{
			UTCTime:  time.Now().UTC().Format(time.RFC3339),
			Started:  boolPtr(true),
			Finished: boolPtr(false),
		},
	}
}

func okPage(t *testing.T, leagueID int, matches []fotmobMatch) *http.Response {
	t.Helper()
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(liveLeaguePageHTML(t, leagueID, "Test League", matches))),
	}
}

func requireActiveLeagues(t *testing.T) []int {
	t.Helper()
	leagues := ActiveLeagues()
	if len(leagues) == 0 {
		t.Skip("no active leagues configured; aggregate behavior cannot be exercised")
	}
	return leagues
}

// A total upstream failure must be reported as an error, not as a successful
// empty result. Agents are told (docs/CLI.md, "No matches vs failure
// semantics") that `count: 0` means the request succeeded, so silently
// classifying every failed league fetch as "no live matches" is a contract
// violation: `golazo live` prints {"status":"ok","count":0,...} and exits 0
// while FotMob is down or blocking requests.
func TestLiveAndUpcoming_AllLeagueFetchesFailReturnsError(t *testing.T) {
	leagues := requireActiveLeagues(t)

	client := liveAggregateTestClient(t, func(leagueID int) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusInternalServerError,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader("upstream exploded")),
		}, nil
	})

	live, upcoming, err := client.LiveAndUpcoming(context.Background())
	if err == nil {
		t.Fatalf("LiveAndUpcoming returned nil error after all %d league fetches failed; live=%d upcoming=%d", len(leagues), len(live), len(upcoming))
	}
	if len(live) != 0 || len(upcoming) != 0 {
		t.Errorf("got %d live / %d upcoming matches alongside an error, want none", len(live), len(upcoming))
	}
}

// Best-effort aggregation must be preserved: one recoverable league is enough
// to return a successful result, and the failing leagues must not prevent the
// working one from being reported.
func TestLiveAndUpcoming_PartialFailureKeepsBestEffortResults(t *testing.T) {
	leagues := requireActiveLeagues(t)
	working := leagues[0]

	client := liveAggregateTestClient(t, func(leagueID int) (*http.Response, error) {
		if leagueID != working {
			return &http.Response{
				StatusCode: http.StatusBadGateway,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader("bad gateway")),
			}, nil
		}
		return okPage(t, leagueID, []fotmobMatch{liveMatchFixture(900001, "Home FC", "Away FC")}), nil
	})

	live, _, err := client.LiveAndUpcoming(context.Background())
	if err != nil {
		t.Fatalf("LiveAndUpcoming returned error despite one successful league: %v", err)
	}
	if len(live) != 1 {
		t.Fatalf("live len = %d, want 1 (the league that succeeded)", len(live))
	}
	if live[0].ID != 900001 {
		t.Errorf("live[0].ID = %d, want 900001", live[0].ID)
	}
}

// A genuinely empty day is a complete answer: every league responded, none had
// a live match, so the empty result must stay a nil-error success.
func TestLiveAndUpcoming_GenuineEmptyResultIsNotAnError(t *testing.T) {
	requireActiveLeagues(t)

	client := liveAggregateTestClient(t, func(leagueID int) (*http.Response, error) {
		return okPage(t, leagueID, nil), nil
	})

	live, upcoming, err := client.LiveAndUpcoming(context.Background())
	if err != nil {
		t.Fatalf("LiveAndUpcoming returned error for a genuinely empty result: %v", err)
	}
	if len(live) != 0 || len(upcoming) != 0 {
		t.Errorf("got %d live / %d upcoming matches, want none", len(live), len(upcoming))
	}
}
