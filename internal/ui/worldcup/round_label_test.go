package worldcup

import "testing"

func TestRoundLabel(t *testing.T) {
	cases := []struct {
		raw  string
		want string
	}{
		{"1", "Group Stage · MD1"},
		{"2", "Group Stage · MD2"},
		{"3", "Group Stage · MD3"},
		{"R32", "Round of 32"},
		{"r16", "Round of 16"},
		{"qf", "Quarter-finals"},
		{"SF", "Semi-finals"},
		{"3RD", "Third Place Play-off"},
		{"final", "Final"},
		{"", ""},
		// Unknown values pass through verbatim so the mock fixtures' "Group A"
		// labels (and any future FotMob value we haven't mapped yet) still
		// render rather than getting swallowed.
		{"Group A", "Group A"},
		{"Play-off", "Play-off"},
	}

	for _, tc := range cases {
		if got := roundLabel(tc.raw); got != tc.want {
			t.Errorf("roundLabel(%q) = %q, want %q", tc.raw, got, tc.want)
		}
	}
}
