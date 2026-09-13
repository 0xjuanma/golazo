package fotmob

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/0xjuanma/golazo/internal/ratelimit"
)

// testDateStr is the date the fixture matches below are played on.
const testDateStr = "2026-03-10"

// leaguePageBody builds a league page carrying one finished match on
// testDateStr, shaped like the pageProps FotMob's Next.js page embeds.
func leaguePageBody(leagueID int) string {
	return fmt.Sprintf(
		`<html><script id="__NEXT_DATA__" type="application/json">{"props":{"pageProps":{`+
			`"details":{"id":%d,"name":"League %d","country":"ESP"},`+
			`"fixtures":{"allMatches":[{"id":"%d01","round":"1",`+
			`"home":{"id":"1","name":"Home %d"},"away":{"id":"2","name":"Away %d"},`+
			`"status":{"utcTime":"%sT20:00:00.000Z","started":true,"finished":true}}]}`+
			`}}}</script></html>`,
		leagueID, leagueID, leagueID, leagueID, leagueID, testDateStr,
	)
}

// matchesCacheTestClient builds a Client whose upstream behaviour the test
// controls. While failing is true every league page request errors; flipping it
// to false simulates the source recovering.
func matchesCacheTestClient(t *testing.T) (client *Client, failing *atomic.Bool, hits *atomic.Int32) {
	t.Helper()
	failing = &atomic.Bool{}
	hits = &atomic.Int32{}

	transport := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		hits.Add(1)
		if failing.Load() {
			return nil, errors.New("simulated upstream failure")
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(leaguePageBody(extractIDFromPath(req.URL.Path)))),
			Request:    req,
			Header:     make(http.Header),
		}, nil
	})

	client = &Client{
		httpClient:    &http.Client{Transport: transport, Timeout: 5 * time.Second},
		baseURL:       baseURL,
		rateLimiter:   ratelimit.New(0),
		cache:         NewResponseCache(DefaultCacheConfig()),
		pageURLs:      make(map[int]string, 10),
		maxConcurrent: make(chan struct{}, 10),
	}
	return client, failing, hits
}

// activeLeaguesForTest reports the league set the client will actually query,
// so these tests hold for any configured selection.
func activeLeaguesForTest(t *testing.T) []int {
	t.Helper()
	leagues := ActiveLeagues()
	if len(leagues) == 0 {
		t.Skip("no active leagues configured; nothing to query")
	}
	return leagues
}

func mustParseTestDate(t *testing.T) time.Time {
	t.Helper()
	date, err := time.Parse("2006-01-02", testDateStr)
	if err != nil {
		t.Fatalf("parse test date: %v", err)
	}
	return date
}

// TestMatchesByDate_TotalFailureIsNotCached is the regression test for the
// reported bug: when every league query fails, the empty aggregate was cached
// under the date and returned with a nil error, so the date stayed empty for
// the whole matches TTL even after the source recovered.
func TestMatchesByDate_TotalFailureIsNotCached(t *testing.T) {
	client, failing, _ := matchesCacheTestClient(t)
	leagues := activeLeaguesForTest(t)
	date := mustParseTestDate(t)

	failing.Store(true)
	matches, err := client.MatchesByDate(context.Background(), date)
	if err == nil {
		t.Error("MatchesByDate returned nil error when every league query failed")
	}
	if len(matches) != 0 {
		t.Errorf("failed fetch returned %d matches, want 0", len(matches))
	}
	if cached := client.cache.Matches(testDateStr); cached != nil {
		t.Errorf("date cached after total failure (%d entries); a later call can never see the recovered data", len(cached))
	}

	// The source recovers: the next call must reach upstream and return data.
	failing.Store(false)
	matches, err = client.MatchesByDate(context.Background(), date)
	if err != nil {
		t.Fatalf("recovered fetch failed: %v", err)
	}
	if len(matches) != len(leagues) {
		t.Errorf("recovered fetch returned %d matches, want %d (one per active league)", len(matches), len(leagues))
	}
}

