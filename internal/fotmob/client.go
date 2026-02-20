package fotmob

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/0xjuanma/golazo/internal/api"
	"github.com/0xjuanma/golazo/internal/data"
)

const (
	baseURL = "https://www.fotmob.com/api"
)

// ActiveLeagues returns the league IDs to use for API calls.
// This respects user settings - if specific leagues are selected, only those are returned.
// If no selection is made, returns all supported leagues.
func ActiveLeagues() []int {
	return data.ActiveLeagueIDs()
}

// SupportedLeagues is kept for backward compatibility but now uses settings.
// Use ActiveLeagues() for dynamic league selection based on user preferences.
var SupportedLeagues = data.AllLeagueIDs()

// Client implements the api.Client interface for FotMob API
type Client struct {
	httpClient  *http.Client
	baseURL     string
	rateLimiter *RateLimiter
	cache       *ResponseCache
	emptyCache  *EmptyResultsCache // Persistent cache for empty league+date combinations
	pageURLs    map[int]string     // Match ID -> page slug mapping for page-based fetching
	pageURLsMu  sync.RWMutex
}

// NewClient creates a new FotMob API client with default configuration.
// Includes minimal rate limiting (200ms between requests) for fast concurrent requests.
// Uses default caching configuration for improved performance.
// Initializes persistent empty results cache to skip known empty league+date combinations.
func NewClient() *Client {
	// Initialize empty results cache (logs error but doesn't fail)
	emptyCache, err := NewEmptyResultsCache()
	if err != nil {
		// If we can't create the cache, create client without it
		emptyCache = nil
	}

	return &Client{
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		baseURL:     baseURL,
		rateLimiter: NewRateLimiter(200 * time.Millisecond), // Minimal delay for concurrent requests
		cache:       NewResponseCache(DefaultCacheConfig()),
		emptyCache:  emptyCache,
		pageURLs:    make(map[int]string),
	}
}

// Cache returns the response cache for external access (e.g., pre-fetching).
func (c *Client) Cache() *ResponseCache {
	return c.cache
}

// StorePageURL stores a match page URL slug for later use by MatchDetails.
func (c *Client) StorePageURL(matchID int, pageURL string) {
	if pageURL == "" {
		return
	}
	c.pageURLsMu.Lock()
	c.pageURLs[matchID] = pageURL
	c.pageURLsMu.Unlock()
}

// getPageURL retrieves the stored page URL slug for a match ID.
func (c *Client) getPageURL(matchID int) string {
	c.pageURLsMu.RLock()
	defer c.pageURLsMu.RUnlock()
	return c.pageURLs[matchID]
}

// SaveEmptyCache persists the empty results cache to disk.
// Should be called periodically or when the application exits.
func (c *Client) SaveEmptyCache() error {
	if c.emptyCache == nil {
		return nil
	}
	return c.emptyCache.Save()
}

// EmptyCacheStats returns statistics about the empty results cache.
func (c *Client) EmptyCacheStats() (total int, expired int) {
	if c.emptyCache == nil {
		return 0, 0
	}
	return c.emptyCache.Stats()
}

// MatchesByDate retrieves all matches for a specific date.
// Since FotMob doesn't have a single endpoint for all matches by date,
// we query each supported league separately and filter by date client-side.
// We query both "fixtures" (upcoming) and "results" (finished) tabs concurrently.
// All requests are made concurrently with minimal rate limiting for maximum speed.
// Results are cached to avoid redundant API calls.
func (c *Client) MatchesByDate(ctx context.Context, date time.Time) ([]api.Match, error) {
	return c.MatchesByDateWithTabs(ctx, date, []string{"fixtures", "results"})
}

