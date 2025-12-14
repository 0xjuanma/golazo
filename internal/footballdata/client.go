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
// These are the most popular leagues that typically have matches
var (
	// SupportedLeagues contains the league IDs that will be queried for matches.
	// API-Sports.io league IDs:
	//   - Premier League: 39
	//   - La Liga: 140
	//   - Bundesliga: 78
	//   - Serie A: 135
	//   - Ligue 1: 61
	//   - Champions League: 2
	//   - MLS: 253
	SupportedLeagues = []int{
		39,  // Premier League
		140, // La Liga
		78,  // Bundesliga
		135, // Serie A
		61,  // Ligue 1
		2,   // Champions League
		253, // MLS
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
// Queries multiple supported leagues and aggregates results.
func (c *Client) FinishedMatchesByDateRange(ctx context.Context, dateFrom, dateTo time.Time) ([]api.Match, error) {
	dateFromStr := dateFrom.Format("2006-01-02")
	dateToStr := dateTo.Format("2006-01-02")

	allMatches := make([]api.Match, 0)

	// Query each supported league and aggregate matches
	// API-Sports.io free tier may require league parameter
	for _, leagueID := range SupportedLeagues {
		// API-Sports.io uses /fixtures endpoint with date range
		// Use 'from' and 'to' parameters for date range, and filter for finished matches (status=FT)
		url := fmt.Sprintf("%s/fixtures?from=%s&to=%s&status=FT&league=%d", c.baseURL, dateFromStr, dateToStr, leagueID)

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			continue // Skip this league on error
		}

		req.Header.Set("x-apisports-key", c.apiKey)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			continue // Skip this league on request error
		}

		if resp.StatusCode != http.StatusOK {
			// Read error response body for debugging (but don't fail completely)
			io.ReadAll(resp.Body) // Discard response body
			resp.Body.Close()
			// Continue to next league instead of failing completely
			continue
		}

		var response footballdataMatchesResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			resp.Body.Close()
			continue // Skip this league on parse error
		}
		resp.Body.Close()

		// Convert all matches (already filtered by status=FT in the API call)
		for _, m := range response.Response {
			apiMatch := m.toAPIMatch()
			// Double-check status is finished (should already be filtered by API)
			if apiMatch.Status == api.MatchStatusFinished {
				allMatches = append(allMatches, apiMatch)
			}
		}
	}

	return allMatches, nil
}

// RecentFinishedMatches retrieves finished matches from the last N days.
// Defaults to last 7 days if days is 0.
func (c *Client) RecentFinishedMatches(ctx context.Context, days int) ([]api.Match, error) {
	if days == 0 {
		days = 7
	}

	dateTo := time.Now()
	dateFrom := dateTo.AddDate(0, 0, -days)

	return c.FinishedMatchesByDateRange(ctx, dateFrom, dateTo)
}

// MatchesByDate retrieves all matches for a specific date.
// Implements api.Client interface.
func (c *Client) MatchesByDate(ctx context.Context, date time.Time) ([]api.Match, error) {
	dateStr := date.Format("2006-01-02")
	url := fmt.Sprintf("%s/fixtures?date=%s", c.baseURL, dateStr)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("x-apisports-key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch matches: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Read error response body for better error messages
		bodyBytes, _ := io.ReadAll(resp.Body)
		bodyStr := string(bodyBytes)
		if len(bodyStr) > 200 {
			bodyStr = bodyStr[:200]
		}
		return nil, fmt.Errorf("unexpected status code: %d, response: %s", resp.StatusCode, bodyStr)
	}

	var response footballdataMatchesResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
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
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("x-apisports-key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch match details: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Read error response body for better error messages
		bodyBytes, _ := io.ReadAll(resp.Body)
		bodyStr := string(bodyBytes)
		if len(bodyStr) > 200 {
			bodyStr = bodyStr[:200]
		}
		return nil, fmt.Errorf("unexpected status code: %d, response: %s", resp.StatusCode, bodyStr)
	}

	var response struct {
		Response []footballdataMatch `json:"response"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
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
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("x-apisports-key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch league matches: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Read error response body for better error messages
		bodyBytes, _ := io.ReadAll(resp.Body)
		bodyStr := string(bodyBytes)
		if len(bodyStr) > 200 {
			bodyStr = bodyStr[:200]
		}
		return nil, fmt.Errorf("unexpected status code: %d, response: %s", resp.StatusCode, bodyStr)
	}

	var response footballdataMatchesResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
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
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("x-apisports-key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch league table: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Read error response body for better error messages
		bodyBytes, _ := io.ReadAll(resp.Body)
		bodyStr := string(bodyBytes)
		if len(bodyStr) > 200 {
			bodyStr = bodyStr[:200]
		}
		return nil, fmt.Errorf("unexpected status code: %d, response: %s", resp.StatusCode, bodyStr)
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
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Response) == 0 || len(response.Response[0].League.Standings) == 0 {
		return []api.LeagueTableEntry{}, nil
	}

	entries := make([]api.LeagueTableEntry, 0)
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
