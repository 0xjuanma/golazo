package fotmob

import (
	"testing"

	"github.com/0xjuanma/golazo/internal/api"
)

func TestParseInt(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want int
	}{
		{"valid positive", "42", 42},
		{"valid zero", "0", 0},
		{"valid negative", "-1", -1},
		{"large number", "4813576", 4813576},
		{"empty string", "", 0},
		{"non-numeric", "abc", 0},
		{"float string", "3.14", 0},
		{"whitespace", " 5 ", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseInt(tt.in)
			if got != tt.want {
				t.Errorf("parseInt(%q) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}

func TestParseScoreStr(t *testing.T) {
	tests := []struct {
		name     string
		in       string
		wantHome int
		wantAway int
		wantOk   bool
	}{
		{"valid score", "4 - 2", 4, 2, true},
		{"zero zero", "0 - 0", 0, 0, true},
		{"large score", "10 - 3", 10, 3, true},
		{"empty string", "", 0, 0, false},
		{"missing separator", "4-2", 0, 0, false},
		{"only home", "4 - ", 0, 0, false},
		{"non-numeric", "a - b", 0, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			home, away, ok := parseScoreStr(tt.in)
			if ok != tt.wantOk {
				t.Errorf("parseScoreStr(%q) ok = %v, want %v", tt.in, ok, tt.wantOk)
			}
			if ok && (home != tt.wantHome || away != tt.wantAway) {
				t.Errorf("parseScoreStr(%q) = (%d, %d), want (%d, %d)", tt.in, home, away, tt.wantHome, tt.wantAway)
			}
		})
	}
}

func TestFormatStatValue(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want string
	}{
		{"string value", "65%", "65%"},
		{"integer float", float64(10), "10"},
		{"decimal float", float64(1.5), "1.5"},
		{"int value", 7, "7"},
		{"nil value", nil, ""},
		{"bool value", true, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatStatValue(tt.in)
			if got != tt.want {
				t.Errorf("formatStatValue(%v) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func boolPtr(b bool) *bool { return &b }

func TestToAPIMatch_BasicFields(t *testing.T) {
	m := fotmobMatch{
		ID:    "4813576",
		Round: "21",
		Home:  team{ID: "8650", Name: "Liverpool", ShortName: "Liverpool"},
		Away:  team{ID: "9825", Name: "Arsenal", ShortName: "Arsenal"},
		Status: status{
			UTCTime:  "2026-01-08T20:00:00Z",
			Started:  boolPtr(false),
			Finished: boolPtr(false),
		},
		League: league{
			ID:          47,
			Name:        "Premier League",
			Country:     "England",
			CountryCode: "GB",
		},
		PageURL: "/matches/liverpool-vs-arsenal/2tmaz7#4813576",
	}

	got := m.toAPIMatch()

	if got.ID != 4813576 {
		t.Errorf("ID = %d, want 4813576", got.ID)
	}
	if got.HomeTeam.ID != 8650 {
		t.Errorf("HomeTeam.ID = %d, want 8650", got.HomeTeam.ID)
	}
	if got.AwayTeam.ID != 9825 {
		t.Errorf("AwayTeam.ID = %d, want 9825", got.AwayTeam.ID)
	}
	if got.Round != "21" {
		t.Errorf("Round = %q, want %q", got.Round, "21")
	}
	if got.League.Name != "Premier League" {
		t.Errorf("League.Name = %q, want %q", got.League.Name, "Premier League")
	}
	if got.Status != api.MatchStatusNotStarted {
		t.Errorf("Status = %q, want %q", got.Status, api.MatchStatusNotStarted)
	}
	if got.PageURL != "/matches/liverpool-vs-arsenal/2tmaz7" {
		t.Errorf("PageURL = %q, want fragment stripped", got.PageURL)
	}
	if got.MatchTime == nil {
		t.Fatal("MatchTime is nil, want parsed time")
	}
	if got.MatchTime.Hour() != 20 {
		t.Errorf("MatchTime hour = %d, want 20", got.MatchTime.Hour())
	}
}

func TestToAPIMatch_StatusVariants(t *testing.T) {
	tests := []struct {
		name   string
		status status
		want   api.MatchStatus
	}{
		{
			"finished",
			status{Finished: boolPtr(true), Started: boolPtr(true)},
			api.MatchStatusFinished,
		},
		{
			"live",
			status{Started: boolPtr(true), Finished: boolPtr(false)},
			api.MatchStatusLive,
		},
		{
			"cancelled",
			status{Cancelled: boolPtr(true)},
			api.MatchStatusCancelled,
		},
		{
			"not started with nil booleans",
			status{},
			api.MatchStatusNotStarted,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := fotmobMatch{Status: tt.status}
			got := m.toAPIMatch()
			if got.Status != tt.want {
				t.Errorf("status = %q, want %q", got.Status, tt.want)
			}
		})
	}
}

func TestToAPIMatch_ScoreFromScoreObj(t *testing.T) {
	m := fotmobMatch{
		Status: status{
			Finished: boolPtr(true),
			Score:    &score{Home: 3, Away: 1},
		},
	}
	got := m.toAPIMatch()
	if got.HomeScore == nil || *got.HomeScore != 3 {
		t.Errorf("HomeScore = %v, want 3", got.HomeScore)
	}
	if got.AwayScore == nil || *got.AwayScore != 1 {
		t.Errorf("AwayScore = %v, want 1", got.AwayScore)
	}
}

func TestToAPIMatch_ScoreFromScoreStr(t *testing.T) {
	m := fotmobMatch{
		Status: status{
			Finished: boolPtr(true),
			ScoreStr: "2 - 0",
		},
	}
	got := m.toAPIMatch()
	if got.HomeScore == nil || *got.HomeScore != 2 {
		t.Errorf("HomeScore = %v, want 2", got.HomeScore)
	}
	if got.AwayScore == nil || *got.AwayScore != 0 {
		t.Errorf("AwayScore = %v, want 0", got.AwayScore)
	}
}

func TestToAPIMatch_MillisecondTimeFormat(t *testing.T) {
	m := fotmobMatch{
		Status: status{
			UTCTime: "2026-01-08T20:00:00.000Z",
		},
	}
	got := m.toAPIMatch()
	if got.MatchTime == nil {
		t.Fatal("MatchTime is nil, want parsed time for .000Z format")
	}
	if got.MatchTime.Hour() != 20 {
		t.Errorf("MatchTime hour = %d, want 20", got.MatchTime.Hour())
	}
}

func TestToAPIMatch_LiveTimeSet(t *testing.T) {
	m := fotmobMatch{
		Status: status{
			Started:  boolPtr(true),
			Finished: boolPtr(false),
			LiveTime: &liveTime{Short: "72'"},
		},
	}
	got := m.toAPIMatch()
	if got.LiveTime == nil || *got.LiveTime != "72'" {
		t.Errorf("LiveTime = %v, want 72'", got.LiveTime)
	}
}

func TestToAPITableEntry(t *testing.T) {
	row := fotmobTableRow{
		ID:          8650,
		Name:        "Liverpool",
		ShortName:   "LIV",
		Idx:         1,
		Played:      20,
		Wins:        15,
		Draws:       3,
		Losses:      2,
		ScoresStr:   "42-17",
		GoalConDiff: 25,
		Pts:         48,
	}
	got := row.toAPITableEntry()
	if got.Position != 1 {
		t.Errorf("Position = %d, want 1", got.Position)
	}
	if got.Team.ID != 8650 {
		t.Errorf("Team.ID = %d, want 8650", got.Team.ID)
	}
	if got.GoalsFor != 42 {
		t.Errorf("GoalsFor = %d, want 42", got.GoalsFor)
	}
	if got.GoalsAgainst != 17 {
		t.Errorf("GoalsAgainst = %d, want 17", got.GoalsAgainst)
	}
	if got.Points != 48 {
		t.Errorf("Points = %d, want 48", got.Points)
	}
}

func TestGetParentLeagueID(t *testing.T) {
	tests := []struct {
		name     string
		league   string
		id       int
		wantID   int
	}{
		{"Champions League match", "Champions League Grp. A", 999, 42},
		{"Europa League match", "Europa League Knockout", 888, 73},
		{"regular league", "Premier League", 47, 47},
		{"unknown league", "Some Random League", 100, 100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getParentLeagueID(tt.league, tt.id)
			if got != tt.wantID {
				t.Errorf("getParentLeagueID(%q, %d) = %d, want %d", tt.league, tt.id, got, tt.wantID)
			}
		})
	}
}
