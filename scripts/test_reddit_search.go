package main

import (
	"fmt"
	"time"

	"github.com/0xjuanma/golazo/internal/reddit"
)

func main() {
	// Create Reddit client
	client, err := reddit.NewClient()
	if err != nil {
		fmt.Printf("Error creating Reddit client: %v\n", err)
		return
	}

	// Test the Everton vs Brentford goal
	goal := reddit.GoalInfo{
		MatchID:    0, // We don't know the match ID yet
		HomeTeam:   "Everton",
		AwayTeam:   "Brentford",
		ScorerName: "Igor Thiago",
		Minute:     11,
		HomeScore:  0,
		AwayScore:  1,
		IsHomeTeam: false, // Igor Thiago scored for Brentford (away team)
		MatchTime:  time.Now().Add(-24 * time.Hour), // Assume yesterday
	}

	fmt.Printf("Searching for goal: %+v\n", goal)

	// Test the search
	link, err := client.GoalLink(goal)
	if err != nil {
		fmt.Printf("Error searching for goal: %v\n", err)
		return
	}

	if link != nil {
		fmt.Printf("Found link: %+v\n", link)
	} else {
		fmt.Println("No link found")
	}
}
