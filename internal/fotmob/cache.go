package fotmob

import (
	"sync"
	"time"

	"github.com/0xjuanma/golazo/internal/api"
	"github.com/0xjuanma/golazo/internal/cache"
)

// CacheConfig holds configuration for API response caching.
type CacheConfig struct {
	MatchesTTL      time.Duration // How long to cache match list results
	MatchDetailsTTL time.Duration // How long to cache match details
	LiveMatchesTTL  time.Duration // How long to cache live matches list
	MaxMatchesCache int           // Maximum number of date entries to cache
	MaxDetailsCache int           // Maximum number of match details to cache
}

// DefaultCacheConfig returns sensible defaults for caching.
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		MatchesTTL:      15 * time.Minute, // Matches list cache (stats view uses client-side filtering)
		MatchDetailsTTL: 5 * time.Minute,  // Details for live matches need fresher data
		LiveMatchesTTL:  2 * time.Minute,  // Live matches list cache (quick nav doesn't re-fetch)
		MaxMatchesCache: 10,               // Cache up to 10 date queries
		MaxDetailsCache: 100,              // Cache up to 100 match details
	}
}

// ResponseCache provides thread-safe caching for API responses.
type ResponseCache struct {
	config       CacheConfig
	matchesCache *cache.Map[string, []api.Match]
	detailsCache *cache.Map[int, *api.MatchDetails]
	liveMu       sync.RWMutex
	liveCache    *liveEntry
}

type liveEntry struct {
	matches   []api.Match
	expiresAt time.Time
}

// NewResponseCache creates a new cache with the given configuration.
func NewResponseCache(config CacheConfig) *ResponseCache {
	return &ResponseCache{
		config:       config,
		matchesCache: cache.NewMap[string, []api.Match](config.MatchesTTL, config.MaxMatchesCache),
		detailsCache: cache.NewMap[int, *api.MatchDetails](config.MatchDetailsTTL, config.MaxDetailsCache),
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

// LiveMatches retrieves cached live matches, returns nil if not cached or expired.
func (c *ResponseCache) LiveMatches() []api.Match {
	c.liveMu.RLock()
	defer c.liveMu.RUnlock()

	if c.liveCache == nil || time.Now().After(c.liveCache.expiresAt) {
		return nil
	}
	return c.liveCache.matches
}

// SetLiveMatches stores live matches in cache with TTL.
func (c *ResponseCache) SetLiveMatches(matches []api.Match) {
	c.liveMu.Lock()
	defer c.liveMu.Unlock()

	c.liveCache = &liveEntry{
		matches:   matches,
		expiresAt: time.Now().Add(c.config.LiveMatchesTTL),
	}
}

// ClearLive invalidates the live matches cache.
// Call this to force a refresh on next fetch.
func (c *ResponseCache) ClearLive() {
	c.liveMu.Lock()
	defer c.liveMu.Unlock()
	c.liveCache = nil
}
