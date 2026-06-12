package worldcup

import (
	"strings"

	"github.com/0xjuanma/golazo/internal/api"
	"github.com/mattn/go-runewidth"
)

// labelTargetWidth is the visual width every team label is padded to so rows
// align across all WC views regardless of which resolution branch produced
// the label or how the terminal measures regional-indicator / tag-sequence
// flag clusters. Matches "   XYZ" (3-cell placeholder + 3-letter code).
const labelTargetWidth = 6

// TeamLabel returns the consistent World Cup display label for a team:
// "<flag-emoji> <CODE>", where <CODE> is the FIFA 3-letter abbreviation.
//
// The code resolution chain is:
//  1. team.ShortName, if non-empty AND it resolves to a registered flag.
//     This guards against FotMob shipping a non-FIFA shortName (e.g. "SOU"
//     for both "South Africa" and "South Korea") that would otherwise mask
//     the correct FIFA code.
//  2. A WC-local Name → code override map for known mismatches where FotMob
//     ships the full English name without a populated short code
//     (e.g. "Netherlands" → "NED") or ships an ambiguous short code.
//  3. team.ShortName as-is (no flag) when the override map has no match.
//  4. A deterministic fallback: uppercase the first 3 letters of the name
//     with spaces stripped (e.g. "Cape Verde" → "CAP"). This is never ideal
//     but keeps every cell aligned even when we receive an unknown country.
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
// described on TeamLabel. The returned code is always truncated to at most
// three characters so every WC view renders teams in the same column width.
//
// ShortName is preferred only when it resolves to a registered flag emoji.
// This guards against FotMob occasionally shipping a non-FIFA shortName
// (e.g. "SOU" for both "South Africa" and "South Korea") that would
// otherwise mask the correct FIFA code available in wcNameToCode.
func teamCode(short, full string) string {
	capped := capCode(strings.ToUpper(strings.TrimSpace(short)))
	if capped != "" && FlagEmoji(capped) != "" {
		return capped
	}
	if c, ok := wcNameToCode[strings.ToLower(strings.TrimSpace(full))]; ok {
		return capCode(c)
	}
	if capped != "" {
		return capped
	}
	stripped := strings.ToUpper(strings.ReplaceAll(full, " ", ""))
	return capCode(stripped)
}

// capCode enforces the 3-letter cap shared by every code-resolution branch.
func capCode(c string) string {
	if len(c) > 3 {
		return c[:3]
	}
	return c
}

// labelWithFlag renders "<emoji> <CODE>", padding to a fixed visual width so
// every row aligns regardless of (a) whether the resolved code has a
// registered flag and (b) how the terminal measures the emoji cluster. The
// no-flag branch reserves a 3-cell placeholder so columns match the with-flag
// branch under any width-table assumption.
func labelWithFlag(code string) string {
	if code == "" {
		return padToLabelWidth("   ")
	}
	if emoji := FlagEmoji(code); emoji != "" {
		return padToLabelWidth(emoji + " " + code)
	}
	return padToLabelWidth("   " + code)
}

// padToLabelWidth right-pads s with spaces until its visual width matches
// labelTargetWidth under runewidth's tables. lipgloss is consulted as a
// second check so labels stay consistent between width metrics; whichever
// table reports the larger width drives the padding floor, and the smaller
// metric is filled to match. Strings already wider than the target are
// returned unchanged.
func padToLabelWidth(s string) string {
	w := runewidth.StringWidth(s)
	if w >= labelTargetWidth {
		return s
	}
	return s + strings.Repeat(" ", labelTargetWidth-w)
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
	// WC 2026 confirmed qualifiers / strong qualifying candidates not in WC 2022.
	"uzbekistan":              "UZB",
	"cape verde":              "CPV",
	"cabo verde":              "CPV",
	"curacao":                 "CUW",
	"curaçao":                 "CUW",
	"haiti":                   "HAI",
	"suriname":                "SUR",
	"new caledonia":           "NCL",
	"dominican republic":      "DOM",
	"guatemala":               "GUA",
	"el salvador":             "SLV",
	"dpr korea":               "PRK",
	"korea dpr":               "PRK",
	// CAF coverage useful in 2026 qualifying rounds and intercontinental playoffs.
	"burkina faso":            "BFA",
	"ethiopia":                "ETH",
	"gabon":                   "GAB",
	"libya":                   "LBY",
	"niger":                   "NIG",
	"madagascar":              "MAD",
	"mozambique":              "MOZ",
	"angola":                  "ANG",
	"zambia":                  "ZAM",
	"sierra leone":            "SLE",
	"equatorial guinea":       "EQG",
	"benin":                   "BEN",
	"togo":                    "TOG",
	"comoros":                 "COM",
	"sudan":                   "SDN",
	"mauritania":              "MTN",
	"namibia":                 "NAM",
	"botswana":                "BOT",
	"rwanda":                  "RWA",
	// AFC / UEFA tail coverage for qualifying rounds.
	"kazakhstan":              "KAZ",
	"tajikistan":              "TJK",
	"kyrgyzstan":              "KGZ",
	"turkmenistan":            "TKM",
	"luxembourg":              "LUX",
	"cyprus":                  "CYP",
	"malta":                   "MLT",
	"latvia":                  "LVA",
	"lithuania":               "LTU",
	"estonia":                 "EST",
	"moldova":                 "MDA",
	"belarus":                 "BLR",
	"faroe islands":           "FRO",
	"liechtenstein":           "LIE",
	"russia":                  "RUS",
	"israel":                  "ISR",
}