// MatchesByDateWithTabs retrieves matches for a specific date, querying only specified tabs.
// tabs can be: ["fixtures"], ["results"], or ["fixtures", "results"]
// This allows optimizing API calls - e.g., only query "results" for past days.
// Results are cached per date (cache key includes all tabs for that date).
func (c *Client) MatchesByDateWithTabs(ctx context.Context, date time.Time, tabs []string) ([]api.Match, error) {
	// Normalize date to UTC for consistent comparison
	requestDateStr := date.UTC().Format("2006-01-02")

	// Check cache first (only if querying both tabs - full cache)
	if len(tabs) == 2 {
		if cached := c.cache.Matches(requestDateStr); cached != nil {
			return cached, nil
		}
	}

	// Use a mutex to protect the shared slice
	var mu sync.Mutex
	var allMatches []api.Match

	// Query leagues concurrently - no stagger delays, just rate limiting
	// Best-effort aggregation: if a league query fails, we skip it and continue with others
	// This allows partial results even if some leagues are unavailable
	var wg sync.WaitGroup

	// Track skipped leagues for logging/debugging
	var skippedFromCache int

	// Get active leagues (respects user settings)
	activeLeagues := ActiveLeagues()

	// Query specified tabs
	for _, tab := range tabs {
		for _, leagueID := range activeLeagues {
			// Check empty cache before spawning goroutine (for "results" tab only)
			// Skip leagues known to have no matches on this date
			if tab == "results" && c.emptyCache != nil && c.emptyCache.IsEmpty(requestDateStr, leagueID) {
				skippedFromCache++
				continue
			}

			wg.Add(1)
			go func(id int, tabName string) {
				defer wg.Done()

				// Apply rate limiting (minimal delay for concurrent requests)
				c.rateLimiter.Wait()

				url := fmt.Sprintf("%s/leagues?id=%d&tab=%s", c.baseURL, id, tabName)

				req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
				if err != nil {
					// Skip this league on error - best effort aggregation
					return
				}

				req.Header.Set("User-Agent", "Mozilla/5.0")

				resp, err := c.httpClient.Do(req)
				if err != nil {
					// Skip this league on request error - best effort aggregation
					return
				}
				defer func() { _ = resp.Body.Close() }()

				var leagueResponse struct {
					Details struct {
						ID          int    `json:"id"`
						Name        string `json:"name"`
						Country     string `json:"country"`
						CountryCode string `json:"countryCode,omitempty"`
					} `json:"details"`
					Fixtures struct {
						AllMatches []fotmobMatch `json:"allMatches"`
					} `json:"fixtures"`
				}

				if err := json.NewDecoder(resp.Body).Decode(&leagueResponse); err != nil {
					// Skip this league on parse error - best effort aggregation
					return
				}

				// Filter matches for the requested date and add league info
				// Note: Matches are sorted chronologically, so we need to check all matches
				var leagueMatches []api.Match
				for _, m := range leagueResponse.Fixtures.AllMatches {
					// Check if match is on the requested date
					if m.Status.UTCTime != "" {
						// Parse the UTC time - FotMob sometimes uses .000Z format
						var matchTime time.Time
						var err error
						matchTime, err = time.Parse(time.RFC3339, m.Status.UTCTime)
						if err != nil {
							// Try alternative format with milliseconds (.000Z)
							matchTime, err = time.Parse("2006-01-02T15:04:05.000Z", m.Status.UTCTime)
						}
						if err == nil {
							// Compare dates in UTC to avoid timezone issues
							matchDateStr := matchTime.UTC().Format("2006-01-02")
							if matchDateStr == requestDateStr {
								// Set league info from the response details
								if m.League.ID == 0 {
									m.League = league{
										ID:          leagueResponse.Details.ID,
										Name:        leagueResponse.Details.Name,
										Country:     leagueResponse.Details.Country,
										CountryCode: leagueResponse.Details.CountryCode,
									}
								}
								apiMatch := m.toAPIMatch()
								c.StorePageURL(apiMatch.ID, apiMatch.PageURL)
								leagueMatches = append(leagueMatches, apiMatch)
							}
						}
					}
				}

				// Mark league+date as empty if no matches found (for results tab only)
				// This will be persisted to avoid future API calls
				if len(leagueMatches) == 0 && tabName == "results" && c.emptyCache != nil {
					c.emptyCache.MarkEmpty(requestDateStr, id)
				}

				// Append to shared slice with mutex protection
				mu.Lock()
				allMatches = append(allMatches, leagueMatches...)
				mu.Unlock()
			}(leagueID, tab)
		}
	}

	// Variable is used below (prevents unused variable error)
	_ = skippedFromCache

	wg.Wait()

	// Cache the results before returning
	c.cache.SetMatches(requestDateStr, allMatches)

	// Persist empty results cache to disk (async, best-effort)
	go func() { _ = c.SaveEmptyCache() }()

	return allMatches, nil
}

