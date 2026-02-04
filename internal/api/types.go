package api

import "time"

// League represents a football league
type League struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Country     string `json:"country"`
	CountryCode string `json:"country_code"`
	Logo        string `json:"logo,omitempty"`
}

// Team represents a football team
type Team struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	ShortName string `json:"short_name"`
	Logo      string `json:"logo,omitempty"`
}

// MatchStatus represents the status of a match
type MatchStatus string

const (
	MatchStatusNotStarted MatchStatus = "not_started"
	MatchStatusLive       MatchStatus = "live"
	MatchStatusFinished   MatchStatus = "finished"
	MatchStatusPostponed  MatchStatus = "postponed"
	MatchStatusCancelled  MatchStatus = "cancelled"
)

// Match represents a football match
type Match struct {
	ID        int         `json:"id"`
	League    League      `json:"league"`
	HomeTeam  Team        `json:"home_team"`
	AwayTeam  Team        `json:"away_team"`
	Status    MatchStatus `json:"status"`
	HomeScore *int        `json:"home_score,omitempty"`
	AwayScore *int        `json:"away_score,omitempty"`
	MatchTime *time.Time  `json:"match_time,omitempty"`
	LiveTime  *string     `json:"live_time,omitempty"` // e.g., "45+2", "HT", "FT"
	Round     string      `json:"round,omitempty"`
}

// MatchEvent represents an event in a match (goal, card, substitution, etc.)
type MatchEvent struct {
	ID            int       `json:"id"`
	Minute        int       `json:"minute"`                   // Base minute (e.g., 45)
	DisplayMinute string    `json:"display_minute,omitempty"` // Formatted minute with stoppage time (e.g., "45+2'")
	Type          string    `json:"type"`                     // "goal", "card", "substitution", etc.
	Team          Team      `json:"team"`
	Player        *string   `json:"player,omitempty"`
	Assist        *string   `json:"assist,omitempty"`
	EventType     *string   `json:"event_type,omitempty"` // "yellow", "red", "in", "out", etc.
	OwnGoal       *bool     `json:"own_goal,omitempty"`   // Indicates if this is an own goal
	Timestamp     time.Time `json:"timestamp"`
}

// MatchStatistic represents a single match statistic (possession, shots, etc.)
type MatchStatistic struct {
	Key       string `json:"key"`        // e.g., "possession", "shots_total"
	Label     string `json:"label"`      // e.g., "Possession", "Total Shots"
	HomeValue string `json:"home_value"` // Value for home team
	AwayValue string `json:"away_value"` // Value for away team
}

// PlayerInfo represents basic player information for lineups
type PlayerInfo struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Number   int    `json:"number,omitempty"`
	Position string `json:"position,omitempty"`
	Rating   string `json:"rating,omitempty"` // Player rating (e.g., "7.2")
}

// MatchDetails contains detailed information about a match
type MatchDetails struct {
	Match
	Events     []MatchEvent `json:"events"`
	HomeLineup []string     `json:"home_lineup,omitempty"`
	AwayLineup []string     `json:"away_lineup,omitempty"`

	// Additional match information
	HalfTimeScore *struct {
		Home *int `json:"home,omitempty"`
		Away *int `json:"away,omitempty"`
	} `json:"half_time_score,omitempty"`
	Venue         string  `json:"venue,omitempty"`          // Stadium name
	Winner        *string `json:"winner,omitempty"`         // "home" or "away"
	MatchDuration int     `json:"match_duration,omitempty"` // 90, 120, etc.
	ExtraTime     bool    `json:"extra_time,omitempty"`     // If match went to extra time
	Penalties     *struct {
		Home *int `json:"home,omitempty"`
		Away *int `json:"away,omitempty"`
	} `json:"penalties,omitempty"`

	// Extended statistics
	Statistics []MatchStatistic `json:"statistics,omitempty"` // Match statistics (possession, shots, etc.)

	// Match context
	Referee    string `json:"referee,omitempty"`    // Referee name
	Attendance int    `json:"attendance,omitempty"` // Stadium attendance

	// Team formations
	HomeFormation string `json:"home_formation,omitempty"` // e.g., "4-3-3"
	AwayFormation string `json:"away_formation,omitempty"` // e.g., "4-4-2"

	// Starting lineups with full details
	HomeStarting []PlayerInfo `json:"home_starting,omitempty"`
	AwayStarting []PlayerInfo `json:"away_starting,omitempty"`

	// Substitutes
	HomeSubstitutes []PlayerInfo `json:"home_substitutes,omitempty"`
	AwaySubstitutes []PlayerInfo `json:"away_substitutes,omitempty"`

	// Momentum/xG data (if available)
	HomeXG *float64 `json:"home_xg,omitempty"` // Expected goals for home team
	AwayXG *float64 `json:"away_xg,omitempty"` // Expected goals for away team

	// Highlight video (if available)
	Highlight *MatchHighlight `json:"highlight,omitempty"` // Official highlight video link
}

// MatchHighlight represents an official highlight video for a match
type MatchHighlight struct {
	URL    string `json:"url"`              // Direct link to highlight video
	Image  string `json:"image,omitempty"`  // Thumbnail image URL
	Source string `json:"source,omitempty"` // Video source (e.g., "www.youtube.com")
	Title  string `json:"title,omitempty"`  // Video title (optional)
}

// LeagueTableEntry represents a team's position in the league table
type LeagueTableEntry struct {
	Position       int  `json:"position"`
	Team           Team `json:"team"`
	Played         int  `json:"played"`
	Won            int  `json:"won"`
	Drawn          int  `json:"drawn"`
	Lost           int  `json:"lost"`
	GoalsFor       int  `json:"goals_for"`
	GoalsAgainst   int  `json:"goals_against"`
	GoalDifference int  `json:"goal_difference"`
	Points         int  `json:"points"`
}
