package data

import (
	"encoding/json"
	"os"

	"github.com/0xjuanma/golazo/internal/api"
)

// MockMatchesData contains multiple matches for testing.
// Uses api.Match directly to ensure single source of truth.
type MockMatchesData struct {
	Matches []api.Match `json:"matches"`
}

// MockMatches returns multiple mock matches for testing.
// This is a legacy function that combines live and finished matches.
func MockMatches() ([]api.Match, error) {
	// Try to load from config directory first
	configPath, err := MockDataPath()
	if err == nil {
		if data, err := os.ReadFile(configPath); err == nil {
			var mockData MockMatchesData
			if err := json.Unmarshal(data, &mockData); err == nil {
				return mockData.Matches, nil
			}
		}
	}

	// Fallback to combining live and finished matches
	return getDefaultMockMatches(), nil
}

// getDefaultMockMatches returns default mock matches by combining live and finished.
func getDefaultMockMatches() []api.Match {
	// Combine live and finished matches for a complete dataset
	matches := []api.Match{}
	matches = append(matches, MockLiveMatches()...)
	matches = append(matches, MockFinishedMatches()...)
	return matches
}

func intPtr(i int) *int {
	return &i
}
