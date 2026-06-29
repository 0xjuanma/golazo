package fotmob

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/0xjuanma/golazo/internal/api"
)

// wcPageResponse is the parsed shape of FotMob's World Cup league page __NEXT_DATA__.
// FotMob serves playoff data at pageProps.playoff (top-level), not pageProps.overview.playoff.
type wcPageResponse struct {
	Table []struct {
		Data struct {
			Composite bool `json:"composite"`
			Tables    []struct {
				LeagueID   int    `json:"leagueId"`
				LeagueName string `json:"leagueName"`
				Table      struct {
					All []fotmobTableRow `json:"all"`
				} `json:"table"`
			} `json:"tables"`
		} `json:"data"`
	} `json:"table"`
	Playoff  wcPlayoff `json:"playoff"`
	Overview struct {
		SelectedSeason string `json:"selectedSeason"`
	} `json:"overview"`
	Stats struct {
		Players []struct {
			Header      string `json:"header"`
			Name        string `json:"name"`
			FetchAllURL string `json:"fetchAllUrl"`
		} `json:"players"`
	} `json:"stats"`
}

// wcStatList is the shape of the data.fotmob.com stats JSON endpoint.
type wcStatList struct {
	TopLists []struct {
		StatList []struct {
			ParticipantName string  `json:"ParticipantName"`
			TeamName        string  `json:"TeamName"`
			StatValue       float64 `json:"StatValue"`
			Rank            int     `json:"Rank"`
		} `json:"StatList"`
	} `json:"TopLists"`
}

type wcPlayoff struct {
	Rounds  []wcPlayoffRound `json:"rounds"`
	Special []wcPlayoffRound `json:"special"`
}

type wcPlayoffRound struct {
	Stage    string        `json:"stage"`
	Matchups []wcMatchupRaw `json:"matchups"`
}

type wcMatchupRaw struct {
	HomeTeam          string `json:"homeTeam"`
	HomeTeamID        int    `json:"homeTeamId"`
	HomeTeamShortName string `json:"homeTeamShortName"`
	AwayTeam          string `json:"awayTeam"`
	AwayTeamID        int    `json:"awayTeamId"`
	AwayTeamShortName string `json:"awayTeamShortName"`
	TBDTeam1          bool   `json:"tbdTeam1"`
	TBDTeam2          bool   `json:"tbdTeam2"`
	Matches           []struct {
		Home struct {
			Score  int  `json:"score"`
			Winner bool `json:"winner"`
		} `json:"home"`
		Away struct {
			Score  int  `json:"score"`
			Winner bool `json:"winner"`
		} `json:"away"`
		Status struct {
			Finished bool `json:"finished"`
		} `json:"status"`
	} `json:"matches"`
}

// WorldCupData fetches and parses the current FIFA World Cup data from FotMob.
// Pass season as "2022", "2026", etc. to fetch a specific year; pass "" for the
// current/latest season.
func (c *Client) WorldCupData(ctx context.Context, season string) (*api.WorldCupData, error) {
	c.rateLimiter.Wait()

	url := "https://www.fotmob.com/leagues/77/overview/world-cup"
	if season != "" {
		url += "?season=" + season
	}
	c.debugLog("WorldCupData: fetching", "url", url, "season", season)

	pageProps, err := fetchWorldCupPage(ctx, c.httpClient, season)
	if err != nil {
		return nil, fmt.Errorf("fetch world cup page: %w", err)
	}

	var resp wcPageResponse
	if err := json.Unmarshal(pageProps, &resp); err != nil {
		return nil, fmt.Errorf("parse world cup page props: %w", err)
	}

	groups := parseWCGroups(resp)
	rounds, bronze := parseWCBracket(resp.Playoff)

	s := season
	if s == "" && resp.Overview.SelectedSeason != "" {
		s = resp.Overview.SelectedSeason
	}

	wcData := &api.WorldCupData{
		Season:         s,
		Name:           fmt.Sprintf("FIFA World Cup %s", s),
		Groups:         groups,
		KnockoutRounds: rounds,
		BronzeFinal:    bronze,
	}
	wcData.Champion, wcData.RunnerUp = wcData.DeriveFinalists()
	return wcData, nil
}

