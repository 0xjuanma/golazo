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

// DeriveFinalists extracts champion and runner-up from the final matchup's WinnerID.
// Returns nil, nil if the final has not been played yet or no final round exists.
func (d *WorldCupData) DeriveFinalists() (*Team, *Team) {
	for _, r := range d.KnockoutRounds {
		if r.Stage != "final" || len(r.Matchups) == 0 {
			continue
		}
		mu := r.Matchups[0]
		if mu.WinnerID == nil {
			return nil, nil
		}
		home := Team{ID: mu.HomeTeamID, Name: mu.HomeTeam, ShortName: mu.HomeShort}
		away := Team{ID: mu.AwayTeamID, Name: mu.AwayTeam, ShortName: mu.AwayShort}
		if *mu.WinnerID == mu.HomeTeamID {
			return &home, &away
		}
		return &away, &home
	}
	return nil, nil
}

// BracketLineCount returns the total number of content lines the bracket view
// produces for this data. Must stay in sync with ui.RenderWorldCupBracket's
// line construction.
func (d *WorldCupData) BracketLineCount() int {
	count := 0
	for _, round := range d.KnockoutRounds {
		count += 2 + len(round.Matchups) + 1 // roundHdr + blank + matchups + trailing blank
	}
	if d.BronzeFinal != nil {
		count += 4 // bronzeHdr + blank + matchup + trailing blank
	}
	if d.Champion != nil {
		count += 2 // blank + champion line
	}
	return count
}

