package data

import (
	"testing"
)

func TestMockWorldCupData_Structure(t *testing.T) {
	d := MockWorldCupData()

	if d == nil {
		t.Fatal("MockWorldCupData() returned nil")
	}

	if d.Season != "2022" {
		t.Errorf("Season = %q, want %q", d.Season, "2022")
	}

	if d.Name == "" {
		t.Error("Name is empty")
	}

	if d.Champion == nil {
		t.Fatal("Champion is nil")
	}
	if d.Champion.ID != 6706 {
		t.Errorf("Champion.ID = %d, want 6706 (Argentina)", d.Champion.ID)
	}

	if d.RunnerUp == nil {
		t.Fatal("RunnerUp is nil")
	}
	if d.RunnerUp.ID != 6723 {
		t.Errorf("RunnerUp.ID = %d, want 6723 (France)", d.RunnerUp.ID)
	}
}

func TestMockWorldCupData_Groups(t *testing.T) {
	d := MockWorldCupData()

	if len(d.Groups) != 8 {
		t.Fatalf("len(Groups) = %d, want 8", len(d.Groups))
	}

	// All groups must have exactly 4 teams
	for _, g := range d.Groups {
		if len(g.Teams) != 4 {
			t.Errorf("Group %s has %d teams, want 4", g.Letter, len(g.Teams))
		}
		if g.Letter == "" {
			t.Errorf("Group %q has empty Letter", g.Name)
		}
	}

	// Letters must be A–H in order
	wantLetters := []string{"A", "B", "C", "D", "E", "F", "G", "H"}
	for i, wl := range wantLetters {
		if d.Groups[i].Letter != wl {
			t.Errorf("Groups[%d].Letter = %q, want %q", i, d.Groups[i].Letter, wl)
		}
	}
}

func TestMockWorldCupData_GroupStandings(t *testing.T) {
	d := MockWorldCupData()

	for _, g := range d.Groups {
		for _, team := range g.Teams {
			// Each team must have played 3 matches
			if team.Played != 3 {
				t.Errorf("Group %s / %s: Played = %d, want 3", g.Letter, team.Team.Name, team.Played)
			}

			// W+D+L must equal Played
			if team.Won+team.Drawn+team.Lost != team.Played {
				t.Errorf("Group %s / %s: W(%d)+D(%d)+L(%d) != Played(%d)",
					g.Letter, team.Team.Name, team.Won, team.Drawn, team.Lost, team.Played)
			}

			// GD must equal GF-GA
			if team.GoalDifference != team.GoalsFor-team.GoalsAgainst {
				t.Errorf("Group %s / %s: GD(%d) != GF(%d)-GA(%d)",
					g.Letter, team.Team.Name, team.GoalDifference, team.GoalsFor, team.GoalsAgainst)
			}

			// Points must equal 3*W + D
			wantPts := 3*team.Won + team.Drawn
			if team.Points != wantPts {
				t.Errorf("Group %s / %s: Points = %d, want %d", g.Letter, team.Team.Name, team.Points, wantPts)
			}

			// Position must be 1–4
			if team.Position < 1 || team.Position > 4 {
				t.Errorf("Group %s / %s: Position = %d, want 1-4", g.Letter, team.Team.Name, team.Position)
			}

			if team.Team.ID == 0 {
				t.Errorf("Group %s / %s: Team.ID is 0", g.Letter, team.Team.Name)
			}
		}

		// Teams within each group must be sorted by position (1,2,3,4)
		for i := 1; i < len(g.Teams); i++ {
			if g.Teams[i].Position <= g.Teams[i-1].Position {
				t.Errorf("Group %s: teams not sorted by position at index %d", g.Letter, i)
			}
		}
	}
}

func TestMockWorldCupData_Bracket(t *testing.T) {
	d := MockWorldCupData()

	if len(d.KnockoutRounds) == 0 {
		t.Fatal("KnockoutRounds is empty")
	}

	wantStages := []struct {
		stage    string
		label    string
		matchups int
	}{
		{"1/8", "Round of 16", 8},
		{"1/4", "Quarterfinals", 4},
		{"1/2", "Semifinals", 2},
		{"final", "Final", 1},
	}

	if len(d.KnockoutRounds) != len(wantStages) {
		t.Fatalf("len(KnockoutRounds) = %d, want %d", len(d.KnockoutRounds), len(wantStages))
	}

	for i, ws := range wantStages {
		r := d.KnockoutRounds[i]
		if r.Stage != ws.stage {
			t.Errorf("KnockoutRounds[%d].Stage = %q, want %q", i, r.Stage, ws.stage)
		}
		if r.Label != ws.label {
			t.Errorf("KnockoutRounds[%d].Label = %q, want %q", i, r.Label, ws.label)
		}
		if len(r.Matchups) != ws.matchups {
			t.Errorf("KnockoutRounds[%d] (%s): len(Matchups) = %d, want %d", i, r.Stage, len(r.Matchups), ws.matchups)
		}
	}
}

func TestMockWorldCupData_BracketMatchups(t *testing.T) {
	d := MockWorldCupData()

	for _, round := range d.KnockoutRounds {
		for _, mu := range round.Matchups {
			if mu.HomeTeamID == 0 {
				t.Errorf("Round %s: matchup %s vs %s has zero HomeTeamID", round.Stage, mu.HomeTeam, mu.AwayTeam)
			}
			if mu.AwayTeamID == 0 {
				t.Errorf("Round %s: matchup %s vs %s has zero AwayTeamID", round.Stage, mu.HomeTeam, mu.AwayTeam)
			}
			if mu.WinnerID == nil {
				t.Errorf("Round %s: matchup %s vs %s has nil WinnerID", round.Stage, mu.HomeTeam, mu.AwayTeam)
				continue
			}
			winnerID := *mu.WinnerID
			if winnerID != mu.HomeTeamID && winnerID != mu.AwayTeamID {
				t.Errorf("Round %s: WinnerID %d is neither HomeTeamID %d nor AwayTeamID %d",
					round.Stage, winnerID, mu.HomeTeamID, mu.AwayTeamID)
			}
			// Penalty matchups must have equal scores
			if mu.IsPenalties {
				if mu.HomeScore == nil || mu.AwayScore == nil {
					t.Errorf("Round %s: penalty matchup %s vs %s has nil scores", round.Stage, mu.HomeTeam, mu.AwayTeam)
				} else if *mu.HomeScore != *mu.AwayScore {
					t.Errorf("Round %s: penalty matchup %s vs %s has unequal scores %d-%d",
						round.Stage, mu.HomeTeam, mu.AwayTeam, *mu.HomeScore, *mu.AwayScore)
				}
			}
		}
	}
}

func TestMockWorldCupData_BronzeFinal(t *testing.T) {
	d := MockWorldCupData()

	if d.BronzeFinal == nil {
		t.Fatal("BronzeFinal is nil")
	}
	bf := d.BronzeFinal
	if bf.HomeTeamID == 0 || bf.AwayTeamID == 0 {
		t.Error("BronzeFinal has zero team ID")
	}
	if bf.WinnerID == nil {
		t.Error("BronzeFinal.WinnerID is nil")
	}
}
