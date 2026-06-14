package app

import (
	"testing"
	"time"

	"github.com/0xjuanma/golazo/internal/api"
)

func TestBuildGoalInfosRunningScore(t *testing.T) {
	home := api.Team{ID: 1, Name: "Australia", ShortName: "AUS"}
	away := api.Team{ID: 2, Name: "Türkiye", ShortName: "TUR"}
	matchTime := time.Date(2025, 11, 10, 16, 0, 0, 0, time.UTC)

	scorer1 := "Nystrom Irankunda"
	scorer2 := "Connor Metcalfe"

	details := &api.MatchDetails{
		Match: api.Match{
			ID:        12345,
			HomeTeam:  home,
			AwayTeam:  away,
			MatchTime: &matchTime,
		},
		Events: []api.MatchEvent{
			{Type: "goal", Minute: 27, Team: home, Player: &scorer1},
			{Type: "card", Minute: 35, Team: away}, // ignored
			{Type: "goal", Minute: 75, Team: home, Player: &scorer2},
		},
	}

	goals := buildGoalInfos(details)
	if len(goals) != 2 {
		t.Fatalf("expected 2 goals, got %d", len(goals))
	}

	if goals[0].HomeScore != 1 || goals[0].AwayScore != 0 {
		t.Errorf("first goal: got %d-%d, want 1-0", goals[0].HomeScore, goals[0].AwayScore)
	}
	if goals[1].HomeScore != 2 || goals[1].AwayScore != 0 {
		t.Errorf("second goal: got %d-%d, want 2-0", goals[1].HomeScore, goals[1].AwayScore)
	}
	if goals[0].ScorerName != scorer1 {
		t.Errorf("first scorer: got %q, want %q", goals[0].ScorerName, scorer1)
	}
	if !goals[0].IsHomeTeam || !goals[1].IsHomeTeam {
		t.Error("both Australia goals should credit home team")
	}
}

func TestBuildGoalInfosAlternatingTeams(t *testing.T) {
	home := api.Team{ID: 1, Name: "France"}
	away := api.Team{ID: 2, Name: "Argentina"}
	matchTime := time.Now()
	mbappe := "Kylian Mbappé"
	messi := "Lionel Messi"

	details := &api.MatchDetails{
		Match: api.Match{ID: 1, HomeTeam: home, AwayTeam: away, MatchTime: &matchTime},
		Events: []api.MatchEvent{
			{Type: "goal", Minute: 23, Team: away, Player: &messi},
			{Type: "goal", Minute: 36, Team: away, Player: &messi},
			{Type: "goal", Minute: 80, Team: home, Player: &mbappe},
			{Type: "goal", Minute: 81, Team: home, Player: &mbappe},
		},
	}

	goals := buildGoalInfos(details)
	want := []struct{ h, a int }{{0, 1}, {0, 2}, {1, 2}, {2, 2}}
	for i, w := range want {
		if goals[i].HomeScore != w.h || goals[i].AwayScore != w.a {
			t.Errorf("goal %d: got %d-%d, want %d-%d", i, goals[i].HomeScore, goals[i].AwayScore, w.h, w.a)
		}
	}
}

func TestBuildGoalInfosOwnGoalCreditsOpposingTeam(t *testing.T) {
	home := api.Team{ID: 1, Name: "Liverpool"}
	away := api.Team{ID: 2, Name: "Everton"}
	matchTime := time.Now()
	defender := "Defender Name"
	yes := true

	details := &api.MatchDetails{
		Match: api.Match{ID: 1, HomeTeam: home, AwayTeam: away, MatchTime: &matchTime},
		Events: []api.MatchEvent{
			// Own goal: scored by the home defender on his own net — credits away.
			{Type: "goal", Minute: 15, Team: home, Player: &defender, OwnGoal: &yes},
		},
	}

	goals := buildGoalInfos(details)
	if len(goals) != 1 {
		t.Fatalf("expected 1 goal, got %d", len(goals))
	}
	if goals[0].HomeScore != 0 || goals[0].AwayScore != 1 {
		t.Errorf("own goal: got %d-%d, want 0-1", goals[0].HomeScore, goals[0].AwayScore)
	}
	if goals[0].IsHomeTeam {
		t.Error("own goal should credit the opposing (away) team, IsHomeTeam should be false")
	}
}

func TestBuildGoalInfosOutOfOrderEventsAreSorted(t *testing.T) {
	home := api.Team{ID: 1, Name: "A"}
	away := api.Team{ID: 2, Name: "B"}
	matchTime := time.Now()
	scorer := "X"

	details := &api.MatchDetails{
		Match: api.Match{ID: 1, HomeTeam: home, AwayTeam: away, MatchTime: &matchTime},
		Events: []api.MatchEvent{
			{Type: "goal", Minute: 75, Team: home, Player: &scorer},
			{Type: "goal", Minute: 27, Team: home, Player: &scorer},
		},
	}

	goals := buildGoalInfos(details)
	if goals[0].Minute != 27 || goals[1].Minute != 75 {
		t.Errorf("expected sorted [27, 75], got [%d, %d]", goals[0].Minute, goals[1].Minute)
	}
	if goals[0].HomeScore != 1 || goals[1].HomeScore != 2 {
		t.Errorf("running score after sort: got [%d, %d], want [1, 2]", goals[0].HomeScore, goals[1].HomeScore)
	}
}

func TestBuildGoalInfosNoGoalsReturnsNil(t *testing.T) {
	home := api.Team{ID: 1, Name: "A"}
	away := api.Team{ID: 2, Name: "B"}
	matchTime := time.Now()

	details := &api.MatchDetails{
		Match: api.Match{ID: 1, HomeTeam: home, AwayTeam: away, MatchTime: &matchTime},
		Events: []api.MatchEvent{
			{Type: "card", Minute: 35, Team: away},
		},
	}
	if got := buildGoalInfos(details); got != nil {
		t.Errorf("expected nil, got %d goals", len(got))
	}
}

func TestBuildGoalInfosNilDetails(t *testing.T) {
	if got := buildGoalInfos(nil); got != nil {
		t.Errorf("expected nil for nil details, got %d goals", len(got))
	}
}
