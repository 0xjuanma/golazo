package app

import (
	"context"
	"fmt"
	"sort"
	"strings"
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

// LiveBatchSize is the number of leagues to fetch concurrently in each batch.
const LiveBatchSize = 4

// fetchLiveBatchData fetches live matches for a batch of leagues concurrently.
// batchIndex: 0, 1, 2, ... (each batch fetches LiveBatchSize leagues in parallel)
// Results appear after each batch completes, giving progressive updates while being fast.
func fetchLiveBatchData(parentCtx context.Context, client *fotmob.Client, useMockData bool, batchIndex int) tea.Cmd {
	return func() tea.Msg {
		totalLeagues := fotmob.TotalLeagues()
		startIdx := batchIndex * LiveBatchSize
		endIdx := startIdx + LiveBatchSize
		endIdx = min(endIdx, totalLeagues)
		isLast := endIdx >= totalLeagues

		// Check if cancelled before starting work
		if parentCtx.Err() != nil {
			return liveBatchDataMsg{batchIndex: batchIndex, isLast: true}
		}

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

		// Fetch all leagues in this batch concurrently.
		// Classification (live vs upcoming) is done inside the fotmob client
		// by status, not by UTC date, so live matches that started before
		// the user's UTC midnight still surface here.
		var wg sync.WaitGroup
		var mu sync.Mutex
		allLive := make([]api.Match, 0, (endIdx-startIdx)*5)
		allUpcoming := make([]api.Match, 0, (endIdx-startIdx)*5)

		for i := startIdx; i < endIdx; i++ {
			wg.Add(1)
			go func(leagueIdx int) {
				defer wg.Done()

				leagueID := fotmob.LeagueIDAtIndex(leagueIdx)
				ctx, cancel := context.WithTimeout(parentCtx, 10*time.Second)
				defer cancel()

				live, upcoming, err := client.LiveAndUpcomingForLeague(ctx, leagueID)
				if err != nil {
					return
				}
				if len(live) == 0 && len(upcoming) == 0 {
					return
				}

				mu.Lock()
				allLive = append(allLive, live...)
				allUpcoming = append(allUpcoming, upcoming...)
				mu.Unlock()
			}(i)
		}

		wg.Wait()

		return liveBatchDataMsg{
			batchIndex: batchIndex,
			isLast:     isLast,
			matches:    allLive,
			upcoming:   allUpcoming,
		}
	}
}

// scheduleLiveRefresh schedules the next live matches refresh after 5 minutes.
// This is used to keep the live matches list current while the user is in the view.
// Fetches both live and upcoming matches so the upcoming section stays current
// as matches transition from upcoming to live.
func scheduleLiveRefresh(client *fotmob.Client, useMockData bool) tea.Cmd {
	return tea.Tick(LiveRefreshInterval, func(t time.Time) tea.Msg {
		if useMockData {
			return liveRefreshMsg{matches: data.MockLiveMatches()}
		}

		if client == nil {
			return liveRefreshMsg{matches: nil}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Classification (live vs upcoming) happens inside the fotmob client
		// by status, not by UTC date. The page-body cache has a short TTL so
		// the 5-minute refresh always lands on stale entries and refetches.
		live, upcoming, err := client.LiveAndUpcoming(ctx)
		if err != nil {
			return liveRefreshMsg{matches: nil}
		}

		return liveRefreshMsg{matches: live, upcoming: upcoming}
	})
}

// refreshLiveNow forces an immediate live-matches refresh by clearing the
// FotMob league-page cache before re-fetching. Wired to the user-initiated
// "r" key in the live view so the user can pull fresh data when FotMob's
// server-rendered page lags realtime around kickoff.
func refreshLiveNow(client *fotmob.Client, useMockData bool) tea.Cmd {
	return func() tea.Msg {
		if useMockData {
			return liveRefreshMsg{matches: data.MockLiveMatches()}
		}

		if client == nil {
			return liveRefreshMsg{matches: nil}
		}

		client.Cache().ClearPages()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		live, upcoming, err := client.LiveAndUpcoming(ctx)
		if err != nil {
			return liveRefreshMsg{matches: nil}
		}

		return liveRefreshMsg{matches: live, upcoming: upcoming}
	}
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
			return matchDetailsMsg{details: nil, err: err}
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
			return matchDetailsMsg{details: nil, err: err}
		}

		return matchDetailsMsg{details: details}
	}
}

