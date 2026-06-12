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
// Covers all 32 WC 2022 teams plus the additional teams confirmed for 2026 and
// the broader set of confederation qualifiers that surface in FotMob payloads
// during the WC 2026 qualifying cycle.
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
	// WC 2026 confirmed qualifiers / strong qualifying candidates not in WC 2022.
	"UZB": "🇺🇿", // Uzbekistan — first-ever WC qualifier (AFC)
	"CPV": "🇨🇻", // Cape Verde — first-ever WC qualifier (CAF)
	"CUW": "🇨🇼", // Curaçao (CONCACAF)
	"HAI": "🇭🇹", // Haiti
	"SUR": "🇸🇷", // Suriname
	"NCL": "🇳🇨", // New Caledonia (OFC playoff candidate)
	"DOM": "🇩🇴", // Dominican Republic
	"GUA": "🇬🇹", // Guatemala
	"SLV": "🇸🇻", // El Salvador
	"PRK": "🇰🇵", // North Korea / DPR Korea
	// CAF coverage useful in 2026 qualifying rounds and intercontinental playoffs.
	"BFA": "🇧🇫", // Burkina Faso
	"ETH": "🇪🇹", // Ethiopia
	"GAB": "🇬🇦", // Gabon
	"LBY": "🇱🇾", // Libya
	"NIG": "🇳🇪", // Niger
	"MAD": "🇲🇬", // Madagascar
	"MOZ": "🇲🇿", // Mozambique
	"ANG": "🇦🇴", // Angola
	"ZAM": "🇿🇲", // Zambia
	"SLE": "🇸🇱", // Sierra Leone
	"EQG": "🇬🇶", // Equatorial Guinea
	"BEN": "🇧🇯", // Benin
	"TOG": "🇹🇬", // Togo
	"COM": "🇰🇲", // Comoros
	"SDN": "🇸🇩", // Sudan
	"MTN": "🇲🇷", // Mauritania
	"NAM": "🇳🇦", // Namibia
	"BOT": "🇧🇼", // Botswana
	"RWA": "🇷🇼", // Rwanda
	// AFC / UEFA tail coverage for qualifying rounds and FotMob's broader payload.
	"KAZ": "🇰🇿", // Kazakhstan
	"TJK": "🇹🇯", // Tajikistan
	"KGZ": "🇰🇬", // Kyrgyzstan
	"TKM": "🇹🇲", // Turkmenistan
	"LUX": "🇱🇺", // Luxembourg
	"CYP": "🇨🇾", // Cyprus
	"MLT": "🇲🇹", // Malta
	"LVA": "🇱🇻", // Latvia
	"LTU": "🇱🇹", // Lithuania
	"EST": "🇪🇪", // Estonia
	"MDA": "🇲🇩", // Moldova
	"BLR": "🇧🇾", // Belarus
	"FRO": "🇫🇴", // Faroe Islands
	"LIE": "🇱🇮", // Liechtenstein
	"RUS": "🇷🇺", // Russia
	// Common alternate codes
	"HOL": "🇳🇱", // Netherlands alternate
	"GBR": "🇬🇧",
	"ISR": "🇮🇱",
}
