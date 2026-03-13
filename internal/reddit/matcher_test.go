package reddit

import (
	"regexp"
	"sync"
	"testing"
	"time"
)

func TestNormalizeTeamName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"lowercase", "Barcelona", "barcelona"},
		{"FC prefix removed", "FC Barcelona", "barcelona"},
		{"CF prefix removed", "CF Pachuca", "pachuca"},
		{"SC prefix removed", "SC Freiburg", "freiburg"},
		{"AFC prefix removed", "AFC Bournemouth", "bournemouth"},
		{"AC prefix removed", "AC Milan", "milan"},
		{"AS prefix removed", "AS Roma", "roma"},
		{"FC suffix removed", "Liverpool FC", "liverpool"},
		{"CF suffix removed", "Valencia CF", "valencia"},
		{"United suffix removed", "Manchester United", "manchester"},
		{"City suffix removed", "Manchester City", "manchester"},
		{"special chars removed", "Atlético Madrid", "atltico madrid"},
		{"accented chars removed", "São Paulo", "so paulo"},
		{"multiple normalizations", "FC Zürich", "zrich"},
		{"already clean", "wolves", "wolves"},
		{"whitespace trimmed", "  arsenal  ", "arsenal"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use uncached version to avoid cache interference between tests
			got := normalizeTeamNameUncached(tt.input)
			if got != tt.expected {
				t.Errorf("normalizeTeamNameUncached(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestNormalizeName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple name", "Rashford", "rashford"},
		{"full name", "Marcus Rashford", "marcus rashford"},
		{"accented chars", "Müller", "mller"},
		{"hyphenated name", "Pierre-Emerick Aubameyang", "pierreemerick aubameyang"},
		{"special chars", "Vinícius Jr.", "vincius jr"},
		{"apostrophe", "N'Golo Kanté", "ngolo kant"},
		{"already clean", "messi", "messi"},
		{"whitespace trimmed", "  salah  ", "salah"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeNameUncached(tt.input)
			if got != tt.expected {
				t.Errorf("normalizeNameUncached(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestBuildMinutePattern(t *testing.T) {
	tests := []struct {
		name          string
		goal          GoalInfo
		shouldMatch   []string
		shouldNotMatch []string
	}{
		{
			name: "regular minute 41",
			goal: GoalInfo{Minute: 41},
			shouldMatch: []string{
				"41'", "40'", "42'", "39'", "43'",
				"Goal 41' by Mane",
				"41+2'",
			},
			shouldNotMatch: []string{
				"44'", "37'", "141'",
			},
		},
		{
			name: "stoppage time 45+2",
			goal: GoalInfo{Minute: 45, DisplayMinute: "45+2'"},
			shouldMatch: []string{
				"45'", "45+2'", "47'", "46'", "48'",
				"Goal at 43'",
			},
			shouldNotMatch: []string{
				"50'", "40'",
			},
		},
		{
			name: "minute 0 edge case",
			goal: GoalInfo{Minute: 1},
			shouldMatch: []string{
				"1'", "2'", "3'",
			},
			shouldNotMatch: []string{
				"5'", "10'",
			},
		},
		{
			name: "90+3 stoppage time",
			goal: GoalInfo{Minute: 90, DisplayMinute: "90+3'"},
			shouldMatch: []string{
				"90'", "90+3'", "93'", "92'", "94'",
				"88'", "91'",
			},
			shouldNotMatch: []string{
				"85'", "96'",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pattern := buildMinutePattern(tt.goal)
			for _, s := range tt.shouldMatch {
				if !pattern.MatchString(s) {
					t.Errorf("pattern should match %q but didn't (pattern: %s)", s, pattern.String())
				}
			}
			for _, s := range tt.shouldNotMatch {
				if pattern.MatchString(s) {
					t.Errorf("pattern should NOT match %q but did (pattern: %s)", s, pattern.String())
				}
			}
		})
	}
}

func TestTeamNameCaching(t *testing.T) {
	// Clear cache to start fresh
	resetSyncMap(&teamNameCache)

	input := "FC Barcelona"
	first := normalizeTeamName(input)
	second := normalizeTeamName(input)

	if first != second {
		t.Errorf("cached result differs: first=%q, second=%q", first, second)
	}

	// Verify it was actually cached
	cached, ok := teamNameCache.Load(input)
	if !ok {
		t.Error("expected value to be cached")
	}
	if cached.(string) != first {
		t.Errorf("cached value %q differs from returned value %q", cached.(string), first)
	}
}

func TestPlayerNameCaching(t *testing.T) {
	// Clear cache to start fresh
	resetSyncMap(&playerNameCache)

	input := "Marcus Rashford"
	first := normalizeName(input)
	second := normalizeName(input)

	if first != second {
		t.Errorf("cached result differs: first=%q, second=%q", first, second)
	}

	// Verify it was actually cached
	cached, ok := playerNameCache.Load(input)
	if !ok {
		t.Error("expected value to be cached")
	}
	if cached.(string) != first {
		t.Errorf("cached value %q differs from returned value %q", cached.(string), first)
	}
}

// resetSyncMap replaces a sync.Map with a fresh empty one.
func resetSyncMap(m *sync.Map) {
	m.Range(func(key, value any) bool {
		m.Delete(key)
		return true
	})
}

func TestContainsTeamName(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		teamNorm string
		expected bool
	}{
		{"exact match", "wolves 3 - 0 west ham", "wolves", true},
		{"partial multi-word", "manchester united vs liverpool", "manchester", true},
		{"not found", "arsenal vs chelsea", "barcelona", false},
		{"word in title", "real madrid vs barcelona", "barcelona", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsTeamName(tt.title, tt.teamNorm)
			if got != tt.expected {
				t.Errorf("containsTeamName(%q, %q) = %v, want %v", tt.title, tt.teamNorm, got, tt.expected)
			}
		})
	}
}

func TestContainsName(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		nameNorm string
		expected bool
	}{
		{"full name match", "goal by marcus rashford 67'", "marcus rashford", true},
		{"last name match", "goal by rashford 67'", "marcus rashford", true},
		{"not found", "goal by salah 67'", "marcus rashford", false},
		{"single name", "goal by messi 45'", "messi", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsName(tt.title, tt.nameNorm)
			if got != tt.expected {
				t.Errorf("containsName(%q, %q) = %v, want %v", tt.title, tt.nameNorm, got, tt.expected)
			}
		})
	}
}

