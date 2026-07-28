package fotmob

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/0xjuanma/golazo/internal/api"
)

const fifaRankingsPath = "/data/fifarankings"

// FIFARanking fetches FotMob's latest official FIFA ranking for men or women.
func (c *Client) FIFARanking(ctx context.Context, gender string) (*api.FIFARanking, error) {
	if gender != "men" && gender != "women" {
		return nil, fmt.Errorf("unsupported FIFA ranking gender %q", gender)
	}

	var periods []struct {
		ID   string `json:"periodId"`
		Name string `json:"periodName"`
	}
	if err := c.getRankingJSON(ctx, fifaRankingsPath+"/period", url.Values{"gender": {gender}}, &periods); err != nil {
		return nil, fmt.Errorf("fetch FIFA ranking periods: %w", err)
	}
	if len(periods) == 0 {
		return nil, fmt.Errorf("FotMob returned no FIFA ranking periods")
	}

	var rows []struct {
		Name           string  `json:"name"`
		ID             int     `json:"id"`
		Rank           int     `json:"rank"`
		TotalPoints    float64 `json:"totalPoints"`
		PreviousPoints float64 `json:"previousPoints"`
		PointsDiff     float64 `json:"pointsDiff"`
		GainedRank     bool    `json:"gainedRank"`
		LostRank       bool    `json:"lostRank"`
	}
	params := url.Values{"gender": {gender}, "periodId": {periods[0].ID}}
	if err := c.getRankingJSON(ctx, fifaRankingsPath+"/ranking", params, &rows); err != nil {
		return nil, fmt.Errorf("fetch FIFA ranking: %w", err)
	}

	ranking := &api.FIFARanking{
		Gender: gender,
		Period: api.FIFARankingPeriod{ID: periods[0].ID, Name: periods[0].Name},
		Teams:  make([]api.FIFARankingEntry, 0, len(rows)),
	}
	for _, row := range rows {
		ranking.Teams = append(ranking.Teams, api.FIFARankingEntry{
			TeamID: row.ID, Name: row.Name, Rank: row.Rank,
			TotalPoints: row.TotalPoints, PreviousPoints: row.PreviousPoints,
			PointsDiff: row.PointsDiff, GainedRank: row.GainedRank, LostRank: row.LostRank,
		})
	}
	return ranking, nil
}

func (c *Client) getRankingJSON(ctx context.Context, path string, params url.Values, dst any) error {
	c.rateLimiter.Wait()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path+"?"+params.Encode(), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", "https://www.fotmob.com/")
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(dst); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}
