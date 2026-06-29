package app

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/0xjuanma/golazo/internal/api"
	"github.com/0xjuanma/golazo/internal/data"
	"github.com/0xjuanma/golazo/internal/fotmob"
	tea "github.com/charmbracelet/bubbletea"
)

// wcUpcomingDays is the number of forward days (including today) considered
// when fetching upcoming World Cup matches.
const wcUpcomingDays = 4

// fetchWorldCupMockData returns hardcoded World Cup data immediately.
// season selects the dataset: "2026" returns the 2026 bracket preview;
// any other value (including "") returns the completed 2022 Qatar data.
func fetchWorldCupMockData(season string) tea.Cmd {
	return func() tea.Msg {
		if season == "2026" {
			return wcDataMsg{data: data.MockWorldCupData2026()}
		}
		return wcDataMsg{data: data.MockWorldCupData()}
	}
}

// fetchWorldCupData fetches live World Cup data from FotMob.
// season is passed to the API (e.g. "2026", "2022"); "" means current/latest.
func fetchWorldCupData(parentCtx context.Context, client *fotmob.Client, season string) tea.Cmd {
	return func() tea.Msg {
		if client == nil {
			return wcDataMsg{data: data.MockWorldCupData()}
		}

		ctx, cancel := context.WithTimeout(parentCtx, 20*time.Second)
		defer cancel()

		wcData, err := client.WorldCupData(ctx, season)
		if err != nil {
			return wcDataMsg{err: err}
		}
		return wcDataMsg{data: wcData}
	}
}

// fetchWCUpcomingMatches fetches World Cup fixtures for today and the next
// wcUpcomingDays-1 days from FotMob, in parallel. Each day is requested via the
// existing per-league fixtures endpoint and only "not started" matches are
// returned (the upstream filter is already applied for the "fixtures" tab).
// Returns matches sorted ascending by kickoff time with duplicates removed.
//
// Falls back to MockWorldCupUpcoming when client is nil.
func fetchWCUpcomingMatches(parentCtx context.Context, client *fotmob.Client) ([]api.Match, error) {
	if client == nil {
		return data.MockWorldCupUpcoming(), nil
	}

	ctx, cancel := context.WithTimeout(parentCtx, 20*time.Second)
	defer cancel()

	today := time.Now()

	var (
		mu  sync.Mutex
		all []api.Match
		wg  sync.WaitGroup
	)

	for i := 0; i < wcUpcomingDays; i++ {
		day := today.AddDate(0, 0, i)
		wg.Add(1)
		go func(d time.Time) {
			defer wg.Done()
			matches, err := client.MatchesForLeagueAndDate(ctx, api.WCFotMobLeagueID, d, "fixtures")
			if err != nil {
				return
			}
			mu.Lock()
			all = append(all, matches...)
			mu.Unlock()
		}(day)
	}
	wg.Wait()

	return sortAndDedupeWCUpcoming(all), nil
}

// sortAndDedupeWCUpcoming returns the input slice sorted ascending by kickoff
// time with duplicate match IDs collapsed (first occurrence wins). Matches
// without a MatchTime are dropped.
func sortAndDedupeWCUpcoming(matches []api.Match) []api.Match {
	seen := make(map[int]struct{}, len(matches))
	out := make([]api.Match, 0, len(matches))
	for _, m := range matches {
		if m.MatchTime == nil {
			continue
		}
		if _, ok := seen[m.ID]; ok {
			continue
		}
		seen[m.ID] = struct{}{}
		out = append(out, m)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].MatchTime.Before(*out[j].MatchTime)
	})
	return out
}

// fetchWorldCupUpcoming wraps fetchWCUpcomingMatches as a tea.Cmd, emitting a
// wcUpcomingMsg with the results (or error). Falls back to the mock when
// client is nil.
func fetchWorldCupUpcoming(parentCtx context.Context, client *fotmob.Client) tea.Cmd {
	return func() tea.Msg {
		matches, err := fetchWCUpcomingMatches(parentCtx, client)
		if err != nil {
			return wcUpcomingMsg{err: err}
		}
		return wcUpcomingMsg{matches: matches}
	}
}

// fetchWorldCupTopScorers fetches the current WC top scorers from FotMob and
// emits a wcTopScorersMsg. Falls back to nil scorers when client is nil.
func fetchWorldCupTopScorers(parentCtx context.Context, client *fotmob.Client, season string) tea.Cmd {
	return func() tea.Msg {
		if client == nil {
			return wcTopScorersMsg{}
		}
		ctx, cancel := context.WithTimeout(parentCtx, 20*time.Second)
		defer cancel()

		scorers, err := client.WorldCupTopScorers(ctx, season)
		if err != nil {
			return wcTopScorersMsg{err: err}
		}
		return wcTopScorersMsg{scorers: scorers}
	}
}
