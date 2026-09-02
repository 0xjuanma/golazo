package reddit

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"
)

// TestPublicJSONFetcherSearchTimestampWindow verifies that the Reddit search
// URL spans the matcher's accepted post-date window (-24h .. +48h around the
// match time) so candidate posts in that range are not pre-filtered out by
// the server-side timestamp query.
func TestPublicJSONFetcherSearchTimestampWindow(t *testing.T) {
	matchTime := time.Date(2025, 11, 10, 16, 0, 0, 0, time.UTC)
	wantStart := matchTime.Add(-24 * time.Hour).Unix()
	wantEnd := matchTime.Add(48 * time.Hour).Unix()

	var capturedURL *url.URL
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"data":{"children":[]}}`)
	}))
	defer server.Close()

	// We need the fetcher to hit the test server instead of www.reddit.com.
	// PublicJSONFetcher hardcodes the URL, so use a transport that rewrites
	// the host.
	fetcher := NewPublicJSONFetcher()
	fetcher.httpClient.Transport = &rewriteTransport{target: server.URL}

	_, err := fetcher.Search("australia turkey 27", 10, matchTime, "relevance")
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if capturedURL == nil {
		t.Fatal("test server did not receive a request")
	}

	q := capturedURL.Query().Get("q")
	want := fmt.Sprintf("timestamp:%d..%d", wantStart, wantEnd)
	if !strings.Contains(q, want) {
		t.Errorf("query missing timestamp window: q=%q, want substring %q", q, want)
	}
}

// rewriteTransport redirects any outbound request to the configured target
// host while preserving the original path and query string.
type rewriteTransport struct {
	target string
}

func (rt *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	targetURL, err := url.Parse(rt.target)
	if err != nil {
		return nil, err
	}
	req.URL.Scheme = targetURL.Scheme
	req.URL.Host = targetURL.Host
	req.Host = targetURL.Host
	return http.DefaultTransport.RoundTrip(req)
}

// TestPublicJSONFetcherSearchHonorsSortParam pins the sort param so future
// refactors don't silently change ranking behavior.
func TestPublicJSONFetcherSearchHonorsSortParam(t *testing.T) {
	var capturedURL *url.URL
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"data":{"children":[]}}`)
	}))
	defer server.Close()

	fetcher := NewPublicJSONFetcher()
	fetcher.httpClient.Transport = &rewriteTransport{target: server.URL}

	_, err := fetcher.Search("q", 10, time.Now(), "top")
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if got := capturedURL.Query().Get("sort"); got != "top" {
		t.Errorf("sort param: got %q, want %q", got, "top")
	}
	if got := capturedURL.Query().Get("limit"); got != strconv.Itoa(10) {
		t.Errorf("limit param: got %q, want %q", got, strconv.Itoa(10))
	}
}

// TestSearchReturnsErrBlockedOn403 pins the typed-error contract for HTTP 403
// responses from Reddit's edge. The queue worker introduced in the goal-link
// rework uses errors.Is(err, ErrBlocked) to enter cooldown; sniffing on
// response body substrings was the previous (fragile) detection mechanism.
func TestSearchReturnsErrBlockedOn403(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = io.WriteString(w, `<html>You've been blocked by network security.</html>`)
	}))
	defer server.Close()

	fetcher := NewPublicJSONFetcher()
	fetcher.httpClient.Transport = &rewriteTransport{target: server.URL}

	_, err := fetcher.Search("anything", 5, time.Now(), "relevance")
	if err == nil {
		t.Fatal("Search returned nil error for 403 response")
	}
	if !errors.Is(err, ErrBlocked) {
		t.Fatalf("Search error %v is not ErrBlocked", err)
	}
}

// recordingFetcher captures queries passed to Search so tests can assert on
// the exact query string searchForGoalOnce constructs (vs. e2e-asserting via
// httptest, which would re-test url-escaping). resultSets is indexed by call
// count — call N returns resultSets[N], or no results if N is out of range —
// so tests can program a first (full-query) call that finds nothing and a
// second (relaxed-retry) call that does.
type recordingFetcher struct {
	queries    []string
	resultSets [][]SearchResult
	err        error
}

func (r *recordingFetcher) Search(query string, _ int, _ time.Time, _ string) ([]SearchResult, error) {
	idx := len(r.queries)
	r.queries = append(r.queries, query)
	if r.err != nil {
		return nil, r.err
	}
	if idx < len(r.resultSets) {
		return r.resultSets[idx], nil
	}
	return nil, nil
}

