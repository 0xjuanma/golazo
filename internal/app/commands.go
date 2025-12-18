package app

import (
	"context"
	"time"

	"github.com/0xjuanma/golazo/internal/api"
	"github.com/0xjuanma/golazo/internal/data"
	"github.com/0xjuanma/golazo/internal/fotmob"
	tea "github.com/charmbracelet/bubbletea"
)

// LiveRefreshInterval is the interval between automatic live matches list refreshes.
const LiveRefreshInterval = 5 * time.Minute

// fetchLiveMatches fetches live matches from the API.
// Returns mock data if useMockData is true, otherwise uses real API.
func fetchLiveMatches(client *fotmob.Client, useMockData bool) tea.Cmd {
	return func() tea.Msg {
		if useMockData {
			return liveMatchesMsg{matches: data.MockLiveMatches()}
		}

		if client == nil {
			return liveMatchesMsg{matches: nil}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		matches, err := client.LiveMatches(ctx)
		if err != nil {
			return liveMatchesMsg{matches: nil}
		}

		return liveMatchesMsg{matches: matches}
	}
}

// scheduleLiveRefresh schedules the next live matches refresh after 5 minutes.
// This is used to keep the live matches list current while the user is in the view.
func scheduleLiveRefresh(client *fotmob.Client, useMockData bool) tea.Cmd {
	return tea.Tick(LiveRefreshInterval, func(t time.Time) tea.Msg {
		if useMockData {
			return liveRefreshMsg{matches: data.MockLiveMatches()}
		}

		if client == nil {
			return liveRefreshMsg{matches: nil}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Force refresh to bypass cache
		matches, err := client.LiveMatchesForceRefresh(ctx)
		if err != nil {
			return liveRefreshMsg{matches: nil}
		}

		return liveRefreshMsg{matches: matches}
	})
}

// fetchMatchDetails fetches match details from the API.
// Returns mock data if useMockData is true, otherwise uses real API.
func fetchMatchDetails(client *fotmob.Client, matchID int, useMockData bool) tea.Cmd {
	return func() tea.Msg {
		if useMockData {
			details, _ := data.MockMatchDetails(matchID)
			return matchDetailsMsg{details: details}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		details, err := client.MatchDetails(ctx, matchID)
		if err != nil {
			return matchDetailsMsg{details: nil}
		}

		return matchDetailsMsg{details: details}
	}
}

// pollMatchDetails polls match details every 90 seconds for live updates.
// Conservative interval to avoid rate limiting.
func pollMatchDetails(client *fotmob.Client, parser *fotmob.LiveUpdateParser, matchID int, lastEvents []api.MatchEvent, useMockData bool) tea.Cmd {
	return tea.Tick(90*time.Second, func(t time.Time) tea.Msg {
		if useMockData {
			details, _ := data.MockMatchDetails(matchID)
			return matchDetailsMsg{details: details}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		details, err := client.MatchDetails(ctx, matchID)
		if err != nil {
			return matchDetailsMsg{details: nil}
		}

		return matchDetailsMsg{details: details}
	})
}

// fetchStatsDayData fetches stats data for a single day (progressive loading).
// dayIndex: 0 = today, 1 = yesterday, etc.
// totalDays: total number of days to fetch (for isLast calculation)
// This enables showing results immediately as each day's data arrives.
func fetchStatsDayData(client *fotmob.Client, useMockData bool, dayIndex int, totalDays int) tea.Cmd {
	return func() tea.Msg {
		isToday := dayIndex == 0
		isLast := dayIndex == totalDays-1

		if useMockData {
			if isToday {
				return statsDayDataMsg{
					dayIndex: dayIndex,
					isToday:  true,
					isLast:   isLast,
					finished: data.MockFinishedMatches(),
					upcoming: nil,
				}
			}
			return statsDayDataMsg{
				dayIndex: dayIndex,
				isToday:  false,
				isLast:   isLast,
				finished: nil,
				upcoming: nil,
			}
		}

		if client == nil {
			return statsDayDataMsg{
				dayIndex: dayIndex,
				isToday:  isToday,
				isLast:   isLast,
				finished: nil,
				upcoming: nil,
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Calculate the date for this day
		today := time.Now().UTC()
		date := today.AddDate(0, 0, -dayIndex)

		var matches []api.Match
		var err error

		if isToday {
			// Today: need both fixtures (upcoming) and results (finished)
			matches, err = client.MatchesByDateWithTabs(ctx, date, []string{"fixtures", "results"})
		} else {
			// Past days: only need results (finished matches)
			matches, err = client.MatchesByDateWithTabs(ctx, date, []string{"results"})
		}

		if err != nil {
			return statsDayDataMsg{
				dayIndex: dayIndex,
				isToday:  isToday,
				isLast:   isLast,
				finished: nil,
				upcoming: nil,
			}
		}

		// Split matches into finished and upcoming
		var finished, upcoming []api.Match
		for _, match := range matches {
			if match.Status == api.MatchStatusFinished {
				finished = append(finished, match)
			} else if match.Status == api.MatchStatusNotStarted && isToday {
				upcoming = append(upcoming, match)
			}
		}

		return statsDayDataMsg{
			dayIndex: dayIndex,
			isToday:  isToday,
			isLast:   isLast,
			finished: finished,
			upcoming: upcoming,
		}
	}
}

// fetchStatsMatchDetailsFotmob fetches match details from FotMob API for stats view.
func fetchStatsMatchDetailsFotmob(client *fotmob.Client, matchID int, useMockData bool) tea.Cmd {
	return func() tea.Msg {
		if useMockData {
			details, _ := data.MockFinishedMatchDetails(matchID)
			return matchDetailsMsg{details: details}
		}

		if client == nil {
			return matchDetailsMsg{details: nil}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		details, err := client.MatchDetails(ctx, matchID)
		if err != nil {
			return matchDetailsMsg{details: nil}
		}

		return matchDetailsMsg{details: details}
	}
}
