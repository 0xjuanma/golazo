package cmd

import (
	"io"
	"os"
	"sort"

	"github.com/0xjuanma/golazo/internal/api"
	"github.com/0xjuanma/golazo/internal/data"
	"github.com/spf13/cobra"
)

// leaguesFlags extends the common flag set with --all.
type leaguesFlags struct {
	cliFlags
	all bool
}

var leaguesFlagSet leaguesFlags

// resolveLeagues returns league metadata for either the active set (default)
// or every supported league (--all). Output is sorted by ID for determinism.
// Pure in-memory read of data.AllSupportedLeagues; no API call.
func resolveLeagues(all bool) []api.League {
	// Build an ID → LeagueInfo lookup over the full catalog.
	catalog := make(map[int]data.LeagueInfo, 200)
	for _, regionLeagues := range data.AllSupportedLeagues {
		for _, info := range regionLeagues {
			catalog[info.ID] = info
		}
	}

	var ids []int
	if all {
		ids = make([]int, 0, len(catalog))
		for id := range catalog {
			ids = append(ids, id)
		}
	} else {
		ids = append(ids, data.ActiveLeagueIDs()...)
	}

	sort.Ints(ids)

	out := make([]api.League, 0, len(ids))
	for _, id := range ids {
		info, ok := catalog[id]
		if !ok {
			// User-selected ID we don't have metadata for; surface as bare ID.
			out = append(out, api.League{ID: id})
			continue
		}
		out = append(out, api.League{
			ID:      info.ID,
			Name:    info.Name,
			Country: info.Country,
		})
	}
	return out
}

// runLeagues is the testable core of the `leagues` subcommand.
func runLeagues(stdout, stderr io.Writer, flags leaguesFlags) int {
	applyPretty(flags.cliFlags)
	leagues := resolveLeagues(flags.all)
	if err := WriteJSON(stdout, leagues); err != nil {
		return WriteError(stderr, ErrCodeUpstreamError, err)
	}
	return ExitOK
}

var leaguesCmd = &cobra.Command{
	Use:           "leagues",
	Short:         "List active (or all) supported leagues as JSON",
	Long: `Prints a JSON envelope listing currently active leagues (or all supported leagues with --all). No network calls. Useful for discovering league IDs to interpret live/finished results.

Example output:
  {"status":"ok","count":3,"data":[{"id":47,"name":"Premier League","country":"England","country_code":""},{"id":87,"name":"La Liga","country":"Spain","country_code":""},{"id":42,"name":"UEFA Champions League","country":"Europe","country_code":""}]}`,
	SilenceUsage:  true,
	SilenceErrors: true,
	Run: func(cmd *cobra.Command, args []string) {
		code := runLeagues(os.Stdout, os.Stderr, leaguesFlagSet)
		if code != ExitOK {
			os.Exit(code)
		}
	},
}

func init() {
	addPrettyOnlyFlag(leaguesCmd, &leaguesFlagSet.cliFlags)
	leaguesCmd.Flags().BoolVar(&leaguesFlagSet.all, "all", false, "List every supported league, not just the active selection")
	rootCmd.AddCommand(leaguesCmd)
}
