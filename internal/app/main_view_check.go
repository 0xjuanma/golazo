package app

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// mainViewCheckMsg is sent after the check delay completes.
type mainViewCheckMsg struct {
	selection int // 0 for Stats, 1 for Live Matches
}

// performMainViewCheck performs a 3-second delay check before navigating.
func performMainViewCheck(selection int) tea.Cmd {
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return mainViewCheckMsg{selection: selection}
	})
}

