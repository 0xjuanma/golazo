package reddit

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"sync"
	"time"
)

// Defaults for the goal-link queue. Exposed as variables (not constants) so
// the queue type can override them in tests without test-only flags in
// production code paths.
const (
	// QueueInterval is the minimum gap between successive fetch attempts the
	// queue worker will perform against Reddit. 30s is conservative; bursts
	// against Reddit's edge are the failure mode this queue exists to avoid.
	QueueInterval = 30 * time.Second

	// CooldownPeriod is the base cooldown the queue pauses fetching for after
	// an ErrBlocked response from Reddit. Repeated blocks grow this
	// exponentially (see nextCooldown) up to maxBackoffMultiplier, since a
	// single fixed window isn't reliably enough to clear an IP-level block
	// that recurs shortly after the previous cooldown elapsed.
	CooldownPeriod = 10 * time.Minute

	// maxBackoffMultiplier caps exponential backoff growth at 2^(n-1) times
	// CooldownPeriod, i.e. 32x at n=6 (~5h20m on the 10m base) — long enough
	// to stop hammering a persistent block, short enough that a multi-day
	// block still gets periodically retried.
	maxBackoffMultiplier = 6

	// backoffJitterFrac is the +/- fraction of the computed backoff duration
	// randomized in, so concurrent installs don't all retry in lockstep.
	backoffJitterFrac = 0.2
)

// fetchFunc is the search hook the queue worker calls per goal. Injected so
// tests can drive the worker without involving httptest. Production wires
// this to Client.searchForGoalOnce.
type fetchFunc func(GoalInfo) (*GoalLink, error)

// queuedWork holds a goal scheduled for fetching plus the reply channels of
// every caller waiting on its result. In-flight de-duplication appends new
// reply channels to an existing entry instead of enqueuing the goal twice.
type queuedWork struct {
	goal    GoalInfo
	replies []chan<- GoalResult
}

// goalQueue is a single-worker FIFO queue for Reddit goal-link fetches. It
// enforces:
//   - one in-flight request at a time
//   - a minimum interval (QueueInterval) between fetches
//   - a global cooldown entered on ErrBlocked, growing exponentially (with
//     jitter) on repeated blocks and resetting on the next successful fetch
//   - in-flight de-duplication by GoalLinkKey
//   - drop-on-block semantics: a fetch that returns ErrBlocked is not cached
//     and not re-enqueued; callers receive a nil GoalResult.Link and the
//     queue stops fetching for the current cooldown duration
//   - cooldown and block-streak state persisted to disk (when a store is
//     configured) so both survive process restarts
//
// The worker goroutine is started lazily on the first Enqueue call (via
// sync.Once) so constructing a Client without using the async API does not
// leak a goroutine.
type goalQueue struct {
	keys     chan GoalLinkKey
	interval time.Duration
	cooldown time.Duration

	mu                sync.Mutex
	items             map[GoalLinkKey]*queuedWork
	cooldownUntil     time.Time
	consecutiveBlocks int

	once  sync.Once
	fetch fetchFunc
	cache *GoalLinkCache
	log   DebugLogger
	store *queueStateStore
}

// newGoalQueue constructs a queue. interval and cooldown default to
// QueueInterval / CooldownPeriod when zero. cache is required (callers rely
// on the queue persisting successful results); fetch is required. store is
// optional — when non-nil, a cooldown already in effect from a prior process
// is restored, and future cooldowns are persisted to it.
func newGoalQueue(fetch fetchFunc, cache *GoalLinkCache, log DebugLogger, interval, cooldown time.Duration, store *queueStateStore) *goalQueue {
	if interval <= 0 {
		interval = QueueInterval
	}
	if cooldown <= 0 {
		cooldown = CooldownPeriod
	}
	q := &goalQueue{
		keys:     make(chan GoalLinkKey, 1024),
		interval: interval,
		cooldown: cooldown,
		items:    make(map[GoalLinkKey]*queuedWork),
		fetch:    fetch,
		cache:    cache,
		log:      log,
		store:    store,
	}
	if store != nil {
		state := store.load()
		q.cooldownUntil = state.CooldownUntil
		q.consecutiveBlocks = state.ConsecutiveBlocks
	}
	return q
}

// nextCooldown computes the next cooldown duration using exponential backoff
// with jitter, based on how many consecutive ErrBlocked responses have been
// seen (including this one). Must be called with q.mu held.
func (q *goalQueue) nextCooldown() time.Duration {
	q.consecutiveBlocks++
	mult := min(q.consecutiveBlocks, maxBackoffMultiplier)
	base := q.cooldown * time.Duration(1<<uint(mult-1))
	jitter := time.Duration(float64(base) * backoffJitterFrac * (rand.Float64()*2 - 1))
	d := base + jitter
	if d <= 0 {
		d = base
	}
	return d
}

