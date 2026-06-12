package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/0xjuanma/golazo/internal/api"
)

func TestWriteJSON_EmptySlice(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteJSON(&buf, []api.Match(nil)); err != nil {
		t.Fatalf("WriteJSON returned error: %v", err)
	}

	var env struct {
		Status string       `json:"status"`
		Count  int          `json:"count"`
		Data   []api.Match  `json:"data"`
	}
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal: %v\nraw: %s", err, buf.String())
	}
	if env.Status != "ok" {
		t.Errorf("status = %q, want ok", env.Status)
	}
	if env.Count != 0 {
		t.Errorf("count = %d, want 0", env.Count)
	}
	if env.Data == nil {
		t.Errorf("data is null; want empty array")
	}
	if len(env.Data) != 0 {
		t.Errorf("data length = %d, want 0", len(env.Data))
	}
	// Verify raw output does not contain "null"
	if bytes.Contains(buf.Bytes(), []byte("null")) {
		t.Errorf("output contains null token: %s", buf.String())
	}
}

func TestWriteJSON_SliceWithItems(t *testing.T) {
	matches := []api.Match{
		{ID: 1, HomeTeam: api.Team{Name: "A"}, AwayTeam: api.Team{Name: "B"}},
		{ID: 2, HomeTeam: api.Team{Name: "C"}, AwayTeam: api.Team{Name: "D"}},
	}
	var buf bytes.Buffer
	if err := WriteJSON(&buf, matches); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}
	var env okEnvelope
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if env.Count != 2 {
		t.Errorf("count = %d, want 2", env.Count)
	}
	if env.Status != "ok" {
		t.Errorf("status = %q", env.Status)
	}
	if env.Degraded {
		t.Errorf("degraded should be false")
	}
}

func TestWriteDegraded(t *testing.T) {
	matches := []api.Match{{ID: 1}}
	var buf bytes.Buffer
	if err := WriteDegraded(&buf, matches, []string{"2026-06-10", "2026-06-11"}); err != nil {
		t.Fatalf("WriteDegraded: %v", err)
	}
	var env struct {
		Status      string   `json:"status"`
		Degraded    bool     `json:"degraded"`
		FailedDates []string `json:"failed_dates"`
		Count       int      `json:"count"`
	}
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !env.Degraded {
		t.Errorf("degraded should be true")
	}
	if len(env.FailedDates) != 2 {
		t.Errorf("failed_dates length = %d, want 2", len(env.FailedDates))
	}
	if env.Count != 1 {
		t.Errorf("count = %d, want 1", env.Count)
	}
}

func TestWriteError_WritesAndReturnsExitCode(t *testing.T) {
	cases := []struct {
		code     ErrorCode
		exitCode int
	}{
		{ErrCodeInvalidArgs, ExitInvalidArgs},
		{ErrCodeNotFound, ExitNotFound},
		{ErrCodeTimeout, ExitTimeout},
		{ErrCodeOffline, ExitOffline},
		{ErrCodeUpstreamError, ExitUpstream},
	}
	for _, tc := range cases {
		t.Run(string(tc.code), func(t *testing.T) {
			var buf bytes.Buffer
			got := WriteError(&buf, tc.code, errors.New("boom"))
			if got != tc.exitCode {
				t.Errorf("exit code = %d, want %d", got, tc.exitCode)
			}
			var env errEnvelope
			if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if env.Status != "error" {
				t.Errorf("status = %q", env.Status)
			}
			if env.Code != tc.code {
				t.Errorf("code = %q, want %q", env.Code, tc.code)
			}
			if env.Message != "boom" {
				t.Errorf("message = %q", env.Message)
			}
		})
	}
}

func TestExitCodeFor_UnknownDefaults(t *testing.T) {
	if got := ExitCodeFor(ErrorCode("nope")); got != ExitUpstream {
		t.Errorf("unknown code exit = %d, want %d", got, ExitUpstream)
	}
}

func TestSortMatches_TimeThenID(t *testing.T) {
	t1 := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 6, 10, 18, 0, 0, 0, time.UTC)
	matches := []api.Match{
		{ID: 9, MatchTime: &t2},
		{ID: 5, MatchTime: nil},
		{ID: 3, MatchTime: &t1},
		{ID: 7, MatchTime: &t1},
	}
	SortMatches(matches)
	wantIDs := []int{3, 7, 9, 5}
	for i, want := range wantIDs {
		if matches[i].ID != want {
			t.Errorf("position %d: got id=%d, want %d (full=%v)", i, matches[i].ID, want, idsOf(matches))
		}
	}
}

func TestClassifyClientError(t *testing.T) {
	if got := ClassifyClientError(errors.New("x"), true); got != ErrCodeTimeout {
		t.Errorf("timedOut=true → %q, want %q", got, ErrCodeTimeout)
	}
	if got := ClassifyClientError(errors.New("x"), false); got != ErrCodeUpstreamError {
		t.Errorf("timedOut=false → %q, want %q", got, ErrCodeUpstreamError)
	}
}

func TestPrettyToggle(t *testing.T) {
	prev := Pretty
	defer func() { Pretty = prev }()

	Pretty = false
	var compact bytes.Buffer
	_ = WriteJSON(&compact, []api.Match{{ID: 1}})

	Pretty = true
	var pretty bytes.Buffer
	_ = WriteJSON(&pretty, []api.Match{{ID: 1}})

	if !bytes.Contains(pretty.Bytes(), []byte("\n  ")) {
		t.Errorf("pretty output not indented: %s", pretty.String())
	}
	if bytes.Contains(compact.Bytes(), []byte("\n  ")) {
		t.Errorf("compact output was indented: %s", compact.String())
	}
}

func idsOf(matches []api.Match) []int {
	out := make([]int, len(matches))
	for i, m := range matches {
		out[i] = m.ID
	}
	return out
}
