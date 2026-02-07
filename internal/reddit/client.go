package reddit

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

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
	userAgent   string
	rateLimiter *rateLimiter
}

// rateLimiter implements simple rate limiting for Reddit API.
type rateLimiter struct {
	mu          sync.Mutex
	lastRequest time.Time
	minInterval time.Duration
}

func newRateLimiter(requestsPerMinute int) *rateLimiter {
	interval := time.Minute / time.Duration(requestsPerMinute)
	return &rateLimiter{
		minInterval: interval,
	}
}

func (r *rateLimiter) wait() {
	r.mu.Lock()
	defer r.mu.Unlock()

	elapsed := time.Since(r.lastRequest)
	if elapsed < r.minInterval {
		time.Sleep(r.minInterval - elapsed)
	}
	r.lastRequest = time.Now()
}

// NewPublicJSONFetcher creates a new fetcher using public Reddit JSON API.
func NewPublicJSONFetcher() *PublicJSONFetcher {
	return &PublicJSONFetcher{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		// Reddit requires a descriptive User-Agent
		userAgent:   "golazo:v1.0.0 (by /u/golazo_app)",
		rateLimiter: newRateLimiter(10), // 10 requests per minute for public API
	}
}

// Search performs a search on r/soccer for Media posts matching the query.
// matchTime is used to filter results to posts created around the match date.
// sort controls the result ordering (e.g., "relevance", "top", "new", "hot").
func (f *PublicJSONFetcher) Search(query string, limit int, matchTime time.Time, sort string) ([]SearchResult, error) {
	f.rateLimiter.wait()

	// Build timestamp range for filtering (match day only ±12 hours)
	// Goal videos are posted very soon after goals happen - limit to match day
	startTime := matchTime.Add(-12 * time.Hour).Unix()
	endTime := matchTime.Add(12 * time.Hour).Unix()

	// Default to relevance if sort is empty
	if sort == "" {
		sort = "relevance"
	}

	// Build search URL for r/soccer with Media flair filter and timestamp
	// Reddit CloudSearch supports timestamp:START..END syntax
	searchURL := fmt.Sprintf(
		"https://www.reddit.com/r/soccer/search.json?q=%s+flair:Media+timestamp:%d..%d&restrict_sr=on&sort=%s&limit=%d",
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
	req.Header.Set("User-Agent", f.userAgent)

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch from reddit: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
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
}

// debugLog is a helper method to safely call the debug logger if it exists
func (c *Client) debugLog(message string) {
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

	// Search Reddit for the goal
	link, err := c.searchForGoal(goal)
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

// BatchSize is the maximum number of goals to fetch per batch.
// Reduced to make requests even more spaced out.
const BatchSize = 3

// BatchDelay is the delay between batches to avoid rate limiting.
const BatchDelay = 5 * time.Second

// GoalLinks retrieves links for multiple goals, using cache where available.
// Goals are de-duplicated and batched to avoid rate limiting.
func (c *Client) GoalLinks(goals []GoalInfo) map[GoalLinkKey]*GoalLink {
	results := make(map[GoalLinkKey]*GoalLink)

	// De-duplicate goals by key and filter out already-cached goals
	seen := make(map[GoalLinkKey]bool)
	var uncachedGoals []GoalInfo

	for _, goal := range goals {
		key := GoalLinkKey{MatchID: goal.MatchID, Minute: goal.Minute}

		// Skip duplicates
		if seen[key] {
			continue
		}
		seen[key] = true

		// Check cache first
		if link := c.cache.Get(key); link != nil {
			if !IsNotFound(link) {
				results[key] = link
			}
			// Skip - already cached (found or not found)
			continue
		}

		uncachedGoals = append(uncachedGoals, goal)
	}

	// Fetch uncached goals in batches with conservative delays
	for i := 0; i < len(uncachedGoals); i += BatchSize {
		// Add delay between batches (not before first batch)
		if i > 0 {
			time.Sleep(BatchDelay)
		}

		// Process batch
		end := i + BatchSize
		end = min(end, len(uncachedGoals))

		for _, goal := range uncachedGoals[i:end] {
			key := GoalLinkKey{MatchID: goal.MatchID, Minute: goal.Minute}
			link, err := c.GoalLink(goal)
			if err == nil && link != nil {
				results[key] = link
			}
		}
	}

	return results
}

// searchForGoal searches Reddit for a specific goal with conservative retry logic.
func (c *Client) searchForGoal(goal GoalInfo) (*GoalLink, error) {
	// Conservative retry logic - Reddit is very aggressive with CAPTCHA detection
	maxRetries := 2               // Reduced from 3
	baseDelay := 60 * time.Second // Increased delay between retries

	var lastErr error
	for attempt := range maxRetries {
		if attempt > 0 {
			// Exponential backoff: 30s, 60s, 120s
			delay := time.Duration(attempt) * baseDelay
			time.Sleep(delay)
		}

		result, err := c.searchForGoalOnce(goal)
		if err == nil {
			return result, nil
		}

		lastErr = err

		// Check if this is a CAPTCHA/rate limit error
		if strings.Contains(err.Error(), "CAPTCHA") ||
			strings.Contains(err.Error(), "blocking requests") ||
			strings.Contains(err.Error(), "rate limit") ||
			strings.Contains(err.Error(), "HTML instead of JSON") {
			// Don't retry CAPTCHA errors - Reddit is very aggressive, just give up
			c.debugLog(fmt.Sprintf("Reddit blocking goal %d:%d: giving up immediately", goal.MatchID, goal.Minute))
			return nil, err
		}

		// For other errors, retry on next attempt
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return nil, nil // No match found after all retries
}

// searchForGoalOnce performs a single search attempt for a goal.
func (c *Client) searchForGoalOnce(goal GoalInfo) (*GoalLink, error) {
	// Strategy 1: Both teams + minute (most specific, try first)
	query1 := fmt.Sprintf("%s %s %d'", goal.HomeTeam, goal.AwayTeam, goal.Minute)
	c.debugLog(fmt.Sprintf("Reddit search query: '%s' for goal %d:%d (%s vs %s)",
		query1, goal.MatchID, goal.Minute, goal.HomeTeam, goal.AwayTeam))
	results1, err := c.fetcher.Search(query1, 15, goal.MatchTime, "relevance")
	if err != nil {
		c.debugLog(fmt.Sprintf("Reddit search failed for query '%s': %v", query1, err))
	} else {
		c.debugLog(fmt.Sprintf("Reddit search returned %d results for query '%s'", len(results1), query1))
		// Debug: log the first few result titles
		for i, result := range results1 {
			if i < 3 { // Log first 3 results
				c.debugLog(fmt.Sprintf("Result %d: '%s'", i+1, result.Title))
			}
		}
	}
	if err == nil {
		// Check if we found a good match with the first strategy
		match := findBestMatch(results1, goal)
		c.debugLog(fmt.Sprintf("findBestMatch result for goal %d:%d (score %d-%d): %v", goal.MatchID, goal.Minute, goal.HomeScore, goal.AwayScore, match != nil))
		if match != nil {
			c.debugLog(fmt.Sprintf("Found goal link for %d:%d: %s (post: %s)", goal.MatchID, goal.Minute, match.URL, match.PostURL))
			// Found a match, return it immediately to avoid additional API calls
			return &GoalLink{
				MatchID:   goal.MatchID,
				Minute:    goal.Minute,
				URL:       match.URL,
				Title:     match.Title,
				PostURL:   match.PostURL,
				FetchedAt: time.Now(),
			}, nil
		}
	}

	// Strategy 1 didn't find a match, try broader searches
	// Only try one additional strategy to balance coverage vs rate limiting
	var allResults []SearchResult
	if err == nil {
		allResults = append(allResults, results1...)
	}

	// Strategy 2: Try with just the scoring team + minute
	// Determine which team scored
	scoringTeam := goal.AwayTeam
	if goal.IsHomeTeam {
		scoringTeam = goal.HomeTeam
	}
	query2 := fmt.Sprintf("%s %d'", scoringTeam, goal.Minute)
	c.debugLog(fmt.Sprintf("Reddit search query (strategy 2): '%s' for goal %d:%d", query2, goal.MatchID, goal.Minute))
	results2, err := c.fetcher.Search(query2, 15, goal.MatchTime, "relevance")
	if err != nil {
		c.debugLog(fmt.Sprintf("Reddit search failed for strategy 2 query '%s': %v", query2, err))
	} else {
		c.debugLog(fmt.Sprintf("Reddit search returned %d results for strategy 2 query '%s'", len(results2), query2))
		allResults = append(allResults, results2...)
	}

	// Remove duplicates based on URL
	seen := make(map[string]bool)
	uniqueResults := make([]SearchResult, 0, len(allResults))
	for _, result := range allResults {
		if !seen[result.URL] {
			seen[result.URL] = true
			uniqueResults = append(uniqueResults, result)
		}
	}

	// Check if strategies 1+2 found a match before trying strategy 3
	match := findBestMatch(uniqueResults, goal)
	if match != nil {
		c.debugLog(fmt.Sprintf("Strategy 1+2 match found for goal %d:%d, skipping strategy 3", goal.MatchID, goal.Minute))
		c.debugLog(fmt.Sprintf("Found goal link for %d:%d: %s (post: %s)", goal.MatchID, goal.Minute, match.URL, match.PostURL))
		return &GoalLink{
			MatchID:   goal.MatchID,
			Minute:    goal.Minute,
			URL:       match.URL,
			Title:     match.Title,
			PostURL:   match.PostURL,
			FetchedAt: time.Now(),
		}, nil
	}

	// Strategy 3: Try with short/alternative team names + sort by top (upvotes)
	// Only runs if strategies 1+2 did not find a match
	homeShort := strings.TrimSpace(goal.HomeTeamShort)
	awayShort := strings.TrimSpace(goal.AwayTeamShort)

	// Skip if short names are empty or identical to full names (avoids redundant API call)
	homeShortDifferent := homeShort != "" && !strings.EqualFold(homeShort, goal.HomeTeam)
	awayShortDifferent := awayShort != "" && !strings.EqualFold(awayShort, goal.AwayTeam)

	if !homeShortDifferent && !awayShortDifferent {
		c.debugLog(fmt.Sprintf("Skipping strategy 3 for goal %d:%d: short names empty or identical to full names", goal.MatchID, goal.Minute))
		return nil, nil // No match found across all strategies
	}

	// Build query using short names where they differ, falling back to full names
	homeQuery := goal.HomeTeam
	if homeShortDifferent {
		homeQuery = homeShort
	}
	awayQuery := goal.AwayTeam
	if awayShortDifferent {
		awayQuery = awayShort
	}

	query3 := fmt.Sprintf("%s %s %d'", homeQuery, awayQuery, goal.Minute)
	c.debugLog(fmt.Sprintf("Reddit search query (strategy 3): '%s' for goal %d:%d", query3, goal.MatchID, goal.Minute))
	results3, err := c.fetcher.Search(query3, 15, goal.MatchTime, "top")
	if err != nil {
		c.debugLog(fmt.Sprintf("Reddit search failed for strategy 3 query '%s': %v", query3, err))
	} else {
		c.debugLog(fmt.Sprintf("Reddit search returned %d results for strategy 3 query '%s'", len(results3), query3))
		// Debug: log the first few result titles
		for i, result := range results3 {
			if i < 3 { // Log first 3 results
				c.debugLog(fmt.Sprintf("Strategy 3 result %d: '%s' (score: %d)", i+1, result.Title, result.Score))
			}
		}
		// Combine with all prior results for best match selection
		for _, result := range results3 {
			if !seen[result.URL] {
				seen[result.URL] = true
				uniqueResults = append(uniqueResults, result)
			}
		}
	}

	// Find the best matching result across all strategies
	match = findBestMatch(uniqueResults, goal)
	c.debugLog(fmt.Sprintf("findBestMatch result (strategy 3) for goal %d:%d: %v", goal.MatchID, goal.Minute, match != nil))
	if match == nil {
		return nil, nil // No match found, but not an error
	}

	c.debugLog(fmt.Sprintf("Found goal link (strategy 3) for %d:%d: %s (post: %s)", goal.MatchID, goal.Minute, match.URL, match.PostURL))
	return &GoalLink{
		MatchID:   goal.MatchID,
		Minute:    goal.Minute,
		URL:       match.URL,
		Title:     match.Title,
		PostURL:   match.PostURL,
		FetchedAt: time.Now(),
	}, nil
}

// ClearCache clears the goal link cache.
func (c *Client) ClearCache() error {
	return c.cache.Clear()
}

// Cache returns the underlying cache for direct access if needed.
func (c *Client) Cache() *GoalLinkCache {
	return c.cache
}
