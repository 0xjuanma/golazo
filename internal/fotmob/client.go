package fotmob

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/0xjuanma/golazo/internal/api"
)

const (
	baseURL = "https://www.fotmob.com/api"
)

// Supported league IDs for match fetching
// Limited to top 5 leagues to match stats view and optimize API calls
var (
	// SupportedLeagues contains the league IDs that will be queried for matches.
	// FotMob league IDs (different from API-Sports.io IDs):
	//   - Premier League: 47
	//   - La Liga: 87
	//   - Bundesliga: 54
	//   - Serie A: 55
	//   - Ligue 1: 53
	SupportedLeagues = []int{
		47, // Premier League
		87, // La Liga
		54, // Bundesliga
		55, // Serie A
		53, // Ligue 1
	}
)

// Client implements the api.Client interface for FotMob API
type Client struct {
	httpClient  *http.Client
	baseURL     string
	rateLimiter *RateLimiter
}

// NewClient creates a new FotMob API client with default configuration.
// Includes conservative rate limiting (2 seconds between requests).
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL:     baseURL,
		rateLimiter: NewRateLimiter(2 * time.Second), // Conservative: 2 seconds between requests
	}
}

// MatchesByDate retrieves all matches for a specific date.
// Since the /api/matches endpoint returns 404, we query the supported leagues
// concurrently (with rate limiting) and aggregate their fixtures for the given date.
func (c *Client) MatchesByDate(ctx context.Context, date time.Time) ([]api.Match, error) {
	// Normalize date to UTC for consistent comparison
	requestDateStr := date.UTC().Format("2006-01-02")

	// Use a mutex to protect the shared slice
	var mu sync.Mutex
	var allMatches []api.Match

	// Query leagues concurrently with rate limiting
	// Best-effort aggregation: if a league query fails, we skip it and continue with others
	// This allows partial results even if some leagues are unavailable
	var wg sync.WaitGroup
	for i, leagueID := range SupportedLeagues {
		wg.Add(1)
		go func(id int, index int) {
			defer wg.Done()

			// Stagger requests slightly to respect rate limits
			if index > 0 {
				time.Sleep(time.Duration(index) * 500 * time.Millisecond)
			}
			c.rateLimiter.Wait()

			url := fmt.Sprintf("%s/leagues?id=%d&tab=fixtures", c.baseURL, id)

			req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
			if err != nil {
				// Skip this league on error - best effort aggregation
				return
			}

			req.Header.Set("User-Agent", "Mozilla/5.0")

			resp, err := c.httpClient.Do(req)
			if err != nil {
				// Skip this league on request error - best effort aggregation
				return
			}
			defer resp.Body.Close()

			var leagueResponse struct {
				Details struct {
					ID          int    `json:"id"`
					Name        string `json:"name"`
					Country     string `json:"country"`
					CountryCode string `json:"countryCode,omitempty"`
				} `json:"details"`
				Fixtures struct {
					AllMatches []fotmobMatch `json:"allMatches"`
				} `json:"fixtures"`
			}

			if err := json.NewDecoder(resp.Body).Decode(&leagueResponse); err != nil {
				// Skip this league on parse error - best effort aggregation
				return
			}

			// Filter matches for the requested date and add league info
			// Note: Matches are sorted chronologically, so we need to check all matches
			var leagueMatches []api.Match
			for _, m := range leagueResponse.Fixtures.AllMatches {
				// Check if match is on the requested date
				if m.Status.UTCTime != "" {
					// Parse the UTC time - FotMob sometimes uses .000Z format
					var matchTime time.Time
					var err error
					matchTime, err = time.Parse(time.RFC3339, m.Status.UTCTime)
					if err != nil {
						// Try alternative format with milliseconds (.000Z)
						matchTime, err = time.Parse("2006-01-02T15:04:05.000Z", m.Status.UTCTime)
					}
					if err == nil {
						// Compare dates in UTC to avoid timezone issues
						matchDateStr := matchTime.UTC().Format("2006-01-02")
						if matchDateStr == requestDateStr {
							// Set league info from the response details
							if m.League.ID == 0 {
								m.League = league{
									ID:          leagueResponse.Details.ID,
									Name:        leagueResponse.Details.Name,
									Country:     leagueResponse.Details.Country,
									CountryCode: leagueResponse.Details.CountryCode,
								}
							}
							leagueMatches = append(leagueMatches, m.toAPIMatch())
						}
					}
				}
			}

			// Append to shared slice with mutex protection
			mu.Lock()
			allMatches = append(allMatches, leagueMatches...)
			mu.Unlock()
		}(leagueID, i)
	}

	wg.Wait()
	return allMatches, nil
}

// MatchDetails retrieves detailed information about a specific match.
func (c *Client) MatchDetails(ctx context.Context, matchID int) (*api.MatchDetails, error) {
	// Apply rate limiting
	c.rateLimiter.Wait()

	url := fmt.Sprintf("%s/matchDetails?matchId=%d", c.baseURL, matchID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request for match %d: %w", matchID, err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch match details for match %d: %w", matchID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d for match %d", resp.StatusCode, matchID)
	}

	var response fotmobMatchDetails

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode match details response for match %d: %w", matchID, err)
	}

	return response.toAPIMatchDetails(), nil
}

// Leagues retrieves available leagues.
func (c *Client) Leagues(ctx context.Context) ([]api.League, error) {
	// FotMob doesn't have a direct leagues endpoint, so we'll return an empty slice
	// In a real implementation, you might need to maintain a list of known leagues
	// or fetch them from a different endpoint
	return []api.League{}, nil
}

// LeagueMatches retrieves matches for a specific league.
func (c *Client) LeagueMatches(ctx context.Context, leagueID int) ([]api.Match, error) {
	// This would require a different endpoint structure
	// For now, we'll return an empty slice
	// In a real implementation, you'd use: /api/leagues?id={leagueID}
	return []api.Match{}, nil
}

// LeagueTable retrieves the league table/standings for a specific league.
func (c *Client) LeagueTable(ctx context.Context, leagueID int) ([]api.LeagueTableEntry, error) {
	// Apply rate limiting
	c.rateLimiter.Wait()

	url := fmt.Sprintf("%s/leagues?id=%d", c.baseURL, leagueID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request for league %d table: %w", leagueID, err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch league table for league %d: %w", leagueID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d for league %d table", resp.StatusCode, leagueID)
	}

	var response struct {
		Data struct {
			Table []fotmobTableRow `json:"table"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode league table response for league %d: %w", leagueID, err)
	}

	entries := make([]api.LeagueTableEntry, 0, len(response.Data.Table))
	for _, row := range response.Data.Table {
		entries = append(entries, row.toAPITableEntry())
	}

	return entries, nil
}
