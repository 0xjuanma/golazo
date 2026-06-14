package worldcup

import "strings"

// roundLabels maps the bare round identifiers FotMob ships on the fixtures
// endpoint to human-readable display strings for the World Cup upcoming view.
//
// FotMob emits numeric matchday strings ("1", "2", "3") during the group stage
// and short codes ("R32", "R16", "QF", "SF", "3RD", "FINAL") for the knockout
// rounds. The verbose labels below match the World Cup 2026 48-team format.
var roundLabels = map[string]string{
	"1":     "Group Stage · MD1",
	"2":     "Group Stage · MD2",
	"3":     "Group Stage · MD3",
	"R32":   "Round of 32",
	"R16":   "Round of 16",
	"QF":    "Quarter-finals",
	"SF":    "Semi-finals",
	"3RD":   "Third Place Play-off",
	"FINAL": "Final",
}

// roundLabel returns a human-readable label for the given FotMob round value.
// Lookup is case-insensitive against roundLabels; unknown values are returned
// verbatim so descriptive labels coming from other code paths (e.g. mock
// fixtures shipping "Group A") still render unchanged.
func roundLabel(raw string) string {
	if raw == "" {
		return ""
	}
	if label, ok := roundLabels[strings.ToUpper(raw)]; ok {
		return label
	}
	return raw
}
