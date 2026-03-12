package data

import (
	"time"

	"github.com/0xjuanma/golazo/internal/api"
)

// MockFinishedMatches returns finished matches for the stats view.
// 9 matches from preferred leagues: Premier League, La Liga, Champions League
func MockFinishedMatches() []api.Match {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	return []api.Match{
		// ═══════════════════════════════════════════════
		// TODAY'S MATCHES (3 matches)
		// ═══════════════════════════════════════════════

		// Today 1: Newcastle 2-1 Aston Villa (Premier League)
		{
			ID: 1010,
			League: api.League{
				ID:   47,
				Name: "Premier League",
			},
			HomeTeam: api.Team{
				ID:        39,
				Name:      "Newcastle United",
				ShortName: "Newcastle",
			},
			AwayTeam: api.Team{
				ID:        66,
				Name:      "Aston Villa",
				ShortName: "Villa",
			},
			Status:    api.MatchStatusFinished,
			HomeScore: intPtr(2),
			AwayScore: intPtr(1),
			MatchTime: timePtr(today.Add(14 * time.Hour)), // 14:00 today
			Round:     "Matchday 18",
		},

		// Today 2: Valencia 0-2 Athletic Bilbao (La Liga)
		{
			ID: 1011,
			League: api.League{
				ID:   87,
				Name: "La Liga",
			},
			HomeTeam: api.Team{
				ID:        532,
				Name:      "Valencia",
				ShortName: "Valencia",
			},
			AwayTeam: api.Team{
				ID:        531,
				Name:      "Athletic Bilbao",
				ShortName: "Athletic",
			},
			Status:    api.MatchStatusFinished,
			HomeScore: intPtr(0),
			AwayScore: intPtr(2),
			MatchTime: timePtr(today.Add(16 * time.Hour)), // 16:00 today
			Round:     "Matchday 18",
		},

		// Today 3: Napoli 3-1 Roma (Champions League)
		{
			ID: 1012,
			League: api.League{
				ID:   42,
				Name: "UEFA Champions League",
			},
			HomeTeam: api.Team{
				ID:        492,
				Name:      "Napoli",
				ShortName: "Napoli",
			},
			AwayTeam: api.Team{
				ID:        497,
				Name:      "AS Roma",
				ShortName: "Roma",
			},
			Status:    api.MatchStatusFinished,
			HomeScore: intPtr(3),
			AwayScore: intPtr(1),
			MatchTime: timePtr(today.Add(18 * time.Hour)), // 18:00 today
			Round:     "Round of 16 - 1st Leg",
		},

		// ═══════════════════════════════════════════════
		// PREMIER LEAGUE (2 matches)
		// ═══════════════════════════════════════════════

		// Match 1: Man City 2-1 Arsenal (2 days ago)
		{
			ID: 1001,
			League: api.League{
				ID:   47,
				Name: "Premier League",
			},
			HomeTeam: api.Team{
				ID:        50,
				Name:      "Manchester City",
				ShortName: "Man City",
			},
			AwayTeam: api.Team{
				ID:        42,
				Name:      "Arsenal",
				ShortName: "Arsenal",
			},
			Status:    api.MatchStatusFinished,
			HomeScore: intPtr(2),
			AwayScore: intPtr(1),
			MatchTime: timePtr(now.AddDate(0, 0, -2)),
			Round:     "Matchday 16",
		},

		// Match 2: Man Utd 0-3 Liverpool (3 days ago)
		{
			ID: 1002,
			League: api.League{
				ID:   47,
				Name: "Premier League",
			},
			HomeTeam: api.Team{
				ID:        33,
				Name:      "Manchester United",
				ShortName: "Man Utd",
			},
			AwayTeam: api.Team{
				ID:        40,
				Name:      "Liverpool",
				ShortName: "Liverpool",
			},
			Status:    api.MatchStatusFinished,
			HomeScore: intPtr(0),
			AwayScore: intPtr(3),
			MatchTime: timePtr(now.AddDate(0, 0, -3)),
			Round:     "Matchday 15",
		},

		// ═══════════════════════════════════════════════
		// LA LIGA (2 matches)
		// ═══════════════════════════════════════════════

		// Match 3: Real Madrid 3-2 Barcelona - El Clasico (1 day ago)
		{
			ID: 1003,
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
				ID:        529,
				Name:      "Barcelona",
				ShortName: "Barcelona",
			},
			Status:    api.MatchStatusFinished,
			HomeScore: intPtr(3),
			AwayScore: intPtr(2),
			MatchTime: timePtr(now.AddDate(0, 0, -1)),
			Round:     "Matchday 17",
		},

		// Match 4: Atletico 1-1 Sevilla (4 days ago)
		{
			ID: 1004,
			League: api.League{
				ID:   87,
				Name: "La Liga",
			},
			HomeTeam: api.Team{
				ID:        531,
				Name:      "Atletico Madrid",
				ShortName: "Atletico",
			},
			AwayTeam: api.Team{
				ID:        536,
				Name:      "Sevilla",
				ShortName: "Sevilla",
			},
			Status:    api.MatchStatusFinished,
			HomeScore: intPtr(1),
			AwayScore: intPtr(1),
			MatchTime: timePtr(now.AddDate(0, 0, -4)),
			Round:     "Matchday 16",
		},

		// ═══════════════════════════════════════════════
		// UEFA CHAMPIONS LEAGUE (2 matches)
		// ═══════════════════════════════════════════════

		// Match 5: PSG 2-3 Bayern (5 days ago)
		{
			ID: 1005,
			League: api.League{
				ID:   42,
				Name: "UEFA Champions League",
			},
			HomeTeam: api.Team{
				ID:        85,
				Name:      "Paris Saint-Germain",
				ShortName: "PSG",
			},
			AwayTeam: api.Team{
				ID:        157,
				Name:      "Bayern Munich",
				ShortName: "Bayern",
			},
			Status:    api.MatchStatusFinished,
			HomeScore: intPtr(2),
			AwayScore: intPtr(3),
			MatchTime: timePtr(now.AddDate(0, 0, -5)),
			Round:     "Round of 16 - 1st Leg",
		},

		// Match 6: Inter 1-0 Dortmund (6 days ago)
		{
			ID: 1006,
			League: api.League{
				ID:   42,
				Name: "UEFA Champions League",
			},
			HomeTeam: api.Team{
				ID:        108,
				Name:      "Inter Milan",
				ShortName: "Inter",
			},
			AwayTeam: api.Team{
				ID:        165,
				Name:      "Borussia Dortmund",
				ShortName: "Dortmund",
			},
			Status:    api.MatchStatusFinished,
			HomeScore: intPtr(1),
			AwayScore: intPtr(0),
			MatchTime: timePtr(now.AddDate(0, 0, -6)),
			Round:     "Round of 16 - 1st Leg",
		},
	}
}