// fetchWorldCupPage fetches the FotMob World Cup league overview page.
func fetchWorldCupPage(ctx context.Context, httpClient *http.Client, season string) (json.RawMessage, error) {
	url := "https://www.fotmob.com/leagues/77/overview/world-cup"
	if season != "" {
		url += "?season=" + season
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create world cup page request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch world cup page: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("world cup page returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read world cup page body: %w", err)
	}

	return extractPageProps(string(body))
}

// parseWCGroups extracts group standings from the page response.
// Handles both 8-group (Qatar 2022) and 12-group (USA 2026) formats.
//
// FotMob occasionally ships pseudo-tables alongside the real groups (e.g.
// "Qualified teams", "Pot teams") whose derived "letter" is a word rather
// than a single character. Those are filtered out so the UI grid stays
// symmetric (#158).
func parseWCGroups(resp wcPageResponse) []api.WCGroup {
	if len(resp.Table) == 0 {
		return nil
	}
	tables := resp.Table[0].Data.Tables
	groups := make([]api.WCGroup, 0, len(tables))

	for _, t := range tables {
		if len(t.Table.All) == 0 {
			continue
		}

		// Derive group letter from leagueName ("Grp. A" → "A")
		letter := wcGroupLetter(t.LeagueName)
		if !isWCGroupLetter(letter) {
			continue
		}

		entries := make([]api.LeagueTableEntry, 0, len(t.Table.All))
		for _, row := range t.Table.All {
			// Skip placeholder rows (TBD slots, qualifier annotations)
			// that FotMob occasionally interleaves with real teams. They
			// would otherwise inflate the rendered cell height and push
			// the grid out of alignment (#158).
			if strings.TrimSpace(row.Name) == "" && strings.TrimSpace(row.ShortName) == "" {
				continue
			}
			entries = append(entries, row.toAPITableEntry())
		}

		groups = append(groups, api.WCGroup{
			ID:     t.LeagueID,
			Letter: letter,
			Name:   "Group " + letter,
			Teams:  entries,
		})
	}
	return groups
}

// isWCGroupLetter reports whether s is a single uppercase ASCII letter, the
// shape every real WC group identifier takes (A–L). Anything else is a
// FotMob pseudo-table (e.g. "teams" from "Qualified teams") and must be
// dropped before reaching the renderer.
func isWCGroupLetter(s string) bool {
	if len(s) != 1 {
		return false
	}
	c := s[0]
	return c >= 'A' && c <= 'Z'
}

// parseWCBracket extracts the knockout bracket from the playoff data.
// Returns ordered rounds (excluding bronze) and the bronze final matchup.
func parseWCBracket(playoff wcPlayoff) ([]api.WCKnockoutRound, *api.WCMatchup) {
	stageOrder := map[string]int{
		"1/32": 0, // Round of 64 (hypothetical)
		"1/16": 1, // Round of 32
		"1/8":  2, // Round of 16
		"1/4":  3, // Quarterfinals
		"1/2":  4, // Semifinals
		"final": 5,
	}
	stageLabels := map[string]string{
		"1/32":  "Round of 64",
		"1/16":  "Round of 32",
		"1/8":   "Round of 16",
		"1/4":   "Quarterfinals",
		"1/2":   "Semifinals",
		"final": "Final",
	}

	// Sort rounds by their defined order
	type indexedRound struct {
		order int
		round api.WCKnockoutRound
	}
	indexed := make([]indexedRound, 0, len(playoff.Rounds))

	for _, r := range playoff.Rounds {
		label, ok := stageLabels[r.Stage]
		if !ok {
			label = r.Stage
		}
		order, ok := stageOrder[r.Stage]
		if !ok {
			order = 99
		}
		round := api.WCKnockoutRound{
			Stage:    r.Stage,
			Label:    label,
			Matchups: convertMatchups(r.Matchups),
		}
		indexed = append(indexed, indexedRound{order: order, round: round})
	}

	sort.Slice(indexed, func(i, j int) bool { return indexed[i].order < indexed[j].order })

	rounds := make([]api.WCKnockoutRound, 0, len(indexed))
	for _, ir := range indexed {
		rounds = append(rounds, ir.round)
	}

	// Extract bronze final from special
	var bronze *api.WCMatchup
	for _, s := range playoff.Special {
		if s.Stage == "bronze" && len(s.Matchups) > 0 {
			m := convertMatchup(s.Matchups[0])
			bronze = &m
			break
		}
	}

	return rounds, bronze
}

