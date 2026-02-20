package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

func main() {
	client := &http.Client{Timeout: 30 * time.Second}
	ua := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"

	// Get matches from Premier League
	fmt.Println("=== Get matches with pageUrl from leagues endpoint ===")
	leagueURL := "https://www.fotmob.com/api/leagues?id=47&tab=results"
	req, _ := http.NewRequest("GET", leagueURL, nil)
	req.Header.Set("User-Agent", ua)
	resp, _ := client.Do(req)
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	var leagueResp struct {
		Fixtures struct {
			AllMatches []struct {
				ID      string `json:"id"`
				PageURL string `json:"pageUrl"`
				Home    struct{ Name string `json:"name"` } `json:"home"`
				Away    struct{ Name string `json:"name"` } `json:"away"`
				Status  struct {
					Finished bool   `json:"finished"`
					ScoreStr string `json:"scoreStr"`
				} `json:"status"`
			} `json:"allMatches"`
		} `json:"fixtures"`
	}
	json.Unmarshal(body, &leagueResp)

	// Print first 5 finished matches to see pageUrl patterns
	count := 0
	for _, m := range leagueResp.Fixtures.AllMatches {
		if m.Status.Finished && count < 5 {
			fmt.Printf("  ID: %s, %s vs %s (%s) -> pageUrl: %s\n",
				m.ID, m.Home.Name, m.Away.Name, m.Status.ScoreStr, m.PageURL)
			count++
		}
	}

	// Take a RECENT match (not the first one which might be old)
	// Find the most recent finished match
	var recentMatch struct {
		ID      string
		PageURL string
		Home    string
		Away    string
		Score   string
	}
	for i := len(leagueResp.Fixtures.AllMatches) - 1; i >= 0; i-- {
		m := leagueResp.Fixtures.AllMatches[i]
		if m.Status.Finished && m.Status.ScoreStr != "" {
			recentMatch.ID = m.ID
			recentMatch.PageURL = m.PageURL
			recentMatch.Home = m.Home.Name
			recentMatch.Away = m.Away.Name
			recentMatch.Score = m.Status.ScoreStr
			break
		}
	}

	fmt.Printf("\nMost recent finished: ID=%s, %s vs %s (%s)\n", recentMatch.ID, recentMatch.Home, recentMatch.Away, recentMatch.Score)
	fmt.Printf("PageURL: %s\n\n", recentMatch.PageURL)

	// Get the slug and test the match page
	slug := recentMatch.PageURL
	if idx := strings.Index(slug, "#"); idx != -1 {
		slug = slug[:idx]
	}

	// Test 1: Page with fragment (server ignores fragment)
	fmt.Println("=== Test: fetch match page and check matchId ===")
	fetchAndCheckMatch(client, ua, "https://www.fotmob.com"+slug, recentMatch.ID)

	// Test 2: Try passing matchId as query parameter
	fmt.Println("\n=== Test: /_next/data route with matchId query param ===")
	// First get buildId from the page
	req2, _ := http.NewRequest("GET", "https://www.fotmob.com"+slug, nil)
	req2.Header.Set("User-Agent", ua)
	resp2, _ := client.Do(req2)
	pageBody, _ := io.ReadAll(resp2.Body)
	resp2.Body.Close()

	pageStr := string(pageBody)
	buildIdx := strings.Index(pageStr, `"buildId":"`)
	buildID := ""
	if buildIdx != -1 {
		start := buildIdx + len(`"buildId":"`)
		end := strings.Index(pageStr[start:], `"`) + start
		buildID = pageStr[start:end]
	}

	if buildID != "" {
		// Try with matchId query param
		slugParts := strings.TrimPrefix(slug, "/")
		dataURL := fmt.Sprintf("https://www.fotmob.com/_next/data/%s/%s.json?matchId=%s", buildID, slugParts, recentMatch.ID)
		fmt.Printf("Trying: %s\n", dataURL)
		fetchAndCheckMatch(client, ua, dataURL, recentMatch.ID)
	}

	// Test 3: Check if La Liga (different league) works the same way
	fmt.Println("\n=== Test: La Liga recent match ===")
	laligaURL := "https://www.fotmob.com/api/leagues?id=87&tab=results"
	req3, _ := http.NewRequest("GET", laligaURL, nil)
	req3.Header.Set("User-Agent", ua)
	resp3, _ := client.Do(req3)
	body3, _ := io.ReadAll(resp3.Body)
	resp3.Body.Close()

	var laligaResp struct {
		Fixtures struct {
			AllMatches []struct {
				ID      string `json:"id"`
				PageURL string `json:"pageUrl"`
				Home    struct{ Name string `json:"name"` } `json:"home"`
				Away    struct{ Name string `json:"name"` } `json:"away"`
				Status  struct {
					Finished bool   `json:"finished"`
					ScoreStr string `json:"scoreStr"`
				} `json:"status"`
			} `json:"allMatches"`
		} `json:"fixtures"`
	}
	json.Unmarshal(body3, &laligaResp)

	for i := len(laligaResp.Fixtures.AllMatches) - 1; i >= 0; i-- {
		m := laligaResp.Fixtures.AllMatches[i]
		if m.Status.Finished && m.Status.ScoreStr != "" {
			fmt.Printf("La Liga match: ID=%s, %s vs %s (%s)\n", m.ID, m.Home.Name, m.Away.Name, m.Status.ScoreStr)
			fmt.Printf("PageURL: %s\n", m.PageURL)
			mSlug := m.PageURL
			if idx := strings.Index(mSlug, "#"); idx != -1 {
				mSlug = mSlug[:idx]
			}
			fetchAndCheckMatch(client, ua, "https://www.fotmob.com"+mSlug, m.ID)
			break
		}
	}
}

