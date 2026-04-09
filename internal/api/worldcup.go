package api

// WCFotMobLeagueID is the FotMob league ID for the FIFA World Cup.
const WCFotMobLeagueID = 77

// WCGroup represents a single World Cup group with its standings.
type WCGroup struct {
	ID     int
	Letter string // "A", "B", "C", etc.
	Name   string // "Group A", "Group B", etc.
	Teams  []LeagueTableEntry
}

// WCMatchup represents a single knockout stage matchup.
type WCMatchup struct {
	HomeTeam    string
	HomeTeamID  int
	HomeShort   string
	AwayTeam    string
	AwayTeamID  int
	AwayShort   string
	HomeScore   *int
	AwayScore   *int
	WinnerID    *int
	IsPenalties bool
	TBDHome     bool
	TBDAway     bool
}

// WCKnockoutRound represents a round in the knockout stage.
type WCKnockoutRound struct {
	Stage    string // FotMob stage key: "1/16", "1/8", "1/4", "1/2", "final"
	Label    string // Human-readable: "Round of 32", "Round of 16", etc.
	Matchups []WCMatchup
}

// WorldCupData contains all World Cup tournament data.
type WorldCupData struct {
	Season         string // "2022", "2026"
	Name           string // "FIFA World Cup 2022"
	Groups         []WCGroup
	KnockoutRounds []WCKnockoutRound // ordered R32/R16 → QF → SF → Final (bronze excluded)
	BronzeFinal    *WCMatchup
	Champion       *Team
	RunnerUp       *Team
}
