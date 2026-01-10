package fotmob

import (
	"context"
	"fmt"
	"time"

	"github.com/0xjuanma/golazo/internal/api"
)

// StatsData holds all matches data for the stats view.
// This is returned by StatsData and contains both finished and upcoming matches.
type StatsData struct {
	// AllFinished contains finished matches for all fetched days (5 days by default)
	AllFinished []api.Match
	// TodayFinished contains only today's finished matches (filtered from AllFinished)
	TodayFinished []api.Match
	// TodayUpcoming contains today's upcoming matches
	TodayUpcoming []api.Match
}

// StatsDataDays is the number of days to fetch for stats view.
// 5 days ensures we have data even during mid-week breaks.
const StatsDataDays = 5

// StatsData fetches all stats data in one call: 5 days of finished matches + today's upcoming.
// This is the primary API for the stats view - always fetches 5 days, then filters client-side.
//
// OPTIMIZATION: Only queries "fixtures" tab for today (upcoming matches).
// Past days only need "results" tab (finished matches).
//
// API calls breakdown:
//   - Today: 14 leagues × 2 tabs = 28 requests (need both fixtures + results)
//   - Past 4 days: 14 leagues × 1 tab × 4 = 56 requests (only results)
//   - Total: 84 requests
//
// Benefits:
// - Single fetch pattern (always 5 days)
// - Covers mid-week breaks when no matches scheduled
// - Instant switching between Today/5d views after initial load
func (c *Client) StatsData(ctx context.Context) (*StatsData, error) {
	today := time.Now().UTC()
	todayStr := today.Format("2006-01-02")

	// Use maps to deduplicate matches by ID
	allFinishedMap := make(map[int]api.Match)
	todayFinishedMap := make(map[int]api.Match)
	todayUpcomingMap := make(map[int]api.Match)
	var lastErr error
	successCount := 0

	// Fetch 5 days of matches (today + last 4 days)
	for i := 0; i < StatsDataDays; i++ {
		date := today.AddDate(0, 0, -i)
		dateStr := date.Format("2006-01-02")
		isToday := dateStr == todayStr

		var matches []api.Match
		var err error

		if isToday {
			// Today: need both fixtures (upcoming) and results (finished)
			matches, err = c.MatchesByDateWithTabs(ctx, date, []string{"fixtures", "results"})
		} else {
			// Past days: only need results (finished matches)
			matches, err = c.MatchesByDateWithTabs(ctx, date, []string{"results"})
		}

		if err != nil {
			lastErr = fmt.Errorf("fetch matches for date %s: %w", dateStr, err)
			continue
		}
		successCount++

		// Process matches for this day - deduplicate by match ID
		for _, match := range matches {
			if match.Status == api.MatchStatusFinished {
				allFinishedMap[match.ID] = match
				// Also track today's finished separately
				if isToday {
					todayFinishedMap[match.ID] = match
				}
			} else if match.Status == api.MatchStatusNotStarted && isToday {
				// Only today has upcoming matches
				todayUpcomingMap[match.ID] = match
			}
		}
	}

	// Return error only if all days failed
	if successCount == 0 {
		return nil, fmt.Errorf("failed to fetch matches for any date: %w", lastErr)
	}

	// Convert maps to slices
	allFinished := make([]api.Match, 0, len(allFinishedMap))
	for _, match := range allFinishedMap {
		allFinished = append(allFinished, match)
	}

	todayFinished := make([]api.Match, 0, len(todayFinishedMap))
	for _, match := range todayFinishedMap {
		todayFinished = append(todayFinished, match)
	}

	todayUpcoming := make([]api.Match, 0, len(todayUpcomingMap))
	for _, match := range todayUpcomingMap {
		todayUpcoming = append(todayUpcoming, match)
	}

	return &StatsData{
		AllFinished:   allFinished,
		TodayFinished: todayFinished,
		TodayUpcoming: todayUpcoming,
	}, nil
}
