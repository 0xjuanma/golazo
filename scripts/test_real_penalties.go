package main

import (
	"context"
	"fmt"

	"github.com/0xjuanma/golazo/internal/fotmob"
)

func main() {
	fmt.Println("Testing real penalty parsing...")

	client := fotmob.NewClient()
	ctx := context.Background()

	// Test with match ID 5073476 which has penalties
	details, err := client.MatchDetails(ctx, 5073476)
	if err != nil {
		fmt.Printf("Error fetching match details: %v\n", err)
		return
	}

	if details == nil {
		fmt.Println("No match details found")
		return
	}

	fmt.Printf("Match: %s vs %s\n", details.HomeTeam.Name, details.AwayTeam.Name)
	fmt.Printf("Score: %d - %d\n", *details.HomeScore, *details.AwayScore)

	if details.Penalties != nil && details.Penalties.Home != nil && details.Penalties.Away != nil {
		fmt.Printf("✅ Penalties found: %d - %d\n", *details.Penalties.Home, *details.Penalties.Away)
	} else {
		fmt.Println("❌ No penalties found")
	}
}
