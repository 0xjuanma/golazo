package api

import "testing"

func intPtrTest(v int) *int { return &v }

// ── DeriveFinalists ───────────────────────────────────────────────────────────

func TestDeriveFinalists_NoRounds(t *testing.T) {
	d := &WorldCupData{}
	champ, runner := d.DeriveFinalists()
	if champ != nil || runner != nil {
		t.Errorf("expected nil, nil for empty rounds; got %v, %v", champ, runner)
	}
}

func TestDeriveFinalists_NoFinalRound(t *testing.T) {
	d := &WorldCupData{
		KnockoutRounds: []WCKnockoutRound{
			{Stage: "1/2", Matchups: []WCMatchup{{HomeTeamID: 1, AwayTeamID: 2, WinnerID: intPtrTest(1)}}},
		},
	}
	champ, runner := d.DeriveFinalists()
	if champ != nil || runner != nil {
		t.Errorf("expected nil, nil when no final round; got %v, %v", champ, runner)
	}
}

func TestDeriveFinalists_FinalNotPlayedYet(t *testing.T) {
	d := &WorldCupData{
		KnockoutRounds: []WCKnockoutRound{
			{Stage: "final", Matchups: []WCMatchup{{HomeTeamID: 1, HomeTeam: "A", AwayTeamID: 2, AwayTeam: "B"}}},
		},
	}
	champ, runner := d.DeriveFinalists()
	if champ != nil || runner != nil {
		t.Errorf("expected nil, nil when WinnerID is nil; got %v, %v", champ, runner)
	}
}

func TestDeriveFinalists_HomeWins(t *testing.T) {
	d := &WorldCupData{
		KnockoutRounds: []WCKnockoutRound{
			{
				Stage: "final",
				Matchups: []WCMatchup{{
					HomeTeam: "Argentina", HomeTeamID: 10, HomeShort: "ARG",
					AwayTeam: "France", AwayTeamID: 20, AwayShort: "FRA",
					WinnerID: intPtrTest(10),
				}},
			},
		},
	}
	champ, runner := d.DeriveFinalists()
	if champ == nil || runner == nil {
		t.Fatal("expected non-nil champion and runner-up")
	}
	if champ.ID != 10 {
		t.Errorf("champion ID = %d, want 10", champ.ID)
	}
	if champ.Name != "Argentina" {
		t.Errorf("champion Name = %q, want \"Argentina\"", champ.Name)
	}
	if runner.ID != 20 {
		t.Errorf("runner-up ID = %d, want 20", runner.ID)
	}
}

func TestDeriveFinalists_AwayWins(t *testing.T) {
	d := &WorldCupData{
		KnockoutRounds: []WCKnockoutRound{
			{
				Stage: "final",
				Matchups: []WCMatchup{{
					HomeTeam: "Argentina", HomeTeamID: 10, HomeShort: "ARG",
					AwayTeam: "France", AwayTeamID: 20, AwayShort: "FRA",
					WinnerID: intPtrTest(20),
				}},
			},
		},
	}
	champ, runner := d.DeriveFinalists()
	if champ == nil || runner == nil {
		t.Fatal("expected non-nil champion and runner-up")
	}
	if champ.ID != 20 {
		t.Errorf("champion ID = %d, want 20 (away winner)", champ.ID)
	}
	if runner.ID != 10 {
		t.Errorf("runner-up ID = %d, want 10", runner.ID)
	}
}

func TestDeriveFinalists_EmptyMatchups(t *testing.T) {
	d := &WorldCupData{
		KnockoutRounds: []WCKnockoutRound{
			{Stage: "final", Matchups: []WCMatchup{}},
		},
	}
	champ, runner := d.DeriveFinalists()
	if champ != nil || runner != nil {
		t.Errorf("expected nil, nil for final with no matchups; got %v, %v", champ, runner)
	}
}

// ── BracketLineCount ──────────────────────────────────────────────────────────

func TestBracketLineCount_Empty(t *testing.T) {
	d := &WorldCupData{}
	if n := d.BracketLineCount(); n != 0 {
		t.Errorf("BracketLineCount() = %d, want 0", n)
	}
}

func TestBracketLineCount_RoundsOnly(t *testing.T) {
	d := &WorldCupData{
		KnockoutRounds: []WCKnockoutRound{
			{Stage: "1/8", Matchups: make([]WCMatchup, 8)}, // 2+8+1 = 11
			{Stage: "1/4", Matchups: make([]WCMatchup, 4)}, // 2+4+1 = 7
		},
	}
	want := 11 + 7
	if n := d.BracketLineCount(); n != want {
		t.Errorf("BracketLineCount() = %d, want %d", n, want)
	}
}

func TestBracketLineCount_WithBronze(t *testing.T) {
	mu := WCMatchup{}
	d := &WorldCupData{
		KnockoutRounds: []WCKnockoutRound{
			{Stage: "final", Matchups: make([]WCMatchup, 1)}, // 2+1+1 = 4
		},
		BronzeFinal: &mu, // +4
	}
	want := 4 + 4
	if n := d.BracketLineCount(); n != want {
		t.Errorf("BracketLineCount() = %d, want %d", n, want)
	}
}

func TestBracketLineCount_WithChampion(t *testing.T) {
	champ := Team{ID: 1, Name: "X"}
	d := &WorldCupData{
		KnockoutRounds: []WCKnockoutRound{
			{Stage: "final", Matchups: make([]WCMatchup, 1)}, // 4
		},
		Champion: &champ, // +2
	}
	want := 4 + 2
	if n := d.BracketLineCount(); n != want {
		t.Errorf("BracketLineCount() = %d, want %d", n, want)
	}
}

func TestBracketLineCount_Full2022(t *testing.T) {
	champ := Team{ID: 1}
	bronze := WCMatchup{}
	d := &WorldCupData{
		KnockoutRounds: []WCKnockoutRound{
			{Stage: "1/8", Matchups: make([]WCMatchup, 8)},  // 11
			{Stage: "1/4", Matchups: make([]WCMatchup, 4)},  // 7
			{Stage: "1/2", Matchups: make([]WCMatchup, 2)},  // 5
			{Stage: "final", Matchups: make([]WCMatchup, 1)}, // 4
		},
		BronzeFinal: &bronze, // +4
		Champion:    &champ,  // +2
	}
	want := 11 + 7 + 5 + 4 + 4 + 2
	if n := d.BracketLineCount(); n != want {
		t.Errorf("BracketLineCount() = %d, want %d", n, want)
	}
}
