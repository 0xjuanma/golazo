package reddit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/0xjuanma/golazo/internal/data"
)

const (
	goalLinksFileName = "goal_links.json"
	// CacheTTL defines how long goal links are stored.
	// 7 days keeps the cache file small while covering recent matches.
	CacheTTL = 7 * 24 * time.Hour // 7 days
	// NotFoundTTL defines how long to cache "not found" results.
	// Shorter than CacheTTL since links might appear later. 1h is a balance:
	// long enough to avoid re-hitting Reddit for the same missing goal within
	// a single viewing session (queue is paced at 30s/request), short enough
	// that late-uploaded videos surface on the next app run.
	NotFoundTTL = 1 * time.Hour
	// NotFoundMarker is a special URL indicating "searched but not found"
	NotFoundMarker = "__NOT_FOUND__"
)

// GoalLinkCache provides persistent storage for goal replay links.
type GoalLinkCache struct {
	mu       sync.RWMutex
	links    map[string]GoalLink // key: "matchID:minute"
	filePath string
}

// NewGoalLinkCache creates a new cache, loading existing data from disk.
func NewGoalLinkCache() (*GoalLinkCache, error) {
	dir, err := data.ConfigDir()
	if err != nil {
		return nil, fmt.Errorf("get config dir: %w", err)
	}

	cache := &GoalLinkCache{
		links:    make(map[string]GoalLink),
		filePath: filepath.Join(dir, goalLinksFileName),
	}

	// Load existing cache from disk (silently ignore errors - start with empty cache)
	_ = cache.load()

	// Clean expired entries on startup to keep file size manageable
	_ = cache.CleanExpired()

	return cache, nil
}

// makeKey creates a cache key from matchID and minute.
func makeKey(key GoalLinkKey) string {
	return fmt.Sprintf("%d:%d", key.MatchID, key.Minute)
}

// Get retrieves a goal link from cache if it exists and is not expired.
// Returns nil if not cached or expired.
// To distinguish "not found" from "not cached", use Exists().
func (c *GoalLinkCache) Get(key GoalLinkKey) *GoalLink {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cacheKey := makeKey(key)
	link, ok := c.links[cacheKey]
	if !ok {
		return nil
	}

	// Check if this is a "not found" marker
	if link.URL == NotFoundMarker {
		// Check TTL for not-found entries (shorter)
		if time.Since(link.FetchedAt) > NotFoundTTL {
			return nil // Expired, allow retry
		}
		return &link // Return marker to indicate "searched but not found"
	}

	// Check if expired for regular entries
	if time.Since(link.FetchedAt) > CacheTTL {
		return nil
	}

	return &link
}

// IsNotFound returns true if the cached entry is a "not found" marker.
func IsNotFound(link *GoalLink) bool {
	return link != nil && link.URL == NotFoundMarker
}

// SetNotFound stores a "not found" marker in the cache.
// This prevents re-fetching goals that weren't found on Reddit.
func (c *GoalLinkCache) SetNotFound(matchID, minute int) error {
	return c.Set(GoalLink{
		MatchID:   matchID,
		Minute:    minute,
		URL:       NotFoundMarker,
		FetchedAt: time.Now(),
	})
}

// Set stores a goal link in the cache and persists to disk.
func (c *GoalLinkCache) Set(link GoalLink) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := makeKey(GoalLinkKey{MatchID: link.MatchID, Minute: link.Minute})
	c.links[key] = link

	return c.saveLocked()
}

// All returns all cached goal links for a match.
func (c *GoalLinkCache) All(matchID int) []GoalLink {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var result []GoalLink
	for _, link := range c.links {
		if link.MatchID == matchID && time.Since(link.FetchedAt) <= CacheTTL {
			result = append(result, link)
		}
	}
	return result
}

// Clear removes all cached goal links.
func (c *GoalLinkCache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.links = make(map[string]GoalLink)
	return c.saveLocked()
}

// CleanExpired removes expired entries from the cache.
// Uses different TTLs for regular links vs "not found" markers.
func (c *GoalLinkCache) CleanExpired() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	cleaned := false
	for key, link := range c.links {
		age := time.Since(link.FetchedAt)

		// Use shorter TTL for "not found" entries
		if link.URL == NotFoundMarker {
			if age > NotFoundTTL {
				delete(c.links, key)
				cleaned = true
			}
		} else {
			if age > CacheTTL {
				delete(c.links, key)
				cleaned = true
			}
		}
	}

	// Only save if something was cleaned
	if cleaned {
		return c.saveLocked()
	}
	return nil
}

// load reads the cache from disk.
func (c *GoalLinkCache) load() error {
	data, err := os.ReadFile(c.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No cache file yet, that's fine
		}
		return fmt.Errorf("read cache file: %w", err)
	}

	var links []GoalLink
	if err := json.Unmarshal(data, &links); err != nil {
		return fmt.Errorf("parse cache file: %w", err)
	}

	// Convert to map
	for _, link := range links {
		key := makeKey(GoalLinkKey{MatchID: link.MatchID, Minute: link.Minute})
		c.links[key] = link
	}

	return nil
}

// saveLocked persists the cache to disk (must hold write lock).
func (c *GoalLinkCache) saveLocked() error {
	// Convert map to slice for JSON
	links := make([]GoalLink, 0, len(c.links))
	for _, link := range c.links {
		links = append(links, link)
	}

	data, err := json.MarshalIndent(links, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal cache: %w", err)
	}

	if err := os.WriteFile(c.filePath, data, 0644); err != nil {
		return fmt.Errorf("write cache file: %w", err)
	}

	return nil
}

// Size returns the number of cached goal links.
func (c *GoalLinkCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.links)
}
