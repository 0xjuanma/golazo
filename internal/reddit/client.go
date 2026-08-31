package reddit

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/0xjuanma/golazo/internal/ratelimit"
)

// ErrBlocked indicates Reddit's edge returned an HTTP 403 (typically the
// "blocked by network security" interstitial). Returned by Search so callers
// can react with errors.Is without sniffing HTML response bodies.
var ErrBlocked = errors.New("reddit: blocked (HTTP 403)")

// DebugLogger is a function type for debug logging
type DebugLogger func(message string)

// Fetcher defines the interface for fetching data from Reddit.
// Uses Reddit's public JSON API for goal link retrieval.
type Fetcher interface {
	Search(query string, limit int, matchTime time.Time, sort string) ([]SearchResult, error)
}

// PublicJSONFetcher uses Reddit's public JSON endpoints (no auth required).
// Uses Reddit's public JSON API with rate limiting.
type PublicJSONFetcher struct {
	httpClient  *http.Client
	rateLimiter *ratelimit.Limiter
}

// browserUserAgent is sent on every request. The queue's 30s pacing and
// cooldown subsume the anti-detection role that User-Agent rotation
// previously played; one fixed browser-shaped UA is enough to coexist with
// Reddit's edge heuristics without introducing additional sources of
// nondeterminism in the request signature.
const browserUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.5 Safari/605.1.15"

// NewPublicJSONFetcher creates a new fetcher using public Reddit JSON API.
func NewPublicJSONFetcher() *PublicJSONFetcher {
	return &PublicJSONFetcher{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		rateLimiter: ratelimit.NewFromRate(10), // 10 requests per minute for public API
	}
}

