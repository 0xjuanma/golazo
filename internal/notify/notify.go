// Package notify provides desktop notification functionality for match events.
// Currently supports macOS, Linux, and Windows via the beeep library.
package notify

import (
	"fmt"
	"os"

	"github.com/0xjuanma/golazo/internal/api"
	"github.com/0xjuanma/golazo/internal/constants"
	"github.com/gen2brain/beeep"
)

// Notifier defines the interface for sending desktop notifications.
// This allows for easy mocking in tests and potential future implementations.
type Notifier interface {
	// Goal sends a notification for a new goal event.
	Goal(event api.MatchEvent, homeTeam, awayTeam api.Team, homeScore, awayScore int) error
}

// DesktopNotifier implements Notifier using native desktop notifications.
type DesktopNotifier struct {
	enabled bool
}

// NewDesktopNotifier creates a new desktop notifier.
// Notifications are enabled by default.
func NewDesktopNotifier() *DesktopNotifier {
	return &DesktopNotifier{
		enabled: true,
	}
}

// SetEnabled enables or disables notifications.
func (n *DesktopNotifier) SetEnabled(enabled bool) {
	n.enabled = enabled
}

// Enabled returns whether notifications are currently enabled.
func (n *DesktopNotifier) Enabled() bool {
	return n.enabled
}

// Goal sends a desktop notification for a new goal event.
// Includes scorer name, minute, team, and current score.
// Always plays a terminal beep as a fallback notification.
func (n *DesktopNotifier) Goal(event api.MatchEvent, homeTeam, awayTeam api.Team, homeScore, awayScore int) error {
	if !n.enabled {
		return nil
	}

	// Play terminal beep via stderr (bypasses bubbletea's stdout capture)
	// This works even when the TUI is active
	_, _ = os.Stderr.WriteString("\a")

	// Build notification content
	title := constants.NotificationTitleGoal
	message := formatGoalMessage(event, homeTeam, awayTeam, homeScore, awayScore)

	// Send notification via beeep (cross-platform)
	// Errors are ignored - OS notification is best-effort, beep already played
	_ = beeep.Notify(title, message, "")

	return nil
}

// formatGoalMessage creates the notification message for a goal.
// Format: "Scorer (Team) 34' | Home 2-1 Away"
func formatGoalMessage(event api.MatchEvent, homeTeam, awayTeam api.Team, homeScore, awayScore int) string {
	scorer := "Unknown"
	if event.Player != nil {
		scorer = *event.Player
	}

	// Determine which team scored
	teamName := event.Team.ShortName
	if teamName == "" {
		teamName = event.Team.Name
	}

	// Build message with assist if available
	assistText := ""
	if event.Assist != nil && *event.Assist != "" {
		assistText = fmt.Sprintf(" (%s)", *event.Assist)
	}

	return fmt.Sprintf("%s%s %d' [%s]\n%s %d - %d %s",
		scorer,
		assistText,
		event.Minute,
		teamName,
		homeTeam.ShortName,
		homeScore,
		awayScore,
		awayTeam.ShortName,
	)
}
