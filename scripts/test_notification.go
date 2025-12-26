// test_notification.go - Test script to verify notifications work
//
// Usage:
//   go run scripts/test_notification.go

package main

import (
	"fmt"
	"time"

	"github.com/gen2brain/beeep"
)

func main() {
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║              Golazo Notification Test                        ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Test 1: Terminal beep
	fmt.Println("1. Testing terminal beep...")
	fmt.Print("\a") // Terminal bell
	fmt.Println("   ✓ Beep sent! (You should have heard a sound)")
	fmt.Println()

	time.Sleep(500 * time.Millisecond)

	// Test 2: OS notification
	fmt.Println("2. Testing OS notification via beeep...")
	err := beeep.Notify(
		"⚽ GOAL! (Test)",
		"Salah 34' [LIV]\nLiverpool 2 - 1 Man City",
		"",
	)

	if err != nil {
		fmt.Printf("   ❌ Error: %v\n", err)
	} else {
		fmt.Println("   ✓ Notification sent!")
	}
	fmt.Println()

	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("RESULTS:")
	fmt.Println("  • Beep: Should always work (check your volume)")
	fmt.Println("  • OS Notification: Requires macOS notification permissions")
	fmt.Println()
	fmt.Println("To enable OS notifications on macOS:")
	fmt.Println("  1. Open System Settings → Notifications")
	fmt.Println("  2. Find your terminal app (Terminal, iTerm2, etc.)")
	fmt.Println("  3. Enable 'Allow Notifications'")
	fmt.Println("  4. Set alert style to 'Banners' or 'Alerts'")
	fmt.Println("═══════════════════════════════════════════════════════════════")
}
