package fotmob

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/0xjuanma/golazo/internal/api"
)

// classifyLeagueMatches splits a league's allMatches into currently-live and
// upcoming sets. "Live" is determined by status only (Started && !Finished &&
// !Cancelled) and is intentionally date-agnostic — a match that kicked off
// before the user's UTC midnight is still live during its second half.
// "Upcoming" is gated to matches scheduled on the same calendar day as `now`
// in `now`'s timezone, so the upcoming list reflects what the user calls
// "today" rather than what UTC calls today.
//
// The classifier is pure and deterministic given a fixed `now` — pass
// time.Now() in production and a fixed clock in tests.
func classifyLeagueMatches(allMatches []fotmobMatch, leagueInfo league, now time.Time) (live, upcoming []api.Match) {
	loc := now.Location()
	todayStr := now.Format("2006-01-02")
	for _, m := range allMatches {
		if m.Status.UTCTime == "" {
			continue
		}
		if m.League.ID == 0 {
			m.League = leagueInfo
		}
		apiM := m.toAPIMatch()
		switch apiM.Status {
		case api.MatchStatusLive:
			live = append(live, apiM)
		case api.MatchStatusNotStarted:
			if apiM.MatchTime == nil {
				continue
			}
			if apiM.MatchTime.In(loc).Format("2006-01-02") != todayStr {
				continue
			}
			upcoming = append(upcoming, apiM)
		}
	}
	return live, upcoming
}

// LiveAndUpcomingForLeague fetches a league's page and returns the matches
// that are currently live (status-only) along with the matches scheduled for
// the user's local "today". This replaces the older UTC-date-filtered path
// that dropped live matches whose UTC date no longer matched the user's UTC
// "today" (e.g. a 22:00Z kickoff during its second half for a user past UTC
// midnight).
func (c *Client) LiveAndUpcomingForLeague(ctx context.Context, leagueID int) (live, upcoming []api.Match, err error) {
	c.rateLimiter.Wait()

	pageProps, err := fetchLeagueFromPage(ctx, c.httpClient, leagueID)
	if err != nil {
		return nil, nil, fmt.Errorf("fetch league %d page: %w", leagueID, err)
	}

	var leagueResponse struct {
		Details struct {
			ID          int    `json:"id"`
			Name        string `json:"name"`
			Country     string `json:"country"`
			CountryCode string `json:"countryCode,omitempty"`
		} `json:"details"`
		Fixtures struct {
			AllMatches []fotmobMatch `json:"allMatches"`
		} `json:"fixtures"`
	}
	if err := json.Unmarshal(pageProps, &leagueResponse); err != nil {
		return nil, nil, fmt.Errorf("decode league %d response: %w", leagueID, err)
	}

	leagueInfo := league{
		ID:          leagueResponse.Details.ID,
		Name:        leagueResponse.Details.Name,
		Country:     leagueResponse.Details.Country,
		CountryCode: leagueResponse.Details.CountryCode,
	}

	live, upcoming = classifyLeagueMatches(leagueResponse.Fixtures.AllMatches, leagueInfo, time.Now())

	for _, m := range live {
		c.StorePageURL(m.ID, m.PageURL)
	}
	for _, m := range upcoming {
		c.StorePageURL(m.ID, m.PageURL)
	}
	return live, upcoming, nil
}

// LiveAndUpcoming fetches live and upcoming matches across all active leagues
// concurrently using the status-only classifier. Best-effort aggregation: a
// league that errors is skipped, the rest still return.
func (c *Client) LiveAndUpcoming(ctx context.Context) (live, upcoming []api.Match, err error) {
	activeLeagues := ActiveLeagues()
	var mu sync.Mutex
	var wg sync.WaitGroup
	for _, leagueID := range activeLeagues {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			c.maxConcurrent <- struct{}{}
			defer func() { <-c.maxConcurrent }()

			liveL, upL, err := c.LiveAndUpcomingForLeague(ctx, id)
			if err != nil {
				return
			}
			mu.Lock()
			live = append(live, liveL...)
			upcoming = append(upcoming, upL...)
			mu.Unlock()
		}(leagueID)
	}
	wg.Wait()
	return live, upcoming, nil
}

