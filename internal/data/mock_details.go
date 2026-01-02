package data

import (
	"time"

	"github.com/0xjuanma/golazo/internal/api"
)

// MockMatchDetails returns detailed match information for a specific match ID.
// Handles both live matches (from MockLiveMatches) and finished matches.
func MockMatchDetails(matchID int) (*api.MatchDetails, error) {
	// Try live matches first
	liveMatches := MockLiveMatches()
	for i := range liveMatches {
		if liveMatches[i].ID == matchID {
			events := generateLiveMatchEvents(matchID, liveMatches[i])
			stats := generateMockStatistics(matchID)
			return &api.MatchDetails{
				Match:      liveMatches[i],
				Events:     events,
				Statistics: stats,
				Venue:      getMockVenue(matchID),
				Referee:    getMockReferee(matchID),
				Attendance: getMockAttendance(matchID),
			}, nil
		}
	}

	// Then try finished matches
	finishedMatches := MockFinishedMatches()
	for i := range finishedMatches {
		if finishedMatches[i].ID == matchID {
			events := generateFinishedMatchEvents(matchID, finishedMatches[i])
			stats := generateMockStatistics(matchID)
			return &api.MatchDetails{
				Match:      finishedMatches[i],
				Events:     events,
				Statistics: stats,
				Venue:      getMockVenue(matchID),
				Referee:    getMockReferee(matchID),
				Attendance: getMockAttendance(matchID),
			}, nil
		}
	}

	return nil, nil
}

