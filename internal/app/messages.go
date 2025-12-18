package app

import (
	"github.com/0xjuanma/golazo/internal/api"
	"github.com/0xjuanma/golazo/internal/fotmob"
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
