package worldcup

import "strings"

// FlagEmoji returns the Unicode flag emoji for a 3-letter team short code.
// Falls back to an empty string when not found so callers can decide whether
// to show a placeholder or nothing at all.
func FlagEmoji(shortName string) string {
	if e, ok := flagEmojis[strings.ToUpper(shortName)]; ok {
		return e
	}
	return ""
}

// flagEmojis maps FIFA 3-letter codes to Unicode regional indicator flag emojis.
// Covers all 32 WC 2022 teams plus the additional 16 teams confirmed for 2026.
var flagEmojis = map[string]string{
	// WC 2022 participants
	"QAT": "🇶🇦",
	"ECU": "🇪🇨",
	"SEN": "🇸🇳",
	"NED": "🇳🇱",
	"ENG": "🏴󠁧󠁢󠁥󠁮󠁧󠁿",
	"IRN": "🇮🇷",
	"WAL": "🏴󠁧󠁢󠁷󠁬󠁳󠁿",
	"USA": "🇺🇸",
	"ARG": "🇦🇷",
	"KSA": "🇸🇦",
	"MEX": "🇲🇽",
	"POL": "🇵🇱",
	"FRA": "🇫🇷",
	"DEN": "🇩🇰",
	"TUN": "🇹🇳",
	"AUS": "🇦🇺",
	"ESP": "🇪🇸",
	"GER": "🇩🇪",
	"JPN": "🇯🇵",
	"CRC": "🇨🇷",
	"BEL": "🇧🇪",
	"CAN": "🇨🇦",
	"MAR": "🇲🇦",
	"CRO": "🇭🇷",
	"BRA": "🇧🇷",
	"SRB": "🇷🇸",
	"SUI": "🇨🇭",
	"CMR": "🇨🇲",
	"POR": "🇵🇹",
	"GHA": "🇬🇭",
	"URU": "🇺🇾",
	"KOR": "🇰🇷",
	// Additional WC 2026 qualifiers / likely participants
	"COL": "🇨🇴",
	"CHI": "🇨🇱",
	"PER": "🇵🇪",
	"VEN": "🇻🇪",
	"PAR": "🇵🇾",
	"BOL": "🇧🇴",
	"HON": "🇭🇳",
	"PAN": "🇵🇦",
	"JAM": "🇯🇲",
	"TRI": "🇹🇹",
	"CUB": "🇨🇺",
	"NGA": "🇳🇬",
	"CIV": "🇨🇮",
	"ALG": "🇩🇿",
	"EGY": "🇪🇬",
	"MLI": "🇲🇱",
	"GNB": "🇬🇳",
	"RSA": "🇿🇦",
	"ZIM": "🇿🇼",
	"COD": "🇨🇩",
	"TAN": "🇹🇿",
	"UGA": "🇺🇬",
	"KEN": "🇰🇪",
	"IRI": "🇮🇷", // alternate code used by FotMob
	"ITA": "🇮🇹",
	"GRE": "🇬🇷",
	"TUR": "🇹🇷",
	"UKR": "🇺🇦",
	"AUT": "🇦🇹",
	"HUN": "🇭🇺",
	"SVK": "🇸🇰",
	"CZE": "🇨🇿",
	"ROU": "🇷🇴",
	"SLO": "🇸🇮",
	"SCO": "🏴󠁧󠁢󠁳󠁣󠁴󠁿",
	"IRL": "🇮🇪",
	"NOR": "🇳🇴",
	"SWE": "🇸🇪",
	"FIN": "🇫🇮",
	"ISL": "🇮🇸",
	"ALB": "🇦🇱",
	"BIH": "🇧🇦",
	"MKD": "🇲🇰",
	"MNE": "🇲🇪",
	"GEO": "🇬🇪",
	"AZE": "🇦🇿",
	"ARM": "🇦🇲",
	"KSV": "🇽🇰", // Kosovo
	"CHN": "🇨🇳",
	"IND": "🇮🇳",
	"IDN": "🇮🇩",
	"PHI": "🇵🇭",
	"THA": "🇹🇭",
	"VIE": "🇻🇳",
	"MYS": "🇲🇾",
	"IRQ": "🇮🇶",
	"SYR": "🇸🇾",
	"JOR": "🇯🇴",
	"PAL": "🇵🇸",
	"LIB": "🇱🇧",
	"UAE": "🇦🇪",
	"OMA": "🇴🇲",
	"BHR": "🇧🇭",
	"KUW": "🇰🇼",
	"NZL": "🇳🇿",
	// Common alternate codes
	"HOL": "🇳🇱", // Netherlands alternate
	"GBR": "🇬🇧",
	"ISR": "🇮🇱",
}
