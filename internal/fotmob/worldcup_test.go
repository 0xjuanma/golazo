package fotmob

import (
	"encoding/json"
	"testing"
)

func TestWCGroupLetter(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"Grp. A", "A"},
		{"Grp. B", "B"},
		{"Grp. L", "L"},
		{"Group A", "A"},
		{"Group Z", "Z"},
		{"A", "A"},
		{"  Grp. C  ", "C"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got := wcGroupLetter(tt.in)
			if got != tt.want {
				t.Errorf("wcGroupLetter(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestParseWCGroups_Empty(t *testing.T) {
	groups := parseWCGroups(wcPageResponse{})
	if groups != nil {
		t.Errorf("expected nil for empty response, got %v", groups)
	}
}

func TestParseWCGroups_EightGroups(t *testing.T) {
	resp := buildMockWCPageResponse(8)
	groups := parseWCGroups(resp)

	if len(groups) != 8 {
		t.Fatalf("len(groups) = %d, want 8", len(groups))
	}
	for i, g := range groups {
		if len(g.Teams) != 4 {
			t.Errorf("groups[%d] has %d teams, want 4", i, len(g.Teams))
		}
		if g.Letter == "" {
			t.Errorf("groups[%d].Letter is empty", i)
		}
		if g.Name == "" {
			t.Errorf("groups[%d].Name is empty", i)
		}
	}
}

func TestParseWCGroups_TwelveGroups(t *testing.T) {
	resp := buildMockWCPageResponse(12)
	groups := parseWCGroups(resp)

	if len(groups) != 12 {
		t.Fatalf("len(groups) = %d, want 12", len(groups))
	}
}

func TestParseWCBracket_Empty(t *testing.T) {
	rounds, bronze := parseWCBracket(wcPlayoff{})
	if len(rounds) != 0 {
		t.Errorf("expected 0 rounds, got %d", len(rounds))
	}
	if bronze != nil {
		t.Errorf("expected nil bronze, got %v", bronze)
	}
}

func TestParseWCBracket_StandardRounds(t *testing.T) {
	playoff := wcPlayoff{
		Rounds: []wcPlayoffRound{
			{Stage: "1/2", Matchups: []wcMatchupRaw{makeMockMatchup(1, 2, 2, 0, 1, false)}},
			{Stage: "1/8", Matchups: []wcMatchupRaw{
				makeMockMatchup(10, 20, 3, 1, 10, false),
				makeMockMatchup(30, 40, 1, 0, 30, false),
			}},
			{Stage: "1/4", Matchups: []wcMatchupRaw{makeMockMatchup(5, 6, 1, 1, 5, true)}},
			{Stage: "final", Matchups: []wcMatchupRaw{makeMockMatchup(1, 5, 3, 3, 1, true)}},
		},
		Special: []wcPlayoffRound{
			{Stage: "bronze", Matchups: []wcMatchupRaw{makeMockMatchup(2, 6, 2, 1, 2, false)}},
		},
	}

	rounds, bronze := parseWCBracket(playoff)

	if len(rounds) != 4 {
		t.Fatalf("len(rounds) = %d, want 4", len(rounds))
	}

	// Rounds must be sorted: 1/8 → 1/4 → 1/2 → final
	wantOrder := []string{"1/8", "1/4", "1/2", "final"}
	for i, want := range wantOrder {
		if rounds[i].Stage != want {
			t.Errorf("rounds[%d].Stage = %q, want %q", i, rounds[i].Stage, want)
		}
	}

	// Verify R16 has 2 matchups as supplied
	if len(rounds[0].Matchups) != 2 {
		t.Errorf("R16 matchups = %d, want 2", len(rounds[0].Matchups))
	}

	if bronze == nil {
		t.Fatal("bronze is nil")
	}
	if bronze.HomeTeamID != 2 {
		t.Errorf("bronze.HomeTeamID = %d, want 2", bronze.HomeTeamID)
	}
}

func TestParseWCBracket_Labels(t *testing.T) {
	stageTests := []struct {
		stage string
		label string
	}{
		{"1/16", "Round of 32"},
		{"1/8", "Round of 16"},
		{"1/4", "Quarterfinals"},
		{"1/2", "Semifinals"},
		{"final", "Final"},
	}

	for _, tt := range stageTests {
		playoff := wcPlayoff{
			Rounds: []wcPlayoffRound{
				{Stage: tt.stage, Matchups: []wcMatchupRaw{makeMockMatchup(1, 2, 1, 0, 1, false)}},
			},
		}
		rounds, _ := parseWCBracket(playoff)
		if len(rounds) != 1 {
			t.Fatalf("stage %q: expected 1 round, got %d", tt.stage, len(rounds))
		}
		if rounds[0].Label != tt.label {
			t.Errorf("stage %q: Label = %q, want %q", tt.stage, rounds[0].Label, tt.label)
		}
	}
}

func TestConvertMatchup_Penalties(t *testing.T) {
	mu := makeMockMatchup(10, 20, 2, 2, 10, true)
	out := convertMatchup(mu)

	if !out.IsPenalties {
		t.Error("IsPenalties = false, want true")
	}
	if out.WinnerID == nil || *out.WinnerID != 10 {
		t.Errorf("WinnerID = %v, want 10", out.WinnerID)
	}
	if out.HomeScore == nil || *out.HomeScore != 2 {
		t.Errorf("HomeScore = %v, want 2", out.HomeScore)
	}
}

func TestConvertMatchup_NotYetPlayed(t *testing.T) {
	raw := wcMatchupRaw{
		HomeTeam:   "Team A",
		HomeTeamID: 1,
		AwayTeam:   "Team B",
		AwayTeamID: 2,
		// No matches — not yet played
	}
	out := convertMatchup(raw)

	if out.HomeScore != nil {
		t.Errorf("HomeScore should be nil for unplayed match, got %v", out.HomeScore)
	}
	if out.AwayScore != nil {
		t.Errorf("AwayScore should be nil for unplayed match, got %v", out.AwayScore)
	}
	if out.WinnerID != nil {
		t.Errorf("WinnerID should be nil for unplayed match, got %v", out.WinnerID)
	}
}

// --- helpers ---

// buildMockWCPageResponse builds a wcPageResponse with n groups, each having 4 teams.
func buildMockWCPageResponse(numGroups int) wcPageResponse {
	letters := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L"}
	tables := make([]struct {
		LeagueID   int    `json:"leagueId"`
		LeagueName string `json:"leagueName"`
		Table      struct {
			All []fotmobTableRow `json:"all"`
		} `json:"table"`
	}, numGroups)

	for i := 0; i < numGroups; i++ {
		tables[i].LeagueID = 900000 + i
		tables[i].LeagueName = "Grp. " + letters[i]
		tables[i].Table.All = []fotmobTableRow{
			{ID: i*100 + 1, Name: "Team1", ShortName: "TM1", Idx: 1, Played: 3, Wins: 2, Draws: 1, Losses: 0, ScoresStr: "5-1", GoalConDiff: 4, Pts: 7},
			{ID: i*100 + 2, Name: "Team2", ShortName: "TM2", Idx: 2, Played: 3, Wins: 1, Draws: 1, Losses: 1, ScoresStr: "3-3", GoalConDiff: 0, Pts: 4},
			{ID: i*100 + 3, Name: "Team3", ShortName: "TM3", Idx: 3, Played: 3, Wins: 1, Draws: 0, Losses: 2, ScoresStr: "2-4", GoalConDiff: -2, Pts: 3},
			{ID: i*100 + 4, Name: "Team4", ShortName: "TM4", Idx: 4, Played: 3, Wins: 0, Draws: 0, Losses: 3, ScoresStr: "1-3", GoalConDiff: -2, Pts: 0},
		}
	}

	// Marshal/unmarshal to get the right nested type via JSON
	type dataBlock struct {
		Composite bool        `json:"composite"`
		Tables    interface{} `json:"tables"`
	}
	type tableBlock struct {
		Data dataBlock `json:"data"`
	}

	raw, _ := json.Marshal(tables)
	var tablesJSON json.RawMessage = raw

	_ = tablesJSON

	resp := wcPageResponse{}
	resp.Table = []struct {
		Data struct {
			Composite bool `json:"composite"`
			Tables    []struct {
				LeagueID   int    `json:"leagueId"`
				LeagueName string `json:"leagueName"`
				Table      struct {
					All []fotmobTableRow `json:"all"`
				} `json:"table"`
			} `json:"tables"`
		} `json:"data"`
	}{
		{Data: struct {
			Composite bool `json:"composite"`
			Tables    []struct {
				LeagueID   int    `json:"leagueId"`
				LeagueName string `json:"leagueName"`
				Table      struct {
					All []fotmobTableRow `json:"all"`
				} `json:"table"`
			} `json:"tables"`
		}{Composite: true, Tables: tables}},
	}

	return resp
}

// makeMockMatchup creates a finished wcMatchupRaw.
func makeMockMatchup(homeID, awayID, homeScore, awayScore, winnerID int, penalties bool) wcMatchupRaw {
	finished := true
	mu := wcMatchupRaw{
		HomeTeam:   "Home",
		HomeTeamID: homeID,
		AwayTeam:   "Away",
		AwayTeamID: awayID,
		HomeScore:  homeScore,
		AwayScore:  awayScore,
		Winner:     winnerID,
		Matches: []struct {
			Status struct {
				Reason struct {
					Short string `json:"short"`
				} `json:"reason"`
				Finished *bool `json:"finished"`
			} `json:"status"`
		}{
			{Status: struct {
				Reason struct {
					Short string `json:"short"`
				} `json:"reason"`
				Finished *bool `json:"finished"`
			}{Finished: &finished}},
		},
	}
	if penalties {
		mu.Matches[0].Status.Reason.Short = "Pen"
	}
	return mu
}
