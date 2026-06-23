package data

import "github.com/0xjuanma/golazo/internal/api"

// MockWorldCupData2026 returns illustrative data for the 2026 FIFA World Cup
// (USA/Canada/Mexico). Group standings reflect the tournament as of mid-group-
// stage; the knockout bracket shows all 32 qualified teams with no scores yet,
// simulating the bracket preview at knockout kick-off.
//
// Team IDs prefixed 900xx are internal placeholders — they match across groups
// and bracket but do not correspond to FotMob's real IDs for those teams.
// Use --mock for verified 2022 data; this flag is for layout preview only.
func MockWorldCupData2026() *api.WorldCupData {
	return &api.WorldCupData{
		Season:         "2026",
		Name:           "FIFA World Cup 2026",
		Groups:         mockWC2026Groups(),
		KnockoutRounds: mockWC2026Bracket(),
		BronzeFinal:    mockWC2026Bronze(),
		Champion:       nil, // tournament in progress
		RunnerUp:       nil,
	}
}

// ── team ID constants ─────────────────────────────────────────────────────────
// Known FotMob IDs reused from real data; 900xx are mock placeholders.

const (
	idUSA        = 6713
	idEngland    = 8491
	idFrance     = 6723
	idArgentina  = 6706
	idNetherlands = 6708
	idSenegal    = 6395
	idEcuador    = 6707
	idMexico     = 6710
	idAustralia  = 6716
	idIran       = 6711
	idWales      = 5790
	idPoland     = 8568
	idTunisia    = 6719

	idBrazil      = 90001
	idPortugal    = 90002
	idCroatia     = 90003
	idNewZealand  = 90004
	idSpain       = 90005
	idCanada      = 90006
	idMorocco     = 90007
	idAlgeria     = 90008
	idColombia    = 90009
	idBolivia     = 90010
	idUruguay     = 90011
	idChile       = 90012
	idPeru        = 90013
	idVenezuela   = 90014
	idGermany     = 90015
	idAustria     = 90016
	idSerbia      = 90017
	idDenmark     = 90018
	idBelgium     = 90019
	idIvoryCoast  = 90020
	idItaly       = 90021
	idSwitzerland = 90022
	idScotland    = 90023
	idJapan       = 90024
	idSouthKorea  = 90025
	idCameroon    = 90026
	idTurkey      = 90027
	idNigeria     = 90028
	idGhana       = 90029
	idPanama      = 90030
	idCostaRica   = 90031
	idParaguay    = 90032
	idHonduras    = 90033
	idJamaica     = 90034
	idSaudiArabia = 90035
)

// ── helpers ───────────────────────────────────────────────────────────────────

func entry(pos, id int, name, short string, played, won, drawn, lost, gf, ga int) api.LeagueTableEntry {
	return api.LeagueTableEntry{
		Position:       pos,
		Team:           api.Team{ID: id, Name: name, ShortName: short},
		Played:         played,
		Won:            won,
		Drawn:          drawn,
		Lost:           lost,
		GoalsFor:       gf,
		GoalsAgainst:   ga,
		GoalDifference: gf - ga,
		Points:         won*3 + drawn,
	}
}

// ── groups ────────────────────────────────────────────────────────────────────