// Search performs a search on r/soccer for Media posts matching the query.
// matchTime is used to filter results to posts created around the match date.
// sort controls the result ordering (e.g., "relevance", "top", "new", "hot").
func (f *PublicJSONFetcher) Search(query string, limit int, matchTime time.Time, sort string) ([]SearchResult, error) {
	f.rateLimiter.Wait()

	// Build timestamp range for filtering. Aligned with the matcher's
	// accepted date window (matcher.go: -24h .. +48h) so search results are
	// not narrower than what the matcher will validate. Late-uploaded goal
	// videos for matches in distant timezones live in this wider band.
	startTime := matchTime.Add(-24 * time.Hour).Unix()
	endTime := matchTime.Add(48 * time.Hour).Unix()

	// Default to relevance if sort is empty
	if sort == "" {
		sort = "relevance"
	}

	// Build search URL for r/soccer with Media flair filter and timestamp.
	// Targets the legacy `old.reddit.com` host: its edge has historically
	// applied laxer bot-detection rules than `www.reddit.com` while serving
	// the same JSON shape. Falls back to www if Reddit eventually retires it.
	searchURL := fmt.Sprintf(
		"https://old.reddit.com/r/soccer/search.json?q=%s+flair:Media+timestamp:%d..%d&restrict_sr=on&sort=%s&limit=%d",
		url.QueryEscape(query),
		startTime,
		endTime,
		url.QueryEscape(sort),
		limit,
	)

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", browserUserAgent)
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch from reddit: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode == http.StatusForbidden {
			return nil, fmt.Errorf("%w: body: %s", ErrBlocked, string(body))
		}
		return nil, fmt.Errorf("reddit API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var searchResp redditSearchResponse
	if err := json.Unmarshal(body, &searchResp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	results := make([]SearchResult, 0, len(searchResp.Data.Children))
	for _, child := range searchResp.Data.Children {
		result := child.Data.toSearchResult()
		// Only include posts with Media flair
		if result.Flair == "Media" {
			results = append(results, result)
		}
	}

	return results, nil
}

// Client provides goal replay link fetching from Reddit r/soccer.
// Uses Reddit's public JSON API for goal link retrieval.
type Client struct {
	fetcher     Fetcher // Reddit public API fetcher
	cache       *GoalLinkCache
	debugLogger DebugLogger // Optional debug logger function

	queueOnce sync.Once
	queue     *goalQueue
}

// DebugLog forwards a message to the configured debug logger if one is wired.
// No-op in non-debug runs. Exported so callers outside the reddit package
// (e.g., the app's goal-link orchestration) emit their goal-related diagnostics
// through the same logger as the reddit client's internal search logs.
func (c *Client) DebugLog(message string) {
	if c.debugLogger != nil {
		c.debugLogger(message)
	}
}

// NewClient creates a new Reddit client with the default public JSON fetcher.
func NewClient() (*Client, error) {
	cache, err := NewGoalLinkCache()
	if err != nil {
		return nil, fmt.Errorf("create cache: %w", err)
	}

	return &Client{
		fetcher: NewPublicJSONFetcher(),
		cache:   cache,
	}, nil
}

// NewClientWithDebug creates a new Reddit client with debug logging enabled.
// Uses public JSON API like main branch.
func NewClientWithDebug(debugLogger DebugLogger) (*Client, error) {
	cache, err := NewGoalLinkCache()
	if err != nil {
		return nil, fmt.Errorf("create cache: %w", err)
	}

	debugLogger("Initializing Reddit client with public API")

	return &Client{
		fetcher:     NewPublicJSONFetcher(),
		cache:       cache,
		debugLogger: debugLogger,
	}, nil
}

// NewClientWithFetcher creates a new Reddit client with a custom fetcher.
// Use this for testing with custom fetchers.
func NewClientWithFetcher(fetcher Fetcher, cache *GoalLinkCache) *Client {
	return &Client{
		fetcher: fetcher,
		cache:   cache,
	}
}

// GoalLink retrieves a cached goal link or fetches from Reddit if not cached.
// Returns nil if the goal link was previously searched but not found.
func (c *Client) GoalLink(goal GoalInfo) (*GoalLink, error) {
	key := GoalLinkKey{MatchID: goal.MatchID, Minute: goal.Minute}

	// Check cache first (includes "not found" markers)
	if link := c.cache.Get(key); link != nil {
		// If this is a "not found" marker, return nil (don't re-search)
		if IsNotFound(link) {
			return nil, nil
		}
		return link, nil
	}

	// Search Reddit for the goal via a single-strategy attempt. The queue
	// (see GoalLinksAsync) is the production path for batched fetches and
	// owns pacing/cooldown; this singular method is preserved for the
	// scripts/test_reddit_search.go ad-hoc tool and any direct caller that
	// wants synchronous single-goal semantics without subscription wiring.
	link, err := c.searchForGoalOnce(goal)
	if err != nil {
		// Don't cache errors - allow retry
		return nil, err
	}

	if link != nil {
		// Cache the result (silently ignore cache errors - best-effort)
		_ = c.cache.Set(*link)
	} else {
		// Cache "not found" to avoid re-searching
		_ = c.cache.SetNotFound(goal.MatchID, goal.Minute)
	}

	return link, nil
}

// GoalLinksAsync schedules a fetch for each goal through the per-Client queue
// and streams results on the returned channel. Cache hits are emitted
// immediately; uncached goals are enqueued for serial fetching at
// QueueInterval pacing. The returned channel is closed once every goal in
// `goals` has produced a result (or been deduplicated against an in-flight
// peer). Each emitted GoalResult.Link is nil when the goal was not found,
// was dropped due to an ErrBlocked cooldown, or hit a transient fetch error.
//
// Use this in preference to the synchronous GoalLinks: it's what the app's
// subscription wiring consumes for progressive per-goal UI updates and is
// the only path that honors the global queue's cooldown semantics.
func (c *Client) GoalLinksAsync(goals []GoalInfo) <-chan GoalResult {
	out := make(chan GoalResult, len(goals))

	// First pass: serve cache hits inline (no queue work) and collect the
	// uncached subset, de-duplicated by key. Same de-dup as the sync path —
	// keeps the queue contract simple (one fetch per key per batch) while
	// in-flight de-dup inside the queue handles cross-batch collisions.
	seen := make(map[GoalLinkKey]bool)
	var work []GoalInfo
	for _, g := range goals {
		key := GoalLinkKey{MatchID: g.MatchID, Minute: g.Minute}
		if seen[key] {
			continue
		}
		seen[key] = true

		if link := c.cache.Get(key); link != nil {
			if !IsNotFound(link) {
				out <- GoalResult{Key: key, Link: link}
			}
			continue
		}
		work = append(work, g)
	}

	if len(work) == 0 {
		close(out)
		return out
	}

	// Reply channel buffered to len(work) so the queue worker never blocks
	// when broadcasting results to this batch. The forwarder goroutine below
	// owns closing `out` once every queued goal has emitted exactly one
	// GoalResult.
	replies := make(chan GoalResult, len(work))
	queue := c.goalQueueLazy()
	for _, g := range work {
		queue.Enqueue(g, replies)
	}

	go func() {
		defer close(out)
		for i := 0; i < len(work); i++ {
			r, ok := <-replies
			if !ok {
				return
			}
			out <- r
		}
	}()

	return out
}

// goalQueueLazy returns the per-Client queue, constructing it on first use.
// Keeping construction lazy means the worker goroutine doesn't start until a
// caller actually opts into the async API.
func (c *Client) goalQueueLazy() *goalQueue {
	c.queueOnce.Do(func() {
		store, err := newQueueStateStore()
		if err != nil {
			c.DebugLog(fmt.Sprintf("failed to init queue state store: %v", err))
			store = nil
		}
		c.queue = newGoalQueue(c.searchForGoalOnce, c.cache, c.debugLogger, 0, 0, store)
	})
	return c.queue
}

// searchForGoalOnce performs a single search attempt for a goal.
//
// Query format: "<home> <homeScore> <awayScore> <away> <scorerLast>".
// Example for the 7' New Zealand goal in Iran 0-1 New Zealand:
//
//	"Iran 0 1 New Zealand Just"
//
// This mirrors verbatim the token sequence that appears in r/soccer goal-post
// titles like "Iran 0 - [1] New Zealand - E. Just 7'" (slug
// `iran_0_1_new_zealand_e_just_7`). The running score uniquely disambiguates
// goals within a single match — searching by minute alone is the weakest
// signal because Reddit's tokenizer handles the apostrophe inconsistently and
// bare numbers are low-entropy.
//
// When ScorerName is empty (own goals, missing data), the scorer token is
// omitted and matching relies on score + team names. Minute validation lives
// in findBestMatch via buildMinutePattern.
func (c *Client) searchForGoalOnce(goal GoalInfo) (*GoalLink, error) {
	// Log any country-alias variants that will be tried during matching.
	// Helps diagnose national-team mismatches (e.g., FotMob "Türkiye" vs
	// Reddit titles using "Turkey") at a glance in golazo_debug.log.
	if aliases := aliasesFor(normalizeTeamName(goal.HomeTeam)); len(aliases) > 0 {
		c.DebugLog(fmt.Sprintf("Reddit alias expansion for home %q -> %v (goal %d:%d)",
			goal.HomeTeam, aliases, goal.MatchID, goal.Minute))
	}
	if aliases := aliasesFor(normalizeTeamName(goal.AwayTeam)); len(aliases) > 0 {
		c.DebugLog(fmt.Sprintf("Reddit alias expansion for away %q -> %v (goal %d:%d)",
			goal.AwayTeam, aliases, goal.MatchID, goal.Minute))
	}

	query := buildGoalQuery(goal)
	c.DebugLog(fmt.Sprintf("Reddit search query: %q for goal %d:%d (%s %d-%d %s)",
		query, goal.MatchID, goal.Minute, goal.HomeTeam, goal.HomeScore, goal.AwayScore, goal.AwayTeam))

	results, err := c.fetcher.Search(query, 15, goal.MatchTime, "relevance")
	if err != nil {
		c.DebugLog(fmt.Sprintf("Reddit search failed for query %q: %v", query, err))
		return nil, err
	}
	c.DebugLog(fmt.Sprintf("Reddit search returned %d results for query %q", len(results), query))
	for i, result := range results {
		if i < 3 {
			c.DebugLog(fmt.Sprintf("Result %d: %q", i+1, result.Title))
		}
	}

	match := findBestMatch(results, goal)
	c.DebugLog(fmt.Sprintf("findBestMatch result for goal %d:%d (score %d-%d): %v",
		goal.MatchID, goal.Minute, goal.HomeScore, goal.AwayScore, match != nil))
	if match == nil {
		return nil, nil
	}

	c.DebugLog(fmt.Sprintf("Found goal link for %d:%d: %s (post: %s)",
		goal.MatchID, goal.Minute, match.URL, match.PostURL))
	return &GoalLink{
		MatchID:   goal.MatchID,
		Minute:    goal.Minute,
		URL:       match.URL,
		Title:     match.Title,
		PostURL:   match.PostURL,
		FetchedAt: time.Now(),
	}, nil
}

// buildGoalQuery returns the single Reddit search query for a goal:
//
//	"<home> <homeScore> <awayScore> <away> <scorerLast>"
//
// Falls back to "<home> <homeScore> <awayScore> <away>" when the scorer is
// unknown (own goals, missing data). The scorer token is the last whitespace-
// separated component of ScorerName with diacritics folded so the query
// matches anglicized title spellings (e.g., "Núñez" → "Nunez").
func buildGoalQuery(goal GoalInfo) string {
	parts := []string{
		goal.HomeTeam,
		strconv.Itoa(goal.HomeScore),
		strconv.Itoa(goal.AwayScore),
		goal.AwayTeam,
	}
	if last := scorerLastToken(goal.ScorerName); last != "" {
		parts = append(parts, last)
	}
	return strings.Join(parts, " ")
}

// scorerLastToken returns the last whitespace-separated token of name with
// diacritics folded. Returns "" when name is empty or has no usable token.
func scorerLastToken(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	folded := foldDiacritics(name)
	fields := strings.Fields(folded)
	if len(fields) == 0 {
		return ""
	}
	return fields[len(fields)-1]
}

// ClearCache clears the goal link cache.
func (c *Client) ClearCache() error {
	return c.cache.Clear()
}

// Cache returns the underlying cache for direct access if needed.
func (c *Client) Cache() *GoalLinkCache {
	return c.cache
}
