package footballdata

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/0xjuanma/golazo/internal/api"
)

const (
	baseURL = "https://v3.football.api-sports.io"
)

// Supported league IDs for API-Sports.io
// Limited to top 5 most popular leagues to optimize API calls
var (
	// SupportedLeagues contains the league IDs that will be queried for matches.
	// API-Sports.io league IDs.
	// Reduced to top 5 leagues to minimize API calls while covering major competitions.
	SupportedLeagues = []int{
		39,  // Premier League
		140, // La Liga
		78,  // Bundesliga
		135, // Serie A
		61,  // Ligue 1
	}
)

// Client implements the api.Client interface for API-Sports.io (free tier)
type Client struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
}

// NewClient creates a new API-Sports.io client.
// apiKey is required for authentication (get one at https://www.api-sports.io/)
func NewClient(apiKey string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: baseURL,
		apiKey:  apiKey,
	}
}

// FinishedMatchesByDateRange retrieves finished matches within a date range.
// This is used for the stats view to show completed matches.
// Queries each date individually and aggregates results, as API-Sports.io date range queries don't work reliably.
func (c *Client) FinishedMatchesByDateRange(ctx context.Context, dateFrom, dateTo time.Time) ([]api.Match, error) {
	var allMatches []api.Match

	// Create a set of supported league IDs for quick lookup
	supportedLeagueSet := make(map[int]bool)
	for _, id := range SupportedLeagues {
		supportedLeagueSet[id] = true
	}

	// API-Sports.io date range queries (from/to) don't work reliably
	// Instead, query each date individually and aggregate results
	// Normalize dates to UTC to avoid timezone issues
	currentDate := dateFrom.UTC()
	dateToUTC := dateTo.UTC()
	var lastErr error
	successCount := 0
	totalDates := 0

	for !currentDate.After(dateToUTC) {
		totalDates++
		dateStr := currentDate.Format("2006-01-02")

		// API-Sports.io league-specific queries don't work reliably (require season parameter which may be incorrect)
		// Instead, query all finished matches for each date and filter by supported leagues
		// This matches the approach used in the analyze script which successfully finds matches
		url := fmt.Sprintf("%s/fixtures?date=%s&status=FT", c.baseURL, dateStr)
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			lastErr = fmt.Errorf("create request for date %s: %w", dateStr, err)
			currentDate = currentDate.AddDate(0, 0, 1)
			continue
		}

		req.Header.Set("x-apisports-key", c.apiKey)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("fetch matches for date %s: %w", dateStr, err)
			currentDate = currentDate.AddDate(0, 0, 1)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			// Read and discard error response body
			io.ReadAll(resp.Body)
			resp.Body.Close()
			lastErr = fmt.Errorf("unexpected status code %d for date %s", resp.StatusCode, dateStr)
			currentDate = currentDate.AddDate(0, 0, 1)
			continue
		}

		var response footballdataMatchesResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			resp.Body.Close()
			lastErr = fmt.Errorf("decode response for date %s: %w", dateStr, err)
			currentDate = currentDate.AddDate(0, 0, 1)
			continue
		}
		resp.Body.Close()

		// Check for API errors (rate limits, etc.)
		if len(response.Errors) > 0 {
			// If we hit rate limit or other errors, continue to allow partial results
			// Track error but don't fail completely
			lastErr = fmt.Errorf("API errors for date %s: %v", dateStr, response.Errors)
			currentDate = currentDate.AddDate(0, 0, 1)
			continue
		}

		successCount++
		// Filter for finished matches in supported leagues
		for _, m := range response.Response {
			// Check if this match is from a supported league
			if supportedLeagueSet[m.League.ID] {
				apiMatch := m.toAPIMatch()
				// Double-check status is finished
				if apiMatch.Status == api.MatchStatusFinished {
					allMatches = append(allMatches, apiMatch)
				}
			}
		}

		// Move to next day
		currentDate = currentDate.AddDate(0, 0, 1)
	}

	// Return error if all dates failed, but allow partial results
	if successCount == 0 && totalDates > 0 {
		return allMatches, fmt.Errorf("failed to fetch matches for any date in range %s to %s: %w", dateFrom.Format("2006-01-02"), dateTo.Format("2006-01-02"), lastErr)
	}

	return allMatches, nil
}

