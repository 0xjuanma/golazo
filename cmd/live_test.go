package cmd

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/0xjuanma/golazo/internal/api"
	"github.com/0xjuanma/golazo/internal/data"
)

func TestRunLive_MockReturnsEnvelopeWithExpectedCount(t *testing.T) {
	t.Setenv(EnvOffline, "")
	t.Setenv(EnvAgent, "")

	var stdout, stderr bytes.Buffer
	code := runLive(&stdout, &stderr, cliFlags{mock: true, timeout: 5 * time.Second})

	if code != ExitOK {
		t.Fatalf("exit code = %d, want %d. stderr=%s", code, ExitOK, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Errorf("stderr not empty: %s", stderr.String())
	}

	var env struct {
		Status string      `json:"status"`
		Count  int         `json:"count"`
		Data   []api.Match `json:"data"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal stdout: %v\nraw: %s", err, stdout.String())
	}
	if env.Status != "ok" {
		t.Errorf("status = %q", env.Status)
	}
	want := len(data.MockLiveMatches())
	if env.Count != want {
		t.Errorf("count = %d, want %d", env.Count, want)
	}
	if len(env.Data) != want {
		t.Errorf("data length = %d, want %d", len(env.Data), want)
	}
}

func TestRunLive_OfflineEnvWritesErrorToStderr(t *testing.T) {
	t.Setenv(EnvOffline, "1")

	var stdout, stderr bytes.Buffer
	code := runLive(&stdout, &stderr, cliFlags{mock: false, timeout: time.Second})

	if code != ExitOffline {
		t.Errorf("exit code = %d, want %d", code, ExitOffline)
	}
	if stdout.Len() != 0 {
		t.Errorf("stdout should be empty on error, got: %s", stdout.String())
	}
	var env errEnvelope
	if err := json.Unmarshal(stderr.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal stderr: %v\nraw: %s", err, stderr.String())
	}
	if env.Code != ErrCodeOffline {
		t.Errorf("error code = %q, want %q", env.Code, ErrCodeOffline)
	}
}

func TestRunLive_MockWorksUnderOffline(t *testing.T) {
	t.Setenv(EnvOffline, "1")

	var stdout, stderr bytes.Buffer
	code := runLive(&stdout, &stderr, cliFlags{mock: true, timeout: time.Second})

	if code != ExitOK {
		t.Errorf("exit code = %d, want %d (offline+mock should succeed)", code, ExitOK)
	}
	if stdout.Len() == 0 {
		t.Errorf("stdout empty under offline+mock")
	}
}

func TestRunLive_AgentModeForcesCompact(t *testing.T) {
	t.Setenv(EnvAgent, "1")
	defer func() { Pretty = false }()

	var stdout, stderr bytes.Buffer
	_ = runLive(&stdout, &stderr, cliFlags{mock: true, pretty: true, timeout: time.Second})

	if bytes.Contains(stdout.Bytes(), []byte("\n  ")) {
		t.Errorf("agent mode should force compact, got indented: %s", stdout.String())
	}
}

func TestRunLive_TimeoutNotSwallowed(t *testing.T) {
	t.Setenv(EnvOffline, "")
	t.Setenv(EnvAgent, "")

	// 1ns timeout — the headless client's context is deadline-exceeded
	// immediately. The underlying LiveAndUpcoming aggregator may return an
	// empty slice with nil error; the CLI must still report timeout.
	var stdout, stderr bytes.Buffer
	code := runLive(&stdout, &stderr, cliFlags{mock: false, timeout: 1})
	if code != ExitTimeout {
		t.Errorf("exit = %d, want %d (stderr=%s, stdout=%s)", code, ExitTimeout, stderr.String(), stdout.String())
	}
	if stdout.Len() != 0 {
		t.Errorf("stdout should be empty on timeout, got: %s", stdout.String())
	}
}
