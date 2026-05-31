package ui

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// OSC 8 hyperlink escape sequences for terminal hyperlinks.
// Supported in: iTerm2, GNOME Terminal, Windows Terminal, Kitty, Alacritty, etc.
// Format: \033]8;;URL\033\\TEXT\033]8;;\033\\
// See: https://gist.github.com/egmontkob/eb114294efbcd5adb1944c9f3cb5feda

const (
	// oscStart begins the hyperlink sequence
	oscStart = "\033]8;;"
	// oscEnd ends the URL portion
	oscEnd = "\033\\"
)

// Hyperlink creates a terminal hyperlink using OSC 8 escape sequences.
// Falls back to plain text with URL suffix if terminal doesn't support OSC 8.
func Hyperlink(text, url string) string {
	if url == "" {
		return text
	}

	if supportsHyperlinks() {
		return fmt.Sprintf("%s%s%s%s%s%s", oscStart, url, oscEnd, text, oscStart, oscEnd)
	}

	// Fallback: render URL as visible text so the terminal's native URL scanner
	// (e.g. Terminal.app right-click → "Open URL") can detect and open it.
	return text + "  [" + url + "]"
}

// HyperlinkWithFallback creates a hyperlink with a visible fallback indicator.
// If OSC 8 is supported, returns a clickable link.
// Otherwise, returns text with a link indicator like [📹].
func HyperlinkWithFallback(text, url, fallbackIndicator string) string {
	if url == "" {
		return text
	}

	if supportsHyperlinks() {
		return Hyperlink(text, url)
	}

	// Fallback: append indicator that user can act on
	if fallbackIndicator != "" {
		return text + " " + fallbackIndicator
	}
	return text
}

// CreateGoalLinkDisplay creates a display string for a goal with replay link.
// Returns the text with hyperlink if available, or plain text if not.
// If the terminal doesn't support hyperlinks OR no URL is provided,
// returns the original goalText unchanged (no visible difference).
func CreateGoalLinkDisplay(goalText, replayURL string) string {
	// Validate URL using helper function
	if !IsValidReplayURL(replayURL) {
		return goalText
	}

	// Only show indicator if terminal supports clickable hyperlinks
	// Otherwise, return unchanged text (no visible change to user)
	if supportsHyperlinks() {
		// Create a clickable indicator
		indicator := ReplayLinkIndicator
		linkedIndicator := Hyperlink(indicator, replayURL)
		if goalText == "" {
			return linkedIndicator
		}
		return goalText + " " + linkedIndicator
	}

	// Fallback: render URL as visible text so the terminal's native URL scanner
	// can detect and offer right-click → "Open URL".
	if goalText == "" {
		return "▶ [" + replayURL + "]"
	}
	return goalText + " ▶ [" + replayURL + "]"
}

// supportsHyperlinks detects if the terminal likely supports OSC 8 hyperlinks.
// This is a best-effort detection based on common terminal identifiers.
func supportsHyperlinks() bool {
	// Check for specific terminal emulators known to support OSC 8
	term := os.Getenv("TERM")
	termProgram := os.Getenv("TERM_PROGRAM")
	wtSession := os.Getenv("WT_SESSION") // Windows Terminal
	kitty := os.Getenv("KITTY_WINDOW_ID")

	// Known supporting terminals
	supportingTerms := []string{
		"xterm-256color",
		"xterm-kitty",
		"alacritty",
	}

	// Explicitly exclude terminals known not to support OSC 8.
	// This must run before the supportingTerms loop because Apple_Terminal sets
	// TERM=xterm-256color which would otherwise match and return true.
	if termProgram == "Apple_Terminal" {
		return false
	}

	supportingPrograms := []string{
		"iTerm.app",
		"vscode",
		"Hyper",
		"WezTerm",
	}

	// Check TERM
	for _, t := range supportingTerms {
		if strings.Contains(term, t) {
			return true
		}
	}

	// Check TERM_PROGRAM
	for _, p := range supportingPrograms {
		if strings.Contains(termProgram, p) {
			return true
		}
	}

	// Windows Terminal
	if wtSession != "" {
		return true
	}

	// Kitty
	if kitty != "" {
		return true
	}

	// Default to true for modern terminals, as most support it now
	// Only return false for very basic terminals
	if term == "dumb" || term == "" {
		return false
	}

	return true
}

// OpenURL opens a URL in the default browser.
// Use this as a fallback when OSC 8 hyperlinks aren't supported.
func OpenURL(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return cmd.Start()
}

// ReplayLinkIndicator is the visual indicator for replay links.
const ReplayLinkIndicator = "[▶REPLAY]"

// ReplayLinkIndicatorAlt is an alternative ASCII indicator for terminals without emoji.
const ReplayLinkIndicatorAlt = "[replay]"

// IsValidReplayURL validates that a URL is a valid HTTP/HTTPS URL and not a marker.
// Returns true only for valid http:// or https:// URLs.
// Filters out empty strings, "__NOT_FOUND__" markers, and invalid URL schemes.
func IsValidReplayURL(url string) bool {
	if url == "" || url == "__NOT_FOUND__" {
		return false
	}
	return strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")
}