// RecentFinishedMatches retrieves finished matches from the last N days.
// Queries from today going back N-1 days (inclusive).
// Uses UTC timezone to ensure consistent date calculations.
func (c *Client) RecentFinishedMatches(ctx context.Context, days int) ([]api.Match, error) {
	if days <= 0 {
		days = 1 // Default to 1 day if invalid
	}
	// Use UTC to avoid timezone issues
	today := time.Now().UTC()
	dateFrom := today.AddDate(0, 0, -(days - 1)) // Go back (days-1) days to include today
	return c.FinishedMatchesByDateRange(ctx, dateFrom, today)
}

// MatchesByDate retrieves all matches for a specific date.
// Implements api.Client interface.
func (c *Client) MatchesByDate(ctx context.Context, date time.Time) ([]api.Match, error) {
	dateStr := date.Format("2006-01-02")
	url := fmt.Sprintf("%s/fixtures?date=%s", c.baseURL, dateStr)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request for matches on date %s: %w", dateStr, err)
	}

	req.Header.Set("x-apisports-key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch matches for date %s: %w", dateStr, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Read error response body for better error messages
		bodyBytes, _ := io.ReadAll(resp.Body)
		bodyStr := string(bodyBytes)
		if len(bodyStr) > 200 {
			bodyStr = bodyStr[:200]
		}
		return nil, fmt.Errorf("unexpected status code %d for matches on date %s, response: %s", resp.StatusCode, dateStr, bodyStr)
	}

	var response footballdataMatchesResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode matches response for date %s: %w", dateStr, err)
	}

	matches := make([]api.Match, 0, len(response.Response))
	for _, m := range response.Response {
		matches = append(matches, m.toAPIMatch())
	}

	return matches, nil
}

