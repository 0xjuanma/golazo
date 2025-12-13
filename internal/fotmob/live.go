package fotmob

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/0xjuanma/golazo/internal/api"
)

// LiveMatches retrieves all currently live matches.
// This fetches today's matches and filters for live ones.
func (c *Client) LiveMatches(ctx context.Context) ([]api.Match, error) {
	today := time.Now()
	matches, err := c.MatchesByDate(ctx, today)
	if err != nil {
		return nil, fmt.Errorf("fetch matches for today: %w", err)
	}

	// Filter for live matches only
	liveMatches := make([]api.Match, 0)
	for _, match := range matches {
		if match.Status == api.MatchStatusLive {
			liveMatches = append(liveMatches, match)
		}
	}

	// Limit to 10 matches
	if len(liveMatches) > 10 {
		liveMatches = liveMatches[:10]
	}

	return liveMatches, nil
}

// BatchMatchDetails fetches details for multiple matches concurrently.
// This is more efficient than fetching them sequentially.
func (c *Client) BatchMatchDetails(ctx context.Context, matchIDs []int) (map[int]*api.MatchDetails, error) {
	if len(matchIDs) == 0 {
		return make(map[int]*api.MatchDetails), nil
	}

	// Use a channel to collect results
	type result struct {
		matchID int
		details *api.MatchDetails
		err     error
	}

	results := make(chan result, len(matchIDs))
	var wg sync.WaitGroup

	// Fetch each match detail concurrently
	for _, matchID := range matchIDs {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			details, err := c.MatchDetails(ctx, id)
			results <- result{
				matchID: id,
				details: details,
				err:     err,
			}
		}(matchID)
	}

	// Close results channel when all goroutines complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	detailsMap := make(map[int]*api.MatchDetails)
	for res := range results {
		if res.err != nil {
			// Log error but continue with other matches
			continue
		}
		detailsMap[res.matchID] = res.details
	}

	return detailsMap, nil
}

// LiveUpdateParser parses match events into live update strings.
type LiveUpdateParser struct{}

// NewLiveUpdateParser creates a new live update parser.
func NewLiveUpdateParser() *LiveUpdateParser {
	return &LiveUpdateParser{}
}

// ParseEvents converts match events into human-readable update strings.
func (p *LiveUpdateParser) ParseEvents(events []api.MatchEvent, homeTeam, awayTeam api.Team) []string {
	updates := make([]string, 0, len(events))

	for _, event := range events {
		update := p.formatEvent(event, homeTeam, awayTeam)
		if update != "" {
			updates = append(updates, update)
		}
	}

	return updates
}

// formatEvent formats a single event into a readable string.
func (p *LiveUpdateParser) formatEvent(event api.MatchEvent, homeTeam, awayTeam api.Team) string {
	teamName := event.Team.ShortName
	if teamName == "" {
		// Fallback to team ID matching
		if event.Team.ID == homeTeam.ID {
			teamName = homeTeam.ShortName
		} else {
			teamName = awayTeam.ShortName
		}
	}

	switch event.Type {
	case "goal":
		player := "Unknown"
		if event.Player != nil {
			player = *event.Player
		}
		assistText := ""
		if event.Assist != nil && *event.Assist != "" {
			assistText = fmt.Sprintf(" (assist: %s)", *event.Assist)
		}
		return fmt.Sprintf("%d' ⚽ Goal! %s%s - %s", event.Minute, player, assistText, teamName)

	case "card":
		player := "Unknown"
		if event.Player != nil {
			player = *event.Player
		}
		cardType := "yellow"
		if event.EventType != nil {
			cardType = *event.EventType
		}
		emoji := "🟨"
		if cardType == "red" {
			emoji = "🟥"
		}
		return fmt.Sprintf("%d' %s %s - %s (%s card)", event.Minute, emoji, player, teamName, cardType)

	case "substitution":
		player := "Unknown"
		if event.Player != nil {
			player = *event.Player
		}
		subType := "substitution"
		if event.EventType != nil {
			subType = *event.EventType
		}
		arrow := "→"
		if subType == "in" {
			arrow = "←"
		}
		return fmt.Sprintf("%d' 🔄 %s %s - %s", event.Minute, arrow, player, teamName)

	default:
		player := ""
		if event.Player != nil {
			player = *event.Player
		}
		if player != "" {
			return fmt.Sprintf("%d' • %s - %s (%s)", event.Minute, player, teamName, event.Type)
		}
		return fmt.Sprintf("%d' • %s - %s", event.Minute, event.Type, teamName)
	}
}

// GetNewEvents compares two event lists and returns only new events.
// This is useful for detecting new updates when polling match details.
func (p *LiveUpdateParser) GetNewEvents(oldEvents, newEvents []api.MatchEvent) []api.MatchEvent {
	// Create a map of old event IDs for quick lookup
	oldEventMap := make(map[int]bool)
	for _, event := range oldEvents {
		oldEventMap[event.ID] = true
	}

	// Find events that don't exist in old events
	newOnly := make([]api.MatchEvent, 0)
	for _, event := range newEvents {
		if !oldEventMap[event.ID] {
			newOnly = append(newOnly, event)
		}
	}

	return newOnly
}