// generateLiveMatchEvents generates events for ongoing/just finished matches in live view.
func generateLiveMatchEvents(matchID int, match api.Match) []api.MatchEvent {
	events := []api.MatchEvent{}

	switch matchID {
	// ═══════════════════════════════════════════════
	// ONGOING MATCHES
	// ═══════════════════════════════════════════════

	case 2001: // Chelsea 2-1 Spurs (67') - Premier League
		events = []api.MatchEvent{
			{ID: 1, Minute: 12, Type: "goal", Team: match.HomeTeam, Player: stringPtr("Palmer"), Timestamp: time.Now()},
			{ID: 2, Minute: 23, Type: "card", Team: match.AwayTeam, Player: stringPtr("Romero"), EventType: stringPtr("yellow"), Timestamp: time.Now()},
			{ID: 3, Minute: 34, Type: "goal", Team: match.AwayTeam, Player: stringPtr("Son"), Assist: stringPtr("Maddison"), Timestamp: time.Now()},
			{ID: 4, Minute: 45, Type: "substitution", Team: match.HomeTeam, Player: stringPtr("Mudryk"), EventType: stringPtr("sub_in"), Timestamp: time.Now()},
			{ID: 5, Minute: 56, Type: "goal", Team: match.HomeTeam, Player: stringPtr("Jackson"), Assist: stringPtr("Palmer"), Timestamp: time.Now()},
			{ID: 6, Minute: 62, Type: "card", Team: match.HomeTeam, Player: stringPtr("Caicedo"), EventType: stringPtr("yellow"), Timestamp: time.Now()},
		}

	case 2002: // Real Madrid 1-1 Atletico (34') - La Liga
		events = []api.MatchEvent{
			{ID: 7, Minute: 8, Type: "goal", Team: match.AwayTeam, Player: stringPtr("Griezmann"), Timestamp: time.Now()},
			{ID: 8, Minute: 18, Type: "card", Team: match.AwayTeam, Player: stringPtr("Savic"), EventType: stringPtr("yellow"), Timestamp: time.Now()},
			{ID: 9, Minute: 28, Type: "goal", Team: match.HomeTeam, Player: stringPtr("Bellingham"), Assist: stringPtr("Vinicius Jr"), Timestamp: time.Now()},
		}

	case 2003: // Man City 3-2 Bayern (56') - Champions League
		events = []api.MatchEvent{
			{ID: 10, Minute: 5, Type: "goal", Team: match.HomeTeam, Player: stringPtr("Haaland"), Timestamp: time.Now()},
			{ID: 11, Minute: 15, Type: "goal", Team: match.AwayTeam, Player: stringPtr("Kane"), Assist: stringPtr("Sane"), Timestamp: time.Now()},
			{ID: 12, Minute: 23, Type: "card", Team: match.HomeTeam, Player: stringPtr("Rodri"), EventType: stringPtr("yellow"), Timestamp: time.Now()},
			{ID: 13, Minute: 34, Type: "goal", Team: match.HomeTeam, Player: stringPtr("De Bruyne"), Timestamp: time.Now()},
			{ID: 14, Minute: 42, Type: "goal", Team: match.AwayTeam, Player: stringPtr("Musiala"), Timestamp: time.Now()},
			{ID: 15, Minute: 45, Type: "substitution", Team: match.AwayTeam, Player: stringPtr("Coman"), EventType: stringPtr("sub_in"), Timestamp: time.Now()},
			{ID: 16, Minute: 52, Type: "goal", Team: match.HomeTeam, Player: stringPtr("Foden"), Assist: stringPtr("Haaland"), Timestamp: time.Now()},
		}

	// ═══════════════════════════════════════════════
	// JUST FINISHED MATCHES (in live view, showing all events)
	// ═══════════════════════════════════════════════

	case 2004: // Arsenal 2-3 Liverpool (FT) - Premier League
		events = []api.MatchEvent{
			{ID: 17, Minute: 8, Type: "goal", Team: match.AwayTeam, Player: stringPtr("Salah"), Timestamp: time.Now()},
			{ID: 18, Minute: 15, Type: "card", Team: match.HomeTeam, Player: stringPtr("Rice"), EventType: stringPtr("yellow"), Timestamp: time.Now()},
			{ID: 19, Minute: 23, Type: "goal", Team: match.HomeTeam, Player: stringPtr("Saka"), Assist: stringPtr("Odegaard"), Timestamp: time.Now()},
			{ID: 20, Minute: 34, Type: "substitution", Team: match.AwayTeam, Player: stringPtr("Gakpo"), EventType: stringPtr("sub_in"), Timestamp: time.Now()},
			{ID: 21, Minute: 45, Type: "goal", Team: match.AwayTeam, Player: stringPtr("Nunez"), Timestamp: time.Now()},
			{ID: 22, Minute: 56, Type: "card", Team: match.AwayTeam, Player: stringPtr("Van Dijk"), EventType: stringPtr("yellow"), Timestamp: time.Now()},
			{ID: 23, Minute: 67, Type: "goal", Team: match.HomeTeam, Player: stringPtr("Martinelli"), Timestamp: time.Now()},
			{ID: 24, Minute: 78, Type: "substitution", Team: match.HomeTeam, Player: stringPtr("Trossard"), EventType: stringPtr("sub_in"), Timestamp: time.Now()},
			{ID: 25, Minute: 85, Type: "card", Team: match.HomeTeam, Player: stringPtr("Gabriel"), EventType: stringPtr("red"), Timestamp: time.Now()},
			{ID: 26, Minute: 89, Type: "goal", Team: match.AwayTeam, Player: stringPtr("Diaz"), Assist: stringPtr("Salah"), Timestamp: time.Now()},
		}

	case 2005: // Barcelona 4-1 Sevilla (FT) - La Liga
		events = []api.MatchEvent{
			{ID: 27, Minute: 12, Type: "goal", Team: match.HomeTeam, Player: stringPtr("Lewandowski"), Timestamp: time.Now()},
			{ID: 28, Minute: 23, Type: "goal", Team: match.HomeTeam, Player: stringPtr("Yamal"), Assist: stringPtr("Pedri"), Timestamp: time.Now()},
			{ID: 29, Minute: 34, Type: "card", Team: match.AwayTeam, Player: stringPtr("Gudelj"), EventType: stringPtr("yellow"), Timestamp: time.Now()},
			{ID: 30, Minute: 45, Type: "goal", Team: match.AwayTeam, Player: stringPtr("Lukebakio"), Timestamp: time.Now()},
			{ID: 31, Minute: 56, Type: "substitution", Team: match.HomeTeam, Player: stringPtr("Ferran Torres"), EventType: stringPtr("sub_in"), Timestamp: time.Now()},
			{ID: 32, Minute: 67, Type: "goal", Team: match.HomeTeam, Player: stringPtr("Raphinha"), Timestamp: time.Now()},
			{ID: 33, Minute: 78, Type: "card", Team: match.HomeTeam, Player: stringPtr("Araujo"), EventType: stringPtr("yellow"), Timestamp: time.Now()},
			{ID: 34, Minute: 89, Type: "goal", Team: match.HomeTeam, Player: stringPtr("Lewandowski"), Assist: stringPtr("Yamal"), Timestamp: time.Now()},
		}
	}

	return events
}

