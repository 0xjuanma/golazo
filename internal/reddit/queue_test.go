package reddit

import (
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// newTestCache returns a GoalLinkCache backed by a temp file so tests don't
// touch ~/.golazo. Mirrors the construction path used by NewGoalLinkCache.
func newTestCache(t *testing.T) *GoalLinkCache {
	t.Helper()
	return &GoalLinkCache{
		links:    make(map[string]GoalLink),
		filePath: filepath.Join(t.TempDir(), "goal_links.json"),
	}
}

// recordingFetchHook is the queue's fetch hook for tests: records timestamps
// of every call, supports per-call result programming, and is safe under the
// queue's single-worker contract (no concurrency expected, mutex just guards
// against accidents).
type recordingFetchHook struct {
	mu        sync.Mutex
	calls     []time.Time
	results   []*GoalLink
	errors    []error
	callCount int32
}

func (h *recordingFetchHook) fetch(_ GoalInfo) (*GoalLink, error) {
	idx := atomic.AddInt32(&h.callCount, 1) - 1
	h.mu.Lock()
	h.calls = append(h.calls, time.Now())
	h.mu.Unlock()

	if int(idx) < len(h.errors) && h.errors[idx] != nil {
		return nil, h.errors[idx]
	}
	if int(idx) < len(h.results) {
		return h.results[idx], nil
	}
	return nil, nil
}

func (h *recordingFetchHook) callTimes() []time.Time {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([]time.Time, len(h.calls))
	copy(out, h.calls)
	return out
}

// TestQueueIntervalPacing verifies the worker's pacing gate inside run():
// successive fetches must be separated by at least the configured interval.
// This is where the QueueInterval guarantee actually lives.
func TestQueueIntervalPacing(t *testing.T) {
	const interval = 50 * time.Millisecond

	hook := &recordingFetchHook{
		results: []*GoalLink{
			{MatchID: 1, Minute: 1, URL: "https://example.com/1"},
			{MatchID: 1, Minute: 2, URL: "https://example.com/2"},
			{MatchID: 1, Minute: 3, URL: "https://example.com/3"},
		},
	}
	q := newGoalQueue(hook.fetch, newTestCache(t), nil, interval, time.Minute, nil)

	replies := make(chan GoalResult, 3)
	for i := 1; i <= 3; i++ {
		q.Enqueue(GoalInfo{MatchID: 1, Minute: i}, replies)
	}

	for i := 0; i < 3; i++ {
		select {
		case <-replies:
		case <-time.After(time.Second):
			t.Fatalf("timed out waiting for reply %d", i)
		}
	}

	times := hook.callTimes()
	if len(times) != 3 {
		t.Fatalf("expected 3 fetch calls, got %d", len(times))
	}
	for i := 1; i < len(times); i++ {
		gap := times[i].Sub(times[i-1])
		if gap+5*time.Millisecond < interval {
			t.Errorf("call %d -> %d gap %v < interval %v", i-1, i, gap, interval)
		}
	}
}

// TestQueueCooldownOnBlocked verifies that an ErrBlocked response sets the
// cooldown window such that subsequent goals are dropped (no fetch attempt)
// while inside the window.
func TestQueueCooldownOnBlocked(t *testing.T) {
	hook := &recordingFetchHook{
		errors: []error{ErrBlocked}, // first (and only) attempt blocked
	}
	q := newGoalQueue(hook.fetch, newTestCache(t), nil, time.Millisecond, time.Hour, nil)

	replies := make(chan GoalResult, 2)
	q.Enqueue(GoalInfo{MatchID: 1, Minute: 1}, replies)
	q.Enqueue(GoalInfo{MatchID: 1, Minute: 2}, replies)

	got := make([]GoalResult, 0, 2)
	for i := 0; i < 2; i++ {
		select {
		case r := <-replies:
			got = append(got, r)
		case <-time.After(time.Second):
			t.Fatalf("timed out waiting for reply %d", i)
		}
	}

	if calls := atomic.LoadInt32(&hook.callCount); calls != 1 {
		t.Errorf("expected exactly 1 fetch attempt (blocked, then cooldown), got %d", calls)
	}
	for _, r := range got {
		if r.Link != nil {
			t.Errorf("expected nil Link during cooldown, got %+v for %+v", r.Link, r.Key)
		}
	}
}

// TestQueueDropsOnBlocked verifies that a blocked goal is not cached
// (neither as a found link nor as a NotFoundMarker) — leaving the next app
// session free to retry.
func TestQueueDropsOnBlocked(t *testing.T) {
	hook := &recordingFetchHook{errors: []error{ErrBlocked}}
	cache := newTestCache(t)
	q := newGoalQueue(hook.fetch, cache, nil, time.Millisecond, time.Hour, nil)

	replies := make(chan GoalResult, 1)
	key := GoalLinkKey{MatchID: 99, Minute: 33}
	q.Enqueue(GoalInfo{MatchID: 99, Minute: 33}, replies)

	select {
	case <-replies:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for blocked reply")
	}

	if link := cache.Get(key); link != nil {
		t.Errorf("expected no cache entry for blocked goal, got %+v", link)
	}
}

// TestQueueDedupesInFlight verifies that enqueuing the same key twice while
// the first is in flight results in a single fetch, and both reply channels
// receive the result.
func TestQueueDedupesInFlight(t *testing.T) {
	// Block the fetch hook on a signal channel so the second Enqueue lands
	// while the first is still in flight.
	gate := make(chan struct{})
	hook := &slowFetchHook{
		gate: gate,
		link: &GoalLink{MatchID: 5, Minute: 10, URL: "https://example.com/dedup"},
	}
	q := newGoalQueue(hook.fetch, newTestCache(t), nil, time.Millisecond, time.Hour, nil)

	r1 := make(chan GoalResult, 1)
	r2 := make(chan GoalResult, 1)
	g := GoalInfo{MatchID: 5, Minute: 10}
	q.Enqueue(g, r1)
	// Tiny pause: ensure the worker has picked up the first item and is
	// blocked inside the fetch hook before the second Enqueue arrives.
	time.Sleep(20 * time.Millisecond)
	q.Enqueue(g, r2)

	close(gate) // release the in-flight fetch

	for _, ch := range []chan GoalResult{r1, r2} {
		select {
		case r := <-ch:
			if r.Link == nil || r.Link.URL != "https://example.com/dedup" {
				t.Errorf("dedup reply missing expected link, got %+v", r.Link)
			}
		case <-time.After(time.Second):
			t.Fatal("dedup reply channel never received result")
		}
	}

	if got := atomic.LoadInt32(&hook.calls); got != 1 {
		t.Errorf("expected 1 fetch for deduped key, got %d", got)
	}
}

type slowFetchHook struct {
	gate  chan struct{}
	link  *GoalLink
	calls int32
}

func (h *slowFetchHook) fetch(_ GoalInfo) (*GoalLink, error) {
	atomic.AddInt32(&h.calls, 1)
	<-h.gate
	return h.link, nil
}

// TestQueueLazyStart verifies the worker goroutine is not started until the
// first Enqueue. Without lazy start, every Client construction would leak an
// idle goroutine. We assert this behaviorally: the fetch hook must not be
// invoked by mere construction, only by an actual Enqueue.
func TestQueueLazyStart(t *testing.T) {
	hook := &recordingFetchHook{
		results: []*GoalLink{{MatchID: 1, Minute: 1, URL: "https://example.com/lazy"}},
	}
	q := newGoalQueue(hook.fetch, newTestCache(t), nil, time.Millisecond, time.Hour, nil)

	// Give a hypothetical eager worker plenty of scheduler ticks to run.
	time.Sleep(50 * time.Millisecond)
	if got := atomic.LoadInt32(&hook.callCount); got != 0 {
		t.Fatalf("fetch hook called %d times before any Enqueue — worker started eagerly", got)
	}

	replies := make(chan GoalResult, 1)
	q.Enqueue(GoalInfo{MatchID: 1, Minute: 1}, replies)
	select {
	case <-replies:
	case <-time.After(time.Second):
		t.Fatal("worker did not start after first Enqueue")
	}
	if got := atomic.LoadInt32(&hook.callCount); got != 1 {
		t.Errorf("expected 1 fetch after Enqueue, got %d", got)
	}
}

// TestQueueCooldownPersistsAcrossRestart verifies that a cooldown entered by
// one goalQueue is visible to a second goalQueue constructed later against
// the same store — simulating a process restart mid-block.
func TestQueueCooldownPersistsAcrossRestart(t *testing.T) {
	store := &queueStateStore{filePath: filepath.Join(t.TempDir(), "reddit_queue_state.json")}

	hookA := &recordingFetchHook{errors: []error{ErrBlocked}}
	qA := newGoalQueue(hookA.fetch, newTestCache(t), nil, time.Millisecond, time.Hour, store)

	repliesA := make(chan GoalResult, 1)
	qA.Enqueue(GoalInfo{MatchID: 1, Minute: 1}, repliesA)
	select {
	case <-repliesA:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for blocked reply from queue A")
	}

	persisted := store.load()
	if persisted.CooldownUntil.IsZero() || !persisted.CooldownUntil.After(time.Now()) {
		t.Fatalf("expected persisted CooldownUntil in the future, got %v", persisted.CooldownUntil)
	}

	// Queue B simulates a fresh process reading the same store: it must
	// inherit the cooldown and drop the goal without ever calling fetch.
	hookB := &recordingFetchHook{
		results: []*GoalLink{{MatchID: 2, Minute: 2, URL: "https://example.com/should-not-be-fetched"}},
	}
	qB := newGoalQueue(hookB.fetch, newTestCache(t), nil, time.Millisecond, time.Hour, store)

	repliesB := make(chan GoalResult, 1)
	qB.Enqueue(GoalInfo{MatchID: 2, Minute: 2}, repliesB)
	select {
	case r := <-repliesB:
		if r.Link != nil {
			t.Errorf("expected nil Link — goal should be dropped under inherited cooldown, got %+v", r.Link)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for reply from queue B")
	}

	if calls := atomic.LoadInt32(&hookB.callCount); calls != 0 {
		t.Errorf("expected 0 fetch attempts on queue B (inherited cooldown), got %d", calls)
	}
}

// waitOutCooldown blocks until q's current cooldown has elapsed, so the next
// Enqueue reaches the fetch hook instead of being dropped by run()'s
// cooldown check.
func waitOutCooldown(t *testing.T, q *goalQueue) {
	t.Helper()
	q.mu.Lock()
	until := q.cooldownUntil
	q.mu.Unlock()
	if wait := time.Until(until); wait > 0 {
		time.Sleep(wait + 5*time.Millisecond)
	}
}

// TestQueueBackoffGrowsOnRepeatedBlocks verifies that consecutive ErrBlocked
// responses grow the cooldown duration exponentially (with jitter) instead
// of reusing the same flat cooldown every time.
func TestQueueBackoffGrowsOnRepeatedBlocks(t *testing.T) {
	const base = 20 * time.Millisecond
	hook := &recordingFetchHook{errors: []error{ErrBlocked, ErrBlocked, ErrBlocked}}
	q := newGoalQueue(hook.fetch, newTestCache(t), nil, time.Millisecond, base, nil)

	var deltas []time.Duration
	for i := range 3 {
		replies := make(chan GoalResult, 1)
		q.Enqueue(GoalInfo{MatchID: 1, Minute: i + 1}, replies)
		select {
		case <-replies:
		case <-time.After(time.Second):
			t.Fatalf("timed out waiting for block %d", i+1)
		}

		q.mu.Lock()
		until := q.cooldownUntil
		streak := q.consecutiveBlocks
		q.mu.Unlock()
		if streak != i+1 {
			t.Fatalf("block %d: expected consecutiveBlocks=%d, got %d", i+1, i+1, streak)
		}
		deltas = append(deltas, time.Until(until))

		waitOutCooldown(t, q)
	}

	wantMults := []float64{1, 2, 4}
	for i, delta := range deltas {
		want := float64(base) * wantMults[i]
		lo, hi := want*0.7, want*1.3 // jitter band (+/-20%) with slop for scheduling
		if got := float64(delta); got < lo || got > hi {
			t.Errorf("block %d cooldown delta = %v, want within [%v, %v] (base*%v)",
				i+1, delta, time.Duration(lo), time.Duration(hi), wantMults[i])
		}
	}
}

// TestQueueBackoffResetsOnSuccess verifies that a successful fetch (no error,
// found or not-found) resets the consecutive-block streak, so a later
// ErrBlocked starts back at the base cooldown instead of continuing to grow
// from a stale streak.
func TestQueueBackoffResetsOnSuccess(t *testing.T) {
	const base = 20 * time.Millisecond
	hook := &recordingFetchHook{
		errors:  []error{ErrBlocked, ErrBlocked, nil, ErrBlocked},
		results: []*GoalLink{nil, nil, {MatchID: 9, Minute: 9, URL: "https://example.com/ok"}, nil},
	}
	q := newGoalQueue(hook.fetch, newTestCache(t), nil, time.Millisecond, base, nil)

	enqueueAndWait := func(minute int) {
		replies := make(chan GoalResult, 1)
		q.Enqueue(GoalInfo{MatchID: 1, Minute: minute}, replies)
		select {
		case <-replies:
		case <-time.After(time.Second):
			t.Fatalf("timed out waiting for goal %d", minute)
		}
	}

	enqueueAndWait(1) // ErrBlocked, streak=1
	waitOutCooldown(t, q)
	enqueueAndWait(2) // ErrBlocked, streak=2
	waitOutCooldown(t, q)
	enqueueAndWait(3) // success — streak should reset to 0

	q.mu.Lock()
	streak := q.consecutiveBlocks
	q.mu.Unlock()
	if streak != 0 {
		t.Fatalf("expected consecutiveBlocks reset to 0 after success, got %d", streak)
	}

	enqueueAndWait(4) // ErrBlocked again — should be back at base, not 4x base

	q.mu.Lock()
	until := q.cooldownUntil
	streakAfter := q.consecutiveBlocks
	q.mu.Unlock()
	if streakAfter != 1 {
		t.Fatalf("expected consecutiveBlocks=1 after post-reset block, got %d", streakAfter)
	}
	delta := time.Until(until)
	want := float64(base)
	lo, hi := want*0.7, want*1.3
	if got := float64(delta); got < lo || got > hi {
		t.Errorf("post-reset cooldown delta = %v, want within [%v, %v] (base)",
			delta, time.Duration(lo), time.Duration(hi))
	}
}
