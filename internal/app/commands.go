package app

import (
	"context"
	"sync"
	"time"

	"github.com/0xjuanma/golazo/internal/api"
	"github.com/0xjuanma/golazo/internal/data"
	"github.com/0xjuanma/golazo/internal/fotmob"
	tea "github.com/charmbracelet/bubbletea"
)

// LiveRefreshInterval is the interval between automatic live matches list refreshes.
const LiveRefreshInterval = 5 * time.Minute

// fetchLiveMatches fetches live matches from the API (used for cache check only now).
// Returns mock data if useMockData is true, otherwise uses real API.
// NOTE: For initial load, use fetchLiveLeagueData for progressive loading.
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

// LiveBatchSize is the number of leagues to fetch concurrently in each batch.
const LiveBatchSize = 4

// fetchLiveBatchData fetches live matches for a batch of leagues concurrently.
// batchIndex: 0, 1, 2, ... (each batch fetches LiveBatchSize leagues in parallel)
// Results appear after each batch completes, giving progressive updates while being fast.
func fetchLiveBatchData(client *fotmob.Client, useMockData bool, batchIndex int) tea.Cmd {
	return func() tea.Msg {
		totalLeagues := fotmob.TotalLeagues()
		startIdx := batchIndex * LiveBatchSize
		endIdx := startIdx + LiveBatchSize
		if endIdx > totalLeagues {
			endIdx = totalLeagues
		}
		isLast := endIdx >= totalLeagues

		if useMockData {
			// Return mock data only on first batch
			if batchIndex == 0 {
				return liveBatchDataMsg{
					batchIndex: batchIndex,
					isLast:     isLast,
					matches:    data.MockLiveMatches(),
				}
			}
			return liveBatchDataMsg{
				batchIndex: batchIndex,
				isLast:     isLast,
				matches:    nil,
			}
		}

		if client == nil {
			return liveBatchDataMsg{
				batchIndex: batchIndex,
				isLast:     isLast,
				matches:    nil,
			}
		}

		// Fetch all leagues in this batch concurrently
		var wg sync.WaitGroup
		var mu sync.Mutex
		var allMatches []api.Match

		for i := startIdx; i < endIdx; i++ {
			wg.Add(1)
			go func(leagueIdx int) {
				defer wg.Done()

				leagueID := fotmob.LeagueIDAtIndex(leagueIdx)
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()

				matches, err := client.LiveMatchesForLeague(ctx, leagueID)
				if err != nil || len(matches) == 0 {
					return
				}

				mu.Lock()
				allMatches = append(allMatches, matches...)
				mu.Unlock()
			}(i)
		}

		wg.Wait()

		return liveBatchDataMsg{
			batchIndex: batchIndex,
			isLast:     isLast,
			matches:    allMatches,
		}
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

// schedulePollTick schedules the next poll after 90 seconds.
// When the tick fires, it sends pollTickMsg which triggers the actual API call.
func schedulePollTick(matchID int) tea.Cmd {
	return tea.Tick(90*time.Second, func(t time.Time) tea.Msg {
		return pollTickMsg{matchID: matchID}
	})
}

// MinPollDisplayTime is the minimum time to show the "Updating..." spinner.
const MinPollDisplayTime = 1 * time.Second

// scheduleMinDisplayTime schedules a message after the minimum display time.
func scheduleMinDisplayTime() tea.Cmd {
	return tea.Tick(MinPollDisplayTime, func(t time.Time) tea.Msg {
		return pollDisplayCompleteMsg{}
	})
}

// fetchPollMatchDetails fetches match details for a poll refresh.
// This is called when pollTickMsg is received, with loading state visible.
func fetchPollMatchDetails(client *fotmob.Client, matchID int, useMockData bool) tea.Cmd {
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
