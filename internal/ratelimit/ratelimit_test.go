package ratelimit

import (
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name        string
		interval    time.Duration
		wantNonNil  bool
	}{
		{"positive interval", 100 * time.Millisecond, true},
		{"zero interval", 0, true},
		{"negative interval clamped to zero", -1 * time.Second, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rl := New(tt.interval)
			if rl == nil {
				t.Fatal("New returned nil")
			}
		})
	}
}

func TestNewFromRate(t *testing.T) {
	rl := NewFromRate(60) // 60 req/min = 1 per second
	if rl == nil {
		t.Fatal("NewFromRate returned nil")
	}
	if rl.minInterval != time.Second {
		t.Errorf("expected interval %v, got %v", time.Second, rl.minInterval)
	}
}

func TestWait_enforcesMinInterval(t *testing.T) {
	rl := New(50 * time.Millisecond)

	start := time.Now()
	rl.Wait() // first call, no wait
	rl.Wait() // second call, should wait ~50ms
	elapsed := time.Since(start)

	if elapsed < 40*time.Millisecond {
		t.Errorf("expected at least ~50ms between calls, got %v", elapsed)
	}
}

func TestWait_zeroInterval_noDelay(t *testing.T) {
	rl := New(0)

	start := time.Now()
	rl.Wait()
	rl.Wait()
	rl.Wait()
	elapsed := time.Since(start)

	if elapsed > 10*time.Millisecond {
		t.Errorf("zero interval should not delay, got %v", elapsed)
	}
}
