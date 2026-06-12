package cmd

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/spf13/pflag"
)

func TestRunCapabilities_EmitsContract(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runCapabilities(&stdout, &stderr, cliFlags{})
	if code != ExitOK {
		t.Fatalf("exit = %d, stderr=%s", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Errorf("stderr should be empty, got: %s", stderr.String())
	}

	var env struct {
		Status string         `json:"status"`
		Count  int            `json:"count"`
		Data   []capabilities `json:"data"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal: %v\nraw: %s", err, stdout.String())
	}
	if env.Status != "ok" {
		t.Errorf("status = %q", env.Status)
	}
	if len(env.Data) != 1 {
		t.Fatalf("expected 1 capabilities entry, got %d", len(env.Data))
	}
	caps := env.Data[0]
	if caps.SchemaVersion != CapabilitiesSchemaVersion {
		t.Errorf("schema_version = %q, want %q", caps.SchemaVersion, CapabilitiesSchemaVersion)
	}
	if caps.Tool != "golazo" {
		t.Errorf("tool = %q, want golazo", caps.Tool)
	}
}

func TestCapabilities_EnumeratesAllSubcommands(t *testing.T) {
	caps := buildCapabilities()
	want := map[string]bool{
		"live":         false,
		"finished":     false,
		"match":        false,
		"leagues":      false,
		"capabilities": false,
	}
	for _, cmd := range caps.Commands {
		if _, ok := want[cmd.Name]; ok {
			want[cmd.Name] = true
		}
	}
	for name, found := range want {
		if !found {
			t.Errorf("subcommand %q missing from capabilities payload", name)
		}
	}
}

func TestCapabilities_ErrorCodesMatchExitCodes(t *testing.T) {
	caps := buildCapabilities()
	// The error_codes map must be consistent with cli_output.go's ExitCodeFor.
	for code, exit := range caps.ErrorCodes {
		got := ExitCodeFor(ErrorCode(code))
		if got != exit {
			t.Errorf("error_codes[%q] = %d, but ExitCodeFor returns %d", code, exit, got)
		}
	}
}

func TestCapabilities_EveryCommandHasExample(t *testing.T) {
	caps := buildCapabilities()
	for _, cmd := range caps.Commands {
		if cmd.Example == "" {
			t.Errorf("command %q has empty example — agents rely on this", cmd.Name)
		}
		if cmd.Description == "" {
			t.Errorf("command %q has empty description", cmd.Name)
		}
	}
}

// TestCapabilities_FlagsMatchCobra protects the contract from drifting away
// from the actual cobra flag set. If you add or remove a flag on a subcommand,
// you must also update buildCapabilities() to match.
func TestCapabilities_FlagsMatchCobra(t *testing.T) {
	caps := buildCapabilities()
	capsByName := map[string]capabilityCommand{}
	for _, c := range caps.Commands {
		capsByName[c.Name] = c
	}

	for _, cobraCmd := range rootCmd.Commands() {
		entry, ok := capsByName[cobraCmd.Name()]
		if !ok {
			continue // cobra builtins (help, completion) aren't in the contract
		}

		cobraFlagNames := map[string]bool{}
		cobraCmd.Flags().VisitAll(func(f *pflag.Flag) {
			cobraFlagNames[f.Name] = true
		})
		// Strip the inherited `help` flag — cobra adds it automatically.
		delete(cobraFlagNames, "help")

		capsFlagNames := map[string]bool{}
		for _, f := range entry.Flags {
			capsFlagNames[f.Name] = true
		}

		for name := range cobraFlagNames {
			if !capsFlagNames[name] {
				t.Errorf("subcommand %q exposes --%s but capabilities contract omits it", cobraCmd.Name(), name)
			}
		}
		for name := range capsFlagNames {
			if !cobraFlagNames[name] {
				t.Errorf("capabilities contract claims subcommand %q has --%s but cobra doesn't expose it", cobraCmd.Name(), name)
			}
		}
	}
}
