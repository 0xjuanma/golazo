package worldcup

import (
	"strings"
	"testing"

	"github.com/0xjuanma/golazo/internal/api"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

func TestTeamLabel(t *testing.T) {
	argFlag := FlagEmoji("ARG")
	nedFlag := FlagEmoji("NED")
	ausFlag := FlagEmoji("AUS")
	korFlag := FlagEmoji("KOR")
	rsaFlag := FlagEmoji("RSA")
	if argFlag == "" || nedFlag == "" || ausFlag == "" || korFlag == "" || rsaFlag == "" {
		t.Fatal("expected ARG/NED/AUS/KOR/RSA to have flag emojis registered")
	}

	tests := []struct {
		name string
		team api.Team
		want string
	}{
		{
			name: "short name present takes precedence",
			team: api.Team{Name: "Argentina", ShortName: "ARG"},
			want: argFlag + " ARG",
		},
		{
			name: "short name empty falls back to name override map",
			team: api.Team{Name: "Netherlands", ShortName: ""},
			want: nedFlag + " NED",
		},
		{
			name: "short name empty + unknown name truncates to 3 letters",
			team: api.Team{Name: "Nowhereland", ShortName: ""},
			want: "   NOW",
		},
		{
			name: "short name lowercase is normalized",
			team: api.Team{Name: "Argentina", ShortName: "arg"},
			want: argFlag + " ARG",
		},
		{
			name: "unknown short code keeps the code but no flag",
			team: api.Team{Name: "Nowhereland", ShortName: "ZZZ"},
			want: "   ZZZ",
		},
		{
			name: "short name with whitespace is trimmed",
			team: api.Team{Name: "Argentina", ShortName: "  ARG  "},
			want: argFlag + " ARG",
		},
		{
			name: "name override is case-insensitive",
			team: api.Team{Name: "NETHERLANDS"},
			want: nedFlag + " NED",
		},
		{
			name: "short name longer than 3 chars is truncated",
			team: api.Team{Name: "Australia", ShortName: "AUST"},
			want: ausFlag + " AUS", // "AUST" → "AUS" → flag resolves
		},
		{
			name: "ambiguous shortname without flag falls back to name override (KOR)",
			team: api.Team{Name: "South Korea", ShortName: "SOU"},
			want: korFlag + " KOR",
		},
		{
			name: "ambiguous shortname without flag falls back to name override (RSA)",
			team: api.Team{Name: "South Africa", ShortName: "SOU"},
			want: rsaFlag + " RSA",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TeamLabel(tt.team)
			if !strings.HasPrefix(got, tt.want) {
				t.Errorf("TeamLabel(%+v) = %q, want prefix %q", tt.team, got, tt.want)
			}
			if !strings.HasPrefix(strings.TrimRight(got, " "), tt.want) &&
				strings.TrimRight(got, " ") != tt.want {
				t.Errorf("TeamLabel(%+v) trimmed = %q, want %q",
					tt.team, strings.TrimRight(got, " "), tt.want)
			}
		})
	}
}

func TestMatchupTeamLabel(t *testing.T) {
	argFlag := FlagEmoji("ARG")
	nedFlag := FlagEmoji("NED")

	tests := []struct {
		name  string
		short string
		full  string
		tbd   bool
		want  string
	}{
		{name: "tbd returns TBD", tbd: true, want: "TBD"},
		{name: "tbd takes precedence over short", short: "ARG", full: "Argentina", tbd: true, want: "TBD"},
		{name: "empty short and full returns TBD", want: "TBD"},
		{name: "short present", short: "ARG", full: "Argentina", want: argFlag + " ARG"},
		{name: "short empty, name in override", full: "Netherlands", want: nedFlag + " NED"},
		{name: "short empty, unknown name truncates", full: "Nowhereland", want: "   NOW"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MatchupTeamLabel(tt.short, tt.full, tt.tbd)
			trimmed := strings.TrimRight(got, " ")
			if trimmed != tt.want && !strings.HasPrefix(got, tt.want) {
				t.Errorf("MatchupTeamLabel(%q, %q, %v) = %q, want %q (or padded)",
					tt.short, tt.full, tt.tbd, got, tt.want)
			}
		})
	}
}

