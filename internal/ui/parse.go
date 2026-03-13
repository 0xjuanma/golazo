package ui

import (
	"strconv"
	"strings"
)

// parseStatValue extracts a numeric value from a stat string.
// Handles formats like "45%", "23 (45%)", plain numbers, etc.
func parseStatValue(s string) float64 {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, "%")

	// Handle formats like "23 (45%)" - take first number
	if idx := strings.Index(s, " "); idx > 0 {
		s = s[:idx]
	}
	if idx := strings.Index(s, "("); idx > 0 {
		s = s[:idx]
	}
	s = strings.TrimSpace(s)

	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return val
}
