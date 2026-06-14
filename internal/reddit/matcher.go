package reddit

import (
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var (
	reNonAlphanumSpace = regexp.MustCompile(`[^a-z0-9\s]`)
	reNonAlphaSpace    = regexp.MustCompile(`[^a-z\s]`)
	teamNameCache      sync.Map // map[string]string
	playerNameCache    sync.Map // map[string]string

	// diacriticFolder folds Unicode diacritics so "Vinícius" → "Vinicius",
	// "Müller" → "Muller", "Türkiye" → "Turkiye" (preserving the base letter)
	// instead of stripping the entire grapheme. Applied before the lowercase
	// regex strip so player/team-name matching stays loose across sources that
	// use anglicized spellings.
	diacriticFolder = transform.Chain(
		norm.NFD,
		runes.Remove(runes.In(unicode.Mn)),
		norm.NFC,
	)
)

// foldDiacritics returns s with combining diacritical marks removed.
func foldDiacritics(s string) string {
	out, _, err := transform.String(diacriticFolder, s)
	if err != nil {
		return s
	}
	return out
}

// countryAliases maps the normalized form of a national-team name (as produced
// by normalizeTeamName) to additional normalized variants that may appear in
// Reddit goal-post titles. Lookup is exact-key on the goal's normalized team
// name, so this never affects club matches whose normalized names are not
// keys in this map. Variants are kept unambiguous (e.g., "korea republic" not
// bare "korea") to avoid cross-team collisions.
var countryAliases = map[string][]string{
	"turkiye":         {"turkey"},
	"turkey":          {"turkiye"},
	"cote divoire":    {"ivory coast"},
	"ivory coast":     {"cote divoire"},
	"czechia":         {"czech republic"},
	"czech republic":  {"czechia"},
	"korea republic":  {"south korea"},
	"south korea":     {"korea republic"},
	"usa":             {"united states"},
	"united states":   {"usa"},
	"north macedonia": {"macedonia"},
}

// aliasesFor returns the list of alternative normalized names registered for
// the given normalized team name. Returns nil if no aliases are known.
func aliasesFor(teamNorm string) []string {
	return countryAliases[teamNorm]
}

// Matcher provides loose matching for Reddit goal post titles.
// Example titles:
//   - "Wolves [3] - 0 West Ham - Mateus Mane 41'"
//   - "Manchester United [2] - 1 Liverpool - Marcus Rashford 67'"
//   - "Barcelona 0 - [1] Real Madrid - Vinicius Jr 89'"

// findBestMatch finds the best matching search result for a goal.
// Uses loose matching: checks for team names, minute, and date proximity.
func findBestMatch(results []SearchResult, goal GoalInfo) *SearchResult {
	if len(results) == 0 {
		return nil
	}

	// Normalize team names for comparison
	homeNorm := normalizeTeamName(goal.HomeTeam)
	awayNorm := normalizeTeamName(goal.AwayTeam)
	minutePattern := buildMinutePattern(goal)

	// Build score pattern for validation (e.g., "1-0", "2-1", etc.)
	scorePattern := buildScorePattern(goal.HomeScore, goal.AwayScore)

	var bestMatch *SearchResult
	bestScore := 0

	for i := range results {
		result := &results[i]
		titleLower := strings.ToLower(result.Title)

		score := 0

		// Filter by date: post must be within reasonable time of match
		// Allow posts from 1 day before to 2 days after the match
		if !goal.MatchTime.IsZero() {
			postDate := result.CreatedAt
			matchStart := goal.MatchTime.Add(-24 * time.Hour)
			matchEnd := goal.MatchTime.Add(48 * time.Hour)

			if postDate.Before(matchStart) || postDate.After(matchEnd) {
				continue // Post is outside the valid date range
			}

			// Bonus for posts very close to match time (within 12 hours)
			if postDate.After(goal.MatchTime.Add(-6*time.Hour)) && postDate.Before(goal.MatchTime.Add(12*time.Hour)) {
				score += 5
			}
		}

		// Check for team names (required)
		homeFound := containsTeamName(titleLower, homeNorm)
		awayFound := containsTeamName(titleLower, awayNorm)

		if !homeFound && !awayFound {
			continue // Must have at least one team name
		}

		if homeFound {
			score += 10
		}
		if awayFound {
			score += 10
		}

		// Check for minute (highly valuable, but strict)
		minuteFound := minutePattern.MatchString(result.Title)
		if minuteFound {
			score += 25
		}

		// Check for score match (advisory bonus only — no penalty)
		// Score format varies wildly across r/soccer titles (bracketed, spaced,
		// contiguous), so a missing match is not strong evidence of a wrong post.
		// Country + time + scorer carry the matching weight.
		if scorePattern.MatchString(result.Title) {
			score += 20
		}

		// Check for scorer name if available
		if goal.ScorerName != "" {
			scorerNorm := normalizeName(goal.ScorerName)
			if containsName(titleLower, scorerNorm) {
				score += 15
			}
		}

		// Prefer higher Reddit score (upvotes) as tiebreaker
		score += min(result.Score/100, 5) // Max 5 points from upvotes

		if score > bestScore {
			bestScore = score
			bestMatch = result
		}
	}

	// Require minimum score for a match, with higher requirement for score matches
	minScore := 45 // Require score match + minute match + team names
	if bestScore < minScore {
		return nil
	}

	return bestMatch
}

// normalizeTeamName converts a team name to a normalized form for matching.
// Results are cached since the same team names are normalized repeatedly.
func normalizeTeamName(name string) string {
	if cached, ok := teamNameCache.Load(name); ok {
		return cached.(string)
	}
	result := normalizeTeamNameUncached(name)
	teamNameCache.Store(name, result)
	return result
}

func normalizeTeamNameUncached(name string) string {
	// Fold diacritics first ("Türkiye" → "Turkiye") so the lowercase strip
	// below preserves the anglicized base letters instead of dropping them.
	norm := foldDiacritics(name)
	norm = strings.ToLower(norm)

	// Remove common prefixes (e.g., "fc barcelona" -> "barcelona")
	prefixes := []string{"fc ", "cf ", "sc ", "afc ", "ac ", "as "}
	for _, prefix := range prefixes {
		norm = strings.TrimPrefix(norm, prefix)
	}

	// Remove common suffixes
	suffixes := []string{" fc", " cf", " sc", " afc", " united", " city"}
	for _, suffix := range suffixes {
		norm = strings.TrimSuffix(norm, suffix)
	}

	// Remove any remaining special characters
	norm = reNonAlphanumSpace.ReplaceAllString(norm, "")

	return strings.TrimSpace(norm)
}

// normalizeName converts a player name to a normalized form for matching.
// Results are cached since the same player names are normalized repeatedly.
func normalizeName(name string) string {
	if cached, ok := playerNameCache.Load(name); ok {
		return cached.(string)
	}
	result := normalizeNameUncached(name)
	playerNameCache.Store(name, result)
	return result
}

func normalizeNameUncached(name string) string {
	// Fold diacritics so "Müller" → "muller", "Vinícius" → "vinicius" — the
	// anglicized base letters survive the lowercase strip below.
	norm := foldDiacritics(name)
	norm = strings.ToLower(norm)
	// Remove remaining special characters but keep spaces.
	norm = reNonAlphaSpace.ReplaceAllString(norm, "")
	return strings.TrimSpace(norm)
}

// containsTeamName checks if a title contains a team name (or part of it).
// Normalizes the title first to handle variations like "FC Barcelona" vs "Barcelona".
// For national teams listed in countryAliases, also checks anglicized/alternate
// variants (e.g., goal team "Türkiye" → also tries "turkey"). Club matches are
// unaffected because their normalized names are never keys in the alias map.
func containsTeamName(title, teamNorm string) bool {
	if containsTeamNameExact(title, teamNorm) {
		return true
	}
	for _, alias := range aliasesFor(teamNorm) {
		if containsTeamNameExact(title, alias) {
			return true
		}
	}
	return false
}

// containsTeamNameExact is the original substring/word-matching logic, without
// alias expansion. Kept separate so containsTeamName can layer alias lookup on
// top without duplicating the matching strategy.
func containsTeamNameExact(title, teamNorm string) bool {
	// Normalize the title for comparison (handles "FC Barcelona" -> "barcelona")
	titleNorm := normalizeTeamName(title)

	// First try exact match on normalized title
	if strings.Contains(titleNorm, teamNorm) {
		return true
	}

	// Try matching individual words (for multi-word team names)
	words := strings.Fields(teamNorm)
	if len(words) > 1 {
		// Check if significant words are present
		for _, word := range words {
			if len(word) > 3 && strings.Contains(titleNorm, word) {
				return true
			}
		}
	}

	// Also check original title (case-insensitive) for better coverage
	titleLower := strings.ToLower(title)
	if strings.Contains(titleLower, teamNorm) {
		return true
	}

	// Check individual words in original title too
	for _, word := range words {
		if len(word) > 3 && strings.Contains(titleLower, word) {
			return true
		}
	}

	return false
}

// containsName checks if a title contains a player name. Both sides are
// normalized (lowercased, diacritics stripped) before substring matching so
// that "Müller" → "mller" can still match a title containing "Müller". Falls
// back to last-name then first-significant-token match because Reddit titles
// often abbreviate or simplify scorer names (e.g., "Vini Jr" for "Vinícius Jr.",
// "Nuñez" vs "Núñez").
func containsName(title, nameNorm string) bool {
	titleNorm := normalizeName(title)

	// Full normalized name
	if strings.Contains(titleNorm, nameNorm) {
		return true
	}

	parts := strings.Fields(nameNorm)
	if len(parts) == 0 {
		return false
	}

	// Last name (usually the most distinctive token)
	lastName := parts[len(parts)-1]
	if len(lastName) > 2 && strings.Contains(titleNorm, lastName) {
		return true
	}

	// First significant token (e.g., "vinicius" for "Vinícius Jr."), helps
	// when titles drop or abbreviate the surname.
	if len(parts) > 1 {
		first := parts[0]
		if len(first) > 3 && strings.Contains(titleNorm, first) {
			return true
		}
	}

	return false
}

// buildMinutePattern creates a regex to match a minute in various formats.
// Matches: "41'", "41" (at word boundary), "41+2'" etc.
// Also checks +/-2 minute tolerance for better matching.
// If DisplayMinute contains stoppage time (e.g., "45+2'"), also searches for total time (47').
func buildMinutePattern(goal GoalInfo) *regexp.Regexp {
	minute := goal.Minute
	// Create patterns for minute ±2 tolerance
	var patterns []string
	for offset := -2; offset <= 2; offset++ {
		targetMinute := minute + offset
		if targetMinute >= 0 {
			patterns = append(patterns, `\b`+strconv.Itoa(targetMinute)+`(\+\d+)?'?\b`)
		}
	}

	// If DisplayMinute contains stoppage time, also search for total time
	// e.g., "45+2'" should also match "47'" (45 + 2 = 47)
	if goal.DisplayMinute != "" {
		// Parse stoppage time like "45+2'" to find total time
		if plusIndex := strings.Index(goal.DisplayMinute, "+"); plusIndex > 0 {
			baseMinuteStr := goal.DisplayMinute[:plusIndex]
			if baseMinute, err := strconv.Atoi(baseMinuteStr); err == nil {
				// Look for patterns after + (like +2, +3, etc.)
				plusPart := goal.DisplayMinute[plusIndex+1:]
				if quoteIndex := strings.Index(plusPart, "'"); quoteIndex > 0 {
					addedTimeStr := plusPart[:quoteIndex]
					if addedTime, err := strconv.Atoi(addedTimeStr); err == nil {
						totalTime := baseMinute + addedTime
						// Add patterns for the total time ±1 tolerance
						for offset := -1; offset <= 1; offset++ {
							targetTotal := totalTime + offset
							if targetTotal >= 0 && targetTotal != baseMinute { // Avoid duplicate with base minute
								patterns = append(patterns, `\b`+strconv.Itoa(targetTotal)+`'?\b`)
							}
						}
					}
				}
			}
		}
	}

	// Join with OR operator
	patternStr := strings.Join(patterns, "|")
	compiled, err := regexp.Compile(patternStr)
	if err != nil {
		// Fallback to original single minute pattern
		return regexp.MustCompile(`\b` + strconv.Itoa(minute) + `(\+\d+)?'?\b`)
	}
	return compiled
}

// buildScorePattern creates a regex to match the score at the time of goal.
// Accepts several formats commonly used in r/soccer goal-video titles:
//   - contiguous: "1-0"
//   - spaced: "1 - 0"
//   - bracketed on either side: "[1] - 0", "1 - [0]", "[1]-0", "1-[0]"
//   - parenthesised on either side: "(1) - 0", "1 - (0)"
// The brackets/parens may wrap either the home or away digit but not both.
func buildScorePattern(homeScore, awayScore int) *regexp.Regexp {
	h := strconv.Itoa(homeScore)
	a := strconv.Itoa(awayScore)
	// Per-digit optional bracket/paren wrappers (each digit may be wrapped
	// independently), arbitrary whitespace around the dash.
	patternStr := `[\[\(]?` + h + `[\]\)]?\s*-\s*[\[\(]?` + a + `[\]\)]?`
	compiled, err := regexp.Compile(patternStr)
	if err != nil {
		// Fallback to exact match
		return regexp.MustCompile(regexp.QuoteMeta(h + "-" + a))
	}
	return compiled
}

// MatchConfidence represents how confident we are in a match.
type MatchConfidence int

const (
	ConfidenceNone   MatchConfidence = 0
	ConfidenceLow    MatchConfidence = 1
	ConfidenceMedium MatchConfidence = 2
	ConfidenceHigh   MatchConfidence = 3
)

// CalculateConfidence returns the confidence level for a match.
func CalculateConfidence(result SearchResult, goal GoalInfo) MatchConfidence {
	titleLower := strings.ToLower(result.Title)
	homeNorm := normalizeTeamName(goal.HomeTeam)
	awayNorm := normalizeTeamName(goal.AwayTeam)

	hasHome := containsTeamName(titleLower, homeNorm)
	hasAway := containsTeamName(titleLower, awayNorm)
	hasMinute := buildMinutePattern(goal).MatchString(result.Title)

	if hasHome && hasAway && hasMinute {
		return ConfidenceHigh
	}
	if (hasHome || hasAway) && hasMinute {
		return ConfidenceMedium
	}
	if hasHome || hasAway {
		return ConfidenceLow
	}
	return ConfidenceNone
}
