package fotmob

import (
	"encoding/json"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/0xjuanma/golazo/internal/api"
)

// fotmobMatch represents a match in FotMob's API format
// Note: FotMob uses string IDs in JSON, but we convert them to ints
type fotmobMatch struct {
	ID     string `json:"id"` // FotMob returns string IDs
	Round  string `json:"round"`
	Home   team   `json:"home"`
	Away   team   `json:"away"`
	Status status `json:"status"`
	League league `json:"league"`
}

type team struct {
	ID        string `json:"id"` // FotMob returns string IDs
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
	UTCTime   string    `json:"utcTime"`   // Can be null/empty
	Started   *bool     `json:"started"`   // Can be null
	Finished  *bool     `json:"finished"`  // Can be null
	Cancelled *bool     `json:"cancelled"` // Can be null
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
	// Convert string IDs to ints
	matchID := parseInt(m.ID)
	homeID := parseInt(m.Home.ID)
	awayID := parseInt(m.Away.ID)

	match := api.Match{
		ID: matchID,
		League: api.League{
			ID:          m.League.ID,
			Name:        m.League.Name,
			Country:     m.League.Country,
			CountryCode: m.League.CountryCode,
		},
		HomeTeam: api.Team{
			ID:        homeID,
			Name:      m.Home.Name,
			ShortName: m.Home.ShortName,
		},
		AwayTeam: api.Team{
			ID:        awayID,
			Name:      m.Away.Name,
			ShortName: m.Away.ShortName,
		},
		Round: m.Round,
	}

	// Parse match time - FotMob uses .000Z format sometimes
	if m.Status.UTCTime != "" {
		var t time.Time
		var err error
		t, err = time.Parse(time.RFC3339, m.Status.UTCTime)
		if err != nil {
			// Try alternative format with milliseconds
			t, err = time.Parse("2006-01-02T15:04:05.000Z", m.Status.UTCTime)
		}
		if err == nil {
			match.MatchTime = &t
		}
	}

	// Determine status - handle null boolean values
	if m.Status.Cancelled != nil && *m.Status.Cancelled {
		match.Status = api.MatchStatusCancelled
	} else if m.Status.Finished != nil && *m.Status.Finished {
		match.Status = api.MatchStatusFinished
	} else if m.Status.Started != nil && *m.Status.Started {
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
// Note: FotMob API returns a nested structure with content.matchFacts containing events
type fotmobMatchDetails struct {
	Header struct {
		Teams []struct {
			ID    int    `json:"id"`
			Name  string `json:"name"`
			Score int    `json:"score"`
		} `json:"teams"`
		Status status `json:"status"`
	} `json:"header"`
	General struct {
		MatchID  string `json:"matchId"`
		Round    string `json:"matchRound"`
		HomeTeam struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"homeTeam"`
		AwayTeam struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"awayTeam"`
		LeagueID   int    `json:"leagueId"`
		LeagueName string `json:"leagueName"`
	} `json:"general"`
	Content struct {
		MatchFacts struct {
			Events struct {
				Events []fotmobEventDetail `json:"events"`
			} `json:"events"`
			InfoBox struct {
				Stadium struct {
					Name string `json:"name"`
				} `json:"Stadium,omitempty"`
				Referee *struct {
					Text string `json:"text"`
				} `json:"Referee,omitempty"`
				Attendance json.RawMessage `json:"Attendance,omitempty"` // Can be int or object
			} `json:"infoBox,omitempty"`
		} `json:"matchFacts"`
		Stats struct {
			Periods struct {
				All struct {
					Stats []fotmobStatCategory `json:"stats"`
				} `json:"all,omitempty"`
			} `json:"periods,omitempty"`
		} `json:"stats,omitempty"`
		Lineup struct {
			Lineup []fotmobTeamLineup `json:"lineup"`
		} `json:"lineup,omitempty"`
	} `json:"content"`
}

// fotmobStatCategory represents a category of match statistics
type fotmobStatCategory struct {
	Title string           `json:"title"`
	Stats []fotmobStatItem `json:"stats"`
}

// fotmobStatItem represents a single statistic item
type fotmobStatItem struct {
	Key       string        `json:"key"`
	Title     string        `json:"title"`
	Stats     []interface{} `json:"stats"` // [homeValue, awayValue] - can be int, float, or string
	Type      string        `json:"type,omitempty"`
	Highlight string        `json:"highlight,omitempty"` // "home" or "away" for who's better
}

// fotmobTeamLineup represents a team's lineup
type fotmobTeamLineup struct {
	TeamID     int                  `json:"teamId"`
	TeamName   string               `json:"teamName"`
	Formation  string               `json:"formation"`
	Bench      []fotmobPlayerInfo   `json:"bench"`
	Players    [][]fotmobPlayerInfo `json:"players"` // Grouped by position rows
	OptaLineup *struct {
		Starting []fotmobPlayerInfo `json:"starting"`
	} `json:"optaLineup,omitempty"`
}

// fotmobPlayerInfo represents player information in lineups
type fotmobPlayerInfo struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Shirt    int    `json:"shirt"`
	Position string `json:"position,omitempty"`
	Role     string `json:"role,omitempty"`
	Rating   *struct {
		Num string `json:"num"`
	} `json:"rating,omitempty"`
}

// fotmobEventDetail represents a single event detail from FotMob
type fotmobEventDetail struct {
	Time    int         `json:"time"`
	TimeStr interface{} `json:"timeStr"` // Can be int or string
	Type    string      `json:"type"`
	EventID int         `json:"eventId"`
	IsHome  bool        `json:"isHome"`
	Player  *struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"player,omitempty"`
	PlayerID  *int   `json:"playerId,omitempty"`
	NameStr   string `json:"nameStr,omitempty"`
	FirstName string `json:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty"`
	FullName  string `json:"fullName,omitempty"`
	HomeScore int    `json:"homeScore"`
	AwayScore int    `json:"awayScore"`
	NewScore  []int  `json:"newScore,omitempty"`
	OwnGoal   *bool  `json:"ownGoal,omitempty"`
	IsPenalty *bool  `json:"isPenalty,omitempty"`
	Card      string `json:"card,omitempty"` // "Yellow" or "Red"
	Swap      []struct {
		Name string `json:"name"`
		ID   string `json:"id"`
	} `json:"swap,omitempty"` // For substitutions
	AssistStr      string `json:"assistStr,omitempty"`
	AssistInput    string `json:"assistInput,omitempty"`
	AssistPlayerID *int   `json:"assistPlayerId,omitempty"`
}

// toAPIMatchDetails converts fotmobMatchDetails to api.MatchDetails
func (m fotmobMatchDetails) toAPIMatchDetails() *api.MatchDetails {
	// Parse match ID from string
	matchID := parseInt(m.General.MatchID)

	// Determine match status from header
	var status api.MatchStatus
	var liveTime *string
	if m.Header.Status.Cancelled != nil && *m.Header.Status.Cancelled {
		status = api.MatchStatusCancelled
	} else if m.Header.Status.Finished != nil && *m.Header.Status.Finished {
		status = api.MatchStatusFinished
	} else if m.Header.Status.Started != nil && *m.Header.Status.Started {
		status = api.MatchStatusLive
		if m.Header.Status.LiveTime != nil {
			liveTime = &m.Header.Status.LiveTime.Short
		}
	} else {
		status = api.MatchStatusNotStarted
	}

	// Parse match time
	var matchTime *time.Time
	if m.Header.Status.UTCTime != "" {
		t, err := time.Parse(time.RFC3339, m.Header.Status.UTCTime)
		if err != nil {
			t, err = time.Parse("2006-01-02T15:04:05.000Z", m.Header.Status.UTCTime)
		}
		if err == nil {
			matchTime = &t
		}
	}

	// Build the base match
	baseMatch := api.Match{
		ID: matchID,
		League: api.League{
			ID:   m.General.LeagueID,
			Name: m.General.LeagueName,
		},
		HomeTeam: api.Team{
			ID:        m.General.HomeTeam.ID,
			Name:      m.General.HomeTeam.Name,
			ShortName: m.General.HomeTeam.Name, // Use full name as short name if not available
		},
		AwayTeam: api.Team{
			ID:        m.General.AwayTeam.ID,
			Name:      m.General.AwayTeam.Name,
			ShortName: m.General.AwayTeam.Name, // Use full name as short name if not available
		},
		Status:    status,
		LiveTime:  liveTime,
		MatchTime: matchTime,
		Round:     m.General.Round,
	}

	details := &api.MatchDetails{
		Match:  baseMatch,
		Events: make([]api.MatchEvent, 0),
	}

	// Populate scores from header.Teams
	if len(m.Header.Teams) >= 2 {
		homeScore := m.Header.Teams[0].Score
		awayScore := m.Header.Teams[1].Score
		details.Match.HomeScore = &homeScore
		details.Match.AwayScore = &awayScore

		// Determine winner for finished matches
		if status == api.MatchStatusFinished {
			if homeScore > awayScore {
				winner := "home"
				details.Winner = &winner
			} else if awayScore > homeScore {
				winner := "away"
				details.Winner = &winner
			}
		}
	}

	// Populate venue from infoBox
	if m.Content.MatchFacts.InfoBox.Stadium.Name != "" {
		details.Venue = m.Content.MatchFacts.InfoBox.Stadium.Name
	}

	// Populate referee
	if m.Content.MatchFacts.InfoBox.Referee != nil {
		details.Referee = m.Content.MatchFacts.InfoBox.Referee.Text
	}

	// Populate attendance
	// Populate attendance (can be int or object with "number" field)
	if len(m.Content.MatchFacts.InfoBox.Attendance) > 0 {
		// Try to parse as int first
		var attendanceInt int
		if err := json.Unmarshal(m.Content.MatchFacts.InfoBox.Attendance, &attendanceInt); err == nil {
			details.Attendance = attendanceInt
		} else {
			// Try to parse as object with "number" field
			var attendanceObj struct {
				Number int `json:"number"`
			}
			if err := json.Unmarshal(m.Content.MatchFacts.InfoBox.Attendance, &attendanceObj); err == nil {
				details.Attendance = attendanceObj.Number
			}
		}
	}

	// Extract half-time score from events (look for "Half" event type)
	// Also set match duration (default to 90, but can be 120 for extra time)
	details.MatchDuration = 90
	for _, e := range m.Content.MatchFacts.Events.Events {
		if e.Type == "Half" && (e.HomeScore > 0 || e.AwayScore > 0) {
			// Found half-time score
			if details.HalfTimeScore == nil {
				details.HalfTimeScore = &struct {
					Home *int `json:"home,omitempty"`
					Away *int `json:"away,omitempty"`
				}{}
			}
			htHome := e.HomeScore
			htAway := e.AwayScore
			details.HalfTimeScore.Home = &htHome
			details.HalfTimeScore.Away = &htAway
		}
		// Check for extra time indicators (events after 90 minutes)
		if e.Time > 90 {
			details.ExtraTime = true
			details.MatchDuration = 120
		}
	}

	// Parse match statistics
	details.Statistics = m.parseStatistics()

	// Parse lineup information
	m.parseLineups(details)

	// Convert events from content.matchFacts.events
	events := make([]api.MatchEvent, 0, len(m.Content.MatchFacts.Events.Events))
	for _, e := range m.Content.MatchFacts.Events.Events {
		// Skip non-event types like "Half"
		if e.Type == "Half" {
			continue
		}

		// Normalize event type to lowercase for consistent matching
		eventType := strings.ToLower(e.Type)

		event := api.MatchEvent{
			ID:        e.EventID,
			Minute:    e.Time,
			Type:      eventType,
			Timestamp: time.Now(),
		}

		// Extract player name
		playerName := ""
		if e.Player != nil && e.Player.Name != "" {
			playerName = e.Player.Name
		} else if e.FullName != "" {
			playerName = e.FullName
		} else if e.NameStr != "" {
			playerName = e.NameStr
		}
		if playerName != "" {
			event.Player = &playerName
		}

		// Extract assist
		if e.AssistInput != "" {
			event.Assist = &e.AssistInput
		}

		// Extract event type details
		eventTypeDetail := ""
		if e.Type == "Card" && e.Card != "" {
			eventTypeDetail = strings.ToLower(e.Card)
		} else if e.Type == "Substitution" && len(e.Swap) >= 2 {
			// Substitution: swap[0] is player going out, swap[1] is player coming in
			// Store player out in Player field, player in in Assist field (repurposed)
			playerOut := e.Swap[0].Name
			playerIn := e.Swap[1].Name
			event.Player = &playerOut
			event.Assist = &playerIn // Repurpose Assist to store player coming in
			eventTypeDetail = "sub"
		} else if strings.ToLower(e.Type) == "addedtime" {
			// Added time event - extract minutes from available fields
			eventTypeDetail = "addedtime"
			// Try multiple sources for added time info
			var addedTimeStr string
			// Check timeStr (can be string or int)
			if timeStrVal, ok := e.TimeStr.(string); ok && timeStrVal != "" {
				addedTimeStr = timeStrVal
			} else if timeStrInt, ok := e.TimeStr.(float64); ok {
				addedTimeStr = strconv.Itoa(int(timeStrInt))
			}
			// Check nameStr which sometimes contains the added time
			if addedTimeStr == "" && e.NameStr != "" {
				addedTimeStr = e.NameStr
			}
			// Check if there's a player field with the info
			if addedTimeStr == "" && e.Player != nil && e.Player.Name != "" {
				addedTimeStr = e.Player.Name
			}
			if addedTimeStr != "" {
				event.Player = &addedTimeStr
			}
		}
		if eventTypeDetail != "" {
			event.EventType = &eventTypeDetail
		}

		// Set team based on isHome flag
		homeIDInt := m.General.HomeTeam.ID
		awayIDInt := m.General.AwayTeam.ID
		if e.IsHome {
			event.Team = api.Team{
				ID:        homeIDInt,
				Name:      m.General.HomeTeam.Name,
				ShortName: m.General.HomeTeam.Name,
			}
		} else {
			event.Team = api.Team{
				ID:        awayIDInt,
				Name:      m.General.AwayTeam.Name,
				ShortName: m.General.AwayTeam.Name,
			}
		}

		events = append(events, event)
	}

	// Sort events by minute (chronological order)
	sort.Slice(events, func(i, j int) bool {
		return events[i].Minute < events[j].Minute
	})

	details.Events = events
	return details
}

// parseStatistics extracts match statistics from FotMob response
func (m fotmobMatchDetails) parseStatistics() []api.MatchStatistic {
	var stats []api.MatchStatistic

	for _, category := range m.Content.Stats.Periods.All.Stats {
		for _, stat := range category.Stats {
			if len(stat.Stats) < 2 {
				continue
			}

			// Convert stat values to strings
			homeVal := formatStatValue(stat.Stats[0])
			awayVal := formatStatValue(stat.Stats[1])

			// Skip empty stats
			if homeVal == "" && awayVal == "" {
				continue
			}

			stats = append(stats, api.MatchStatistic{
				Key:       stat.Key,
				Label:     stat.Title,
				HomeValue: homeVal,
				AwayValue: awayVal,
			})
		}
	}

	return stats
}

// formatStatValue converts a stat value (can be int, float, or string) to string
func formatStatValue(val interface{}) string {
	switch v := val.(type) {
	case string:
		return v
	case float64:
		// Check if it's a whole number
		if v == float64(int(v)) {
			return strconv.Itoa(int(v))
		}
		return strconv.FormatFloat(v, 'f', 1, 64)
	case int:
		return strconv.Itoa(v)
	default:
		return ""
	}
}

// parseLineups extracts lineup information from FotMob response
func (m fotmobMatchDetails) parseLineups(details *api.MatchDetails) {
	for _, lineup := range m.Content.Lineup.Lineup {
		isHome := lineup.TeamID == m.General.HomeTeam.ID

		// Set formation
		if isHome {
			details.HomeFormation = lineup.Formation
		} else {
			details.AwayFormation = lineup.Formation
		}

		// Extract starting players from the nested players array
		var starting []api.PlayerInfo
		for _, row := range lineup.Players {
			for _, p := range row {
				player := api.PlayerInfo{
					ID:       p.ID,
					Name:     p.Name,
					Number:   p.Shirt,
					Position: p.Position,
				}
				if p.Rating != nil {
					player.Rating = p.Rating.Num
				}
				starting = append(starting, player)
			}
		}

		// Extract substitutes
		var substitutes []api.PlayerInfo
		for _, p := range lineup.Bench {
			player := api.PlayerInfo{
				ID:       p.ID,
				Name:     p.Name,
				Number:   p.Shirt,
				Position: p.Position,
			}
			if p.Rating != nil {
				player.Rating = p.Rating.Num
			}
			substitutes = append(substitutes, player)
		}

		if isHome {
			details.HomeStarting = starting
			details.HomeSubstitutes = substitutes
		} else {
			details.AwayStarting = starting
			details.AwaySubstitutes = substitutes
		}
	}
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
		Position: r.Rank,
		Team: api.Team{
			ID:        r.ID,
			Name:      r.Name,
			ShortName: r.ShortName,
		},
		Played:         r.Played,
		Won:            r.Wins,
		Drawn:          r.Draws,
		Lost:           r.Losses,
		GoalsFor:       r.GoalsFor,
		GoalsAgainst:   r.GoalsAgainst,
		GoalDifference: r.GoalDifference,
		Points:         r.Points,
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
// Returns 0 if parsing fails (for required fields)
func parseInt(s string) int {
	if s == "" {
		return 0
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return val
}
