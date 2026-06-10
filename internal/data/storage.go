// Package data provides utilities for loading mock football match data.
package data

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	configDir = ".golazo"
)

// ConfigDir returns the path to the golazo config directory.
// On Linux, follows XDG Base Directory spec (~/.config/golazo).
// On other systems (macOS, Windows), uses ~/.golazo.
func ConfigDir() (string, error) {
	var configPath string

	if runtime.GOOS == "linux" {
		if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
			configPath = filepath.Join(xdgConfig, "golazo")
		} else {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("get home directory: %w", err)
			}
			configPath = filepath.Join(homeDir, ".config", "golazo")
		}
	} else {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("get home directory: %w", err)
		}
		configPath = filepath.Join(homeDir, configDir)
	}

	if err := os.MkdirAll(configPath, 0755); err != nil {
		return "", fmt.Errorf("create config directory: %w", err)
	}

	return configPath, nil
}

// DebugLogPath returns the user-facing path to the debug log file, with the
// home directory represented as "~" for display purposes. It is platform-aware
// and mirrors the location used by ConfigDir, but performs no filesystem I/O
// and does not create any directories. Safe to call from init() / flag help.
func DebugLogPath() string {
	const logFile = "golazo_debug.log"

	if runtime.GOOS == "linux" {
		if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
			return filepath.Join(xdg, "golazo", logFile)
		}
		return filepath.Join("~", ".config", "golazo", logFile)
	}

	return filepath.Join("~", configDir, logFile)
}

// CacheDir returns the path to the golazo cache directory.
// Uses os.UserCacheDir() which returns:
//   - Linux: ~/.cache/golazo (or $XDG_CACHE_HOME/golazo)
//   - macOS: ~/Library/Caches/golazo
//   - Windows: %LocalAppData%/golazo
func CacheDir() (string, error) {
	userCache, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("get user cache directory: %w", err)
	}

	cachePath := filepath.Join(userCache, "golazo")
	if err := os.MkdirAll(cachePath, 0755); err != nil {
		return "", fmt.Errorf("create cache directory: %w", err)
	}

	return cachePath, nil
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

// LoadLatestVersion reads the latest known version from storage.
// Returns empty string if file doesn't exist or can't be read.
func LoadLatestVersion() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}

	versionFile := filepath.Join(dir, "latest_version.txt")
	data, err := os.ReadFile(versionFile)
	if err != nil {
		return "", nil // Return empty if file doesn't exist
	}

	return strings.TrimSpace(string(data)), nil
}

// SaveLatestVersion saves the latest version to storage.
func SaveLatestVersion(version string) error {
	dir, err := ConfigDir()
	if err != nil {
		return err
	}

	versionFile := filepath.Join(dir, "latest_version.txt")
	return os.WriteFile(versionFile, []byte(strings.TrimSpace(version)), 0644)
}

// CheckLatestVersion fetches the latest version from GitHub releases.
// Uses GitHub's redirect URL which is simpler than the API.
// Returns the version tag (e.g., "v1.2.3").
func CheckLatestVersion() (string, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get("https://github.com/0xjuanma/golazo/releases/latest")
	if err != nil {
		return "", fmt.Errorf("fetch latest release: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// GitHub redirects to: https://github.com/0xjuanma/golazo/releases/tag/v1.2.3
	// Extract version from the final URL
	finalURL := resp.Request.URL.String()

	// Look for "/releases/tag/" in the URL
	if idx := strings.LastIndex(finalURL, "/releases/tag/"); idx != -1 {
		version := finalURL[idx+len("/releases/tag/"):]
		if version != "" {
			return version, nil
		}
	}

	// Fallback: try to read from response body (in case redirect doesn't work)
	body, err := io.ReadAll(resp.Body)
	if err == nil && len(body) > 0 {
		// This is a fallback and might not work, but better than nothing
		bodyStr := string(body)
		if idx := strings.Index(bodyStr, "/releases/tag/"); idx != -1 {
			start := idx + len("/releases/tag/")
			end := strings.Index(bodyStr[start:], "\"")
			if end != -1 {
				version := bodyStr[start : start+end]
				if version != "" {
					return version, nil
				}
			}
		}
	}

	return "", fmt.Errorf("could not extract version from GitHub response")
}

// ShouldCheckVersion returns true if we should check for a new version.
// Checks if the latest_version.txt file is older than 24 hours.
func ShouldCheckVersion() bool {
	dir, err := ConfigDir()
	if err != nil {
		return false
	}

	versionFile := filepath.Join(dir, "latest_version.txt")
	info, err := os.Stat(versionFile)
	if err != nil {
		return true // File doesn't exist, should check
	}

	// Check if file is older than 24 hours
	return time.Since(info.ModTime()) > 24*time.Hour
}
