package app

import (
	"context"
	"time"

	"github.com/0xjuanma/golazo/internal/data"
	"github.com/0xjuanma/golazo/internal/fotmob"
	tea "github.com/charmbracelet/bubbletea"
)

// fetchWorldCupMockData returns the hardcoded Qatar 2022 World Cup data immediately.
func fetchWorldCupMockData() tea.Cmd {
	return func() tea.Msg {
		return wcDataMsg{data: data.MockWorldCupData()}
	}
}

// fetchWorldCupData fetches live World Cup data from FotMob.
// Uses the current/latest season (2026).
func fetchWorldCupData(parentCtx context.Context, client *fotmob.Client) tea.Cmd {
	return func() tea.Msg {
		if client == nil {
			return wcDataMsg{data: data.MockWorldCupData()}
		}

		ctx, cancel := context.WithTimeout(parentCtx, 20*time.Second)
		defer cancel()

		wcData, err := client.WorldCupData(ctx, "")
		if err != nil {
			return wcDataMsg{err: err}
		}
		return wcDataMsg{data: wcData}
	}
}
