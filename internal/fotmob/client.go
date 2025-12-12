package fotmob

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/0xjuanma/golazo/internal/api"
)

const (
	baseURL = "https://www.fotmob.com/api"
)

// Client implements the api.Client interface for FotMob API
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a new FotMob API client
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: baseURL,
	}
}

// GetMatchesByDate retrieves all matches for a specific date
func (c *Client) GetMatchesByDate(ctx context.Context, date time.Time) ([]api.Match, error) {
	dateStr := date.Format("20060102")
	url := fmt.Sprintf("%s/matches?date=%s", c.baseURL, dateStr)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch matches: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var response struct {
		Matches []fotmobMatch `json:"matches"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	matches := make([]api.Match, 0, len(response.Matches))
	for _, m := range response.Matches {
		matches = append(matches, m.toAPIMatch())
	}

	return matches, nil
}

// GetMatchDetails retrieves detailed information about a specific match
func (c *Client) GetMatchDetails(ctx context.Context, matchID int) (*api.MatchDetails, error) {
	url := fmt.Sprintf("%s/matchDetails?matchId=%d", c.baseURL, matchID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch match details: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var response fotmobMatchDetails

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.toAPIMatchDetails(), nil
}

// GetLeagues retrieves available leagues
func (c *Client) GetLeagues(ctx context.Context) ([]api.League, error) {
	// FotMob doesn't have a direct leagues endpoint, so we'll return an empty slice
	// In a real implementation, you might need to maintain a list of known leagues
	// or fetch them from a different endpoint
	return []api.League{}, nil
}

// GetLeagueMatches retrieves matches for a specific league
func (c *Client) GetLeagueMatches(ctx context.Context, leagueID int) ([]api.Match, error) {
	// This would require a different endpoint structure
	// For now, we'll return an empty slice
	// In a real implementation, you'd use: /api/leagues?id={leagueID}
	return []api.Match{}, nil
}

// GetLeagueTable retrieves the league table/standings for a specific league
func (c *Client) GetLeagueTable(ctx context.Context, leagueID int) ([]api.LeagueTableEntry, error) {
	url := fmt.Sprintf("%s/leagues?id=%d", c.baseURL, leagueID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch league table: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var response struct {
		Data struct {
			Table []fotmobTableRow `json:"table"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	entries := make([]api.LeagueTableEntry, 0, len(response.Data.Table))
	for _, row := range response.Data.Table {
		entries = append(entries, row.toAPITableEntry())
	}

	return entries, nil
}

