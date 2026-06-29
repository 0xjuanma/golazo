package app

import (
	"github.com/0xjuanma/golazo/internal/api"
	"github.com/0xjuanma/golazo/internal/fotmob"
	"github.com/0xjuanma/golazo/internal/reddit"
)

// liveUpdateMsg contains a live update string for match events.
type liveUpdateMsg struct {
	update string
}

// matchDetailsMsg contains match details from API response.
type matchDetailsMsg struct {
	details *api.MatchDetails
	err     error
}

// liveMatchesMsg contains live matches from API response.
type liveMatchesMsg struct {
	matches []api.Match
}

// liveRefreshMsg is sent when live matches are refreshed (periodic 5-min timer).
type liveRefreshMsg struct {
	matches  []api.Match
	upcoming []api.Match
}

// liveBatchDataMsg contains live matches for a batch of leagues (parallel loading).
// Sent when a batch of leagues completes, allowing progressive UI updates.
type liveBatchDataMsg struct {
	batchIndex int         // Which batch (0, 1, 2, ...)
	isLast     bool        // true if this is the last batch
	matches    []api.Match // live matches from all leagues in this batch
	upcoming   []api.Match // upcoming (not started) matches from this batch
	err        error
}

// statsDataMsg contains all stats data (5 days finished + today upcoming) from API response.
// This is the unified message for stats view - always fetches 5 days, filters client-side.
type statsDataMsg struct {
	data *fotmob.StatsData
}

// statsDayDataMsg contains stats data for a single day (progressive loading).
// Sent as each day's API calls complete, allowing immediate UI updates.
type statsDayDataMsg struct {
	dayIndex int         // 0 = today, 1 = yesterday, etc.
	isToday  bool        // true if this is today's data
	isLast   bool        // true if this is the last day to fetch
	finished []api.Match // finished matches for this day
	upcoming []api.Match // upcoming matches (only for today)
	err      error
}

// pollTickMsg is sent when the 90-second poll interval elapses.
// This triggers the actual API call with loading state visible.
type pollTickMsg struct {
	matchID int
	gen     int // generation at scheduling time; dropped if model has moved on
}

// pollDisplayCompleteMsg is sent after minimum display time (1 second) has elapsed.
// This allows the "Updating..." spinner to be visible for at least 1 second.
type pollDisplayCompleteMsg struct{}

// goalLinkStreamMsg hands a freshly-opened goal-link subscription channel to
// the Update loop. The handler stashes the channel on the model (keyed by
// matchID) and arms the first reader Cmd. Emitted once per match-details
// load, before any goalLinkMsg for that match.
type goalLinkStreamMsg struct {
	matchID int
	ch      <-chan reddit.GoalResult
}

// goalLinkMsg streams a single goal-link outcome from the reddit queue's
// subscription channel. The match-level goalLinksMsg above remains for
// initial cache-hit batches; goalLinkMsg carries per-goal results that
// arrive at the queue's 30s cadence so the UI can render replay links
// progressively instead of waiting for the entire match's goals to resolve.
// Link is nil when the goal was searched but not found, was dropped because
// Reddit returned ErrBlocked, or hit a transient fetch error.
type goalLinkMsg struct {
	matchID int
	key     reddit.GoalLinkKey
	link    *reddit.GoalLink
}

// goalLinksDoneMsg signals that the reddit queue's subscription channel for
// a given match has closed — every queued goal has produced exactly one
// goalLinkMsg. Used to stop the subscription tea.Cmd loop without leaking
// goroutines.
type goalLinksDoneMsg struct {
	matchID int
}

// standingsMsg contains league standings from API response.
// Used to populate the standings dialog.
type standingsMsg struct {
	leagueID   int
	leagueName string
	standings  []api.LeagueTableEntry
	homeTeamID int
	awayTeamID int
}

// wcDataMsg contains World Cup data fetched from FotMob or mock.
type wcDataMsg struct {
	data *api.WorldCupData
	err  error
}

// wcUpcomingMsg contains upcoming World Cup matches for the next few days,
// fetched from FotMob or mock.
type wcUpcomingMsg struct {
	matches []api.Match
	err     error
}

// wcTopScorersMsg contains World Cup top scorer data fetched from FotMob.
type wcTopScorersMsg struct {
	scorers []api.WCTopScorer
	err     error
}
