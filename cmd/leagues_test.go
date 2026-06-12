package cmd

import (
	"bytes"
	"encoding/json"
	"slices"
	"sort"
	"testing"
	"time"

	"github.com/0xjuanma/golazo/internal/api"
	"github.com/0xjuanma/golazo/internal/data"
)

func TestResolveLeagues_DefaultsToActive(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_CONFIG_HOME", tmp)

	got := resolveLeagues(false)

	gotIDs := make([]int, len(got))
	for i, l := range got {
		gotIDs[i] = l.ID
	}

	wantIDs := append([]int(nil), data.DefaultLeagueIDs...)
	sort.Ints(wantIDs)

	if !slices.Equal(gotIDs, wantIDs) {
		t.Errorf("got IDs=%v, want %v", gotIDs, wantIDs)
	}
}

func TestResolveLeagues_AllListsEverySupportedLeague(t *testing.T) {
	got := resolveLeagues(true)
	if len(got) != len(data.AllLeagueIDs()) {
		t.Errorf("len = %d, want %d", len(got), len(data.AllLeagueIDs()))
	}
	// Verify sorted by ID for determinism.
	for i := 1; i < len(got); i++ {
		if got[i-1].ID > got[i].ID {
			t.Errorf("unsorted at %d: %d > %d", i, got[i-1].ID, got[i].ID)
		}
	}
}

func TestResolveLeagues_EnrichesNameAndCountry(t *testing.T) {
	got := resolveLeagues(true)
	// Premier League (47) must be present with name+country.
	var pl *api.League
	for i := range got {
		if got[i].ID == 47 {
			pl = &got[i]
			break
		}
	}
	if pl == nil {
		t.Fatalf("Premier League (47) not in result")
	}
	if pl.Name == "" || pl.Country == "" {
		t.Errorf("missing enrichment: %+v", *pl)
	}
}

func TestRunLeagues_EmitsEnvelope(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_CONFIG_HOME", tmp)

	var stdout, stderr bytes.Buffer
	code := runLeagues(&stdout, &stderr, leaguesFlags{cliFlags: cliFlags{timeout: time.Second}})
	if code != ExitOK {
		t.Fatalf("exit = %d, stderr=%s", code, stderr.String())
	}
	var env struct {
		Status string       `json:"status"`
		Count  int          `json:"count"`
		Data   []api.League `json:"data"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal: %v\nraw: %s", err, stdout.String())
	}
	if env.Status != "ok" {
		t.Errorf("status = %q", env.Status)
	}
	if env.Count != len(data.DefaultLeagueIDs) {
		t.Errorf("count = %d, want %d", env.Count, len(data.DefaultLeagueIDs))
	}
}
