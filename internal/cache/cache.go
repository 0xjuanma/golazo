package cache

import (
	"sync"
	"time"
)

// entry holds a cached value with its expiration time.
type entry[V any] struct {
	value     V
	expiresAt time.Time
}

// Map is a generic, thread-safe, in-memory cache with TTL and max size eviction.
type Map[K comparable, V any] struct {
	mu      sync.RWMutex
	entries map[K]entry[V]
	ttl     time.Duration
	maxSize int
}

// NewMap creates a new cache with the given TTL and maximum number of entries.
// When maxSize is reached, expired entries are purged first, then the oldest entry is evicted.
func NewMap[K comparable, V any](ttl time.Duration, maxSize int) *Map[K, V] {
	return &Map[K, V]{
		entries: make(map[K]entry[V]),
		ttl:     ttl,
		maxSize: maxSize,
	}
}

// Get retrieves a value from the cache. Returns the zero value and false if
// the key is missing or expired.
func (c *Map[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	e, ok := c.entries[key]
	if !ok || time.Now().After(e.expiresAt) {
		var zero V
		return zero, false
	}
	return e.value, true
}

// Set stores a value in the cache with the default TTL.
func (c *Map[K, V]) Set(key K, value V) {
	c.SetWithTTL(key, value, c.ttl)
}

// SetWithTTL stores a value in the cache with a custom TTL.
func (c *Map[K, V]) SetWithTTL(key K, value V, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.entries) >= c.maxSize {
		c.evictLocked()
	}

	c.entries[key] = entry[V]{
		value:     value,
		expiresAt: time.Now().Add(ttl),
	}
}

// Delete removes a specific key from the cache.
func (c *Map[K, V]) Delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, key)
}

// Clear removes all entries from the cache.
func (c *Map[K, V]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[K]entry[V])
}

// Keys returns all non-expired keys in the cache.
func (c *Map[K, V]) Keys() []K {
	c.mu.RLock()
	defer c.mu.RUnlock()

	now := time.Now()
	keys := make([]K, 0, len(c.entries))
	for k, e := range c.entries {
		if !now.After(e.expiresAt) {
			keys = append(keys, k)
		}
	}
	return keys
}

// evictLocked removes expired entries, then the oldest if still at capacity.
// Must hold write lock.
func (c *Map[K, V]) evictLocked() {
	now := time.Now()
	var oldestKey K
	var oldestTime time.Time
	first := true

	for key, e := range c.entries {
		if now.After(e.expiresAt) {
			delete(c.entries, key)
			continue
		}
		if first || e.expiresAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = e.expiresAt
			first = false
		}
	}

	if len(c.entries) >= c.maxSize && !first {
		delete(c.entries, oldestKey)
	}
}
