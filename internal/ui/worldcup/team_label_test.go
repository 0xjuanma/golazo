package worldcup

import (
	"strings"
	"testing"

	"github.com/0xjuanma/golazo/internal/api"
)

func TestTeamLabel(t *testing.T) {
	argFlag := FlagEmoji("ARG")
	nedFlag := FlagEmoji("NED")
	if argFlag == "" || nedFlag == "" {
		t.Fatal("expected ARG/NED to have flag emojis registered")
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TeamLabel(tt.team); got != tt.want {
				t.Errorf("TeamLabel(%+v) = %q, want %q", tt.team, got, tt.want)
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
			if got := MatchupTeamLabel(tt.short, tt.full, tt.tbd); got != tt.want {
				t.Errorf("MatchupTeamLabel(%q, %q, %v) = %q, want %q",
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
