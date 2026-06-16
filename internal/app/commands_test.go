package app

import (
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/0xjuanma/golazo/internal/api"
	"github.com/0xjuanma/golazo/internal/reddit"
)

// testLogger returns a slog.Logger that discards output, matching the
// debugLog-disabled path of initLogger. Required because handleGoalLink
// calls m.debugLog which panics with nil m.logger.
func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

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

// TestHandleGoalLinkMergesSingle exercises the per-goal merge that the
// subscription path relies on: a single goalLinkMsg must be applied to
// m.goalLinks[key] and the reader Cmd must be re-armed against the same
// channel. The behavior under test is the merge inside handleGoalLink — not
// any wrapping fetch logic — so the test wires a channel directly without
// invoking GoalLinksAsync.
func TestHandleGoalLinkMergesSingle(t *testing.T) {
	ch := make(chan reddit.GoalResult, 1)
	defer close(ch)

	m := model{
		goalLinks:     make(map[reddit.GoalLinkKey]*reddit.GoalLink),
		goalLinkChans: map[int]<-chan reddit.GoalResult{42: ch},
		logger:        testLogger(),
	}

	link := &reddit.GoalLink{
		MatchID: 42,
		Minute:  17,
		URL:     "https://example.com/replay",
		Title:   "Iran 0 - [1] New Zealand - E. Just 7'",
	}
	key := reddit.GoalLinkKey{MatchID: 42, Minute: 17}

	newModel, cmd := m.handleGoalLink(goalLinkMsg{matchID: 42, key: key, link: link})
	mm, ok := newModel.(model)
	if !ok {
		t.Fatalf("handleGoalLink returned wrong model type: %T", newModel)
	}
	if got := mm.goalLinks[key]; got != link {
		t.Errorf("goalLinks[%+v] = %+v, want %+v", key, got, link)
	}
	if cmd == nil {
		t.Fatal("handleGoalLink returned nil Cmd; expected re-armed waitForGoalLink")
	}
}

// TestHandleGoalLinkNotFoundIsRecorded verifies that nil/not-found results
// are still stored in goalLinks so the UI can distinguish "search resolved
// to no match" from "search still pending".
func TestHandleGoalLinkNotFoundIsRecorded(t *testing.T) {
	ch := make(chan reddit.GoalResult, 1)
	defer close(ch)

	m := model{
		goalLinks:     make(map[reddit.GoalLinkKey]*reddit.GoalLink),
		goalLinkChans: map[int]<-chan reddit.GoalResult{99: ch},
		logger:        testLogger(),
	}
	key := reddit.GoalLinkKey{MatchID: 99, Minute: 45}

	newModel, _ := m.handleGoalLink(goalLinkMsg{matchID: 99, key: key, link: nil})
	mm := newModel.(model)
	if _, ok := mm.goalLinks[key]; !ok {
		t.Errorf("expected nil-link sentinel entry in goalLinks for %+v", key)
	}
}