func fetchAndCheckMatch(client *http.Client, ua string, url string, expectedMatchID string) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", ua)
	req.Header.Set("Accept", "text/html,application/json")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("  Status: %d, Size: %d bytes\n", resp.StatusCode, len(body))

	if resp.StatusCode != 200 {
		if len(body) < 200 {
			fmt.Printf("  Body: %s\n", string(body))
		}
		return
	}

	pageStr := string(body)

	// Try to parse as JSON first (for /_next/data/ routes)
	var jsonResp map[string]any
	if err := json.Unmarshal(body, &jsonResp); err == nil {
		if pageProps, ok := jsonResp["pageProps"].(map[string]any); ok {
			if general, ok := pageProps["general"].(map[string]any); ok {
				matchID := fmt.Sprintf("%v", general["matchId"])
				match := expectedMatchID == matchID
				fmt.Printf("  matchId in response: %s (expected: %s, match: %v)\n", matchID, expectedMatchID, match)
				if header, ok := pageProps["header"].(map[string]any); ok {
					if teams, ok := header["teams"].([]any); ok && len(teams) >= 2 {
						t0 := teams[0].(map[string]any)
						t1 := teams[1].(map[string]any)
						fmt.Printf("  Match: %s %v - %v %s\n", t0["name"], t0["score"], t1["score"], t1["name"])
					}
				}
			}
			return
		}
	}

	// HTML response - extract __NEXT_DATA__
	nextIdx := strings.Index(pageStr, `__NEXT_DATA__`)
	if nextIdx == -1 {
		fmt.Println("  No __NEXT_DATA__ found")
		return
	}
	startIdx := strings.Index(pageStr[nextIdx:], ">") + nextIdx + 1
	endIdx := strings.Index(pageStr[startIdx:], "</script>") + startIdx
	nextDataJSON := pageStr[startIdx:endIdx]

	var nextData map[string]any
	json.Unmarshal([]byte(nextDataJSON), &nextData)

	props := nextData["props"].(map[string]any)
	if pageProps, ok := props["pageProps"].(map[string]any); ok {
		if general, ok := pageProps["general"].(map[string]any); ok {
			matchID := fmt.Sprintf("%v", general["matchId"])
			match := expectedMatchID == matchID
			fmt.Printf("  matchId in response: %s (expected: %s, match: %v)\n", matchID, expectedMatchID, match)
			if header, ok := pageProps["header"].(map[string]any); ok {
				if teams, ok := header["teams"].([]any); ok && len(teams) >= 2 {
					t0 := teams[0].(map[string]any)
					t1 := teams[1].(map[string]any)
					fmt.Printf("  Match: %s %v - %v %s\n", t0["name"], t0["score"], t1["score"], t1["name"])
				}
			}
		}
	}
}
