package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/0xjuanma/golazo/internal/api"
	"github.com/0xjuanma/golazo/internal/data"
)

func TestRunFinished_MockReturnsExpectedCount(t *testing.T) {
	t.Setenv(EnvOffline, "")
	t.Setenv(EnvAgent, "")

	var stdout, stderr bytes.Buffer
	code := runFinished(&stdout, &stderr, finishedFlags{cliFlags: cliFlags{mock: true, timeout: time.Second}, days: 1})

	if code != ExitOK {
		t.Fatalf("exit code = %d, want %d. stderr=%s", code, ExitOK, stderr.String())
	}
	var env struct {
		Status string      `json:"status"`
		Count  int         `json:"count"`
		Data   []api.Match `json:"data"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal: %v\nraw: %s", err, stdout.String())
	}
	if env.Count != len(data.MockFinishedMatches()) {
		t.Errorf("count = %d, want %d", env.Count, len(data.MockFinishedMatches()))
	}
}

func TestRunFinished_InvalidDays(t *testing.T) {
	cases := []int{0, -1, 8, 100}
	for _, days := range cases {
		t.Run("", func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := runFinished(&stdout, &stderr, finishedFlags{cliFlags: cliFlags{mock: true, timeout: time.Second}, days: days})
			if code != ExitInvalidArgs {
				t.Errorf("days=%d: exit = %d, want %d", days, code, ExitInvalidArgs)
			}
			if stdout.Len() != 0 {
				t.Errorf("stdout should be empty on invalid args, got: %s", stdout.String())
			}
			var env errEnvelope
			if err := json.Unmarshal(stderr.Bytes(), &env); err != nil {
				t.Fatalf("unmarshal stderr: %v", err)
			}
			if env.Code != ErrCodeInvalidArgs {
				t.Errorf("code = %q, want %q", env.Code, ErrCodeInvalidArgs)
			}
		})
	}
}

func TestCollectFinished_DedupsByID(t *testing.T) {
	now := time.Date(2026, 6, 12, 12, 0, 0, 0, time.UTC)
	calls := 0
	fetch := func(ctx context.Context, date time.Time, tabs []string) ([]api.Match, error) {
		calls++
		// Return the same finished match each day; expect dedup to keep it once.
		return []api.Match{
			{ID: 100, Status: api.MatchStatusFinished},
			{ID: 101, Status: api.MatchStatusFinished},
			{ID: 200, Status: api.MatchStatusNotStarted}, // must be filtered
		}, nil
	}

	matches, failed, err := collectFinished(context.Background(), fetch, now, 3)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if len(failed) != 0 {
		t.Errorf("failedDates = %v, want empty", failed)
	}
	if calls != 3 {
		t.Errorf("fetch called %d times, want 3", calls)
	}
	if len(matches) != 2 {
		t.Errorf("returned %d matches, want 2 (dedup keeps 100,101; drops 200)", len(matches))
	}
	for _, m := range matches {
		if m.Status != api.MatchStatusFinished {
			t.Errorf("non-finished match leaked: %+v", m)
		}
	}
}

func TestCollectFinished_PartialFailureFlagsDegraded(t *testing.T) {
	now := time.Date(2026, 6, 12, 12, 0, 0, 0, time.UTC)
	callCount := 0
	fetch := func(ctx context.Context, date time.Time, tabs []string) ([]api.Match, error) {
		callCount++
		if callCount == 2 {
			return nil, errors.New("upstream blew up")
		}
		return []api.Match{
			{ID: callCount, Status: api.MatchStatusFinished},
		}, nil
	}

	matches, failed, err := collectFinished(context.Background(), fetch, now, 3)
	if err != nil {
		t.Fatalf("partial-success should not return err, got %v", err)
	}
	if len(failed) != 1 {
		t.Errorf("failedDates = %v, want 1 entry", failed)
	}
	if len(matches) != 2 {
		t.Errorf("matches len = %d, want 2", len(matches))
	}
}

func TestCollectFinished_AllFailureReturnsError(t *testing.T) {
	now := time.Date(2026, 6, 12, 12, 0, 0, 0, time.UTC)
	fetch := func(ctx context.Context, date time.Time, tabs []string) ([]api.Match, error) {
		return nil, errors.New("nope")
	}

	matches, failed, err := collectFinished(context.Background(), fetch, now, 2)
	if err == nil {
		t.Fatalf("expected err when all days fail")
	}
	if matches != nil {
		t.Errorf("matches not nil on total failure: %v", matches)
	}
	if len(failed) != 2 {
		t.Errorf("failedDates = %v, want 2", failed)
	}
}

func TestCollectFinished_TodayUsesFixturesAndResults(t *testing.T) {
	now := time.Date(2026, 6, 12, 12, 0, 0, 0, time.UTC)
	gotTabs := [][]string{}
	fetch := func(ctx context.Context, date time.Time, tabs []string) ([]api.Match, error) {
		gotTabs = append(gotTabs, tabs)
		return nil, nil
	}

	_, _, _ = collectFinished(context.Background(), fetch, now, 2)
	if len(gotTabs) != 2 {
		t.Fatalf("expected 2 fetches, got %d", len(gotTabs))
	}
	// Day 0 (today) → fixtures+results
	if len(gotTabs[0]) != 2 {
		t.Errorf("today tabs = %v, want fixtures+results", gotTabs[0])
	}
	// Day 1 → results only
	if len(gotTabs[1]) != 1 || gotTabs[1][0] != "results" {
		t.Errorf("past-day tabs = %v, want [results]", gotTabs[1])
	}
}
