package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/spf13/cobra"
)

// TestExecute_UnknownCommandEnveloped verifies that cobra's "unknown command"
// errors reach the JSON envelope rather than printing plain text. This is the
// contract Issue 3 fixes: agents that parse stderr as JSON shouldn't crash on
// a typo'd subcommand.
//
// We can't exec the binary here, but we can exercise the same envelope path
// the Execute() wrapper uses.
func TestExecute_UnknownCommandEnvelopedAsInvalidArgs(t *testing.T) {
	// Build a throwaway cobra root with one subcommand so we can synthesize
	// the same error type that the real rootCmd would return.
	root := &cobra.Command{Use: "golazo", SilenceErrors: true, SilenceUsage: true}
	root.AddCommand(&cobra.Command{Use: "live", Run: func(cmd *cobra.Command, args []string) {}})
	root.SetArgs([]string{"bogus"})

	var cobraOut bytes.Buffer
	root.SetOut(&cobraOut)
	root.SetErr(&cobraOut)

	err := root.Execute()
	if err == nil {
		t.Fatalf("expected error from unknown subcommand")
	}

	// This is what Execute() does in production.
	var envBuf bytes.Buffer
	exit := WriteError(&envBuf, ErrCodeInvalidArgs, err)
	if exit != ExitInvalidArgs {
		t.Errorf("exit = %d, want %d", exit, ExitInvalidArgs)
	}
	var env errEnvelope
	if jerr := json.Unmarshal(envBuf.Bytes(), &env); jerr != nil {
		t.Fatalf("envelope not valid JSON: %v\nraw: %s", jerr, envBuf.String())
	}
	if env.Code != ErrCodeInvalidArgs {
		t.Errorf("code = %q, want %q", env.Code, ErrCodeInvalidArgs)
	}
	if env.Message == "" {
		t.Errorf("message empty; agents need a useful description")
	}
}

func TestExecute_FlagErrorEnveloped(t *testing.T) {
	// Sanity: a synthetic flag-parse error follows the same path.
	err := errors.New("unknown flag: --foo")
	var envBuf bytes.Buffer
	exit := WriteError(&envBuf, ErrCodeInvalidArgs, err)
	if exit != ExitInvalidArgs {
		t.Errorf("exit = %d, want %d", exit, ExitInvalidArgs)
	}
	if !bytes.Contains(envBuf.Bytes(), []byte("unknown flag")) {
		t.Errorf("envelope missing original error: %s", envBuf.String())
	}
}
