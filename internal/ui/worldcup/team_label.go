package worldcup

import (
	"strings"

	"github.com/0xjuanma/golazo/internal/api"
)

// TeamLabel returns the consistent World Cup display label for a team:
// "<flag-emoji> <CODE>", where <CODE> is the FIFA 3-letter abbreviation.
//
// The code resolution chain is:
//  1. team.ShortName, if non-empty (already a 3-letter code from FotMob).
//  2. A WC-local Name → code override map for known mismatches where FotMob
//     ships the full English name without a short code (e.g. "Netherlands" → "NED").
//  3. A deterministic fallback: uppercase the first 3 letters of the name with
//     spaces stripped (e.g. "Cape Verde" → "CAP"). This is never ideal but
//     keeps every cell aligned even when we receive an unknown country.
//
// When no flag emoji is registered for the resolved code, the emoji slot is
// padded with two spaces so that columns stay aligned across rows.
func TeamLabel(t api.Team) string {
	code := teamCode(t.ShortName, t.Name)
	return labelWithFlag(code)
}

// MatchupTeamLabel is the matchup-shape variant: bracket matchups carry
// (short, full, tbd) as separate fields on api.WCMatchup rather than an
// api.Team value. It applies the same code resolution chain as TeamLabel
// and returns "TBD" for unresolved bracket slots.
func MatchupTeamLabel(short, full string, tbd bool) string {
	if tbd {
		return "TBD"
	}
	if short == "" && full == "" {
		return "TBD"
	}
	code := teamCode(short, full)
	return labelWithFlag(code)
}

// teamCode resolves a team to its canonical 3-letter code using the chain
// described on TeamLabel.
func teamCode(short, full string) string {
	if c := strings.ToUpper(strings.TrimSpace(short)); c != "" {
		return c
	}
	if c, ok := wcNameToCode[strings.ToLower(strings.TrimSpace(full))]; ok {
		return c
	}
	stripped := strings.ToUpper(strings.ReplaceAll(full, " ", ""))
	if len(stripped) > 3 {
		stripped = stripped[:3]
	}
	return stripped
}

// labelWithFlag renders "<emoji> <CODE>", padding the emoji slot with two
// spaces when no flag is registered so columns line up across rows.
func labelWithFlag(code string) string {
	if code == "" {
		return "   " // 3-cell placeholder consistent with emoji + space slot
	}
	if emoji := FlagEmoji(code); emoji != "" {
		return emoji + " " + code
	}
	return "   " + code
}

// wcNameToCode covers WC teams whose FotMob payloads sometimes ship a full
// English name without a populated shortName. Keep keys lowercase for the
// case-insensitive lookup in teamCode. Coverage tracks flagEmojis (WC 2022
// participants + likely 2026 qualifiers) so a new entry there should add a
// matching entry here when the name → code mapping is non-obvious.
var wcNameToCode = map[string]string{
	// Names that don't naive-truncate correctly.
	"netherlands":       "NED",
	"holland":           "NED",
	"saudi arabia":      "KSA",
	"south korea":       "KOR",
	"korea republic":    "KOR",
	"republic of korea": "KOR",
	"north korea":       "PRK",
	"costa rica":        "CRC",
	"switzerland":       "SUI",
	"croatia":           "CRO",
	"serbia":            "SRB",
	"poland":            "POL",
	"portugal":          "POR",
	"germany":           "GER",
	"denmark":           "DEN",
	"belgium":           "BEL",
	"morocco":           "MAR",
	"senegal":           "SEN",
	"cameroon":          "CMR",
	"ghana":             "GHA",
	"uruguay":           "URU",
	"australia":         "AUS",
	"ecuador":           "ECU",
	"qatar":             "QAT",
	"iran":              "IRN",
	"ir iran":           "IRN",
	"wales":             "WAL",
	"england":           "ENG",
	"scotland":          "SCO",
	"northern ireland":  "NIR",
	"czech republic":    "CZE",
	"czechia":           "CZE",
	"slovakia":          "SVK",
	"slovenia":          "SLO",
	"romania":           "ROU",
	"hungary":           "HUN",
	"austria":           "AUT",
	"ukraine":           "UKR",
	"turkey":            "TUR",
	"türkiye":           "TUR",
	"greece":            "GRE",
	"ireland":           "IRL",
	"republic of ireland": "IRL",
	"iceland":           "ISL",
	"norway":            "NOR",
	"sweden":            "SWE",
	"finland":           "FIN",
	"bosnia & herzegovina": "BIH",
	"bosnia and herzegovina": "BIH",
	"north macedonia":   "MKD",
	"montenegro":        "MNE",
	"albania":           "ALB",
	"kosovo":            "KSV",
	"georgia":           "GEO",
	"azerbaijan":        "AZE",
	"armenia":           "ARM",
	// Americas.
	"united states":     "USA",
	"usa":               "USA",
	"mexico":            "MEX",
	"canada":            "CAN",
	"argentina":         "ARG",
	"brazil":            "BRA",
	"chile":             "CHI",
	"peru":              "PER",
	"colombia":          "COL",
	"venezuela":         "VEN",
	"paraguay":          "PAR",
	"bolivia":           "BOL",
	"honduras":          "HON",
	"panama":            "PAN",
	"jamaica":           "JAM",
	"trinidad and tobago": "TRI",
	"cuba":              "CUB",
	// Africa.
	"nigeria":           "NGA",
	"ivory coast":       "CIV",
	"côte d'ivoire":     "CIV",
	"cote d'ivoire":     "CIV",
	"algeria":           "ALG",
	"egypt":             "EGY",
	"mali":              "MLI",
	"guinea-bissau":     "GNB",
	"guinea bissau":     "GNB",
	"south africa":      "RSA",
	"zimbabwe":          "ZIM",
	"dr congo":          "COD",
	"congo dr":          "COD",
	"tanzania":          "TAN",
	"uganda":            "UGA",
	"kenya":             "KEN",
	// Asia/Oceania.
	"japan":             "JPN",
	"china":             "CHN",
	"china pr":          "CHN",
	"india":             "IND",
	"indonesia":         "IDN",
	"philippines":       "PHI",
	"thailand":          "THA",
	"vietnam":           "VIE",
	"malaysia":          "MYS",
	"iraq":              "IRQ",
	"syria":             "SYR",
	"jordan":            "JOR",
	"palestine":         "PAL",
	"lebanon":           "LIB",
	"united arab emirates": "UAE",
	"oman":              "OMA",
	"bahrain":           "BHR",
	"kuwait":            "KUW",
	"new zealand":       "NZL",
	// Europe big-five and others not already truncating right.
	"spain":             "ESP",
	"france":            "FRA",
	"italy":             "ITA",
	"tunisia":           "TUN",
}