func TestTeamLabel_AlwaysContainsCode(t *testing.T) {
	// Every label must include the resolved code so callers don't need to
	// re-derive it (e.g. for column-width estimation).
	cases := []api.Team{
		{Name: "Argentina", ShortName: "ARG"},
		{Name: "Netherlands"},
		{Name: "Nowhereland"},
	}
	for _, c := range cases {
		label := TeamLabel(c)
		code := teamCode(c.ShortName, c.Name)
		if !strings.Contains(label, code) {
			t.Errorf("TeamLabel(%+v) = %q must contain code %q", c, label, code)
		}
	}
}

// TestTeamLabel_WC2026Qualifiers asserts a representative sample of teams
// added for the WC 2026 qualifying cycle resolves to "<flag> <CODE>". This
// guards against either the flagEmojis or wcNameToCode entries being
// silently removed.
func TestTeamLabel_WC2026Qualifiers(t *testing.T) {
	cases := []struct {
		name string
		code string
	}{
		{"Uzbekistan", "UZB"},
		{"Cape Verde", "CPV"},
		{"Curaçao", "CUW"},
		{"Curacao", "CUW"},
		{"Haiti", "HAI"},
		{"Suriname", "SUR"},
		{"New Caledonia", "NCL"},
		{"Dominican Republic", "DOM"},
		{"Guatemala", "GUA"},
		{"El Salvador", "SLV"},
		{"North Korea", "PRK"},
		{"Burkina Faso", "BFA"},
		{"Madagascar", "MAD"},
		{"Kazakhstan", "KAZ"},
		{"Luxembourg", "LUX"},
		{"Israel", "ISR"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			flag := FlagEmoji(tc.code)
			if flag == "" {
				t.Fatalf("missing flagEmojis entry for %s (%s)", tc.name, tc.code)
			}
			got := TeamLabel(api.Team{Name: tc.name})
			wantPrefix := flag + " " + tc.code
			if !strings.HasPrefix(got, wantPrefix) {
				t.Errorf("TeamLabel({Name:%q}) = %q, want prefix %q", tc.name, got, wantPrefix)
			}
		})
	}
}

// TestLabelWidthInvariant locks in the fix for #158: every team label must
// occupy the same visual width regardless of which resolution branch
// produced it. Without padding, with-flag labels measure 5 cells under
// runewidth's tables while placeholder labels measure 6, which caused
// per-row drift in terminals following the legacy width table after the
// v0.27.0 flag/name-override backfill.
func TestLabelWidthInvariant(t *testing.T) {
	probes := []api.Team{
		// Registered flag (regional indicator pair)
		{Name: "Mexico", ShortName: "MEX"},
		{Name: "Brazil", ShortName: "BRA"},
		{Name: "South Korea", ShortName: "SOU"}, // exercises override fallback
		// Registered flag (tag-sequence subdivision)
		{Name: "England", ShortName: "ENG"},
		{Name: "Wales", ShortName: "WAL"},
		{Name: "Scotland", ShortName: "SCO"},
		// No registered flag → placeholder branch
		{Name: "Nowhereland", ShortName: "ZZZ"},
		{Name: "Nowhereland", ShortName: ""},
	}

	for _, team := range probes {
		t.Run(team.Name+"/"+team.ShortName, func(t *testing.T) {
			label := TeamLabel(team)
			if w := runewidth.StringWidth(label); w != labelTargetWidth {
				t.Errorf("runewidth.StringWidth(%q) = %d, want %d",
					label, w, labelTargetWidth)
			}
			if w := lipgloss.Width(label); w < labelTargetWidth {
				t.Errorf("lipgloss.Width(%q) = %d, want >= %d",
					label, w, labelTargetWidth)
			}
		})
	}
}
