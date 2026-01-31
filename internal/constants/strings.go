package constants

// Menu items
const (
	MenuStats       = "Finished Matches"
	MenuLiveMatches = "Live Matches"
	MenuSettings    = "Settings"
)

// Panel titles
const (
	PanelLiveMatches       = "Live Matches"
	PanelFinishedMatches   = "Finished Matches"
	PanelMatchDetails      = "Match Details"
	PanelMatchList         = "Match List"
	PanelUpcomingMatches   = "Upcoming Matches"
	PanelMinuteByMinute    = "Minute-by-minute"
	PanelMatchStatistics   = "Match Statistics"
	PanelUpdates           = "Updates"
	PanelLeaguePreferences = "League Preferences"
)

// Empty state messages
const (
	EmptyNoLiveMatches     = "No live matches"
	EmptyNoFinishedMatches = "No finished matches"
	EmptySelectMatch       = "Select a match"
	EmptyNoUpdates         = "No updates"
	EmptyNoMatches         = "No matches available"
)

// Help text
const (
	HelpMainMenu           = "↑/↓: navigate  Enter: select  q: quit"
	HelpMatchesView        = "↑/↓: navigate  r: refresh details  /: filter  Esc: back  q: quit"
	HelpSettingsView       = "↑/↓: navigate  ←/→: switch tabs  Space: toggle  /: filter  Enter: save  Esc: back"
	HelpStatsView          = "h/l: date range  j/k: navigate  Tab: focus details  ↑/↓: scroll when focused  r: refresh details  /: filter  Esc: back"
	HelpStatsViewUnfocused = "Tab: focus details"
	HelpStatsViewFocused   = "Tab: unfocus  s: standings  f: formations  x: all statistics  ↑/↓: scroll"
	HelpStandingsDialog    = "Esc: close"
	HelpFormationsDialog   = "Tab/←/→: switch team  Esc: close"
	HelpStatisticsDialog   = "↑/↓: navigate  Esc: close"
)

// Status text
const (
	StatusLive            = "LIVE"
	StatusFinished        = "FT"
	StatusNotStarted      = "VS"
	StatusNotStartedShort = "NS"
	StatusFinishedText    = "Finished"
)

// Loading text
const (
	LoadingFetching = "Fetching..."
)

// Notification text
const (
	// NotificationTitleGoal is the title shown in goal notifications.
	NotificationTitleGoal = "⚽ GOLAZO!"
)

// Stats labels
const (
	LabelStatus = "Status: "
	LabelScore  = "Score: "
	LabelLeague = "League: "
	LabelDate   = "Date: "
	LabelVenue  = "Venue: "
)
