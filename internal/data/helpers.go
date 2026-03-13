package data

import "time"

// intPtr returns a pointer to the given int value.
func intPtr(i int) *int {
	return &i
}

// stringPtr returns a pointer to the given string value.
func stringPtr(s string) *string {
	return &s
}

// timePtr returns a pointer to the given time.Time value.
func timePtr(t time.Time) *time.Time {
	return &t
}
