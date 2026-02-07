// Package reddit provides functionality to fetch goal replay links from r/soccer.
package reddit

import "time"

// GoalLink represents a cached goal replay link from Reddit.
type GoalLink struct {
	MatchID   int       `json:"match_id"`
	Minute    int       `json:"minute"`
	URL       string    `json:"url"`
	Title     string    `json:"title"`
	PostURL   string    `json:"post_url"`
	FetchedAt time.Time `json:"fetched_at"`
}

// GoalLinkKey creates a unique key for a goal (matchID + minute).
type GoalLinkKey struct {
	MatchID int
	Minute  int
}

// SearchResult represents a Reddit search result from r/soccer.
type SearchResult struct {
	Title     string
	URL       string // The media URL (video/gif link)
	PostURL   string // The Reddit post URL
	Flair     string // e.g., "Media"
	CreatedAt time.Time
	Score     int
}

// redditSearchResponse represents the JSON structure from Reddit's search API.
type redditSearchResponse struct {
	Data struct {
		Children []struct {
			Data redditPost `json:"data"`
		} `json:"children"`
	} `json:"data"`
}

// redditPost represents a single post from Reddit's API.
type redditPost struct {
	Title         string  `json:"title"`
	URL           string  `json:"url"`
	Permalink     string  `json:"permalink"`
	LinkFlairText string  `json:"link_flair_text"`
	CreatedUTC    float64 `json:"created_utc"`
	Score         int     `json:"score"`
	Domain        string  `json:"domain"`
	IsSelf        bool    `json:"is_self"`
	// Media fields for various embed types
	SecureMedia *struct {
		RedditVideo *struct {
			FallbackURL string `json:"fallback_url"`
		} `json:"reddit_video"`
	} `json:"secure_media"`
	Preview *struct {
		Images []struct {
			Source struct {
				URL string `json:"url"`
			} `json:"source"`
		} `json:"images"`
	} `json:"preview"`
}

// toSearchResult converts a redditPost to SearchResult.
func (p *redditPost) toSearchResult() SearchResult {
	// Extract the best available media URL
	mediaURL := p.URL

	// Try to get Reddit video fallback URL if available
	if p.SecureMedia != nil && p.SecureMedia.RedditVideo != nil {
		if p.SecureMedia.RedditVideo.FallbackURL != "" {
			mediaURL = p.SecureMedia.RedditVideo.FallbackURL
		}
	}

	return SearchResult{
		Title:     p.Title,
		URL:       mediaURL,
		PostURL:   "https://www.reddit.com" + p.Permalink,
		Flair:     p.LinkFlairText,
		CreatedAt: time.Unix(int64(p.CreatedUTC), 0),
		Score:     p.Score,
	}
}

// GoalInfo contains information about a goal to search for.
type GoalInfo struct {
	MatchID       int
	HomeTeam      string
	AwayTeam      string
	HomeTeamShort string // Short/alternative name (e.g., "Wolves", "Man Utd")
	AwayTeamShort string // Short/alternative name (e.g., "Wolves", "Man Utd")
	ScorerName    string
	Minute        int
	DisplayMinute string // e.g., "45+2'" for stoppage time display
	HomeScore     int
	AwayScore     int
	IsHomeTeam    bool
	MatchTime     time.Time
}