// TestMatchesByDate_PartialFailureIsNotCached covers one league failing while
// the others answer. The best-effort result is still returned, but it is an
// incomplete answer for the date, so it must not be cached.
func TestMatchesByDate_PartialFailureIsNotCached(t *testing.T) {
	leagues := activeLeaguesForTest(t)
	if len(leagues) < 2 {
		t.Skip("need at least two active leagues to fail one and succeed another")
	}
	date := mustParseTestDate(t)
	brokenLeague := leagues[0]

	transport := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		id := extractIDFromPath(req.URL.Path)
		if id == brokenLeague {
			return nil, errors.New("simulated upstream failure")
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(leaguePageBody(id))),
			Request:    req,
			Header:     make(http.Header),
		}, nil
	})
	client := &Client{
		httpClient:    &http.Client{Transport: transport, Timeout: 5 * time.Second},
		baseURL:       baseURL,
		rateLimiter:   ratelimit.New(0),
		cache:         NewResponseCache(DefaultCacheConfig()),
		pageURLs:      make(map[int]string, 10),
		maxConcurrent: make(chan struct{}, 10),
	}

	matches, err := client.MatchesByDate(context.Background(), date)
	if err != nil {
		t.Fatalf("partial failure should still return the successful leagues: %v", err)
	}
	if want := len(leagues) - 1; len(matches) != want {
		t.Errorf("partial fetch returned %d matches, want %d", len(matches), want)
	}
	if cached := client.cache.Matches(testDateStr); cached != nil {
		t.Errorf("date cached after a partial failure (%d entries); the missing league could never recover", len(cached))
	}
}

// TestMatchesByDate_DecodeFailureIsNotCached covers a league page that is
// fetched successfully but cannot be decoded. That is a failure too, so the
// aggregate must not be cached.
func TestMatchesByDate_DecodeFailureIsNotCached(t *testing.T) {
	activeLeaguesForTest(t)
	date := mustParseTestDate(t)

	transport := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		body := `<html><script id="__NEXT_DATA__" type="application/json">` +
			`{"props":{"pageProps":{"details":[]}}}</script></html>`
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(body)),
			Request:    req,
			Header:     make(http.Header),
		}, nil
	})
	client := &Client{
		httpClient:    &http.Client{Transport: transport, Timeout: 5 * time.Second},
		baseURL:       baseURL,
		rateLimiter:   ratelimit.New(0),
		cache:         NewResponseCache(DefaultCacheConfig()),
		pageURLs:      make(map[int]string, 10),
		maxConcurrent: make(chan struct{}, 10),
	}

	if _, err := client.MatchesByDate(context.Background(), date); err == nil {
		t.Error("MatchesByDate returned nil error when every league response failed to decode")
	}
	if cached := client.cache.Matches(testDateStr); cached != nil {
		t.Errorf("date cached after decode failures (%d entries)", len(cached))
	}
}

// TestMatchesByDate_GenuineEmptyDayIsCached guards the other side of the fix:
// a day on which every league answers and simply has no matches is a complete
// result, so it must still be cached and served without a second round trip.
func TestMatchesByDate_GenuineEmptyDayIsCached(t *testing.T) {
	client, failing, hits := matchesCacheTestClient(t)
	activeLeaguesForTest(t)
	failing.Store(false)

	// The fixtures are on testDateStr, so any other date is genuinely empty.
	emptyDate, err := time.Parse("2006-01-02", "2026-03-11")
	if err != nil {
		t.Fatalf("parse date: %v", err)
	}

	matches, err := client.MatchesByDate(context.Background(), emptyDate)
	if err != nil {
		t.Fatalf("empty-day fetch failed: %v", err)
	}
	if len(matches) != 0 {
		t.Fatalf("empty-day fetch returned %d matches, want 0", len(matches))
	}
	if cached := client.cache.Matches("2026-03-11"); cached == nil {
		t.Error("a genuinely empty day was not cached; every call would re-query every league")
	}

	// Page bodies are cached too, so assert on the date cache: a second call
	// must not re-enter the aggregation path's network fetches.
	before := hits.Load()
	if _, err := client.MatchesByDate(context.Background(), emptyDate); err != nil {
		t.Fatalf("second empty-day fetch failed: %v", err)
	}
	if got := hits.Load(); got != before {
		t.Errorf("second call made %d extra request(s); the cached empty day should have served it", got-before)
	}
}