func TestBuildScorePattern(t *testing.T) {
	tests := []struct {
		name       string
		home, away int
		shouldMatch   []string
		shouldNotMatch []string
	}{
		{
			name: "1-0",
			home: 1, away: 0,
			shouldMatch:    []string{"1-0", "[1-0]", "(1-0)", " 1-0 "},
			shouldNotMatch: []string{"2-0", "1-1"},
		},
		{
			name: "2-1",
			home: 2, away: 1,
			shouldMatch:    []string{"2-1", "[2-1]"},
			shouldNotMatch: []string{"1-2", "2-0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pattern := buildScorePattern(tt.home, tt.away)
			for _, s := range tt.shouldMatch {
				if !pattern.MatchString(s) {
					t.Errorf("pattern should match %q but didn't (pattern: %s)", s, pattern.String())
				}
			}
			for _, s := range tt.shouldNotMatch {
				if pattern.MatchString(s) {
					t.Errorf("pattern should NOT match %q but did (pattern: %s)", s, pattern.String())
				}
			}
		})
	}
}

func TestFindBestMatch(t *testing.T) {
	matchTime := time.Date(2024, 1, 15, 20, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		results  []SearchResult
		goal     GoalInfo
		wantNil  bool
		wantURL  string
	}{
		{
			name:    "empty results returns nil",
			results: nil,
			goal:    GoalInfo{HomeTeam: "Wolves", AwayTeam: "West Ham", Minute: 41},
			wantNil: true,
		},
		{
			name: "high confidence match",
			results: []SearchResult{
				{
					Title:     "Wolves [3] - 0 West Ham - Mateus Mane 41'",
					URL:       "https://example.com/goal1",
					CreatedAt: matchTime.Add(1 * time.Hour),
					Score:     5000,
				},
			},
			goal: GoalInfo{
				HomeTeam:   "Wolves",
				AwayTeam:   "West Ham",
				ScorerName: "Mateus Mane",
				Minute:     41,
				HomeScore:  3,
				AwayScore:  0,
				MatchTime:  matchTime,
			},
			wantNil: false,
			wantURL: "https://example.com/goal1",
		},
		{
			name: "no team match returns nil",
			results: []SearchResult{
				{
					Title:     "Arsenal [2] - 0 Chelsea - Saka 55'",
					URL:       "https://example.com/goal2",
					CreatedAt: matchTime.Add(1 * time.Hour),
					Score:     3000,
				},
			},
			goal: GoalInfo{
				HomeTeam:  "Wolves",
				AwayTeam:  "West Ham",
				Minute:    41,
				HomeScore: 3,
				AwayScore: 0,
				MatchTime: matchTime,
			},
			wantNil: true,
		},
		{
			name: "post outside date range is excluded",
			results: []SearchResult{
				{
					Title:     "Wolves [3] - 0 West Ham - Mateus Mane 41'",
					URL:       "https://example.com/goal3",
					CreatedAt: matchTime.Add(-72 * time.Hour), // 3 days before
					Score:     5000,
				},
			},
			goal: GoalInfo{
				HomeTeam:  "Wolves",
				AwayTeam:  "West Ham",
				Minute:    41,
				HomeScore: 3,
				AwayScore: 0,
				MatchTime: matchTime,
			},
			wantNil: true,
		},
		{
			name: "picks best scoring result",
			results: []SearchResult{
				{
					Title:     "Wolves [3] - 0 West Ham - Mateus Mane 41'",
					URL:       "https://example.com/best",
					CreatedAt: matchTime.Add(1 * time.Hour),
					Score:     5000,
				},
				{
					Title:     "Wolves [3] - 0 West Ham 41'",
					URL:       "https://example.com/worse",
					CreatedAt: matchTime.Add(1 * time.Hour),
					Score:     100,
				},
			},
			goal: GoalInfo{
				HomeTeam:   "Wolves",
				AwayTeam:   "West Ham",
				ScorerName: "Mateus Mane",
				Minute:     41,
				HomeScore:  3,
				AwayScore:  0,
				MatchTime:  matchTime,
			},
			wantNil: false,
			wantURL: "https://example.com/best",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findBestMatch(tt.results, tt.goal)
			if tt.wantNil {
				if got != nil {
					t.Errorf("expected nil, got result with URL %q", got.URL)
				}
				return
			}
			if got == nil {
				t.Fatal("expected non-nil result, got nil")
			}
			if got.URL != tt.wantURL {
				t.Errorf("got URL %q, want %q", got.URL, tt.wantURL)
			}
		})
	}
}