// Enqueue schedules a fetch for goal. The result is delivered on reply. If a
// goal with the same (MatchID, Minute) key is already in flight, reply is
// attached to the existing work and no second fetch occurs.
//
// Reply is sent exactly one GoalResult. The caller is responsible for the
// reply channel having capacity (or a reader). The queue uses a non-blocking
// send and logs a drop if no slot is available.
func (q *goalQueue) Enqueue(goal GoalInfo, reply chan<- GoalResult) {
	q.once.Do(q.start)

	key := GoalLinkKey{MatchID: goal.MatchID, Minute: goal.Minute}

	q.mu.Lock()
	if existing, ok := q.items[key]; ok {
		existing.replies = append(existing.replies, reply)
		q.mu.Unlock()
		return
	}
	q.items[key] = &queuedWork{goal: goal, replies: []chan<- GoalResult{reply}}
	q.mu.Unlock()

	q.keys <- key
}

// start runs the single worker. Invoked exactly once per queue via sync.Once
// from the first Enqueue call.
func (q *goalQueue) start() {
	go q.run()
}

func (q *goalQueue) run() {
	var lastFetch time.Time

	for key := range q.keys {
		q.mu.Lock()
		item := q.items[key]
		until := q.cooldownUntil
		q.mu.Unlock()

		if item == nil {
			continue // defensive: shouldn't happen
		}

		// Honor cooldown. While inside a cooldown window we still drop the
		// goal (no fetch, no cache, nil result) per the design contract:
		// blocked goals are not re-enqueued and not cached, leaving a future
		// app session free to retry.
		if d := time.Until(until); d > 0 {
			q.debugf("queue: dropping goal %d:%d — inside cooldown for %v",
				key.MatchID, key.Minute, d.Round(time.Second))
			q.complete(key, nil, false)
			continue
		}

		// Pacing: wait out the interval since the last fetch attempt. The
		// gate applies to every fetch, not only the first, so the queue's
		// rate-limit guarantee holds even after a cache-hit-skipped item.
		if !lastFetch.IsZero() {
			if wait := q.interval - time.Since(lastFetch); wait > 0 {
				time.Sleep(wait)
			}
		}

		link, err := q.fetch(item.goal)
		lastFetch = time.Now()

		if err != nil && errors.Is(err, ErrBlocked) {
			q.mu.Lock()
			cooldown := q.nextCooldown()
			q.cooldownUntil = time.Now().Add(cooldown)
			state := queueState{CooldownUntil: q.cooldownUntil, ConsecutiveBlocks: q.consecutiveBlocks}
			q.mu.Unlock()
			if q.store != nil {
				if err := q.store.save(state); err != nil {
					q.debugf("queue: failed to persist cooldown state: %v", err)
				}
			}
			q.debugf("queue: goal %d:%d hit ErrBlocked (streak %d) — cooldown for %v",
				key.MatchID, key.Minute, state.ConsecutiveBlocks, cooldown.Round(time.Second))
			q.complete(key, nil, false)
			continue
		}
		if err != nil {
			// Transient (network) error: don't cache, don't re-enqueue. The
			// next call that produces this goal will retry naturally; logging
			// keeps the failure visible.
			q.debugf("queue: goal %d:%d fetch error: %v", key.MatchID, key.Minute, err)
			q.complete(key, nil, false)
			continue
		}

		// Successful fetch: any prior block streak is over. Reset so the next
		// ErrBlocked (if any) starts back at the base cooldown instead of
		// continuing to grow from a stale streak.
		q.mu.Lock()
		hadStreak := q.consecutiveBlocks > 0
		q.consecutiveBlocks = 0
		state := queueState{CooldownUntil: q.cooldownUntil, ConsecutiveBlocks: 0}
		q.mu.Unlock()
		if hadStreak && q.store != nil {
			if err := q.store.save(state); err != nil {
				q.debugf("queue: failed to persist reset block streak: %v", err)
			}
		}

		q.complete(key, link, true)
	}
}

// complete sends the result to every waiting reply channel and removes the
// in-flight entry. When cache is true, persists the outcome (found-link with
// the 7d TTL, or NotFoundMarker with the 1h TTL).
func (q *goalQueue) complete(key GoalLinkKey, link *GoalLink, persist bool) {
	if persist && q.cache != nil {
		if link != nil {
			_ = q.cache.Set(*link)
		} else {
			_ = q.cache.SetNotFound(key.MatchID, key.Minute)
		}
	}

	q.mu.Lock()
	item := q.items[key]
	delete(q.items, key)
	q.mu.Unlock()

	if item == nil {
		return
	}
	result := GoalResult{Key: key, Link: link}
	for _, r := range item.replies {
		select {
		case r <- result:
		default:
			q.debugf("queue: dropped result for %d:%d — reply channel full",
				key.MatchID, key.Minute)
		}
	}
}

func (q *goalQueue) debugf(format string, args ...any) {
	if q.log != nil {
		q.log(fmt.Sprintf(format, args...))
	}
}