// TestSearchForGoalOnceQueryFormat pins the query format produced by
// buildGoalQuery / searchForGoalOnce: "<home> <hScore> <aScore> <away>
// <scorerLast>", with the scorer token omitted when ScorerName is empty.
// Each scorer-present case programs a matching first-call result so it
// resolves without triggering the relaxed retry (that path is covered
// separately by TestSearchForGoalOnceFallbackOnNoMatch) — asserts exactly
// one fetcher.Search call.
func TestSearchForGoalOnceQueryFormat(t *testing.T) {
	cases := []struct {
		name        string
		goal        GoalInfo
		wantQuery   string
		firstResult []SearchResult // nil is fine when no retry is possible anyway (empty ScorerName)
	}{
		{
			name: "scorer present uses last token",
			goal: GoalInfo{
				MatchID:    4667791,
				HomeTeam:   "Iran",
				AwayTeam:   "New Zealand",
				HomeScore:  0,
				AwayScore:  1,
				ScorerName: "Elijah Just",
				Minute:     7,
				IsHomeTeam: false,
				MatchTime:  time.Now(),
			},
			wantQuery: "Iran 0 1 New Zealand Just",
			firstResult: []SearchResult{{
				Title:     "Iran 0 - [1] New Zealand - E. Just 7'",
				URL:       "https://example.com/iran-nz",
				PostURL:   "https://www.reddit.com/r/soccer/comments/iran-nz",
				CreatedAt: time.Now(),
			}},
		},
		{
			name: "empty scorer falls back to teams + score",
			goal: GoalInfo{
				MatchID:    1,
				HomeTeam:   "Iran",
				AwayTeam:   "New Zealand",
				HomeScore:  1,
				AwayScore:  1,
				ScorerName: "",
				Minute:     50,
				MatchTime:  time.Now(),
			},
			wantQuery: "Iran 1 1 New Zealand",
		},
		{
			name: "scorer with diacritics is folded",
			goal: GoalInfo{
				MatchID:    2,
				HomeTeam:   "Liverpool",
				AwayTeam:   "Wolves",
				HomeScore:  2,
				AwayScore:  0,
				ScorerName: "Darwin Núñez",
				Minute:     12,
				MatchTime:  time.Now(),
			},
			wantQuery: "Liverpool 2 0 Wolves Nunez",
			firstResult: []SearchResult{{
				Title:     "Liverpool [2] - 0 Wolves - Darwin Nunez 12'",
				URL:       "https://example.com/liv-wolves",
				PostURL:   "https://www.reddit.com/r/soccer/comments/liv-wolves",
				CreatedAt: time.Now(),
			}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fetcher := &recordingFetcher{resultSets: [][]SearchResult{tc.firstResult}}
			client := NewClientWithFetcher(fetcher, &GoalLinkCache{links: make(map[string]GoalLink)})

			_, err := client.searchForGoalOnce(tc.goal)
			if err != nil {
				t.Fatalf("searchForGoalOnce returned err: %v", err)
			}
			if len(fetcher.queries) != 1 {
				t.Fatalf("expected exactly 1 fetcher.Search call (matched on first attempt), got %d: %v",
					len(fetcher.queries), fetcher.queries)
			}
			if fetcher.queries[0] != tc.wantQuery {
				t.Errorf("query mismatch:\n  got  %q\n  want %q", fetcher.queries[0], tc.wantQuery)
			}
		})
	}
}

// TestSearchForGoalOnceFallbackOnNoMatch verifies that when the first (full)
// query finds no match and the goal has a scorer name, searchForGoalOnce
// retries once with the scorer token dropped and uses the retry's match.
func TestSearchForGoalOnceFallbackOnNoMatch(t *testing.T) {
	goal := GoalInfo{
		MatchID:    3,
		HomeTeam:   "Wolves",
		AwayTeam:   "West Ham",
		HomeScore:  3,
		AwayScore:  0,
		ScorerName: "Mateus Mane",
		Minute:     41,
		MatchTime:  time.Now(),
	}
	fetcher := &recordingFetcher{
		resultSets: [][]SearchResult{
			{}, // full query (with scorer token): no results
			{{ // relaxed retry (scorer token dropped): a match
				Title:     "Wolves [3] - 0 West Ham - Mateus Mane 41'",
				URL:       "https://example.com/relaxed-match",
				PostURL:   "https://www.reddit.com/r/soccer/comments/relaxed",
				CreatedAt: goal.MatchTime,
			}},
		},
	}
	client := NewClientWithFetcher(fetcher, &GoalLinkCache{links: make(map[string]GoalLink)})

	link, err := client.searchForGoalOnce(goal)
	if err != nil {
		t.Fatalf("searchForGoalOnce returned err: %v", err)
	}
	if len(fetcher.queries) != 2 {
		t.Fatalf("expected exactly 2 fetcher.Search calls (full + relaxed retry), got %d: %v",
			len(fetcher.queries), fetcher.queries)
	}
	if want := "Wolves 3 0 West Ham Mane"; fetcher.queries[0] != want {
		t.Errorf("first query = %q, want %q", fetcher.queries[0], want)
	}
	if want := "Wolves 3 0 West Ham"; fetcher.queries[1] != want {
		t.Errorf("retry query = %q, want %q (scorer token dropped)", fetcher.queries[1], want)
	}
	if link == nil {
		t.Fatal("expected a link from the relaxed retry, got nil")
	}
	if link.URL != "https://example.com/relaxed-match" {
		t.Errorf("link URL = %q, want the relaxed-retry result", link.URL)
	}
}

// TestSearchForGoalOnceNoFallbackWhenScorerAlreadyEmpty verifies that a goal
// with no scorer name never triggers the relaxed retry: its first query is
// already the relaxed shape (buildGoalQuery omits an empty scorer token), so
// retrying it would be a wasted duplicate request against Reddit.
func TestSearchForGoalOnceNoFallbackWhenScorerAlreadyEmpty(t *testing.T) {
	goal := GoalInfo{
		MatchID:   4,
		HomeTeam:  "Barcelona",
		AwayTeam:  "Real Madrid",
		HomeScore: 1,
		AwayScore: 1,
		Minute:    60,
		MatchTime: time.Now(),
	}
	fetcher := &recordingFetcher{} // no results at any call index
	client := NewClientWithFetcher(fetcher, &GoalLinkCache{links: make(map[string]GoalLink)})

	link, err := client.searchForGoalOnce(goal)
	if err != nil {
		t.Fatalf("searchForGoalOnce returned err: %v", err)
	}
	if link != nil {
		t.Errorf("expected nil link (no results), got %+v", link)
	}
	if len(fetcher.queries) != 1 {
		t.Fatalf("expected exactly 1 fetcher.Search call (no retry for empty-scorer goal), got %d: %v",
			len(fetcher.queries), fetcher.queries)
	}
}
