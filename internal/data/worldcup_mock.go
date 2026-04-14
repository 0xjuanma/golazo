package data

import "github.com/0xjuanma/golazo/internal/api"

// MockWorldCupData returns real completed data from the 2022 FIFA World Cup (Qatar).
// Sourced from FotMob. Used as sample data for development and the --mock flag.
func MockWorldCupData() *api.WorldCupData {
	arg := api.Team{ID: 6706, Name: "Argentina", ShortName: "ARG"}
	fra := api.Team{ID: 6723, Name: "France", ShortName: "FRA"}

	return &api.WorldCupData{
		Season:         "2022",
		Name:           "FIFA World Cup 2022",
		Groups:         mockWC2022Groups(),
		KnockoutRounds: mockWC2022Bracket(),
		BronzeFinal:    mockWC2022Bronze(),
		Champion:       &arg,
		RunnerUp:       &fra,
	}
}

func mockWC2022Groups() []api.WCGroup {
	return []api.WCGroup{
		{
			ID: 868712, Letter: "A", Name: "Group A",
			Teams: []api.LeagueTableEntry{
				{Position: 1, Team: api.Team{ID: 6708, Name: "Netherlands", ShortName: "NED"}, Played: 3, Won: 2, Drawn: 1, Lost: 0, GoalsFor: 5, GoalsAgainst: 1, GoalDifference: 4, Points: 7},
				{Position: 2, Team: api.Team{ID: 6395, Name: "Senegal", ShortName: "SEN"}, Played: 3, Won: 2, Drawn: 0, Lost: 1, GoalsFor: 5, GoalsAgainst: 4, GoalDifference: 1, Points: 6},
				{Position: 3, Team: api.Team{ID: 6707, Name: "Ecuador", ShortName: "ECU"}, Played: 3, Won: 1, Drawn: 1, Lost: 1, GoalsFor: 4, GoalsAgainst: 3, GoalDifference: 1, Points: 4},
				{Position: 4, Team: api.Team{ID: 5902, Name: "Qatar", ShortName: "QAT"}, Played: 3, Won: 0, Drawn: 0, Lost: 3, GoalsFor: 1, GoalsAgainst: 7, GoalDifference: -6, Points: 0},
			},
		},
		{
			ID: 868713, Letter: "B", Name: "Group B",
			Teams: []api.LeagueTableEntry{
				{Position: 1, Team: api.Team{ID: 8491, Name: "England", ShortName: "ENG"}, Played: 3, Won: 2, Drawn: 1, Lost: 0, GoalsFor: 9, GoalsAgainst: 2, GoalDifference: 7, Points: 7},
				{Position: 2, Team: api.Team{ID: 6713, Name: "USA", ShortName: "USA"}, Played: 3, Won: 1, Drawn: 2, Lost: 0, GoalsFor: 2, GoalsAgainst: 1, GoalDifference: 1, Points: 5},
				{Position: 3, Team: api.Team{ID: 6711, Name: "Iran", ShortName: "IRN"}, Played: 3, Won: 1, Drawn: 0, Lost: 2, GoalsFor: 4, GoalsAgainst: 7, GoalDifference: -3, Points: 3},
				{Position: 4, Team: api.Team{ID: 5790, Name: "Wales", ShortName: "WAL"}, Played: 3, Won: 0, Drawn: 1, Lost: 2, GoalsFor: 1, GoalsAgainst: 6, GoalDifference: -5, Points: 1},
			},
		},
		{
			ID: 868714, Letter: "C", Name: "Group C",
			Teams: []api.LeagueTableEntry{
				{Position: 1, Team: api.Team{ID: 6706, Name: "Argentina", ShortName: "ARG"}, Played: 3, Won: 2, Drawn: 0, Lost: 1, GoalsFor: 5, GoalsAgainst: 2, GoalDifference: 3, Points: 6},
				{Position: 2, Team: api.Team{ID: 8568, Name: "Poland", ShortName: "POL"}, Played: 3, Won: 1, Drawn: 1, Lost: 1, GoalsFor: 2, GoalsAgainst: 2, GoalDifference: 0, Points: 4},
				{Position: 3, Team: api.Team{ID: 6710, Name: "Mexico", ShortName: "MEX"}, Played: 3, Won: 1, Drawn: 1, Lost: 1, GoalsFor: 2, GoalsAgainst: 3, GoalDifference: -1, Points: 4},
				{Position: 4, Team: api.Team{ID: 7795, Name: "Saudi Arabia", ShortName: "KSA"}, Played: 3, Won: 1, Drawn: 0, Lost: 2, GoalsFor: 3, GoalsAgainst: 5, GoalDifference: -2, Points: 3},
			},
		},
		{
			ID: 868715, Letter: "D", Name: "Group D",
			Teams: []api.LeagueTableEntry{
				{Position: 1, Team: api.Team{ID: 6723, Name: "France", ShortName: "FRA"}, Played: 3, Won: 2, Drawn: 0, Lost: 1, GoalsFor: 6, GoalsAgainst: 3, GoalDifference: 3, Points: 6},
				{Position: 2, Team: api.Team{ID: 6716, Name: "Australia", ShortName: "AUS"}, Played: 3, Won: 2, Drawn: 0, Lost: 1, GoalsFor: 3, GoalsAgainst: 4, GoalDifference: -1, Points: 6},
				{Position: 3, Team: api.Team{ID: 6719, Name: "Tunisia", ShortName: "TUN"}, Played: 3, Won: 1, Drawn: 1, Lost: 1, GoalsFor: 1, GoalsAgainst: 1, GoalDifference: 0, Points: 4},
				{Position: 4, Team: api.Team{ID: 8238, Name: "Denmark", ShortName: "DEN"}, Played: 3, Won: 0, Drawn: 1, Lost: 2, GoalsFor: 1, GoalsAgainst: 3, GoalDifference: -2, Points: 1},
			},
		},
		{
			ID: 868716, Letter: "E", Name: "Group E",
			Teams: []api.LeagueTableEntry{
				{Position: 1, Team: api.Team{ID: 6715, Name: "Japan", ShortName: "JPN"}, Played: 3, Won: 2, Drawn: 0, Lost: 1, GoalsFor: 4, GoalsAgainst: 3, GoalDifference: 1, Points: 6},
				{Position: 2, Team: api.Team{ID: 6720, Name: "Spain", ShortName: "ESP"}, Played: 3, Won: 1, Drawn: 1, Lost: 1, GoalsFor: 9, GoalsAgainst: 3, GoalDifference: 6, Points: 4},
				{Position: 3, Team: api.Team{ID: 8570, Name: "Germany", ShortName: "GER"}, Played: 3, Won: 1, Drawn: 1, Lost: 1, GoalsFor: 6, GoalsAgainst: 5, GoalDifference: 1, Points: 4},
				{Position: 4, Team: api.Team{ID: 6705, Name: "Costa Rica", ShortName: "CRC"}, Played: 3, Won: 1, Drawn: 0, Lost: 2, GoalsFor: 3, GoalsAgainst: 11, GoalDifference: -8, Points: 3},
			},
		},
		{
			ID: 868719, Letter: "F", Name: "Group F",
			Teams: []api.LeagueTableEntry{
				{Position: 1, Team: api.Team{ID: 6262, Name: "Morocco", ShortName: "MAR"}, Played: 3, Won: 2, Drawn: 1, Lost: 0, GoalsFor: 4, GoalsAgainst: 1, GoalDifference: 3, Points: 7},
				{Position: 2, Team: api.Team{ID: 10155, Name: "Croatia", ShortName: "CRO"}, Played: 3, Won: 1, Drawn: 2, Lost: 0, GoalsFor: 4, GoalsAgainst: 1, GoalDifference: 3, Points: 5},
				{Position: 3, Team: api.Team{ID: 8263, Name: "Belgium", ShortName: "BEL"}, Played: 3, Won: 1, Drawn: 1, Lost: 1, GoalsFor: 1, GoalsAgainst: 2, GoalDifference: -1, Points: 4},
				{Position: 4, Team: api.Team{ID: 5810, Name: "Canada", ShortName: "CAN"}, Played: 3, Won: 0, Drawn: 0, Lost: 3, GoalsFor: 2, GoalsAgainst: 7, GoalDifference: -5, Points: 0},
			},
		},
		{
			ID: 868718, Letter: "G", Name: "Group G",
			Teams: []api.LeagueTableEntry{
				{Position: 1, Team: api.Team{ID: 8256, Name: "Brazil", ShortName: "BRA"}, Played: 3, Won: 2, Drawn: 0, Lost: 1, GoalsFor: 3, GoalsAgainst: 1, GoalDifference: 2, Points: 6},
				{Position: 2, Team: api.Team{ID: 6717, Name: "Switzerland", ShortName: "SUI"}, Played: 3, Won: 2, Drawn: 0, Lost: 1, GoalsFor: 4, GoalsAgainst: 3, GoalDifference: 1, Points: 6},
				{Position: 3, Team: api.Team{ID: 6629, Name: "Cameroon", ShortName: "CMR"}, Played: 3, Won: 1, Drawn: 1, Lost: 1, GoalsFor: 4, GoalsAgainst: 4, GoalDifference: 0, Points: 4},
				{Position: 4, Team: api.Team{ID: 8205, Name: "Serbia", ShortName: "SRB"}, Played: 3, Won: 0, Drawn: 1, Lost: 2, GoalsFor: 5, GoalsAgainst: 8, GoalDifference: -3, Points: 1},
			},
		},
		{
			ID: 868717, Letter: "H", Name: "Group H",
			Teams: []api.LeagueTableEntry{
				{Position: 1, Team: api.Team{ID: 8361, Name: "Portugal", ShortName: "POR"}, Played: 3, Won: 2, Drawn: 0, Lost: 1, GoalsFor: 6, GoalsAgainst: 4, GoalDifference: 2, Points: 6},
				{Position: 2, Team: api.Team{ID: 7804, Name: "South Korea", ShortName: "KOR"}, Played: 3, Won: 1, Drawn: 1, Lost: 1, GoalsFor: 4, GoalsAgainst: 4, GoalDifference: 0, Points: 4},
				{Position: 3, Team: api.Team{ID: 5796, Name: "Uruguay", ShortName: "URU"}, Played: 3, Won: 1, Drawn: 1, Lost: 1, GoalsFor: 2, GoalsAgainst: 2, GoalDifference: 0, Points: 4},
				{Position: 4, Team: api.Team{ID: 6714, Name: "Ghana", ShortName: "GHA"}, Played: 3, Won: 1, Drawn: 0, Lost: 2, GoalsFor: 5, GoalsAgainst: 7, GoalDifference: -2, Points: 3},
			},
		},
	}
}

