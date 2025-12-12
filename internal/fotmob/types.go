package fotmob

import (
	"strconv"
	"time"

	"github.com/0xjuanma/golazo/internal/api"
)

// fotmobMatch represents a match in FotMob's API format
type fotmobMatch struct {
	ID      int    `json:"id"`
	Round   string `json:"round"`
	Home    team   `json:"home"`
	Away    team   `json:"away"`
	Status  status `json:"status"`
	League  league `json:"league"`
}

type team struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	ShortName string `json:"shortName"`
}

type league struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Country     string `json:"country"`
	CountryCode string `json:"countryCode"`
}

type status struct {
	UTCTime  string `json:"utcTime"`
	Started  bool   `json:"started"`
	Finished bool   `json:"finished"`
	Cancelled bool  `json:"cancelled"`
	LiveTime  *liveTime `json:"liveTime,omitempty"`
	Score     *score    `json:"score,omitempty"`
}

type liveTime struct {
	Short string `json:"short"`
}

type score struct {
	Home int `json:"home"`
	Away int `json:"away"`
}

// toAPIMatch converts a fotmobMatch to api.Match
func (m fotmobMatch) toAPIMatch() api.Match {
	match := api.Match{
		ID: m.ID,
		League: api.League{
			ID:          m.League.ID,
			Name:        m.League.Name,
			Country:     m.League.Country,
			CountryCode: m.League.CountryCode,
		},
		HomeTeam: api.Team{
			ID:        m.Home.ID,
			Name:      m.Home.Name,
			ShortName: m.Home.ShortName,
		},
		AwayTeam: api.Team{
			ID:        m.Away.ID,
			Name:      m.Away.Name,
			ShortName: m.Away.ShortName,
		},
		Round: m.Round,
	}

	// Parse match time
	if m.Status.UTCTime != "" {
		if t, err := time.Parse(time.RFC3339, m.Status.UTCTime); err == nil {
			match.MatchTime = &t
		}
	}

	// Determine status
	if m.Status.Cancelled {
		match.Status = api.MatchStatusCancelled
	} else if m.Status.Finished {
		match.Status = api.MatchStatusFinished
	} else if m.Status.Started {
		match.Status = api.MatchStatusLive
		if m.Status.LiveTime != nil {
			match.LiveTime = &m.Status.LiveTime.Short
		}
	} else {
		match.Status = api.MatchStatusNotStarted
	}

	// Set scores if available
	if m.Status.Score != nil {
		match.HomeScore = &m.Status.Score.Home
		match.AwayScore = &m.Status.Score.Away
	}

	return match
}

// fotmobMatchDetails represents detailed match information from FotMob
type fotmobMatchDetails struct {
	ID      int    `json:"id"`
	Round   string `json:"round"`
	Home    team   `json:"home"`
	Away    team   `json:"away"`
	Status  status `json:"status"`
	League  league `json:"league"`
	Events  []event `json:"events"`
}

type event struct {
	ID        int    `json:"id"`
	Minute    int    `json:"minute"`
	Type      string `json:"type"`
	TeamID    int    `json:"teamId"`
	Player    string `json:"player,omitempty"`
	Assist    string `json:"assist,omitempty"`
	EventType string `json:"eventType,omitempty"`
}

// toAPIMatchDetails converts fotmobMatchDetails to api.MatchDetails
func (m fotmobMatchDetails) toAPIMatchDetails() *api.MatchDetails {
	baseMatch := fotmobMatch{
		ID:     m.ID,
		Round:  m.Round,
		Home:   m.Home,
		Away:   m.Away,
		Status: m.Status,
		League: m.League,
	}.toAPIMatch()

	details := &api.MatchDetails{
		Match: baseMatch,
		Events: make([]api.MatchEvent, 0, len(m.Events)),
	}

	// Convert events
	for _, e := range m.Events {
		event := api.MatchEvent{
			ID:        e.ID,
			Minute:    e.Minute,
			Type:      e.Type,
			Timestamp: time.Now(), // FotMob doesn't always provide timestamp
		}

		if e.Player != "" {
			event.Player = &e.Player
		}
		if e.Assist != "" {
			event.Assist = &e.Assist
		}
		if e.EventType != "" {
			event.EventType = &e.EventType
		}

		// Set team (simplified - would need team lookup in real implementation)
		event.Team = api.Team{
			ID: e.TeamID,
		}

		details.Events = append(details.Events, event)
	}

	return details
}

// fotmobTableRow represents a single row in the league table from FotMob
type fotmobTableRow struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	ShortName      string `json:"shortName"`
	Rank           int    `json:"rank"`
	Played         int    `json:"played"`
	Wins           int    `json:"wins"`
	Draws          int    `json:"draws"`
	Losses         int    `json:"losses"`
	GoalsFor       int    `json:"goalsFor"`
	GoalsAgainst   int    `json:"goalsAgainst"`
	GoalDifference int    `json:"goalDifference"`
	Points         int    `json:"points"`
}

// toAPITableEntry converts fotmobTableRow to api.LeagueTableEntry
func (r fotmobTableRow) toAPITableEntry() api.LeagueTableEntry {
	return api.LeagueTableEntry{
		Position:      r.Rank,
		Team: api.Team{
			ID:        r.ID,
			Name:      r.Name,
			ShortName: r.ShortName,
		},
		Played:        r.Played,
		Won:           r.Wins,
		Drawn:         r.Draws,
		Lost:          r.Losses,
		GoalsFor:      r.GoalsFor,
		GoalsAgainst:  r.GoalsAgainst,
		GoalDifference: r.GoalDifference,
		Points:        r.Points,
	}
}

// Helper function to parse time from various formats
func parseTime(timeStr string) *time.Time {
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return &t
		}
	}
	return nil
}

// Helper function to parse int from string
func parseInt(s string) *int {
	if s == "" {
		return nil
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}
	return &val
}