// LiveMatches retrieves all currently live matches for today.
// Fetches matches from supported leagues and filters for those that have started but not finished.
// Only queries "fixtures" tab since live matches are not in "results" (50% fewer API calls).
// Results are cached for 2 minutes to avoid redundant fetches on quick navigation.
func (c *Client) LiveMatches(ctx context.Context) ([]api.Match, error) {
	// Check cache first (2-min TTL for quick nav in/out)
	if cached := c.cache.LiveMatches(); cached != nil {
		return cached, nil
	}

	today := time.Now()

	// Only query "fixtures" tab - live matches are in fixtures, not results
	// This reduces API calls from 28 (14 leagues × 2 tabs) to 14 (14 leagues × 1 tab)
	matches, err := c.MatchesByDateWithTabs(ctx, today, []string{"fixtures"})
	if err != nil {
		return nil, fmt.Errorf("fetch matches for date %s: %w", today.Format("2006-01-02"), err)
	}

	// Filter for live matches only (started but not finished)
	var liveMatches []api.Match
	for _, match := range matches {
		if match.Status == api.MatchStatusLive {
			liveMatches = append(liveMatches, match)
		}
	}

	// Cache the result
	c.cache.SetLiveMatches(liveMatches)

	return liveMatches, nil
}

// LiveMatchesForceRefresh fetches live matches, bypassing the cache.
// Use this for periodic refreshes to get the latest data.
func (c *Client) LiveMatchesForceRefresh(ctx context.Context) ([]api.Match, error) {
	c.cache.ClearLive()
	return c.LiveMatches(ctx)
}

// LiveMatchesForLeague fetches live matches for a single league.
// Used for progressive loading - results appear as each league responds.
func (c *Client) LiveMatchesForLeague(ctx context.Context, leagueID int) ([]api.Match, error) {
	today := time.Now()
	dateStr := today.Format("2006-01-02")

	// Fetch from API for this specific league
	matches, err := c.MatchesForLeagueAndDate(ctx, leagueID, today, "fixtures")
	if err != nil {
		return nil, err
	}

	// Filter for live matches only
	var liveMatches []api.Match
	for _, match := range matches {
		// Verify match is for today and is live
		if match.MatchTime != nil {
			//Compare with local times
			matchDate := match.MatchTime.Local().Format("2006-01-02")
			if matchDate == dateStr && match.Status == api.MatchStatusLive {
				liveMatches = append(liveMatches, match)
			}
		}
	}

	return liveMatches, nil
}

// TotalLeagues returns the number of active leagues (respects user settings).
func TotalLeagues() int {
	return len(ActiveLeagues())
}

// LeagueIDAtIndex returns the league ID at the given index from active leagues.
func LeagueIDAtIndex(index int) int {
	activeLeagues := ActiveLeagues()
	if index < 0 || index >= len(activeLeagues) {
		return 0
	}
	return activeLeagues[index]
}

// LiveUpdateParser parses match events into live update strings.
type LiveUpdateParser struct{}

// NewLiveUpdateParser creates a new live update parser.
func NewLiveUpdateParser() *LiveUpdateParser {
	return &LiveUpdateParser{}
}

// ParseEvents converts match events into human-readable update strings.
// Events are sorted by minute in descending order (most recent first).
func (p *LiveUpdateParser) ParseEvents(events []api.MatchEvent, homeTeam, awayTeam api.Team) []string {
	// Sort events by minute descending (most recent first)
	sort.Slice(events, func(i, j int) bool {
		return events[i].Minute > events[j].Minute
	})

	updates := make([]string, 0, len(events))
	for _, event := range events {
		update := p.formatEvent(event, homeTeam, awayTeam)
		if update != "" {
			updates = append(updates, update)
		}
	}

	return updates
}