func mockWC2022Bracket() []api.WCKnockoutRound {
	return []api.WCKnockoutRound{
		{
			Stage: "1/8",
			Label: "Round of 16",
			Matchups: []api.WCMatchup{
				{HomeTeam: "Netherlands", HomeTeamID: 6708, HomeShort: "NED", AwayTeam: "USA", AwayTeamID: 6713, AwayShort: "USA", HomeScore: intPtr(3), AwayScore: intPtr(1), WinnerID: intPtr(6708)},
				{HomeTeam: "Argentina", HomeTeamID: 6706, HomeShort: "ARG", AwayTeam: "Australia", AwayTeamID: 6716, AwayShort: "AUS", HomeScore: intPtr(2), AwayScore: intPtr(1), WinnerID: intPtr(6706)},
				{HomeTeam: "Japan", HomeTeamID: 6715, HomeShort: "JPN", AwayTeam: "Croatia", AwayTeamID: 10155, AwayShort: "CRO", HomeScore: intPtr(1), AwayScore: intPtr(1), WinnerID: intPtr(10155), IsPenalties: true},
				{HomeTeam: "Brazil", HomeTeamID: 8256, HomeShort: "BRA", AwayTeam: "South Korea", AwayTeamID: 7804, AwayShort: "KOR", HomeScore: intPtr(4), AwayScore: intPtr(1), WinnerID: intPtr(8256)},
				{HomeTeam: "England", HomeTeamID: 8491, HomeShort: "ENG", AwayTeam: "Senegal", AwayTeamID: 6395, AwayShort: "SEN", HomeScore: intPtr(3), AwayScore: intPtr(0), WinnerID: intPtr(8491)},
				{HomeTeam: "France", HomeTeamID: 6723, HomeShort: "FRA", AwayTeam: "Poland", AwayTeamID: 8568, AwayShort: "POL", HomeScore: intPtr(3), AwayScore: intPtr(1), WinnerID: intPtr(6723)},
				{HomeTeam: "Morocco", HomeTeamID: 6262, HomeShort: "MAR", AwayTeam: "Spain", AwayTeamID: 6720, AwayShort: "ESP", HomeScore: intPtr(0), AwayScore: intPtr(0), WinnerID: intPtr(6262), IsPenalties: true},
				{HomeTeam: "Portugal", HomeTeamID: 8361, HomeShort: "POR", AwayTeam: "Switzerland", AwayTeamID: 6717, AwayShort: "SUI", HomeScore: intPtr(6), AwayScore: intPtr(1), WinnerID: intPtr(8361)},
			},
		},
		{
			Stage: "1/4",
			Label: "Quarterfinals",
			Matchups: []api.WCMatchup{
				{HomeTeam: "Netherlands", HomeTeamID: 6708, HomeShort: "NED", AwayTeam: "Argentina", AwayTeamID: 6706, AwayShort: "ARG", HomeScore: intPtr(2), AwayScore: intPtr(2), WinnerID: intPtr(6706), IsPenalties: true},
				{HomeTeam: "Croatia", HomeTeamID: 10155, HomeShort: "CRO", AwayTeam: "Brazil", AwayTeamID: 8256, AwayShort: "BRA", HomeScore: intPtr(1), AwayScore: intPtr(1), WinnerID: intPtr(10155), IsPenalties: true},
				{HomeTeam: "England", HomeTeamID: 8491, HomeShort: "ENG", AwayTeam: "France", AwayTeamID: 6723, AwayShort: "FRA", HomeScore: intPtr(1), AwayScore: intPtr(2), WinnerID: intPtr(6723)},
				{HomeTeam: "Morocco", HomeTeamID: 6262, HomeShort: "MAR", AwayTeam: "Portugal", AwayTeamID: 8361, AwayShort: "POR", HomeScore: intPtr(1), AwayScore: intPtr(0), WinnerID: intPtr(6262)},
			},
		},
		{
			Stage: "1/2",
			Label: "Semifinals",
			Matchups: []api.WCMatchup{
				{HomeTeam: "Argentina", HomeTeamID: 6706, HomeShort: "ARG", AwayTeam: "Croatia", AwayTeamID: 10155, AwayShort: "CRO", HomeScore: intPtr(3), AwayScore: intPtr(0), WinnerID: intPtr(6706)},
				{HomeTeam: "France", HomeTeamID: 6723, HomeShort: "FRA", AwayTeam: "Morocco", AwayTeamID: 6262, AwayShort: "MAR", HomeScore: intPtr(2), AwayScore: intPtr(0), WinnerID: intPtr(6723)},
			},
		},
		{
			Stage: "final",
			Label: "Final",
			Matchups: []api.WCMatchup{
				{HomeTeam: "Argentina", HomeTeamID: 6706, HomeShort: "ARG", AwayTeam: "France", AwayTeamID: 6723, AwayShort: "FRA", HomeScore: intPtr(3), AwayScore: intPtr(3), WinnerID: intPtr(6706), IsPenalties: true},
			},
		},
	}
}

func mockWC2022Bronze() *api.WCMatchup {
	return &api.WCMatchup{
		HomeTeam: "Croatia", HomeTeamID: 10155, HomeShort: "CRO",
		AwayTeam: "Morocco", AwayTeamID: 6262, AwayShort: "MAR",
		HomeScore: intPtr(2), AwayScore: intPtr(1), WinnerID: intPtr(10155),
	}
}
