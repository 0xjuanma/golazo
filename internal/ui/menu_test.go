package ui

import (
	"strings"
	"testing"

	"github.com/0xjuanma/golazo/internal/constants"
	"github.com/charmbracelet/lipgloss"
)

// TestRenderMenuHelp_NarrowWidthDoesNotOverflow verifies that the main-menu
// nav-tip/help footer never renders wider than the available terminal width.
// At narrow widths (small terminals / large fonts) the long help string must be
// truncated to a single line rather than wrapped at arbitrary points. See issue
// #124 (Navigation tips wrap unproperly with different font sizes).
func TestRenderMenuHelp_NarrowWidthDoesNotOverflow(t *testing.T) {
	// HelpMainMenu is long enough to overflow a narrow terminal.
	widths := []int{10, 20, 30, 40}
	for _, width := range widths {
		got := renderMenuHelp(constants.HelpMainMenu, width)
		for _, line := range strings.Split(got, "\n") {
			if w := lipgloss.Width(line); w > width {
				t.Errorf("renderMenuHelp(%q, %d): line %q has width %d > %d (overflow)",
					constants.HelpMainMenu, width, line, w, width)
			}
		}
	}
}

// TestRenderMenuHelp_LongStringIsTruncatedToOneLine verifies that a help string
// longer than the width collapses to a single truncated line (with an ellipsis)
// instead of wrapping into multiple lines.
func TestRenderMenuHelp_LongStringIsTruncatedToOneLine(t *testing.T) {
	const width = 20
	got := renderMenuHelp(constants.HelpMainMenu, width)
	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	if len(lines) != 1 {
		t.Errorf("renderMenuHelp at width %d produced %d lines, want 1 (no wrap):\n%q",
			width, len(lines), got)
	}
	if !strings.Contains(got, "...") {
		t.Errorf("renderMenuHelp at width %d should truncate with ellipsis, got %q", width, got)
	}
}

// TestRenderMenuHelp_WideWidthKeepsFullText verifies that when the terminal is
// wide enough, the full help string is preserved (no truncation).
func TestRenderMenuHelp_WideWidthKeepsFullText(t *testing.T) {
	got := renderMenuHelp(constants.HelpMainMenu, 120)
	if strings.Contains(got, "...") {
		t.Errorf("renderMenuHelp at wide width should not truncate, got %q", got)
	}
	if !strings.Contains(got, "navigate") {
		t.Errorf("renderMenuHelp at wide width should contain full text, got %q", got)
	}
}
