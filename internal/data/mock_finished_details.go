package data

import (
	"time"

	"github.com/0xjuanma/golazo/internal/api"
)

// MockFinishedMatchDetails returns detailed match information for finished matches.
// These matches have rich event data for testing the "all events" view.
func MockFinishedMatchDetails(matchID int) (*api.MatchDetails, error) {
	finishedMatches := MockFinishedMatches()

	// Find the match
	var match *api.Match
	for i := range finishedMatches {
		if finishedMatches[i].ID == matchID {
			match = &finishedMatches[i]
			break
		}
	}

	if match == nil {
		return nil, nil
	}

	// Generate events and stats
	events := generateFinishedMatchEvents(matchID, *match)
	stats := generateMockStatistics(matchID)

	details := &api.MatchDetails{
		Match:      *match,
		Events:     events,
		Statistics: stats,
		Venue:      getMockVenue(matchID),
		Referee:    getMockReferee(matchID),
		Attendance: getMockAttendance(matchID),
	}

	// Add mock highlights for some matches to demonstrate the feature
	if highlight := getMockHighlight(matchID); highlight != nil {
		details.Highlight = highlight
	}

	return details, nil
}

// generateFinishedMatchEvents generates comprehensive events for finished matches.
// These include goals, cards, and substitutions to test the full "all events" view.
func generateFinishedMatchEvents(matchID int, match api.Match) []api.MatchEvent {
	events := []api.MatchEvent{}

	switch matchID {
	// ═══════════════════════════════════════════════
	// TODAY'S MATCHES
	// ═══════════════════════════════════════════════

	case 1010: // Newcastle 2-1 Aston Villa
		events = []api.MatchEvent{
			{ID: 51, Minute: 18, Type: "goal", Team: match.HomeTeam, Player: stringPtr("Isak"), Timestamp: time.Now()},
			{ID: 52, Minute: 34, Type: "card", Team: match.AwayTeam, Player: stringPtr("Konsa"), EventType: stringPtr("yellow"), Timestamp: time.Now()},
			{ID: 53, Minute: 56, Type: "goal", Team: match.AwayTeam, Player: stringPtr("Watkins"), Assist: stringPtr("McGinn"), Timestamp: time.Now()},
			{ID: 54, Minute: 78, Type: "goal", Team: match.HomeTeam, Player: stringPtr("Gordon"), Timestamp: time.Now()},
		}

	case 1011: // Valencia 0-2 Athletic Bilbao
		events = []api.MatchEvent{
			{ID: 55, Minute: 23, Type: "card", Team: match.HomeTeam, Player: stringPtr("Mosquera"), EventType: stringPtr("yellow"), Timestamp: time.Now()},
			{ID: 56, Minute: 45, Type: "goal", Team: match.AwayTeam, Player: stringPtr("Williams"), Timestamp: time.Now()},
			{ID: 57, Minute: 67, Type: "substitution", Team: match.HomeTeam, Player: stringPtr("Hugo Duro"), EventType: stringPtr("sub_in"), Timestamp: time.Now()},
			{ID: 58, Minute: 82, Type: "goal", Team: match.AwayTeam, Player: stringPtr("Sancet"), Assist: stringPtr("Williams"), Timestamp: time.Now()},
		}

	case 1012: // Napoli 3-1 Roma
		events = []api.MatchEvent{
			{ID: 59, Minute: 12, Type: "goal", Team: match.HomeTeam, Player: stringPtr("Osimhen"), Timestamp: time.Now()},
			{ID: 60, Minute: 28, Type: "goal", Team: match.AwayTeam, Player: stringPtr("Dybala"), Timestamp: time.Now()},
			{ID: 61, Minute: 45, Type: "card", Team: match.AwayTeam, Player: stringPtr("Cristante"), EventType: stringPtr("yellow"), Timestamp: time.Now()},
			{ID: 62, Minute: 56, Type: "goal", Team: match.HomeTeam, Player: stringPtr("Kvaratskhelia"), Assist: stringPtr("Osimhen"), Timestamp: time.Now()},
			{ID: 63, Minute: 78, Type: "substitution", Team: match.HomeTeam, Player: stringPtr("Simeone"), EventType: stringPtr("sub_in"), Timestamp: time.Now()},
			{ID: 64, Minute: 89, Type: "goal", Team: match.HomeTeam, Player: stringPtr("Politano"), Timestamp: time.Now()},
		}

	// ═══════════════════════════════════════════════
	// PREMIER LEAGUE
	// ═══════════════════════════════════════════════

	case 1001: // Man City 2-1 Arsenal
		events = []api.MatchEvent{
			{ID: 1, Minute: 12, Type: "goal", Team: match.HomeTeam, Player: stringPtr("Haaland"), Timestamp: time.Now()},
			{ID: 2, Minute: 23, Type: "card", Team: match.AwayTeam, Player: stringPtr("Rice"), EventType: stringPtr("yellow"), Timestamp: time.Now()},
			{ID: 3, Minute: 34, Type: "goal", Team: match.AwayTeam, Player: stringPtr("Saka"), Assist: stringPtr("Odegaard"), Timestamp: time.Now()},
			{ID: 4, Minute: 45, Type: "substitution", Team: match.HomeTeam, Player: stringPtr("Grealish"), EventType: stringPtr("sub_in"), Timestamp: time.Now()},
			{ID: 5, Minute: 56, Type: "card", Team: match.HomeTeam, Player: stringPtr("Rodri"), EventType: stringPtr("yellow"), Timestamp: time.Now()},
			{ID: 6, Minute: 67, Type: "goal", Team: match.HomeTeam, Player: stringPtr("De Bruyne"), Assist: stringPtr("Foden"), Timestamp: time.Now()},
			{ID: 7, Minute: 78, Type: "substitution", Team: match.AwayTeam, Player: stringPtr("Trossard"), EventType: stringPtr("sub_in"), Timestamp: time.Now()},
			{ID: 8, Minute: 85, Type: "card", Team: match.AwayTeam, Player: stringPtr("Saliba"), EventType: stringPtr("yellow"), Timestamp: time.Now()},
		}

	case 1002: // Man Utd 0-3 Liverpool
		events = []api.MatchEvent{
			{ID: 9, Minute: 5, Type: "goal", Team: match.AwayTeam, Player: stringPtr("Salah"), Timestamp: time.Now()},
			{ID: 10, Minute: 15, Type: "card", Team: match.HomeTeam, Player: stringPtr("Casemiro"), EventType: stringPtr("yellow"), Timestamp: time.Now()},
			{ID: 11, Minute: 23, Type: "substitution", Team: match.HomeTeam, Player: stringPtr("Garnacho"), EventType: stringPtr("sub_in"), Timestamp: time.Now()},
			{ID: 12, Minute: 34, Type: "card", Team: match.AwayTeam, Player: stringPtr("Mac Allister"), EventType: stringPtr("yellow"), Timestamp: time.Now()},
			{ID: 13, Minute: 45, Type: "goal", Team: match.AwayTeam, Player: stringPtr("Nunez"), Assist: stringPtr("Salah"), Timestamp: time.Now()},
			{ID: 14, Minute: 56, Type: "substitution", Team: match.AwayTeam, Player: stringPtr("Gakpo"), EventType: stringPtr("sub_in"), Timestamp: time.Now()},
			{ID: 15, Minute: 67, Type: "card", Team: match.HomeTeam, Player: stringPtr("Martinez"), EventType: stringPtr("red"), Timestamp: time.Now()},
			{ID: 16, Minute: 78, Type: "substitution", Team: match.HomeTeam, Player: stringPtr("Hojlund"), EventType: stringPtr("sub_in"), Timestamp: time.Now()},
			{ID: 17, Minute: 89, Type: "goal", Team: match.AwayTeam, Player: stringPtr("Diaz"), Timestamp: time.Now()},
		}

	// ═══════════════════════════════════════════════
	// LA LIGA
	// ═══════════════════════════════════════════════

	case 1003: // Real Madrid 3-2 Barcelona (El Clasico)
		events = []api.MatchEvent{
			{ID: 18, Minute: 8, Type: "goal", Team: match.AwayTeam, Player: stringPtr("Lewandowski"), Timestamp: time.Now()},
			{ID: 19, Minute: 15, Type: "card", Team: match.HomeTeam, Player: stringPtr("Tchouameni"), EventType: stringPtr("yellow"), Timestamp: time.Now()},
			{ID: 20, Minute: 23, Type: "goal", Team: match.HomeTeam, Player: stringPtr("Vinicius Jr"), Timestamp: time.Now()},
			{ID: 21, Minute: 34, Type: "goal", Team: match.HomeTeam, Player: stringPtr("Bellingham"), Assist: stringPtr("Modric"), Timestamp: time.Now()},
			{ID: 22, Minute: 45, Type: "substitution", Team: match.AwayTeam, Player: stringPtr("Ferran Torres"), EventType: stringPtr("sub_in"), Timestamp: time.Now()},
			{ID: 23, Minute: 52, Type: "card", Team: match.AwayTeam, Player: stringPtr("Gavi"), EventType: stringPtr("yellow"), Timestamp: time.Now()},
			{ID: 24, Minute: 56, Type: "goal", Team: match.AwayTeam, Player: stringPtr("Pedri"), Timestamp: time.Now()},
			{ID: 25, Minute: 67, Type: "substitution", Team: match.HomeTeam, Player: stringPtr("Camavinga"), EventType: stringPtr("sub_in"), Timestamp: time.Now()},
			{ID: 26, Minute: 78, Type: "goal", Team: match.HomeTeam, Player: stringPtr("Rodrygo"), Assist: stringPtr("Vinicius Jr"), Timestamp: time.Now()},
			{ID: 27, Minute: 85, Type: "card", Team: match.AwayTeam, Player: stringPtr("Araujo"), EventType: stringPtr("yellow"), Timestamp: time.Now()},
		}

	case 1004: // Atletico 1-1 Sevilla
		events = []api.MatchEvent{
			{ID: 28, Minute: 23, Type: "goal", Team: match.HomeTeam, Player: stringPtr("Griezmann"), Assist: stringPtr("Morata"), Timestamp: time.Now()},
			{ID: 29, Minute: 34, Type: "card", Team: match.AwayTeam, Player: stringPtr("Gudelj"), EventType: stringPtr("yellow"), Timestamp: time.Now()},
			{ID: 30, Minute: 45, Type: "substitution", Team: match.HomeTeam, Player: stringPtr("Correa"), EventType: stringPtr("sub_in"), Timestamp: time.Now()},
			{ID: 31, Minute: 56, Type: "card", Team: match.HomeTeam, Player: stringPtr("Koke"), EventType: stringPtr("yellow"), Timestamp: time.Now()},
			{ID: 32, Minute: 78, Type: "goal", Team: match.AwayTeam, Player: stringPtr("Lukebakio"), Timestamp: time.Now()},
			{ID: 33, Minute: 89, Type: "card", Team: match.AwayTeam, Player: stringPtr("Acuna"), EventType: stringPtr("yellow"), Timestamp: time.Now()},
		}

	// ═══════════════════════════════════════════════
	// UEFA CHAMPIONS LEAGUE
	// ═══════════════════════════════════════════════

	case 1005: // PSG 2-3 Bayern
		events = []api.MatchEvent{
			{ID: 34, Minute: 8, Type: "goal", Team: match.AwayTeam, Player: stringPtr("Kane"), Timestamp: time.Now()},
			{ID: 35, Minute: 18, Type: "goal", Team: match.HomeTeam, Player: stringPtr("Mbappe"), Timestamp: time.Now()},
			{ID: 36, Minute: 28, Type: "card", Team: match.AwayTeam, Player: stringPtr("Upamecano"), EventType: stringPtr("yellow"), Timestamp: time.Now()},
			{ID: 37, Minute: 34, Type: "goal", Team: match.AwayTeam, Player: stringPtr("Musiala"), Assist: stringPtr("Sane"), Timestamp: time.Now()},
			{ID: 38, Minute: 45, Type: "substitution", Team: match.HomeTeam, Player: stringPtr("Kolo Muani"), EventType: stringPtr("sub_in"), Timestamp: time.Now()},
			{ID: 39, Minute: 56, Type: "card", Team: match.HomeTeam, Player: stringPtr("Vitinha"), EventType: stringPtr("yellow"), Timestamp: time.Now()},
			{ID: 40, Minute: 67, Type: "goal", Team: match.HomeTeam, Player: stringPtr("Dembele"), Assist: stringPtr("Mbappe"), Timestamp: time.Now()},
			{ID: 41, Minute: 78, Type: "substitution", Team: match.AwayTeam, Player: stringPtr("Coman"), EventType: stringPtr("sub_in"), Timestamp: time.Now()},
			{ID: 42, Minute: 85, Type: "goal", Team: match.AwayTeam, Player: stringPtr("Kane"), Assist: stringPtr("Muller"), Timestamp: time.Now()},
			{ID: 43, Minute: 90, Type: "card", Team: match.HomeTeam, Player: stringPtr("Marquinhos"), EventType: stringPtr("yellow"), Timestamp: time.Now()},
		}

	case 1006: // Inter 1-0 Dortmund
		events = []api.MatchEvent{
			{ID: 44, Minute: 15, Type: "card", Team: match.AwayTeam, Player: stringPtr("Hummels"), EventType: stringPtr("yellow"), Timestamp: time.Now()},
			{ID: 45, Minute: 34, Type: "substitution", Team: match.HomeTeam, Player: stringPtr("Thuram"), EventType: stringPtr("sub_in"), Timestamp: time.Now()},
			{ID: 46, Minute: 45, Type: "card", Team: match.HomeTeam, Player: stringPtr("Barella"), EventType: stringPtr("yellow"), Timestamp: time.Now()},
			{ID: 47, Minute: 56, Type: "substitution", Team: match.AwayTeam, Player: stringPtr("Malen"), EventType: stringPtr("sub_in"), Timestamp: time.Now()},
			{ID: 48, Minute: 67, Type: "goal", Team: match.HomeTeam, Player: stringPtr("Lautaro"), Assist: stringPtr("Calhanoglu"), Timestamp: time.Now()},
			{ID: 49, Minute: 78, Type: "card", Team: match.AwayTeam, Player: stringPtr("Sabitzer"), EventType: stringPtr("yellow"), Timestamp: time.Now()},
			{ID: 50, Minute: 89, Type: "substitution", Team: match.HomeTeam, Player: stringPtr("Arnautovic"), EventType: stringPtr("sub_in"), Timestamp: time.Now()},
		}
	}

	return events
}

// getMockHighlight returns mock highlight data for testing the highlights feature.
// Only some matches have highlights to simulate real-world availability.
func getMockHighlight(matchID int) *api.MatchHighlight {
	switch matchID {
	case 1001: // Man City 2-1 Arsenal (finished with highlights)
		return &api.MatchHighlight{
			URL:    "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
			Image:  "https://i.ytimg.com/vi/dQw4w9WgXcQ/maxresdefault.jpg",
			Source: "www.youtube.com",
			Title:  "Man City 2-1 Arsenal | Premier League Highlights",
		}
	case 1005: // PSG 2-3 Bayern (Champions League highlights)
		return &api.MatchHighlight{
			URL:    "https://www.youtube.com/watch?v=example123",
			Image:  "https://i.ytimg.com/vi/example123/maxresdefault.jpg",
			Source: "www.youtube.com",
			Title:  "PSG 2-3 Bayern | Champions League Highlights",
		}
	default:
		return nil // No highlights available for this match
	}
}