// schedulePollTick schedules the next poll after 90 seconds.
// When the tick fires, it sends pollTickMsg which triggers the actual API call.
func schedulePollTick(matchID, gen int) tea.Cmd {
	return tea.Tick(90*time.Second, func(t time.Time) tea.Msg {
		return pollTickMsg{matchID: matchID, gen: gen}
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
			return matchDetailsMsg{details: nil, err: err}
		}

		return matchDetailsMsg{details: details}
	}
}

// fetchStatsDayData fetches stats data for a single day (progressive loading).
// dayIndex: 0 = today, 1 = yesterday, etc.
// totalDays: total number of days to fetch (for isLast calculation)
// This enables showing results immediately as each day's data arrives.
func fetchStatsDayData(parentCtx context.Context, client *fotmob.Client, useMockData bool, dayIndex int, totalDays int) tea.Cmd {
	return func() tea.Msg {
		isToday := dayIndex == 0
		isLast := dayIndex == totalDays-1

		// Check if cancelled before starting work
		if parentCtx.Err() != nil {
			return statsDayDataMsg{dayIndex: dayIndex, isToday: isToday, isLast: true}
		}

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

		ctx, cancel := context.WithTimeout(parentCtx, 30*time.Second)
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
				err:      err,
			}
		}

		// Split matches into finished and upcoming
		finished := make([]api.Match, 0, len(matches)/2)
		upcoming := make([]api.Match, 0, len(matches)/4)
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
			return matchDetailsMsg{details: nil, err: err}
		}

		return matchDetailsMsg{details: details}
	}
}

// fetchGoalLinks opens a streaming subscription to the reddit queue for all
// goals in `details`. Returns a tea.Cmd that emits a goalLinkStreamMsg
// carrying the result channel; the Update loop stashes the channel and arms
// successive waitForGoalLink reader Cmds at the queue's cadence.
//
// Goals already in the persistent cache surface immediately (they're written
// into the channel by GoalLinksAsync before queueing begins); uncached goals
// arrive at QueueInterval cadence. Each emitted goalLinkMsg carries a single
// result so the UI updates progressively instead of waiting for the entire
// match's goals to resolve.
func fetchGoalLinks(redditClient *reddit.Client, details *api.MatchDetails) tea.Cmd {
	if redditClient == nil || details == nil {
		return nil
	}

	goals := buildGoalInfos(details)
	if len(goals) == 0 {
		return nil
	}

	// Log the per-goal running scores so the corrected matcher inputs are
	// observable in golazo_debug.log without trawling individual search
	// queries. Useful for verifying World Cup / national-team retrievals.
	redditClient.DebugLog(fmt.Sprintf("fetchGoalLinks: match=%d %s vs %s — %d goals: %s",
		details.ID, details.HomeTeam.Name, details.AwayTeam.Name, len(goals),
		formatGoalSummary(goals)))

	matchID := details.ID
	results := redditClient.GoalLinksAsync(goals)

	return func() tea.Msg {
		return goalLinkStreamMsg{matchID: matchID, ch: results}
	}
}

// waitForGoalLink returns a tea.Cmd that blocks until the next GoalResult
// arrives on results (or the channel closes), then emits a goalLinkMsg or a
// terminal goalLinksDoneMsg. The Update handler re-arms this Cmd after each
// goalLinkMsg, forming the subscription loop.
func waitForGoalLink(matchID int, results <-chan reddit.GoalResult) tea.Cmd {
	return func() tea.Msg {
		r, ok := <-results
		if !ok {
			return goalLinksDoneMsg{matchID: matchID}
		}
		return goalLinkMsg{matchID: matchID, key: r.Key, link: r.Link}
	}
}

