package fotmob

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/0xjuanma/golazo/internal/api"
	"github.com/0xjuanma/golazo/internal/ratelimit"
)

// newTestClient creates a Client pointing at the given test server URL.
func newTestClient(baseURL string) *Client {
	return &Client{
		httpClient:    &http.Client{Timeout: 5 * time.Second},
		baseURL:       baseURL,
		rateLimiter:   ratelimit.New(0), // no delay in tests
		cache:         NewResponseCache(DefaultCacheConfig()),
		pageURLs:      make(map[int]string, 10),
		maxConcurrent: make(chan struct{}, 10),
	}
}

// minimalMatchDetailsJSON returns a valid fotmobMatchDetails JSON for match 12345.
func minimalMatchDetailsJSON() []byte {
	finished := true
	resp := fotmobMatchDetails{}
	resp.General.MatchID = "12345"
	resp.General.HomeTeam.ID = 1
	resp.General.HomeTeam.Name = "Home FC"
	resp.General.AwayTeam.ID = 2
	resp.General.AwayTeam.Name = "Away FC"
	resp.General.LeagueID = 47
	resp.General.LeagueName = "Premier League"
	resp.General.Round = "10"
	resp.Header.Status.UTCTime = "2026-03-10T20:00:00Z"
	resp.Header.Status.Finished = &finished
	resp.Header.Teams = []struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Score int    `json:"score"`
	}{
		{ID: 1, Name: "Home FC", Score: 2},
		{ID: 2, Name: "Away FC", Score: 1},
	}

	data, _ := json.Marshal(resp)
	return data
}

func TestMatchDetails_ValidResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(minimalMatchDetailsJSON())
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	details, err := client.matchDetailsFromAPI(context.Background(), 12345)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if details == nil {
		t.Fatal("details is nil")
	}
	if details.ID != 12345 {
		t.Errorf("ID = %d, want 12345", details.ID)
	}
	if details.HomeTeam.Name != "Home FC" {
		t.Errorf("HomeTeam.Name = %q, want %q", details.HomeTeam.Name, "Home FC")
	}
	if details.AwayTeam.Name != "Away FC" {
		t.Errorf("AwayTeam.Name = %q, want %q", details.AwayTeam.Name, "Away FC")
	}
	if details.Status != api.MatchStatusFinished {
		t.Errorf("Status = %q, want %q", details.Status, api.MatchStatusFinished)
	}
	if details.HomeScore == nil || *details.HomeScore != 2 {
		t.Errorf("HomeScore = %v, want 2", details.HomeScore)
	}
	if details.AwayScore == nil || *details.AwayScore != 1 {
		t.Errorf("AwayScore = %v, want 1", details.AwayScore)
	}
	if details.Winner == nil || *details.Winner != "home" {
		t.Errorf("Winner = %v, want 'home'", details.Winner)
	}
}

func TestMatchDetails_CacheHit(t *testing.T) {
	var callCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount.Add(1)
		w.Header().Set("Content-Type", "application/json")
		w.Write(minimalMatchDetailsJSON())
	}))
	defer server.Close()

	client := newTestClient(server.URL)

	// First call hits the server.
	_, err := client.matchDetailsFromAPI(context.Background(), 12345)
	if err != nil {
		t.Fatalf("first call error: %v", err)
	}
	if callCount.Load() != 1 {
		t.Fatalf("expected 1 server call, got %d", callCount.Load())
	}

	// Second call should be served from cache (matchDetailsFromAPI caches results).
	_, err = client.matchDetailsFromAPI(context.Background(), 12345)
	if err != nil {
		t.Fatalf("second call error: %v", err)
	}

	// The second call goes through matchDetailsFromAPI which also caches,
	// but since it doesn't check cache itself (MatchDetails does), we verify
	// through the MatchDetails method instead.
	// Reset to test the full MatchDetails path with cache.
	callCount.Store(0)
	client2 := newTestClient(server.URL)

	// First call via MatchDetails (no pageURL, so falls through to API).
	details1, err := client2.MatchDetails(context.Background(), 12345)
	if err != nil {
		t.Fatalf("MatchDetails first call error: %v", err)
	}
	if details1 == nil {
		t.Fatal("first call returned nil")
	}
	firstCount := callCount.Load()

	// Second call should hit cache - no additional server call.
	details2, err := client2.MatchDetails(context.Background(), 12345)
	if err != nil {
		t.Fatalf("MatchDetails second call error: %v", err)
	}
	if details2 == nil {
		t.Fatal("second call returned nil")
	}
	if callCount.Load() != firstCount {
		t.Errorf("expected cache hit (server calls %d), but got %d server calls", firstCount, callCount.Load())
	}
}

