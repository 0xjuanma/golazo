package ui

import (
	"strings"
	"testing"
)

func TestSupportsHyperlinks(t *testing.T) {
	tests := []struct {
		name        string
		term        string
		termProgram string
		wtSession   string
		kitty       string
		want        bool
	}{
		{
			name:        "Apple_Terminal is excluded even with xterm-256color TERM",
			term:        "xterm-256color",
			termProgram: "Apple_Terminal",
			want:        false,
		},
		{
			name:        "iTerm2 is supported",
			termProgram: "iTerm.app",
			want:        true,
		},
		{
			name:        "xterm-256color with no TERM_PROGRAM is supported",
			term:        "xterm-256color",
			termProgram: "",
			want:        true,
		},
		{
			name: "dumb terminal is not supported",
			term: "dumb",
			want: false,
		},
		{
			name: "empty TERM is not supported",
			term: "",
			want: false,
		},
		{
			name:      "Windows Terminal (WT_SESSION) is supported",
			wtSession: "some-session-id",
			want:      true,
		},
		{
			name:  "Kitty is supported",
			kitty: "1",
			want:  true,
		},
		{
			name:        "WezTerm is supported",
			termProgram: "WezTerm",
			want:        true,
		},
		{
			name:        "vscode is supported",
			termProgram: "vscode",
			want:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("TERM", tt.term)
			t.Setenv("TERM_PROGRAM", tt.termProgram)
			t.Setenv("WT_SESSION", tt.wtSession)
			t.Setenv("KITTY_WINDOW_ID", tt.kitty)

			got := supportsHyperlinks()
			if got != tt.want {
				t.Errorf("supportsHyperlinks() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHyperlinkFallbackContainsURL(t *testing.T) {
	t.Setenv("TERM", "dumb")
	t.Setenv("TERM_PROGRAM", "")
	t.Setenv("WT_SESSION", "")
	t.Setenv("KITTY_WINDOW_ID", "")

	result := Hyperlink("link text", "https://example.com")
	if !strings.Contains(result, "https://example.com") {
		t.Errorf("Hyperlink() fallback = %q, want it to contain the URL", result)
	}
}

func TestCreateGoalLinkDisplayFallbackContainsURL(t *testing.T) {
	t.Setenv("TERM", "dumb")
	t.Setenv("TERM_PROGRAM", "")
	t.Setenv("WT_SESSION", "")
	t.Setenv("KITTY_WINDOW_ID", "")

	url := "https://reddit.com/r/soccer/comments/abc"

	t.Run("empty goalText", func(t *testing.T) {
		result := CreateGoalLinkDisplay("", url)
		if !strings.Contains(result, url) {
			t.Errorf("CreateGoalLinkDisplay() fallback = %q, want it to contain the URL", result)
		}
	})

	t.Run("with goalText", func(t *testing.T) {
		result := CreateGoalLinkDisplay("Messi 45'", url)
		if !strings.Contains(result, url) {
			t.Errorf("CreateGoalLinkDisplay() fallback = %q, want it to contain the URL", result)
		}
		if !strings.Contains(result, "Messi 45'") {
			t.Errorf("CreateGoalLinkDisplay() fallback = %q, want it to contain the goalText", result)
		}
	})
}
