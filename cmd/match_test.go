package cmd

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/0xjuanma/golazo/internal/api"
)

func TestRunMatch_MissingArg(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runMatch(&stdout, &stderr, cliFlags{mock: true, timeout: time.Second}, nil)
	if code != ExitInvalidArgs {
		t.Errorf("exit = %d, want %d", code, ExitInvalidArgs)
	}
	var env errEnvelope
	if err := json.Unmarshal(stderr.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal stderr: %v", err)
	}
	if env.Code != ErrCodeInvalidArgs {
		t.Errorf("code = %q, want %q", env.Code, ErrCodeInvalidArgs)
	}
}

func TestRunMatch_InvalidID(t *testing.T) {
	cases := []string{"abc", "0", "-1", ""}
	for _, arg := range cases {
		t.Run(arg, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := runMatch(&stdout, &stderr, cliFlags{mock: true, timeout: time.Second}, []string{arg})
			if code != ExitInvalidArgs {
				t.Errorf("arg=%q exit = %d, want %d", arg, code, ExitInvalidArgs)
			}
		})
	}
}

func TestRunMatch_MockUnknownIDReturnsNotFound(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runMatch(&stdout, &stderr, cliFlags{mock: true, timeout: time.Second}, []string{"99999999"})
	if code != ExitNotFound {
		t.Errorf("exit = %d, want %d", code, ExitNotFound)
	}
	var env errEnvelope
	if err := json.Unmarshal(stderr.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal stderr: %v", err)
	}
	if env.Code != ErrCodeNotFound {
		t.Errorf("code = %q, want %q", env.Code, ErrCodeNotFound)
	}
}

func TestRunMatch_MockKnownIDReturnsDetails(t *testing.T) {
	// 2001 is the first mock live match ID (see internal/data/mock_live_matches.go).
	var stdout, stderr bytes.Buffer
	code := runMatch(&stdout, &stderr, cliFlags{mock: true, timeout: time.Second}, []string{"2001"})
	if code != ExitOK {
		t.Fatalf("exit = %d, want %d. stderr=%s", code, ExitOK, stderr.String())
	}
	var env struct {
		Status string             `json:"status"`
		Count  int                `json:"count"`
		Data   []api.MatchDetails `json:"data"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal: %v\nraw: %s", err, stdout.String())
	}
	if env.Count != 1 || len(env.Data) != 1 {
		t.Fatalf("expected count=1 and one entry, got count=%d len=%d", env.Count, len(env.Data))
	}
	if env.Data[0].ID != 2001 {
		t.Errorf("data[0].ID = %d, want 2001", env.Data[0].ID)
	}
}
