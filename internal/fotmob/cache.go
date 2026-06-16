package fotmob

import (
	"encoding/json"
	"time"

	"github.com/0xjuanma/golazo/internal/api"
	"github.com/0xjuanma/golazo/internal/cache"
)

// CacheConfig holds configuration for API response caching.
type CacheConfig struct {
	MatchesTTL      time.Duration // How long to cache match list results
	MatchDetailsTTL time.Duration // How long to cache match details
	PageBodyTTL     time.Duration // How long to cache raw FotMob league page JSON bodies
	MaxMatchesCache int           // Maximum number of date entries to cache
	MaxDetailsCache int           // Maximum number of match details to cache
	MaxPageCache    int           // Maximum number of league pages to cache
}

// DefaultCacheConfig returns sensible defaults for caching.
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		MatchesTTL:      15 * time.Minute, // Matches list cache (stats view uses client-side filtering)
		MatchDetailsTTL: 5 * time.Minute,  // Details for live matches need fresher data
		PageBodyTTL:     60 * time.Second, // Raw league page bodies — short to keep live data fresh
		MaxMatchesCache: 10,               // Cache up to 10 date queries
		MaxDetailsCache: 100,              // Cache up to 100 match details
		MaxPageCache:    30,               // Cache up to 30 league pages (well above any plausible active-leagues count)
	}
}

// ResponseCache provides thread-safe caching for API responses.
type ResponseCache struct {
	config       CacheConfig
	matchesCache *cache.Map[string, []api.Match]
	detailsCache *cache.Map[int, *api.MatchDetails]
	pageCache    *cache.Map[int, json.RawMessage]
}

// NewResponseCache creates a new cache with the given configuration.
func NewResponseCache(config CacheConfig) *ResponseCache {
	return &ResponseCache{
		config:       config,
		matchesCache: cache.NewMap[string, []api.Match](config.MatchesTTL, config.MaxMatchesCache),
		detailsCache: cache.NewMap[int, *api.MatchDetails](config.MatchDetailsTTL, config.MaxDetailsCache),
		pageCache:    cache.NewMap[int, json.RawMessage](config.PageBodyTTL, config.MaxPageCache),
	}
}

// Matches retrieves cached matches for a date, returns nil if not cached or expired.
func (c *ResponseCache) Matches(dateKey string) []api.Match {
	matches, ok := c.matchesCache.Get(dateKey)
	if !ok {
		return nil
	}
	return matches
}

// SetMatches stores matches in cache with TTL.
func (c *ResponseCache) SetMatches(dateKey string, matches []api.Match) {
	c.matchesCache.Set(dateKey, matches)
}

// Details retrieves cached match details, returns nil if not cached or expired.
func (c *ResponseCache) Details(matchID int) *api.MatchDetails {
	details, ok := c.detailsCache.Get(matchID)
	if !ok {
		return nil
	}
	return details
}

// SetDetails stores match details in cache with TTL.
// For finished matches, uses a longer TTL since the data won't change.
func (c *ResponseCache) SetDetails(matchID int, details *api.MatchDetails) {
	ttl := c.config.MatchDetailsTTL
	if details != nil && details.Status == api.MatchStatusFinished {
		ttl = 30 * time.Minute
	}
	c.detailsCache.SetWithTTL(matchID, details, ttl)
}

// ClearMatchDetails removes a specific match from the details cache.
// Use this to force a refresh on next fetch for a specific match.
func (c *ResponseCache) ClearMatchDetails(matchID int) {
	c.detailsCache.Delete(matchID)
}

// Page retrieves a cached league-page JSON body, returns nil if not cached or expired.
func (c *ResponseCache) Page(leagueID int) json.RawMessage {
	body, ok := c.pageCache.Get(leagueID)
	if !ok {
		return nil
	}
	return body
}

// SetPage stores a league-page JSON body in the cache with the configured TTL.
func (c *ResponseCache) SetPage(leagueID int, body json.RawMessage) {
	c.pageCache.Set(leagueID, body)
}

// ClearPages invalidates all cached league-page bodies. Use this when the
// caller needs a forced refresh (e.g. user-initiated reload of the live list).
func (c *ResponseCache) ClearPages() {
	c.pageCache.Clear()
}

