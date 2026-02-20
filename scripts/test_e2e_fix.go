package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/0xjuanma/golazo/internal/api"
	"github.com/0xjuanma/golazo/internal/fotmob"
)

func main() {
	client := fotmob.NewClient()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Try multiple recent dates to find one with matches
	fmt.Println("=== End-to-End Test ===\n")

	fmt.Println("--- Test 1: Fetch matches with scores ---")
	var matches []api.Match
	var testDate time.Time
	for daysBack := 1; daysBack <= 5; daysBack++ {
		testDate = time.Now().AddDate(0, 0, -daysBack)
		fmt.Printf("Trying %s...\n", testDate.Format("2006-01-02"))
		var err2 error
		matches, err2 = client.MatchesByDateWithTabs(ctx, testDate, []string{"results"})
		if err2 != nil {
			fmt.Printf("  Error: %v\n", err2)
			continue
		}
		if len(matches) > 0 {
			fmt.Printf("  Found %d matches\n", len(matches))
			break
		}
	}
	fmt.Printf("\nTotal matches: %d\n", len(matches))

	var matchWithScores int
	var matchWithPageURL int
	var testMatchID int
	for _, m := range matches {
		if m.HomeScore != nil && m.AwayScore != nil {
			matchWithScores++
		}
		if m.PageURL != "" {
			matchWithPageURL++
		}
		// Pick a finished match with a pageURL for details test
		if testMatchID == 0 && m.PageURL != "" && m.HomeScore != nil {
			testMatchID = m.ID
			fmt.Printf("  Sample: %s %d - %d %s (ID: %d, PageURL: %s)\n",
				m.HomeTeam.Name, *m.HomeScore, *m.AwayScore, m.AwayTeam.Name, m.ID, m.PageURL)
		}
	}

	fmt.Printf("  Matches with scores: %d/%d\n", matchWithScores, len(matches))
	fmt.Printf("  Matches with pageURL: %d/%d\n", matchWithPageURL, len(matches))

	if matchWithScores == 0 && len(matches) > 0 {
		fmt.Println("FAIL: No matches have scores populated")
		os.Exit(1)
	}
	if matchWithPageURL == 0 && len(matches) > 0 {
		fmt.Println("FAIL: No matches have pageURL populated")
		os.Exit(1)
	}
	fmt.Println("PASS: Scores and pageURLs are populated")

	// Test 2: Fetch match details via page-based fetching
	if testMatchID == 0 {
		fmt.Println("\nSKIP: No finished match found for details test")
		os.Exit(0)
	}

	fmt.Printf("\n--- Test 2: Fetch match details (ID: %d) ---\n", testMatchID)
	details, err := client.MatchDetails(ctx, testMatchID)
	if err != nil {
		fmt.Printf("FAIL: MatchDetails: %v\n", err)
		os.Exit(1)
	}
	if details == nil {
		fmt.Println("FAIL: MatchDetails returned nil")
		os.Exit(1)
	}

	fmt.Printf("  %s vs %s\n", details.HomeTeam.Name, details.AwayTeam.Name)
	if details.HomeScore != nil && details.AwayScore != nil {
		fmt.Printf("  Score: %d - %d\n", *details.HomeScore, *details.AwayScore)
	}
	fmt.Printf("  Status: %s\n", details.Status)
	fmt.Printf("  Events: %d\n", len(details.Events))
	fmt.Printf("  Statistics: %d\n", len(details.Statistics))
	fmt.Printf("  Venue: %s\n", details.Venue)
	if details.HomeFormation != "" {
		fmt.Printf("  Formations: %s vs %s\n", details.HomeFormation, details.AwayFormation)
	}
	if details.HomeXG != nil {
		fmt.Printf("  xG: %.2f - %.2f\n", *details.HomeXG, *details.AwayXG)
	}
	fmt.Printf("  Home starting: %d players\n", len(details.HomeStarting))
	fmt.Printf("  Away starting: %d players\n", len(details.AwayStarting))

	fmt.Println("PASS: Match details fetched successfully via page-based method")
	fmt.Println("\n=== All tests passed ===")
}