func mockWC2026Groups() []api.WCGroup {
	return []api.WCGroup{
		{ID: 926201, Letter: "A", Name: "Group A", Teams: []api.LeagueTableEntry{
			entry(1, idUSA, "United States", "USA", 3, 2, 1, 0, 5, 2),
			entry(2, idBrazil, "Brazil", "BRA", 3, 2, 0, 1, 6, 3),
			entry(3, idHonduras, "Honduras", "HON", 3, 1, 0, 2, 2, 5),
			entry(4, idJamaica, "Jamaica", "JAM", 3, 0, 1, 2, 1, 4),
		}},
		{ID: 926202, Letter: "B", Name: "Group B", Teams: []api.LeagueTableEntry{
			entry(1, idPortugal, "Portugal", "POR", 3, 3, 0, 0, 8, 2),
			entry(2, idMexico, "Mexico", "MEX", 3, 1, 1, 1, 4, 5),
			entry(3, idCroatia, "Croatia", "CRO", 3, 1, 0, 2, 3, 5),
			entry(4, idNewZealand, "New Zealand", "NZL", 3, 0, 1, 2, 2, 5),
		}},
		{ID: 926203, Letter: "C", Name: "Group C", Teams: []api.LeagueTableEntry{
			entry(1, idSpain, "Spain", "ESP", 3, 2, 1, 0, 7, 2),
			entry(2, idCanada, "Canada", "CAN", 3, 2, 0, 1, 5, 4),
			entry(3, idMorocco, "Morocco", "MAR", 3, 1, 1, 1, 4, 4),
			entry(4, idAlgeria, "Algeria", "ALG", 3, 0, 0, 3, 1, 7),
		}},
		{ID: 926204, Letter: "D", Name: "Group D", Teams: []api.LeagueTableEntry{
			entry(1, idFrance, "France", "FRA", 3, 2, 1, 0, 6, 2),
			entry(2, idEngland, "England", "ENG", 3, 2, 0, 1, 5, 3),
			entry(3, idSenegal, "Senegal", "SEN", 3, 1, 1, 1, 3, 4),
			entry(4, idSaudiArabia, "Saudi Arabia", "KSA", 3, 0, 0, 3, 1, 6),
		}},
		{ID: 926205, Letter: "E", Name: "Group E", Teams: []api.LeagueTableEntry{
			entry(1, idArgentina, "Argentina", "ARG", 3, 2, 1, 0, 7, 2),
			entry(2, idColombia, "Colombia", "COL", 3, 2, 0, 1, 5, 3),
			entry(3, idEcuador, "Ecuador", "ECU", 3, 1, 0, 2, 3, 5),
			entry(4, idBolivia, "Bolivia", "BOL", 3, 0, 1, 2, 2, 7),
		}},
		{ID: 926206, Letter: "F", Name: "Group F", Teams: []api.LeagueTableEntry{
			entry(1, idUruguay, "Uruguay", "URU", 3, 2, 0, 1, 5, 3),
			entry(2, idChile, "Chile", "CHI", 3, 1, 2, 0, 5, 4),
			entry(3, idPeru, "Peru", "PER", 3, 1, 0, 2, 3, 5),
			entry(4, idVenezuela, "Venezuela", "VEN", 3, 0, 2, 1, 2, 3),
		}},
		{ID: 926207, Letter: "G", Name: "Group G", Teams: []api.LeagueTableEntry{
			entry(1, idGermany, "Germany", "GER", 3, 3, 0, 0, 9, 2),
			entry(2, idNetherlands, "Netherlands", "NED", 3, 2, 0, 1, 6, 4),
			entry(3, idAustria, "Austria", "AUT", 3, 1, 0, 2, 3, 6),
			entry(4, idSerbia, "Serbia", "SRB", 3, 0, 0, 3, 1, 7),
		}},
		{ID: 926208, Letter: "H", Name: "Group H", Teams: []api.LeagueTableEntry{
			entry(1, idBelgium, "Belgium", "BEL", 3, 2, 1, 0, 7, 3),
			entry(2, idDenmark, "Denmark", "DEN", 3, 2, 0, 1, 6, 4),
			entry(3, idPoland, "Poland", "POL", 3, 1, 1, 1, 4, 4),
			entry(4, idIvoryCoast, "Ivory Coast", "CIV", 3, 0, 0, 3, 2, 8),
		}},
		{ID: 926209, Letter: "I", Name: "Group I", Teams: []api.LeagueTableEntry{
			entry(1, idItaly, "Italy", "ITA", 3, 2, 0, 1, 5, 3),
			entry(2, idSwitzerland, "Switzerland", "SUI", 3, 1, 2, 0, 4, 3),
			entry(3, idScotland, "Scotland", "SCO", 3, 1, 0, 2, 3, 5),
			entry(4, idIran, "Iran", "IRN", 3, 0, 2, 1, 2, 3),
		}},
		{ID: 926210, Letter: "J", Name: "Group J", Teams: []api.LeagueTableEntry{
			entry(1, idJapan, "Japan", "JPN", 3, 2, 1, 0, 6, 2),
			entry(2, idSouthKorea, "South Korea", "KOR", 3, 1, 2, 0, 4, 3),
			entry(3, idAustralia, "Australia", "AUS", 3, 1, 0, 2, 3, 5),
			entry(4, idCameroon, "Cameroon", "CMR", 3, 0, 1, 2, 2, 5),
		}},
		{ID: 926211, Letter: "K", Name: "Group K", Teams: []api.LeagueTableEntry{
			entry(1, idTurkey, "Turkey", "TUR", 3, 2, 1, 0, 6, 3),
			entry(2, idNigeria, "Nigeria", "NGA", 3, 2, 0, 1, 5, 3),
			entry(3, idGhana, "Ghana", "GHA", 3, 1, 1, 1, 4, 4),
			entry(4, idPanama, "Panama", "PAN", 3, 0, 0, 3, 1, 6),
		}},
		{ID: 926212, Letter: "L", Name: "Group L", Teams: []api.LeagueTableEntry{
			entry(1, idWales, "Wales", "WAL", 3, 2, 0, 1, 5, 4),
			entry(2, idCostaRica, "Costa Rica", "CRC", 3, 1, 2, 0, 4, 3),
			entry(3, idParaguay, "Paraguay", "PAR", 3, 1, 0, 2, 3, 5),
			entry(4, idTunisia, "Tunisia", "TUN", 3, 0, 2, 1, 2, 3),
		}},
	}
}