// generateMockStatistics generates mock statistics for a match.
func generateMockStatistics(matchID int) []api.MatchStatistic {
	// Different stats based on match ID for variety
	switch matchID {
	case 2001, 1001: // Chelsea matches
		return []api.MatchStatistic{
			{Key: "possession", Label: "Possession %", HomeValue: "58", AwayValue: "42"},
			{Key: "shots_total", Label: "Total Shots", HomeValue: "14", AwayValue: "8"},
			{Key: "shots_on_target", Label: "Shots on Target", HomeValue: "6", AwayValue: "3"},
			{Key: "corners", Label: "Corners", HomeValue: "7", AwayValue: "4"},
			{Key: "fouls", Label: "Fouls", HomeValue: "9", AwayValue: "12"},
		}
	case 2002, 1003: // Madrid matches
		return []api.MatchStatistic{
			{Key: "possession", Label: "Possession %", HomeValue: "52", AwayValue: "48"},
			{Key: "shots_total", Label: "Total Shots", HomeValue: "11", AwayValue: "9"},
			{Key: "shots_on_target", Label: "Shots on Target", HomeValue: "4", AwayValue: "4"},
			{Key: "corners", Label: "Corners", HomeValue: "5", AwayValue: "6"},
			{Key: "fouls", Label: "Fouls", HomeValue: "11", AwayValue: "14"},
		}
	case 2003, 1005: // Champions League matches
		return []api.MatchStatistic{
			{Key: "possession", Label: "Possession %", HomeValue: "62", AwayValue: "38"},
			{Key: "shots_total", Label: "Total Shots", HomeValue: "18", AwayValue: "10"},
			{Key: "shots_on_target", Label: "Shots on Target", HomeValue: "8", AwayValue: "5"},
			{Key: "corners", Label: "Corners", HomeValue: "9", AwayValue: "3"},
			{Key: "fouls", Label: "Fouls", HomeValue: "7", AwayValue: "10"},
		}
	default:
		return []api.MatchStatistic{
			{Key: "possession", Label: "Possession %", HomeValue: "50", AwayValue: "50"},
			{Key: "shots_total", Label: "Total Shots", HomeValue: "12", AwayValue: "10"},
			{Key: "shots_on_target", Label: "Shots on Target", HomeValue: "5", AwayValue: "4"},
			{Key: "corners", Label: "Corners", HomeValue: "6", AwayValue: "5"},
			{Key: "fouls", Label: "Fouls", HomeValue: "10", AwayValue: "11"},
		}
	}
}

func getMockVenue(matchID int) string {
	venues := map[int]string{
		2001: "Stamford Bridge",
		2002: "Santiago Bernabeu",
		2003: "Etihad Stadium",
		2004: "Emirates Stadium",
		2005: "Camp Nou",
		1001: "Etihad Stadium",
		1002: "Old Trafford",
		1003: "Santiago Bernabeu",
		1004: "Civitas Metropolitano",
		1005: "Parc des Princes",
		1006: "San Siro",
	}
	if v, ok := venues[matchID]; ok {
		return v
	}
	return "Stadium"
}

func getMockReferee(matchID int) string {
	referees := map[int]string{
		2001: "Michael Oliver",
		2002: "Mateu Lahoz",
		2003: "Daniele Orsato",
		2004: "Anthony Taylor",
		2005: "Jesus Gil Manzano",
		1001: "Stuart Attwell",
		1002: "Michael Oliver",
		1003: "Juan Martinez Munuera",
		1004: "Carlos del Cerro Grande",
		1005: "Clement Turpin",
		1006: "Felix Brych",
	}
	if r, ok := referees[matchID]; ok {
		return r
	}
	return "Unknown"
}

func getMockAttendance(matchID int) int {
	attendances := map[int]int{
		2001: 40341,
		2002: 81044,
		2003: 53400,
		2004: 60260,
		2005: 85764,
		1001: 53400,
		1002: 74310,
		1003: 81044,
		1004: 68456,
		1005: 48583,
		1006: 75923,
	}
	if a, ok := attendances[matchID]; ok {
		return a
	}
	return 50000
}

func stringPtr(s string) *string {
	return &s
}
