package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/0xjuanma/golazo/internal/api"
	"github.com/0xjuanma/golazo/internal/data"
	"github.com/0xjuanma/golazo/internal/fotmob"
	"github.com/spf13/cobra"
)

// MaxFinishedDays is the maximum supported lookback window for `golazo finished`.
const MaxFinishedDays = 7

// finishedDayFetcher abstracts MatchesByDateWithTabs for testing.
type finishedDayFetcher func(ctx context.Context, date time.Time, tabs []string) ([]api.Match, error)

// collectFinished iterates `days` calendar days ending today, calling the
// per-day fetcher for each. It deduplicates by Match.ID and returns:
//   - the union of finished matches
//   - the list of date strings (YYYY-MM-DD) whose fetch failed
//   - an error iff ALL days failed (callers may then return upstream_error)
//
// This is the layer where the multi-day, dedup, and degraded-surface behavior
// lives, so tests target it directly.
//
// When includeUpcoming is true, today's not-yet-started matches are also
// included in the result. Past days never contain not_started matches.
func collectFinished(ctx context.Context, fetch finishedDayFetcher, now time.Time, days int, includeUpcoming bool) ([]api.Match, []string, error) {
	dedup := make(map[int]api.Match, days*10)
	var failedDates []string
	successCount := 0
	var lastErr error

	for i := 0; i < days; i++ {
		date := now.AddDate(0, 0, -i).UTC()
		dateStr := date.Format("2006-01-02")
		isToday := i == 0

		// Today needs fixtures+results (TUI behavior); past days only need results.
		tabs := []string{"results"}
		if isToday {
			tabs = []string{"fixtures", "results"}
		}

		matches, err := fetch(ctx, date, tabs)
		if err != nil {
			failedDates = append(failedDates, dateStr)
			lastErr = err
			continue
		}
		successCount++
		for _, m := range matches {
			switch m.Status {
			case api.MatchStatusFinished:
				dedup[m.ID] = m
			case api.MatchStatusNotStarted:
				if includeUpcoming && isToday {
					dedup[m.ID] = m
				}
			}
		}
	}

	if successCount == 0 {
		return nil, failedDates, lastErr
	}

	out := make([]api.Match, 0, len(dedup))
	for _, m := range dedup {
		out = append(out, m)
	}
	return out, failedDates, nil
}

func defaultFinishedFetcher(c *fotmob.Client) finishedDayFetcher {
	return c.MatchesByDateWithTabs
}

// finishedFlags extends the common flag set with --days and --include-upcoming.
type finishedFlags struct {
	cliFlags
	days            int
	includeUpcoming bool
}

var finishedFlagSet finishedFlags

// runFinished is the testable core of the `finished` subcommand.
func runFinished(stdout, stderr io.Writer, flags finishedFlags) int {
	applyPretty(flags.cliFlags)

	if flags.days < 1 || flags.days > MaxFinishedDays {
		return WriteError(stderr, ErrCodeInvalidArgs,
			NewInvalidArg("--days must be between 1 and %d, got %d", MaxFinishedDays, flags.days))
	}

	client, ctx, cancel, err := newHeadlessClient(runtimeOpts{
		mock:    flags.mock,
		debug:   flags.debug,
		timeout: flags.timeout,
	})
	defer cancel()
	if err == ErrOffline {
		return WriteError(stderr, ErrCodeOffline, err)
	}
	if err != nil {
		return WriteError(stderr, ErrCodeUpstreamError, err)
	}

	var (
		matches     []api.Match
		failedDates []string
	)

	if flags.mock {
		// Mock data is single-day; serve it regardless of --days.
		matches = data.MockFinishedMatches()
	} else {
		matches, failedDates, err = collectFinished(ctx, defaultFinishedFetcher(client), time.Now(), flags.days, flags.includeUpcoming)
		if err != nil {
			return WriteError(stderr, ClassifyClientError(err, isTimeout(ctx)), err)
		}
		// Guard against silent timeouts: per-day fetches may all swallow the
		// cancellation and return empty without an error. Surface as timeout
		// so agents can distinguish "no finished matches" from "timed out".
		if isTimeout(ctx) {
			return WriteError(stderr, ErrCodeTimeout,
				fmt.Errorf("finished matches fetch timed out after %s", flags.timeout))
		}
	}

	SortMatches(matches)

	var writeErr error
	if len(failedDates) > 0 {
		writeErr = WriteDegraded(stdout, matches, failedDates)
	} else {
		writeErr = WriteJSON(stdout, matches)
	}
	if writeErr != nil {
		return WriteError(stderr, ErrCodeUpstreamError, writeErr)
	}
	return ExitOK
}

var finishedCmd = &cobra.Command{
	Use:           "finished",
	Short:         "List finished matches over a day window as JSON",
	Long: `Fetches finished matches for the last --days days (default 1 = today) across active leagues. Use --include-upcoming to also include today's not-yet-started matches. Partial failures surface as degraded:true with failed_dates listed.

Example output:
  {"status":"ok","count":2,"data":[{"id":4506420,"league":{"id":47,"name":"Premier League","country":"England"},"home_team":{"name":"Liverpool","short_name":"Liverpool"},"away_team":{"name":"Arsenal","short_name":"Arsenal"},"status":"finished","home_score":3,"away_score":1,"match_time":"2026-06-12T15:00:00Z"}]}

Degraded example (one date failed but others succeeded):
  {"status":"ok","degraded":true,"failed_dates":["2026-06-10"],"count":12,"data":[...]}`,
	SilenceUsage:  true,
	SilenceErrors: true,
	Run: func(cmd *cobra.Command, args []string) {
		code := runFinished(os.Stdout, os.Stderr, finishedFlagSet)
		if code != ExitOK {
			os.Exit(code)
		}
	},
}

func init() {
	addCommonCLIFlags(finishedCmd, &finishedFlagSet.cliFlags)
	finishedCmd.Flags().IntVar(&finishedFlagSet.days, "days", 1, "Number of days to look back (1..7)")
	finishedCmd.Flags().BoolVar(&finishedFlagSet.includeUpcoming, "include-upcoming", false, "Also include today's not-yet-started matches in the result")
	rootCmd.AddCommand(finishedCmd)
}
