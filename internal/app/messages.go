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
}

// liveMatchesMsg contains live matches from API response.
type liveMatchesMsg struct {
	matches []api.Match
}

// liveRefreshMsg is sent when live matches are refreshed (periodic 5-min timer).
type liveRefreshMsg struct {
	matches []api.Match
}

// liveBatchDataMsg contains live matches for a batch of leagues (parallel loading).
// Sent when a batch of leagues completes, allowing progressive UI updates.
type liveBatchDataMsg struct {
	batchIndex int         // Which batch (0, 1, 2, ...)
	isLast     bool        // true if this is the last batch
	matches    []api.Match // live matches from all leagues in this batch
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
}

// pollTickMsg is sent when the 90-second poll interval elapses.
// This triggers the actual API call with loading state visible.
type pollTickMsg struct {
	matchID int
}

// pollDisplayCompleteMsg is sent after minimum display time (1 second) has elapsed.
// This allows the "Updating..." spinner to be visible for at least 1 second.
type pollDisplayCompleteMsg struct{}

// goalLinksMsg contains goal replay links fetched from Reddit.
// Sent after searching r/soccer for Media posts matching goal events.
type goalLinksMsg struct {
	matchID int
	links   map[reddit.GoalLinkKey]*reddit.GoalLink
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