// MatchesForLeagueAndDate fetches matches for a single league on a specific date.
// Used for progressive loading - allows fetching one league at a time.
func (c *Client) MatchesForLeagueAndDate(ctx context.Context, leagueID int, date time.Time, tab string) ([]api.Match, error) {
	requestDateStr := date.UTC().Format("2006-01-02")

	// Apply rate limiting
	c.rateLimiter.Wait()

	url := fmt.Sprintf("%s/leagues?id=%d&tab=%s", c.baseURL, leagueID, tab)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request for league %d: %w", leagueID, err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch league %d: %w", leagueID, err)
	}
	defer func() { _ = resp.Body.Close() }()

	var leagueResponse struct {
		Details struct {
			ID          int    `json:"id"`
			Name        string `json:"name"`
			Country     string `json:"country"`
			CountryCode string `json:"countryCode,omitempty"`
		} `json:"details"`
		Fixtures struct {
			AllMatches []fotmobMatch `json:"allMatches"`
		} `json:"fixtures"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&leagueResponse); err != nil {
		return nil, fmt.Errorf("decode league %d response: %w", leagueID, err)
	}

	// Filter matches for the requested date
	var matches []api.Match
	for _, m := range leagueResponse.Fixtures.AllMatches {
		if m.Status.UTCTime != "" {
			var matchTime time.Time
			var parseErr error
			matchTime, parseErr = time.Parse(time.RFC3339, m.Status.UTCTime)
			if parseErr != nil {
				matchTime, parseErr = time.Parse("2006-01-02T15:04:05.000Z", m.Status.UTCTime)
			}
			if parseErr == nil {
				matchDateStr := matchTime.UTC().Format("2006-01-02")
				if matchDateStr == requestDateStr {
					if m.League.ID == 0 {
						m.League = league{
							ID:          leagueResponse.Details.ID,
							Name:        leagueResponse.Details.Name,
							Country:     leagueResponse.Details.Country,
							CountryCode: leagueResponse.Details.CountryCode,
						}
					}
					apiMatch := m.toAPIMatch()
				c.StorePageURL(apiMatch.ID, apiMatch.PageURL)
				matches = append(matches, apiMatch)
				}
			}
		}
	}

	return matches, nil
}

// MatchDetails retrieves detailed information about a specific match.
// Results are cached to avoid redundant API calls.
//
// Uses page-based fetching (match page HTML with __NEXT_DATA__) as the primary
// method, since the /api/matchDetails endpoint now requires Cloudflare Turnstile
// verification. Falls back to the direct API endpoint if page fetching fails.
func (c *Client) MatchDetails(ctx context.Context, matchID int) (*api.MatchDetails, error) {
	// Check cache first
	if cached := c.cache.Details(matchID); cached != nil {
		return cached, nil
	}

	// Apply rate limiting
	c.rateLimiter.Wait()

	// Try page-based fetching first (primary method)
	if pageSlug := c.getPageURL(matchID); pageSlug != "" {
		details, err := fetchMatchDetailsFromPage(ctx, c.httpClient, pageSlug)
		if err == nil && details != nil {
			c.cache.SetDetails(matchID, details)
			return details, nil
		}
		// Page fetch failed, fall through to direct API
	}

	// Fallback: direct API endpoint (may return 403 if Turnstile is required)
	return c.matchDetailsFromAPI(ctx, matchID)
}

// matchDetailsFromAPI fetches match details from the /api/matchDetails endpoint.
// This is the original fetching method, kept as a fallback in case the endpoint
// becomes accessible again or for specific match IDs without a page URL.
func (c *Client) matchDetailsFromAPI(ctx context.Context, matchID int) (*api.MatchDetails, error) {
	url := fmt.Sprintf("%s/matchDetails?matchId=%d", c.baseURL, matchID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request for match %d: %w", matchID, err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch match details for match %d: %w", matchID, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d for match %d", resp.StatusCode, matchID)
	}

	var response fotmobMatchDetails

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode match details response for match %d: %w", matchID, err)
	}

	details := response.toAPIMatchDetails()

	// Cache the result
	c.cache.SetDetails(matchID, details)

	return details, nil
}

// MatchDetailsForceRefresh fetches match details, bypassing the cache.
// Use this for polling live matches to ensure fresh data.
func (c *Client) MatchDetailsForceRefresh(ctx context.Context, matchID int) (*api.MatchDetails, error) {
	c.cache.ClearMatchDetails(matchID)
	return c.MatchDetails(ctx, matchID)
}

// BatchMatchDetails retrieves details for multiple matches concurrently.
// Uses caching and rate limiting to balance speed with API limits.
// Returns a map of matchID -> details (nil if fetch failed).
func (c *Client) BatchMatchDetails(ctx context.Context, matchIDs []int) map[int]*api.MatchDetails {
	results := make(map[int]*api.MatchDetails)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, id := range matchIDs {
		wg.Add(1)
		go func(matchID int) {
			defer wg.Done()

			details, err := c.MatchDetails(ctx, matchID)
			if err != nil {
				// Store nil for failed fetches
				mu.Lock()
				results[matchID] = nil
				mu.Unlock()
				return
			}

			mu.Lock()
			results[matchID] = details
			mu.Unlock()
		}(id)
	}

	wg.Wait()
	return results
}

// PreFetchMatchDetails fetches details for the first N matches in the background.
// This improves perceived performance by pre-loading details before user selection.
// maxConcurrent limits how many concurrent requests to make.
func (c *Client) PreFetchMatchDetails(ctx context.Context, matchIDs []int, maxPrefetch int) {
	if len(matchIDs) == 0 {
		return
	}

	// Limit the number of matches to prefetch
	if maxPrefetch > 0 && len(matchIDs) > maxPrefetch {
		matchIDs = matchIDs[:maxPrefetch]
	}

	// Filter out already cached matches
	var uncachedIDs []int
	for _, id := range matchIDs {
		if c.cache.Details(id) == nil {
			uncachedIDs = append(uncachedIDs, id)
		}
	}

	if len(uncachedIDs) == 0 {
		return
	}

	// Fetch uncached matches in the background (fire and forget)
	go func() {
		c.BatchMatchDetails(ctx, uncachedIDs)
	}()
}

// Leagues retrieves available leagues.
func (c *Client) Leagues(ctx context.Context) ([]api.League, error) {
	// FotMob doesn't have a direct leagues endpoint, so we'll return an empty slice
	// In a real implementation, you might need to maintain a list of known leagues
	// or fetch them from a different endpoint
	return []api.League{}, nil
}

// LeagueMatches retrieves matches for a specific league.
func (c *Client) LeagueMatches(ctx context.Context, leagueID int) ([]api.Match, error) {
	// This would require a different endpoint structure
	// For now, we'll return an empty slice
	// In a real implementation, you'd use: /api/leagues?id={leagueID}
	return []api.Match{}, nil
}

// parentLeagueByName maps league name patterns to their parent league IDs.
// Some competitions have sub-leagues for different stages/seasons that don't have
// their own standings - we detect these by name and use the parent league.
// This is more robust than mapping sub-league IDs which change each stage/season.
var parentLeagueByName = map[string]int{
	"Champions League":  42,
	"Europa League":     73,
	"Conference League": 10216,
	"Libertadores":      45,
	"Sudamericana":      299,
}

// getParentLeagueID returns the parent league ID if the league name matches a known pattern.
// Returns the original leagueID if no parent match is found.
func getParentLeagueID(leagueName string, leagueID int) int {
	for pattern, parentID := range parentLeagueByName {
		if strings.Contains(leagueName, pattern) {
			return parentID
		}
	}
	return leagueID
}

// LeagueTable retrieves the league table/standings for a specific league.
// Handles both regular league tables and knockout competition tables (e.g., Champions League).
// Uses parentLeagueID (from FotMob match details) when available, then falls back to
// league name pattern matching for knockout competitions.
// Multi-season leagues (e.g., Liga MX, Liga Profesional) have sub-league IDs per season
// that don't have standings — the parentLeagueID points to the main league that does.
func (c *Client) LeagueTable(ctx context.Context, leagueID int, leagueName string) ([]api.LeagueTableEntry, error) {
	// Determine the effective league ID for standings lookup.
	// Priority: parentLeagueID from match details > name pattern matching > original ID
	effectiveID := getParentLeagueID(leagueName, leagueID)

	// Fetch standings using the effective league ID
	return c.fetchLeagueTable(ctx, effectiveID)
}

// LeagueTableWithParent retrieves the league table/standings, using the parent league ID
// when available. This is the preferred method when match details provide a parentLeagueId.
// Multi-season leagues (e.g., Liga MX Clausura, Liga Profesional Apertura) return sub-league
// IDs in match details that have no standings — the parentLeagueID points to the main league.
func (c *Client) LeagueTableWithParent(ctx context.Context, leagueID int, leagueName string, parentLeagueID int) ([]api.LeagueTableEntry, error) {
	effectiveID := leagueID

	// Use parentLeagueID if it differs from leagueID (indicates a sub-season league)
	if parentLeagueID > 0 && parentLeagueID != leagueID {
		effectiveID = parentLeagueID
	} else {
		// Fall back to name-based parent league detection for knockout competitions
		effectiveID = getParentLeagueID(leagueName, leagueID)
	}

	return c.fetchLeagueTable(ctx, effectiveID)
}

// fetchLeagueTable fetches the league table for a specific league ID.
func (c *Client) fetchLeagueTable(ctx context.Context, leagueID int) ([]api.LeagueTableEntry, error) {
	// Apply rate limiting
	c.rateLimiter.Wait()

	url := fmt.Sprintf("%s/leagues?id=%d", c.baseURL, leagueID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request for league %d table: %w", leagueID, err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch league table for league %d: %w", leagueID, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d for league %d table", resp.StatusCode, leagueID)
	}

	// FotMob returns table data in several formats:
	// 1. Regular leagues (EPL, La Liga): table[0].data.table.all[]
	// 2. Knockout competitions (Champions League): table[0].data.tables[0].table.all[]
	// 3. Multi-season leagues (Liga MX, Liga Profesional): table[0].data.tables[] with
	//    multiple sub-tables (e.g., Clausura + Apertura, or Group A + Group B).
	//    The first sub-table is typically the current/most relevant season.
	var response struct {
		Table []struct {
			Data struct {
				// Regular league table
				Table struct {
					All []fotmobTableRow `json:"all"`
				} `json:"table"`
				// Multi-table format: knockout competitions and multi-season leagues
				// Examples:
				//   - Champions League: single table with all teams
				//   - Liga MX: Clausura + Apertura tables
				//   - Liga Profesional: Apertura Group A + Group B tables
				Tables []struct {
					Table struct {
						All []fotmobTableRow `json:"all"`
					} `json:"table"`
					LeagueName string `json:"leagueName"` // e.g., "Clausura", "Apertura - Group A"
				} `json:"tables"`
			} `json:"data"`
		} `json:"table"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode league table response for league %d: %w", leagueID, err)
	}

	// Extract table rows - try regular format first, then multi-table format
	var tableData []fotmobTableRow
	if len(response.Table) > 0 {
		data := response.Table[0].Data
		// Try regular league format first (single table with all teams)
		if len(data.Table.All) > 0 {
			tableData = data.Table.All
		} else if len(data.Tables) > 0 {
			// Multi-table format: use first sub-table (current/most relevant season)
			// This covers both knockout competitions and multi-season leagues
			for _, subTable := range data.Tables {
				if len(subTable.Table.All) > 0 {
					tableData = subTable.Table.All
					break
				}
			}
		}
	}

	if len(tableData) == 0 {
		return nil, fmt.Errorf("no table data available for league %d", leagueID)
	}

	entries := make([]api.LeagueTableEntry, 0, len(tableData))
	for _, row := range tableData {
		entries = append(entries, row.toAPITableEntry())
	}

	return entries, nil
}
