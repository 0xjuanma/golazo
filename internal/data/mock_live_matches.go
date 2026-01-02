package data

import (
	"time"

	"github.com/0xjuanma/golazo/internal/api"
)

// MockLiveMatches returns live matches for the live matches view.
// 5 matches total: 3 ongoing with events, 2 just finished (showing "all events")
func MockLiveMatches() []api.Match {
	now := time.Now()

	return []api.Match{
		// ═══════════════════════════════════════════════
		// ONGOING MATCHES (3) - with live events
		// ═══════════════════════════════════════════════

		// Match 1: Premier League - Chelsea vs Tottenham (67')
		{
			ID: 2001,
			League: api.League{
				ID:   47,
				Name: "Premier League",
			},
			HomeTeam: api.Team{
				ID:        49,
				Name:      "Chelsea",
				ShortName: "Chelsea",
			},
			AwayTeam: api.Team{
				ID:        66,
				Name:      "Tottenham",
				ShortName: "Spurs",
			},
			Status:    api.MatchStatusLive,
			HomeScore: intPtr(2),
			AwayScore: intPtr(1),
			LiveTime:  stringPtr("67'"),
			MatchTime: &now,
			Round:     "Matchday 17",
		},

		// Match 2: La Liga - Real Madrid vs Atletico (34')
		{
			ID: 2002,
			League: api.League{
				ID:   87,
				Name: "La Liga",
			},
			HomeTeam: api.Team{
				ID:        541,
				Name:      "Real Madrid",
				ShortName: "Real Madrid",
			},
			AwayTeam: api.Team{
				ID:        531,
				Name:      "Atletico Madrid",
				ShortName: "Atletico",
			},
			Status:    api.MatchStatusLive,
			HomeScore: intPtr(1),
			AwayScore: intPtr(1),
			LiveTime:  stringPtr("34'"),
			MatchTime: &now,
			Round:     "Matchday 18",
		},

		// Match 3: Champions League - Man City vs Bayern (56')
		{
			ID: 2003,
			League: api.League{
				ID:   42,
				Name: "UEFA Champions League",
			},
			HomeTeam: api.Team{
				ID:        50,
				Name:      "Manchester City",
				ShortName: "Man City",
			},
			AwayTeam: api.Team{
				ID:        157,
				Name:      "Bayern Munich",
				ShortName: "Bayern",
			},
			Status:    api.MatchStatusLive,
			HomeScore: intPtr(3),
			AwayScore: intPtr(2),
			LiveTime:  stringPtr("56'"),
			MatchTime: &now,
			Round:     "Round of 16",
		},

		// ═══════════════════════════════════════════════
		// JUST FINISHED MATCHES (2) - for "all events" view
		// ═══════════════════════════════════════════════

		// Match 4: Premier League - Arsenal vs Liverpool (FT)
		{
			ID: 2004,
			League: api.League{
				ID:   47,
				Name: "Premier League",
			},
			HomeTeam: api.Team{
				ID:        42,
				Name:      "Arsenal",
				ShortName: "Arsenal",
			},
			AwayTeam: api.Team{
				ID:        40,
				Name:      "Liverpool",
				ShortName: "Liverpool",
			},
			Status:    api.MatchStatusFinished,
			HomeScore: intPtr(2),
			AwayScore: intPtr(3),
			LiveTime:  stringPtr("FT"),
			MatchTime: &now,
			Round:     "Matchday 17",
		},

		// Match 5: La Liga - Barcelona vs Sevilla (FT)
		{
			ID: 2005,
			League: api.League{
				ID:   87,
				Name: "La Liga",
			},
			HomeTeam: api.Team{
				ID:        529,
				Name:      "Barcelona",
				ShortName: "Barcelona",
			},
			AwayTeam: api.Team{
				ID:        536,
				Name:      "Sevilla",
				ShortName: "Sevilla",
			},
			Status:    api.MatchStatusFinished,
			HomeScore: intPtr(4),
			AwayScore: intPtr(1),
			LiveTime:  stringPtr("FT"),
			MatchTime: &now,
			Round:     "Matchday 18",
		},
	}
}