func TestMatchDetails_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	_, err := client.matchDetailsFromAPI(context.Background(), 99999)
	if err == nil {
		t.Fatal("expected error for 500 response, got nil")
	}
}

func TestMatchDetails_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{invalid json`))
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	_, err := client.matchDetailsFromAPI(context.Background(), 99999)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestMatchDetails_ContextCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.Write(minimalMatchDetailsJSON())
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err := client.matchDetailsFromAPI(ctx, 12345)
	if err == nil {
		t.Fatal("expected error for cancelled context, got nil")
	}
}

func TestStoreAndGetPageURL(t *testing.T) {
	client := newTestClient("http://unused")

	client.StorePageURL(123, "/matches/a-vs-b/slug")
	got := client.getPageURL(123)
	if got != "/matches/a-vs-b/slug" {
		t.Errorf("getPageURL(123) = %q, want /matches/a-vs-b/slug", got)
	}

	// Empty URL should not be stored.
	client.StorePageURL(456, "")
	got = client.getPageURL(456)
	if got != "" {
		t.Errorf("getPageURL(456) = %q, want empty (should not store empty)", got)
	}

	// Missing key returns empty.
	got = client.getPageURL(789)
	if got != "" {
		t.Errorf("getPageURL(789) = %q, want empty", got)
	}
}

func TestExtractPageProps(t *testing.T) {
	validHTML := `<html><head><script id="__NEXT_DATA__" type="application/json">{"props":{"pageProps":{"general":{"matchId":"123"}}}}</script></head></html>`

	props, err := extractPageProps(validHTML)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(props, &parsed); err != nil {
		t.Fatalf("failed to parse pageProps: %v", err)
	}
	gen, ok := parsed["general"].(map[string]any)
	if !ok {
		t.Fatal("expected general key in pageProps")
	}
	if gen["matchId"] != "123" {
		t.Errorf("matchId = %v, want 123", gen["matchId"])
	}
}

func TestExtractPageProps_MissingMarker(t *testing.T) {
	_, err := extractPageProps("<html><body>no next data</body></html>")
	if err == nil {
		t.Fatal("expected error for missing __NEXT_DATA__, got nil")
	}
}

func TestResponseCache_MatchDetails(t *testing.T) {
	cache := NewResponseCache(DefaultCacheConfig())

	// Miss
	if got := cache.Details(1); got != nil {
		t.Error("expected nil for cache miss")
	}

	// Set and hit
	details := &api.MatchDetails{}
	details.ID = 1
	details.Status = api.MatchStatusLive
	cache.SetDetails(1, details)

	got := cache.Details(1)
	if got == nil {
		t.Fatal("expected cache hit")
	}
	if got.ID != 1 {
		t.Errorf("cached ID = %d, want 1", got.ID)
	}

	// Clear and miss
	cache.ClearMatchDetails(1)
	if got := cache.Details(1); got != nil {
		t.Error("expected nil after clear")
	}
}

func TestResponseCache_Matches(t *testing.T) {
	cache := NewResponseCache(DefaultCacheConfig())

	// Miss
	if got := cache.Matches("2026-03-10"); got != nil {
		t.Error("expected nil for cache miss")
	}

	// Set and hit
	matches := []api.Match{{ID: 1}, {ID: 2}}
	cache.SetMatches("2026-03-10", matches)

	got := cache.Matches("2026-03-10")
	if got == nil {
		t.Fatal("expected cache hit")
	}
	if len(got) != 2 {
		t.Errorf("cached matches len = %d, want 2", len(got))
	}
}
