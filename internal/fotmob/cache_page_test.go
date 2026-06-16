package fotmob

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/0xjuanma/golazo/internal/ratelimit"
)

// roundTripperFunc adapts a plain function to http.RoundTripper for tests.
type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// pageCacheTestClient builds a Client with a controllable TTL and a counting
// HTTP transport. It bypasses NewClient's emptyCache wiring (not needed here).
func pageCacheTestClient(t *testing.T, ttl time.Duration) (*Client, *atomic.Int32) {
	t.Helper()
	var hits atomic.Int32

	transport := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		if !strings.Contains(req.URL.Host, "fotmob.com") {
			return nil, fmt.Errorf("unexpected host: %s", req.URL.Host)
		}
		hits.Add(1)
		body := fmt.Sprintf(
			`<html><script id="__NEXT_DATA__" type="application/json">{"props":{"pageProps":{"details":{"id":%d,"name":"Test League"}}}}</script></html>`,
			extractIDFromPath(req.URL.Path),
		)
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(body)),
			Request:    req,
			Header:     make(http.Header),
		}, nil
	})

	cfg := DefaultCacheConfig()
	cfg.PageBodyTTL = ttl

	client := &Client{
		httpClient:    &http.Client{Transport: transport, Timeout: 5 * time.Second},
		baseURL:       baseURL,
		rateLimiter:   ratelimit.New(0),
		cache:         NewResponseCache(cfg),
		pageURLs:      make(map[int]string, 10),
		maxConcurrent: make(chan struct{}, 10),
	}
	return client, &hits
}

func TestFetchLeaguePage_CacheHitSkipsNetwork(t *testing.T) {
	client, hits := pageCacheTestClient(t, 5*time.Second)

	for i := 0; i < 3; i++ {
		body, err := client.fetchLeaguePage(context.Background(), 77)
		if err != nil {
			t.Fatalf("fetchLeaguePage call %d failed: %v", i, err)
		}
		if len(body) == 0 {
			t.Fatalf("call %d returned empty body", i)
		}
	}

	if got := hits.Load(); got != 1 {
		t.Errorf("network hits = %d, want 1 (cache should have served calls 2 and 3)", got)
	}
}

func TestFetchLeaguePage_ExpiredEntryRefetches(t *testing.T) {
	// 50ms TTL so the test stays fast; we sleep past it before the second call.
	client, hits := pageCacheTestClient(t, 50*time.Millisecond)

	if _, err := client.fetchLeaguePage(context.Background(), 77); err != nil {
		t.Fatalf("first fetch failed: %v", err)
	}
	time.Sleep(80 * time.Millisecond)
	if _, err := client.fetchLeaguePage(context.Background(), 77); err != nil {
		t.Fatalf("second fetch failed: %v", err)
	}

	if got := hits.Load(); got != 2 {
		t.Errorf("network hits = %d, want 2 (TTL expiry should force refetch)", got)
	}
}

func TestFetchLeaguePage_DifferentLeaguesDoNotShareCache(t *testing.T) {
	client, hits := pageCacheTestClient(t, 5*time.Second)

	if _, err := client.fetchLeaguePage(context.Background(), 77); err != nil {
		t.Fatalf("fetch league 77 failed: %v", err)
	}
	if _, err := client.fetchLeaguePage(context.Background(), 47); err != nil {
		t.Fatalf("fetch league 47 failed: %v", err)
	}
	// Repeat both — should both hit cache now.
	if _, err := client.fetchLeaguePage(context.Background(), 77); err != nil {
		t.Fatalf("refetch league 77 failed: %v", err)
	}
	if _, err := client.fetchLeaguePage(context.Background(), 47); err != nil {
		t.Fatalf("refetch league 47 failed: %v", err)
	}

	if got := hits.Load(); got != 2 {
		t.Errorf("network hits = %d, want 2 (one per league, cache for repeats)", got)
	}
}

func TestFetchLeaguePage_ClearPagesForcesRefetch(t *testing.T) {
	client, hits := pageCacheTestClient(t, 5*time.Second)

	if _, err := client.fetchLeaguePage(context.Background(), 77); err != nil {
		t.Fatalf("first fetch failed: %v", err)
	}
	if _, err := client.fetchLeaguePage(context.Background(), 77); err != nil {
		t.Fatalf("cached fetch failed: %v", err)
	}
	if got := hits.Load(); got != 1 {
		t.Fatalf("network hits before clear = %d, want 1", got)
	}

	client.cache.ClearPages()

	if _, err := client.fetchLeaguePage(context.Background(), 77); err != nil {
		t.Fatalf("post-clear fetch failed: %v", err)
	}
	if got := hits.Load(); got != 2 {
		t.Errorf("network hits after ClearPages = %d, want 2 (force-refresh path)", got)
	}
}

// extractIDFromPath pulls the league ID from "/leagues/{id}" so each cached
// body carries a distinct identifier (helps catch any cross-key contamination).
func extractIDFromPath(path string) int {
	const prefix = "/leagues/"
	if !strings.HasPrefix(path, prefix) {
		return 0
	}
	rest := strings.TrimPrefix(path, prefix)
	if idx := strings.IndexAny(rest, "/?"); idx >= 0 {
		rest = rest[:idx]
	}
	var id int
	if _, err := fmt.Sscanf(rest, "%d", &id); err != nil {
		return 0
	}
	return id
}
