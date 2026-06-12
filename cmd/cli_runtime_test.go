package cmd

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestNewHeadlessClient_OfflineReturnsError(t *testing.T) {
	t.Setenv(EnvOffline, "1")
	_, _, cancel, err := newHeadlessClient(runtimeOpts{mock: false, timeout: time.Second})
	defer cancel()
	if !errors.Is(err, ErrOffline) {
		t.Errorf("err = %v, want ErrOffline", err)
	}
}

func TestNewHeadlessClient_OfflineButMockSucceeds(t *testing.T) {
	t.Setenv(EnvOffline, "1")
	client, ctx, cancel, err := newHeadlessClient(runtimeOpts{mock: true, timeout: time.Second})
	defer cancel()
	if err != nil {
		t.Fatalf("mock under offline should succeed, got err=%v", err)
	}
	if client == nil {
		t.Errorf("client is nil")
	}
	if ctx == nil {
		t.Errorf("ctx is nil")
	}
}

func TestNewHeadlessClient_DefaultsToTimeout(t *testing.T) {
	t.Setenv(EnvOffline, "")
	_, ctx, cancel, err := newHeadlessClient(runtimeOpts{timeout: 0})
	defer cancel()
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatalf("ctx has no deadline")
	}
	// Default is 15s; allow slack.
	if d := time.Until(deadline); d < 10*time.Second || d > 20*time.Second {
		t.Errorf("default timeout deadline = %v from now, want ~15s", d)
	}
}

func TestNewHeadlessClient_RespectsCustomTimeout(t *testing.T) {
	t.Setenv(EnvOffline, "")
	_, ctx, cancel, err := newHeadlessClient(runtimeOpts{timeout: 5 * time.Millisecond})
	defer cancel()
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	// Wait for the context to expire and verify isTimeout reports it.
	<-ctx.Done()
	if !isTimeout(ctx) {
		t.Errorf("isTimeout returned false on deadline-exceeded ctx (err=%v)", ctx.Err())
	}
}

func TestIsTimeout_CancelledIsNotTimeout(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if isTimeout(ctx) {
		t.Errorf("cancelled ctx classified as timeout")
	}
}

func TestAgentMode(t *testing.T) {
	t.Setenv(EnvAgent, "1")
	if !agentMode() {
		t.Errorf("agentMode = false with GOLAZO_AGENT=1")
	}
	t.Setenv(EnvAgent, "")
	if agentMode() {
		t.Errorf("agentMode = true with empty env")
	}
}

func TestNewStderrLogger_NoopWhenDisabled(t *testing.T) {
	t.Setenv(EnvAgent, "")
	logger := newStderrLogger(false)
	if logger == nil {
		t.Fatalf("logger nil")
	}
	// We can't easily intercept stderr; assert that the handler is non-nil
	// and that the logger does not panic. The hard contract (stderr only)
	// is enforced by code review of the constructor.
	logger.Debug("noop")
}