// ── bracket ───────────────────────────────────────────────────────────────────
// Round of 32: all 32 teams known (top 2 from each group + 8 best 3rd-place),
// no matches played yet. Higher rounds are fully TBD.
//
// 3rd-place qualifiers used: ECU, CRO, MAR, SCO, POL, AUS, GHA, PER.

func mockWC2026Bracket() []api.WCKnockoutRound {
	return []api.WCKnockoutRound{
		{
			Stage: "1/16",
			Label: "Round of 32",
			Matchups: []api.WCMatchup{
				// A1 USA vs L2 Costa Rica
				{HomeTeam: "United States", HomeTeamID: idUSA, HomeShort: "USA",
					AwayTeam: "Costa Rica", AwayTeamID: idCostaRica, AwayShort: "CRC"},
				// B1 Portugal vs K3 Ghana
				{HomeTeam: "Portugal", HomeTeamID: idPortugal, HomeShort: "POR",
					AwayTeam: "Ghana", AwayTeamID: idGhana, AwayShort: "GHA"},
				// C1 Spain vs F3 Peru
				{HomeTeam: "Spain", HomeTeamID: idSpain, HomeShort: "ESP",
					AwayTeam: "Peru", AwayTeamID: idPeru, AwayShort: "PER"},
				// D1 France vs I3 Scotland
				{HomeTeam: "France", HomeTeamID: idFrance, HomeShort: "FRA",
					AwayTeam: "Scotland", AwayTeamID: idScotland, AwayShort: "SCO"},
				// E1 Argentina vs J3 Australia
				{HomeTeam: "Argentina", HomeTeamID: idArgentina, HomeShort: "ARG",
					AwayTeam: "Australia", AwayTeamID: idAustralia, AwayShort: "AUS"},
				// F1 Uruguay vs C3 Morocco
				{HomeTeam: "Uruguay", HomeTeamID: idUruguay, HomeShort: "URU",
					AwayTeam: "Morocco", AwayTeamID: idMorocco, AwayShort: "MAR"},
				// G1 Germany vs H3 Poland
				{HomeTeam: "Germany", HomeTeamID: idGermany, HomeShort: "GER",
					AwayTeam: "Poland", AwayTeamID: idPoland, AwayShort: "POL"},
				// H1 Belgium vs E3 Ecuador
				{HomeTeam: "Belgium", HomeTeamID: idBelgium, HomeShort: "BEL",
					AwayTeam: "Ecuador", AwayTeamID: idEcuador, AwayShort: "ECU"},
				// I1 Italy vs B3 Croatia
				{HomeTeam: "Italy", HomeTeamID: idItaly, HomeShort: "ITA",
					AwayTeam: "Croatia", AwayTeamID: idCroatia, AwayShort: "CRO"},
				// J1 Japan vs H2 Denmark
				{HomeTeam: "Japan", HomeTeamID: idJapan, HomeShort: "JPN",
					AwayTeam: "Denmark", AwayTeamID: idDenmark, AwayShort: "DEN"},
				// K1 Turkey vs F2 Chile
				{HomeTeam: "Turkey", HomeTeamID: idTurkey, HomeShort: "TUR",
					AwayTeam: "Chile", AwayTeamID: idChile, AwayShort: "CHI"},
				// L1 Wales vs E2 Colombia
				{HomeTeam: "Wales", HomeTeamID: idWales, HomeShort: "WAL",
					AwayTeam: "Colombia", AwayTeamID: idColombia, AwayShort: "COL"},
				// A2 Brazil vs K2 Nigeria
				{HomeTeam: "Brazil", HomeTeamID: idBrazil, HomeShort: "BRA",
					AwayTeam: "Nigeria", AwayTeamID: idNigeria, AwayShort: "NGA"},
				// B2 Mexico vs J2 South Korea
				{HomeTeam: "Mexico", HomeTeamID: idMexico, HomeShort: "MEX",
					AwayTeam: "South Korea", AwayTeamID: idSouthKorea, AwayShort: "KOR"},
				// C2 Canada vs G2 Netherlands
				{HomeTeam: "Canada", HomeTeamID: idCanada, HomeShort: "CAN",
					AwayTeam: "Netherlands", AwayTeamID: idNetherlands, AwayShort: "NED"},
				// D2 England vs I2 Switzerland
				{HomeTeam: "England", HomeTeamID: idEngland, HomeShort: "ENG",
					AwayTeam: "Switzerland", AwayTeamID: idSwitzerland, AwayShort: "SUI"},
			},
		},
		{
			Stage: "1/8",
			Label: "Round of 16",
			Matchups: []api.WCMatchup{
				{TBDHome: true, TBDAway: true},
				{TBDHome: true, TBDAway: true},
				{TBDHome: true, TBDAway: true},
				{TBDHome: true, TBDAway: true},
				{TBDHome: true, TBDAway: true},
				{TBDHome: true, TBDAway: true},
				{TBDHome: true, TBDAway: true},
				{TBDHome: true, TBDAway: true},
			},
		},
		{
			Stage: "1/4",
			Label: "Quarterfinals",
			Matchups: []api.WCMatchup{
				{TBDHome: true, TBDAway: true},
				{TBDHome: true, TBDAway: true},
				{TBDHome: true, TBDAway: true},
				{TBDHome: true, TBDAway: true},
			},
		},
		{
			Stage: "1/2",
			Label: "Semifinals",
			Matchups: []api.WCMatchup{
				{TBDHome: true, TBDAway: true},
				{TBDHome: true, TBDAway: true},
			},
		},
		{
			Stage: "final",
			Label: "Final",
			Matchups: []api.WCMatchup{
				{TBDHome: true, TBDAway: true},
			},
		},
	}
}

func mockWC2026Bronze() *api.WCMatchup {
	mu := api.WCMatchup{TBDHome: true, TBDAway: true}
	return &mu
}

// MockWorldCupUpcoming2026 is intentionally empty — upcoming fixtures are
// fetched live from FotMob even in --wc-season 2026 mode.
func MockWorldCupUpcoming2026() []api.Match {
	return nil
}

