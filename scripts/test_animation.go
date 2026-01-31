//go:build ignore
// +build ignore

// Test script to demonstrate all logo animation styles.
// Run with: go run scripts/test_animation.go
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/0xjuanma/golazo/internal/ui/logo"
)

const (
	// Animation duration in milliseconds
	durationMs = 600
	// Tick interval to simulate the app's tick rate
	tickInterval = 70 * time.Millisecond
	// Pause between animations
	pauseBetween = 2 * time.Second
)

func main() {
	// Clear screen
	fmt.Print("\033[2J\033[H")

	fmt.Println("╔════════════════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    GOLAZO Logo Animation Test                                  ║")
	fmt.Println("║                                                                                ║")
	fmt.Println("║  This script demonstrates all available animation styles.                      ║")
	fmt.Println("║  Each animation will play once, then move to the next.                         ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	time.Sleep(2 * time.Second)

	// Test all animation types
	for _, animType := range logo.AllAnimationTypes() {
		demonstrateAnimation(animType)
		time.Sleep(pauseBetween)
	}

	fmt.Println()
	fmt.Println("╔════════════════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                         All animations complete!                               ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════════════════════╝")
}

func demonstrateAnimation(animType logo.AnimationType) {
	// Clear screen
	fmt.Print("\033[2J\033[H")

	// Show animation name
	fmt.Printf("╔════════════════════════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║  Animation: %-67s ║\n", animType.String())
	fmt.Printf("╚════════════════════════════════════════════════════════════════════════════════╝\n")
	fmt.Println()

	// Create animated logo with this animation type
	opts := logo.DefaultOpts()
	opts.Width = 80
	anim := logo.NewAnimatedLogoWithType("v0.14.0", false, opts, durationMs, 1, animType)

	// Animate until complete
	startTime := time.Now()
	for !anim.IsComplete() {
		// Clear the logo area (move cursor up and clear)
		fmt.Print("\033[5;0H") // Move to row 5, column 0

		// Render current frame
		fmt.Println(anim.View())

		// Tick the animation
		anim.Tick()

		// Wait for next tick
		time.Sleep(tickInterval)

		// Safety timeout
		if time.Since(startTime) > 5*time.Second {
			fmt.Fprintln(os.Stderr, "Animation timeout!")
			break
		}
	}

	// Show final frame
	fmt.Print("\033[5;0H")
	fmt.Println(anim.View())

	fmt.Println()
	fmt.Printf("  Animation '%s' complete in %v\n", animType.String(), time.Since(startTime).Round(time.Millisecond))
}
