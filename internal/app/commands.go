package app

import (
	"context"
	"sync"
	"time"

	"github.com/0xjuanma/golazo/internal/api"
	"github.com/0xjuanma/golazo/internal/data"
	"github.com/0xjuanma/golazo/internal/fotmob"
	"github.com/0xjuanma/golazo/internal/reddit"
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

// fetchMatchDetailsForceRefresh fetches match details with cache bypass.
// Forces fresh data from the API, ignoring any cached data.
func fetchMatchDetailsForceRefresh(client *fotmob.Client, matchID int, useMockData bool) tea.Cmd {
	return func() tea.Msg {
		if useMockData {
			details, _ := data.MockMatchDetails(matchID)
			return matchDetailsMsg{details: details}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		details, err := client.MatchDetailsForceRefresh(ctx, matchID)
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

// PollSpinnerDuration is how long to show the "Updating..." spinner.
const PollSpinnerDuration = 1 * time.Second

// schedulePollSpinnerHide schedules hiding the spinner after the display duration.
func schedulePollSpinnerHide() tea.Cmd {
	return tea.Tick(PollSpinnerDuration, func(t time.Time) tea.Msg {
		return pollDisplayCompleteMsg{}
	})
}

// fetchPollMatchDetails fetches match details for a poll refresh.
// This is called when pollTickMsg is received, with loading state visible.
// Uses force refresh to bypass cache and ensure fresh data for live matches.
func fetchPollMatchDetails(client *fotmob.Client, matchID int, useMockData bool) tea.Cmd {
	return func() tea.Msg {
		if useMockData {
			details, _ := data.MockMatchDetails(matchID)
			return matchDetailsMsg{details: details}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Force refresh to bypass cache - live matches need fresh data
		details, err := client.MatchDetailsForceRefresh(ctx, matchID)
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

// fetchGoalLinks fetches goal replay links from Reddit for all goals in a match.
// This is called on-demand when match details are loaded/displayed.
// Links are cached persistently to avoid redundant API calls.
func fetchGoalLinks(redditClient *reddit.Client, details *api.MatchDetails) tea.Cmd {
	return func() tea.Msg {
		if redditClient == nil || details == nil {
			return goalLinksMsg{matchID: 0, links: nil}
		}

		// Extract goal events from match details
		var goals []reddit.GoalInfo
		for _, event := range details.Events {
			if event.Type != "goal" {
				continue
			}

			// Debug log goal extraction (will be logged when redditClient.GoalLinks is called)

			scorer := ""
			if event.Player != nil {
				scorer = *event.Player
			}

			// Determine if goal is for home team
			isHome := event.Team.ID == details.HomeTeam.ID

			// Get scores at the time of goal (approximate)
			homeScore := 0
			awayScore := 0
			if details.HomeScore != nil {
				homeScore = *details.HomeScore
			}
			if details.AwayScore != nil {
				awayScore = *details.AwayScore
			}

			// Get match time for date-based Reddit search
			matchTime := time.Now() // Default to now for live matches
			if details.MatchTime != nil {
				matchTime = *details.MatchTime
			}

			goals = append(goals, reddit.GoalInfo{
				MatchID:      details.ID,
				HomeTeam:     details.HomeTeam.Name,
				AwayTeam:     details.AwayTeam.Name,
				ScorerName:   scorer,
				Minute:       event.Minute,
				DisplayMinute: event.DisplayMinute,
				HomeScore:    homeScore,
				AwayScore:    awayScore,
				IsHomeTeam:   isHome,
				MatchTime:    matchTime,
			})
		}

		if len(goals) == 0 {
			return goalLinksMsg{matchID: details.ID, links: nil}
		}

		// Fetch links for all goals (uses cache internally)
		links := redditClient.GoalLinks(goals)

		return goalLinksMsg{matchID: details.ID, links: links}
	}
}

// fetchGoalLinksForGoals fetches goal replay links from Reddit for specific goals.
// This is used when new goals are detected that don't have cached links yet.
// Uses the same batching and rate limiting as the full match fetch.
func fetchGoalLinksForGoals(redditClient *reddit.Client, goals []reddit.GoalInfo) tea.Cmd {
	return func() tea.Msg {
		if redditClient == nil || len(goals) == 0 {
			return goalLinksMsg{matchID: 0, links: nil}
		}

		// Fetch links for the specific goals (uses cache internally)
		links := redditClient.GoalLinks(goals)

		// Use the match ID from the first goal (all goals should be from same match)
		matchID := goals[0].MatchID

		return goalLinksMsg{matchID: matchID, links: links}
	}
}

// fetchStandings fetches league standings for a specific league.
// Used to populate the standings dialog.
func fetchStandings(client *fotmob.Client, leagueID int, leagueName string, homeTeamID, awayTeamID int) tea.Cmd {
	return func() tea.Msg {
		if client == nil {
			return standingsMsg{leagueID: leagueID, standings: nil}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		standings, err := client.LeagueTable(ctx, leagueID)
		if err != nil {
			return standingsMsg{leagueID: leagueID, standings: nil}
		}

		return standingsMsg{
			leagueID:   leagueID,
			leagueName: leagueName,
			standings:  standings,
			homeTeamID: homeTeamID,
			awayTeamID: awayTeamID,
		}
	}
}


