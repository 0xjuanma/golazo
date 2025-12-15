// Package data provides utilities for loading mock football match data.
package data

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	configDir = ".golazo"
)

// ConfigDir returns the path to the golazo config directory.
func ConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home directory: %w", err)
	}

	configPath := filepath.Join(homeDir, configDir)
	if err := os.MkdirAll(configPath, 0755); err != nil {
		return "", fmt.Errorf("create config directory: %w", err)
	}

	return configPath, nil
}

// MockDataPath returns the path to the mock data file.
func MockDataPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "matches.json"), nil
}

// LiveUpdate represents a single live update string.
type LiveUpdate struct {
	MatchID int
	Update  string
	Time    time.Time
}

// SaveLiveUpdate appends a live update to the storage.
func SaveLiveUpdate(matchID int, update string) error {
	dir, err := ConfigDir()
	if err != nil {
		return err
	}

	updatesFile := filepath.Join(dir, fmt.Sprintf("updates_%d.json", matchID))

	var updates []LiveUpdate
	if data, err := os.ReadFile(updatesFile); err == nil {
		// Best effort to load existing updates; if unmarshal fails, start with empty slice
		if err := json.Unmarshal(data, &updates); err != nil {
			// Invalid JSON in file - start fresh with empty slice
			updates = []LiveUpdate{}
		}
	}

	updates = append(updates, LiveUpdate{
		MatchID: matchID,
		Update:  update,
		Time:    time.Now(),
	})

	data, err := json.Marshal(updates)
	if err != nil {
		return fmt.Errorf("marshal updates: %w", err)
	}

	return os.WriteFile(updatesFile, data, 0644)
}

// LiveUpdates retrieves live updates for a match.
func LiveUpdates(matchID int) ([]string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return nil, err
	}

	updatesFile := filepath.Join(dir, fmt.Sprintf("updates_%d.json", matchID))
	data, err := os.ReadFile(updatesFile)
	if err != nil {
		return []string{}, nil // Return empty if file doesn't exist
	}

	var updates []LiveUpdate
	if err := json.Unmarshal(data, &updates); err != nil {
		return nil, fmt.Errorf("unmarshal updates: %w", err)
	}

	result := make([]string, 0, len(updates))
	for _, update := range updates {
		result = append(result, update.Update)
	}

	return result, nil
}