// convertMatchups converts raw FotMob matchups to API matchups.
func convertMatchups(raw []wcMatchupRaw) []api.WCMatchup {
	out := make([]api.WCMatchup, 0, len(raw))
	for _, r := range raw {
		out = append(out, convertMatchup(r))
	}
	return out
}

// convertMatchup converts a single raw FotMob matchup to an API matchup.
func convertMatchup(r wcMatchupRaw) api.WCMatchup {
	m := api.WCMatchup{
		HomeTeam:   r.HomeTeam,
		HomeTeamID: r.HomeTeamID,
		HomeShort:  r.HomeTeamShortName,
		AwayTeam:   r.AwayTeam,
		AwayTeamID: r.AwayTeamID,
		AwayShort:  r.AwayTeamShortName,
		TBDHome:    r.TBDTeam1,
		TBDAway:    r.TBDTeam2,
	}

	if len(r.Matches) > 0 && r.Matches[0].Status.Finished {
		m.HomeScore = intPtr(r.Matches[0].Home.Score)
		m.AwayScore = intPtr(r.Matches[0].Away.Score)

		if r.Matches[0].Home.Winner {
			m.WinnerID = intPtr(r.HomeTeamID)
		} else if r.Matches[0].Away.Winner {
			m.WinnerID = intPtr(r.AwayTeamID)
		}

		// Detect penalties: scores level at final whistle but there's a winner
		if m.WinnerID != nil && *m.HomeScore == *m.AwayScore {
			m.IsPenalties = true
		}
	}

	return m
}

// wcGroupLetter derives the group letter from FotMob's group name.
// "Grp. A" → "A", "Group A" → "A"
func wcGroupLetter(name string) string {
	name = strings.TrimSpace(name)
	// "Grp. A" → last non-space char sequence
	parts := strings.Fields(name)
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return name
}

func intPtr(v int) *int { i := v; return &i }

// WorldCupTopScorers fetches the top scorers for the current World Cup season.
// It reuses the same page fetch as WorldCupData to obtain the stats URL, then
// makes one additional call to data.fotmob.com for the full ranked list.
func (c *Client) WorldCupTopScorers(ctx context.Context, season string) ([]api.WCTopScorer, error) {
	c.rateLimiter.Wait()

	pageProps, err := fetchWorldCupPage(ctx, c.httpClient, season)
	if err != nil {
		return nil, fmt.Errorf("fetch world cup page for scorers: %w", err)
	}

	var resp wcPageResponse
	if err := json.Unmarshal(pageProps, &resp); err != nil {
		return nil, fmt.Errorf("parse world cup page props for scorers: %w", err)
	}

	// Find the "Top scorer" / "goals" stat entry to get the full-list URL.
	fetchURL := ""
	for _, p := range resp.Stats.Players {
		if p.Name == "goals" {
			fetchURL = p.FetchAllURL
			break
		}
	}
	if fetchURL == "" {
		return nil, fmt.Errorf("top scorer stat URL not found in world cup page")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", fetchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create top scorers request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "https://www.fotmob.com/")

	statResp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch top scorers data: %w", err)
	}
	defer func() { _ = statResp.Body.Close() }()

	if statResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("top scorers endpoint returned status %d", statResp.StatusCode)
	}

	var statList wcStatList
	if err := json.NewDecoder(statResp.Body).Decode(&statList); err != nil {
		return nil, fmt.Errorf("decode top scorers response: %w", err)
	}

	if len(statList.TopLists) == 0 {
		return nil, nil
	}
	return parseWCTopScorers(statList), nil
}

// parseWCTopScorers converts a raw stat list response into API top scorer entries.
func parseWCTopScorers(statList wcStatList) []api.WCTopScorer {
	if len(statList.TopLists) == 0 {
		return nil
	}
	entries := statList.TopLists[0].StatList
	scorers := make([]api.WCTopScorer, 0, len(entries))
	for _, e := range entries {
		scorers = append(scorers, api.WCTopScorer{
			PlayerName: e.ParticipantName,
			Team:       e.TeamName,
			Goals:      int(e.StatValue),
		})
	}
	return scorers
}