// MatchDetails retrieves detailed information about a specific match.
// Implements api.Client interface.
func (c *Client) MatchDetails(ctx context.Context, matchID int) (*api.MatchDetails, error) {
	url := fmt.Sprintf("%s/fixtures?id=%d", c.baseURL, matchID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request for match %d: %w", matchID, err)
	}

	req.Header.Set("x-apisports-key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch match details for match %d: %w", matchID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Read error response body for better error messages
		bodyBytes, _ := io.ReadAll(resp.Body)
		bodyStr := string(bodyBytes)
		if len(bodyStr) > 200 {
			bodyStr = bodyStr[:200]
		}
		return nil, fmt.Errorf("unexpected status code %d for match %d, response: %s", resp.StatusCode, matchID, bodyStr)
	}

	var response struct {
		Response []footballdataMatch `json:"response"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode match details response for match %d: %w", matchID, err)
	}

	if len(response.Response) == 0 {
		return nil, fmt.Errorf("match not found")
	}

	match := response.Response[0]
	baseMatch := match.toAPIMatch()
	details := &api.MatchDetails{
		Match:  baseMatch,
		Events: []api.MatchEvent{}, // Events would need separate endpoint call
	}

	// Add half-time score
	if match.Score.HalfTime.Home != nil || match.Score.HalfTime.Away != nil {
		details.HalfTimeScore = &struct {
			Home *int `json:"home,omitempty"`
			Away *int `json:"away,omitempty"`
		}{
			Home: match.Score.HalfTime.Home,
			Away: match.Score.HalfTime.Away,
		}
	}

	// Add venue
	if match.Fixture.Venue.Name != "" {
		details.Venue = match.Fixture.Venue.Name
	}

	// Add winner indicator
	if match.Teams.Home.Winner != nil && *match.Teams.Home.Winner {
		winner := "home"
		details.Winner = &winner
	} else if match.Teams.Away.Winner != nil && *match.Teams.Away.Winner {
		winner := "away"
		details.Winner = &winner
	}

	// Add match duration and extra time/penalties info
	if match.Fixture.Status.Short == "AET" || match.Fixture.Status.Short == "PEN" {
		details.ExtraTime = true
		details.MatchDuration = 120
	} else {
		details.MatchDuration = 90
	}

	// Add penalties if available
	if match.Score.Penalty.Home != nil || match.Score.Penalty.Away != nil {
		details.Penalties = &struct {
			Home *int `json:"home,omitempty"`
			Away *int `json:"away,omitempty"`
		}{
			Home: match.Score.Penalty.Home,
			Away: match.Score.Penalty.Away,
		}
	}

	return details, nil
}

// Leagues retrieves available leagues.
// Implements api.Client interface.
func (c *Client) Leagues(ctx context.Context) ([]api.League, error) {
	// Football-Data.org doesn't have a simple leagues endpoint
	// Would need to query competitions endpoint
	return []api.League{}, nil
}

// LeagueMatches retrieves matches for a specific league.
// Implements api.Client interface.
func (c *Client) LeagueMatches(ctx context.Context, leagueID int) ([]api.Match, error) {
	url := fmt.Sprintf("%s/fixtures?league=%d", c.baseURL, leagueID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request for league %d matches: %w", leagueID, err)
	}

	req.Header.Set("x-apisports-key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch matches for league %d: %w", leagueID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Read error response body for better error messages
		bodyBytes, _ := io.ReadAll(resp.Body)
		bodyStr := string(bodyBytes)
		if len(bodyStr) > 200 {
			bodyStr = bodyStr[:200]
		}
		return nil, fmt.Errorf("unexpected status code %d for league %d matches, response: %s", resp.StatusCode, leagueID, bodyStr)
	}

	var response footballdataMatchesResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode league matches response for league %d: %w", leagueID, err)
	}

	matches := make([]api.Match, 0, len(response.Response))
	for _, m := range response.Response {
		matches = append(matches, m.toAPIMatch())
	}

	return matches, nil
}

// LeagueTable retrieves the league table/standings for a specific league.
// Implements api.Client interface.
func (c *Client) LeagueTable(ctx context.Context, leagueID int) ([]api.LeagueTableEntry, error) {
	url := fmt.Sprintf("%s/standings?league=%d&season=2024", c.baseURL, leagueID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request for league %d table: %w", leagueID, err)
	}

	req.Header.Set("x-apisports-key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch league table for league %d: %w", leagueID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Read error response body for better error messages
		bodyBytes, _ := io.ReadAll(resp.Body)
		bodyStr := string(bodyBytes)
		if len(bodyStr) > 200 {
			bodyStr = bodyStr[:200]
		}
		return nil, fmt.Errorf("unexpected status code %d for league %d table, response: %s", resp.StatusCode, leagueID, bodyStr)
	}

	var response struct {
		Response []struct {
			League struct {
				Standings [][]struct {
					Rank int `json:"rank"`
					Team struct {
						ID   int    `json:"id"`
						Name string `json:"name"`
						Logo string `json:"logo"`
					} `json:"team"`
					All struct {
						Played int `json:"played"`
						Win    int `json:"win"`
						Draw   int `json:"draw"`
						Lose   int `json:"lose"`
						Goals  struct {
							For     int `json:"for"`
							Against int `json:"against"`
						} `json:"goals"`
					} `json:"all"`
					GoalsDiff int `json:"goalsDiff"`
					Points    int `json:"points"`
				} `json:"standings"`
			} `json:"league"`
		} `json:"response"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode league table response for league %d: %w", leagueID, err)
	}

	if len(response.Response) == 0 || len(response.Response[0].League.Standings) == 0 {
		return []api.LeagueTableEntry{}, nil
	}

	var entries []api.LeagueTableEntry
	for _, row := range response.Response[0].League.Standings[0] {
		entries = append(entries, api.LeagueTableEntry{
			Position: row.Rank,
			Team: api.Team{
				ID:   row.Team.ID,
				Name: row.Team.Name,
				Logo: row.Team.Logo,
			},
			Played:         row.All.Played,
			Won:            row.All.Win,
			Drawn:          row.All.Draw,
			Lost:           row.All.Lose,
			GoalsFor:       row.All.Goals.For,
			GoalsAgainst:   row.All.Goals.Against,
			GoalDifference: row.GoalsDiff,
			Points:         row.Points,
		})
	}

	return entries, nil
}