// Event type prefixes for visual identification (used by UI for coloring)
const (
	EventPrefixGoal        = "●" // Solid circle - goals (red)
	EventPrefixYellowCard  = "▪" // Square - yellow card (cyan)
	EventPrefixRedCard     = "■" // Filled square - red card (red)
	EventPrefixSubstitution = "↔" // Arrow - substitution (dim)
	EventPrefixOther       = "·" // Small dot - other events (dim)
)

// formatEvent formats a single event into a readable string with symbol prefix and label.
// Format: SYMBOL TIME' [LABEL] details [H] or [A]
// Symbol prefixes are used by the UI to apply appropriate colors.
// [H] or [A] suffix indicates home or away team for UI alignment.
func (p *LiveUpdateParser) formatEvent(event api.MatchEvent, homeTeam, awayTeam api.Team) string {
	// Determine if this is a home or away team event
	isHome := event.Team.ID == homeTeam.ID
	if event.Team.ID == 0 && event.Team.ShortName != "" {
		// Fallback to short name matching if ID not set
		isHome = event.Team.ShortName == homeTeam.ShortName
	}
	teamMarker := "[A]"
	if isHome {
		teamMarker = "[H]"
	}

	switch strings.ToLower(event.Type) {
	case "goal":
		player := "Unknown"
		if event.Player != nil {
			player = *event.Player
		}
		label := "[GOAL]"
		if event.OwnGoal != nil && *event.OwnGoal {
			label = "[OWN GOAL]"
		}
		return fmt.Sprintf("%s %d' %s %s %s", EventPrefixGoal, event.Minute, label, player, teamMarker)

	case "card":
		player := "Unknown"
		if event.Player != nil {
			player = *event.Player
		}
		cardType := "yellow"
		if event.EventType != nil {
			cardType = strings.ToLower(*event.EventType)
		}
		prefix := EventPrefixYellowCard
		if cardType == "red" || cardType == "redcard" || cardType == "secondyellow" {
			prefix = EventPrefixRedCard
		}
		return fmt.Sprintf("%s %d' [CARD] %s %s", prefix, event.Minute, player, teamMarker)

	case "substitution":
		// Player = player going out, Assist = player coming in (repurposed)
		playerOut := "Unknown"
		playerIn := "Unknown"
		if event.Player != nil && *event.Player != "" {
			playerOut = *event.Player
		}
		if event.Assist != nil && *event.Assist != "" {
			playerIn = *event.Assist
		}
		// Format: show both players - "OUT→ Player | IN← Player"
		// Using special markers for UI to color-code: {OUT} and {IN}
		return fmt.Sprintf("%s %d' [SUB] {OUT}%s {IN}%s %s", EventPrefixSubstitution, event.Minute, playerOut, playerIn, teamMarker)

	case "addedtime":
		// Skip added time events - not useful
		return ""

	default:
		player := ""
		if event.Player != nil {
			player = *event.Player
		}
		if player != "" {
			return fmt.Sprintf("%s %d' %s %s", EventPrefixOther, event.Minute, player, teamMarker)
		}
		return fmt.Sprintf("%s %d' %s %s", EventPrefixOther, event.Minute, event.Type, teamMarker)
	}
}

// NewEvents compares two event lists and returns only new events.
// This is useful for detecting new updates when polling match details.
func (p *LiveUpdateParser) NewEvents(oldEvents, newEvents []api.MatchEvent) []api.MatchEvent {
	// Create a map of old event IDs for quick lookup
	oldEventMap := make(map[int]bool)
	for _, event := range oldEvents {
		oldEventMap[event.ID] = true
	}

	// Find events that don't exist in old events
	var newOnly []api.MatchEvent
	for _, event := range newEvents {
		if !oldEventMap[event.ID] {
			newOnly = append(newOnly, event)
		}
	}

	return newOnly
}
