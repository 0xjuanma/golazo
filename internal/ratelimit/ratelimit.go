package ratelimit

import (
	"sync"
	"time"
)

// Limiter provides interval-based rate limiting for API requests.
type Limiter struct {
	mu              sync.Mutex
	lastRequestTime time.Time
	minInterval     time.Duration
}

// New creates a rate limiter with the given minimum interval between requests.
func New(minInterval time.Duration) *Limiter {
	if minInterval < 0 {
		minInterval = 0
	}
	return &Limiter{
		minInterval: minInterval,
	}
}

// NewFromRate creates a rate limiter from a requests-per-minute value.
func NewFromRate(requestsPerMinute int) *Limiter {
	interval := time.Minute / time.Duration(requestsPerMinute)
	return New(interval)
}

// Wait blocks until the minimum interval has elapsed since the last request.
func (l *Limiter) Wait() {
	l.mu.Lock()
	defer l.mu.Unlock()

	elapsed := time.Since(l.lastRequestTime)
	if elapsed < l.minInterval {
		time.Sleep(l.minInterval - elapsed)
	}
	l.lastRequestTime = time.Now()
}