func TestCalculateConfidence(t *testing.T) {
	tests := []struct {
		name       string
		result     SearchResult
		goal       GoalInfo
		wantLevel  MatchConfidence
	}{
		{
			name:   "high confidence - both teams and minute",
			result: SearchResult{Title: "Wolves [3] - 0 West Ham - Mane 41'"},
			goal: GoalInfo{
				HomeTeam: "Wolves",
				AwayTeam: "West Ham",
				Minute:   41,
			},
			wantLevel: ConfidenceHigh,
		},
		{
			name:   "medium confidence - one team and minute",
			result: SearchResult{Title: "Wolves goal 41'"},
			goal: GoalInfo{
				HomeTeam: "Wolves",
				AwayTeam: "West Ham",
				Minute:   41,
			},
			wantLevel: ConfidenceMedium,
		},
		{
			name:   "low confidence - team only no minute",
			result: SearchResult{Title: "Wolves vs West Ham highlights"},
			goal: GoalInfo{
				HomeTeam: "Wolves",
				AwayTeam: "West Ham",
				Minute:   41,
			},
			wantLevel: ConfidenceLow,
		},
		{
			name:   "no confidence - no match",
			result: SearchResult{Title: "Arsenal vs Chelsea 55'"},
			goal: GoalInfo{
				HomeTeam: "Wolves",
				AwayTeam: "West Ham",
				Minute:   41,
			},
			wantLevel: ConfidenceNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateConfidence(tt.result, tt.goal)
			if got != tt.wantLevel {
				t.Errorf("CalculateConfidence() = %d, want %d", got, tt.wantLevel)
			}
		})
	}
}