// formatGoalSummary renders a compact "27' AUS 1-0 (Irankunda)" list for debug
// logging. Kept private to commands.go since it's purely a log-helper.
func formatGoalSummary(goals []reddit.GoalInfo) string {
	parts := make([]string, 0, len(goals))
	for _, g := range goals {
		team := g.AwayTeam
		if g.IsHomeTeam {
			team = g.HomeTeam
		}
		scorer := g.ScorerName
		if scorer == "" {
			scorer = "?"
		}
		parts = append(parts, fmt.Sprintf("%d' %s %d-%d (%s)",
			g.Minute, team, g.HomeScore, g.AwayScore, scorer))
	}
	return strings.Join(parts, ", ")
}

// buildGoalInfos converts match details into Reddit goal-search inputs with a
// per-goal running score. Goals are processed in minute order so each GoalInfo
// carries the score AT THE TIME of that goal (not the final score), which is
// what r/soccer goal-video titles encode. Own goals credit the opposing team.
func buildGoalInfos(details *api.MatchDetails) []reddit.GoalInfo {
	if details == nil {
		return nil
	}

	// Collect goal events with their original index for stable ordering when
	// minutes tie. Defensive-sort because event ordering is not guaranteed
	// across data sources.
	type indexedEvent struct {
		event api.MatchEvent
		idx   int
	}
	var goalEvents []indexedEvent
	for i, ev := range details.Events {
		if ev.Type == "goal" {
			goalEvents = append(goalEvents, indexedEvent{event: ev, idx: i})
		}
	}
	sort.SliceStable(goalEvents, func(i, j int) bool {
		if goalEvents[i].event.Minute != goalEvents[j].event.Minute {
			return goalEvents[i].event.Minute < goalEvents[j].event.Minute
		}
		return goalEvents[i].idx < goalEvents[j].idx
	})

	matchTime := time.Now()
	if details.MatchTime != nil {
		matchTime = *details.MatchTime
	}

	var goals []reddit.GoalInfo
	runningHome, runningAway := 0, 0
	for _, ge := range goalEvents {
		ev := ge.event
		isHome := ev.Team.ID == details.HomeTeam.ID
		isOwnGoal := ev.OwnGoal != nil && *ev.OwnGoal
		// Own goals are recorded against the scoring player's own team in the
		// event stream, but the goal credits the opposing side on the scoreboard.
		creditsHome := isHome
		if isOwnGoal {
			creditsHome = !isHome
		}
		if creditsHome {
			runningHome++
		} else {
			runningAway++
		}

		scorer := ""
		if ev.Player != nil {
			scorer = *ev.Player
		}

		goals = append(goals, reddit.GoalInfo{
			MatchID:       details.ID,
			HomeTeam:      details.HomeTeam.Name,
			AwayTeam:      details.AwayTeam.Name,
			HomeTeamShort: details.HomeTeam.ShortName,
			AwayTeamShort: details.AwayTeam.ShortName,
			ScorerName:    scorer,
			Minute:        ev.Minute,
			DisplayMinute: ev.DisplayMinute,
			HomeScore:     runningHome,
			AwayScore:     runningAway,
			IsHomeTeam:    creditsHome,
			MatchTime:     matchTime,
		})
	}
	return goals
}

// fetchStandings fetches league standings for a specific league.
// Used to populate the standings dialog.
// parentLeagueID is used for multi-season leagues (e.g., Liga MX Clausura -> Liga MX)
// where the sub-league ID has no standings but the parent league does.
func fetchStandings(client *fotmob.Client, leagueID int, leagueName string, parentLeagueID int, homeTeamID, awayTeamID int) tea.Cmd {
	return func() tea.Msg {
		if client == nil {
			return standingsMsg{leagueID: leagueID, standings: nil}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		standings, err := client.LeagueTableWithParent(ctx, leagueID, leagueName, parentLeagueID)
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