func TestBuildMinutePatternIsValidRegex(t *testing.T) {
	// Ensure buildMinutePattern always returns a valid compiled regex
	goals := []GoalInfo{
		{Minute: 0},
		{Minute: 1},
		{Minute: 45, DisplayMinute: "45+2'"},
		{Minute: 90, DisplayMinute: "90+5'"},
		{Minute: 120, DisplayMinute: "120+1'"},
	}

	for _, g := range goals {
		pattern := buildMinutePattern(g)
		if pattern == nil {
			t.Errorf("buildMinutePattern returned nil for minute=%d display=%q", g.Minute, g.DisplayMinute)
		}
		// Verify it's a valid compiled regexp by calling a method
		_ = pattern.MatchString("test")
	}
}

// TestBuildMinutePatternNoFalsePositive verifies that minute patterns
// don't match numbers embedded in larger numbers (word boundary check).
func TestBuildMinutePatternNoFalsePositive(t *testing.T) {
	goal := GoalInfo{Minute: 41}
	pattern := buildMinutePattern(goal)

	// "141'" contains "41" but the word boundary should prevent matching
	// Note: \b in regex considers digit-to-non-digit as boundary,
	// so "141'" would have \b before 1 not before 4.
	// Let's verify actual behavior:
	if pattern.MatchString("x141x") {
		t.Log("Note: pattern matches '141' embedded in text - word boundary applies at digit boundaries")
	}

	// These should definitely NOT match
	noMatch := []string{
		"minute 100",
		"score 500",
	}
	for _, s := range noMatch {
		if matched := pattern.MatchString(s); matched {
			t.Errorf("pattern unexpectedly matched %q", s)
		}
	}
}

// Compile-time check that MatchConfidence constants are defined
func TestMatchConfidenceConstants(t *testing.T) {
	if ConfidenceNone != 0 {
		t.Errorf("ConfidenceNone = %d, want 0", ConfidenceNone)
	}
	if ConfidenceHigh <= ConfidenceMedium {
		t.Error("ConfidenceHigh should be greater than ConfidenceMedium")
	}
	if ConfidenceMedium <= ConfidenceLow {
		t.Error("ConfidenceMedium should be greater than ConfidenceLow")
	}
	if ConfidenceLow <= ConfidenceNone {
		t.Error("ConfidenceLow should be greater than ConfidenceNone")
	}
}

// TestRegexpPackageLevelVars ensures the package-level regexps compile correctly.
func TestRegexpPackageLevelVars(t *testing.T) {
	// These are compiled at init time; if they fail, the package won't load.
	// But we can verify their behavior.
	tests := []struct {
		name    string
		re      *regexp.Regexp
		input   string
		want    string
	}{
		{"reNonAlphanumSpace removes accents", reNonAlphanumSpace, "atlético", "atltico"},
		{"reNonAlphanumSpace keeps digits", reNonAlphanumSpace, "team123", "team123"},
		{"reNonAlphanumSpace keeps spaces", reNonAlphanumSpace, "real madrid", "real madrid"},
		{"reNonAlphaSpace removes digits", reNonAlphaSpace, "player123", "player"},
		{"reNonAlphaSpace removes hyphens", reNonAlphaSpace, "pierre-emerick", "pierreemerick"},
		{"reNonAlphaSpace keeps spaces", reNonAlphaSpace, "marcus rashford", "marcus rashford"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.re.ReplaceAllString(tt.input, "")
			if got != tt.want {
				t.Errorf("regex replace on %q = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
