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

	// Get a match from leagues endpoint
	fmt.Println("=== Fetching match from leagues endpoint ===")
	leagueURL := "https://www.fotmob.com/api/leagues?id=47&tab=results"
	req, _ := http.NewRequest("GET", leagueURL, nil)
	req.Header.Set("User-Agent", ua)
	resp, _ := client.Do(req)
	leagueBody, _ := io.ReadAll(resp.Body)
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
	json.Unmarshal(leagueBody, &leagueResp)

	// Find recent finished match
	var pageURL string
	for _, m := range leagueResp.Fixtures.AllMatches {
		if m.Status.Finished && m.Status.ScoreStr != "" {
			pageURL = m.PageURL
			fmt.Printf("Match: %s vs %s (ID: %s, Score: %s)\n", m.Home.Name, m.Away.Name, m.ID, m.Status.ScoreStr)
			fmt.Printf("PageURL: %s\n\n", m.PageURL)
			break
		}
	}

	// Fetch via match page HTML and extract __NEXT_DATA__
	fmt.Println("=== Fetching match page for __NEXT_DATA__ ===")
	slug := pageURL
	if idx := strings.Index(slug, "#"); idx != -1 {
		slug = slug[:idx]
	}
	fullURL := "https://www.fotmob.com" + slug
	req2, _ := http.NewRequest("GET", fullURL, nil)
	req2.Header.Set("User-Agent", ua)
	resp2, _ := client.Do(req2)
	pageBody, _ := io.ReadAll(resp2.Body)
	resp2.Body.Close()
	fmt.Printf("Status: %d, Size: %d\n", resp2.StatusCode, len(pageBody))

	// Extract __NEXT_DATA__
	pageStr := string(pageBody)
	nextIdx := strings.Index(pageStr, `__NEXT_DATA__`)
	startIdx := strings.Index(pageStr[nextIdx:], ">") + nextIdx + 1
	endIdx := strings.Index(pageStr[startIdx:], "</script>") + startIdx
	nextDataJSON := pageStr[startIdx:endIdx]

	var nextData map[string]any
	json.Unmarshal([]byte(nextDataJSON), &nextData)

	props := nextData["props"].(map[string]any)
	pageProps := props["pageProps"].(map[string]any)

	// Verify the structure matches what fotmobMatchDetails expects
	fmt.Println("\n=== Verifying structure matches fotmobMatchDetails ===")

	// Check 'header' structure
	if header, ok := pageProps["header"].(map[string]any); ok {
		fmt.Println("\n--- header ---")
		if teams, ok := header["teams"].([]any); ok {
			fmt.Printf("  teams count: %d\n", len(teams))
			for i, t := range teams {
				team := t.(map[string]any)
				fmt.Printf("  team[%d]: name=%v, score=%v, id=%v\n", i, team["name"], team["score"], team["id"])
			}
		}
		if status, ok := header["status"].(map[string]any); ok {
			fmt.Printf("  status: finished=%v, started=%v, utcTime=%v\n", status["finished"], status["started"], status["utcTime"])
			if lt, ok := status["liveTime"]; ok {
				fmt.Printf("  liveTime: %v\n", lt)
			}
			if sc, ok := status["scoreStr"]; ok {
				fmt.Printf("  scoreStr: %v\n", sc)
			}
		}
	}

	// Check 'general' structure
	if general, ok := pageProps["general"].(map[string]any); ok {
		fmt.Println("\n--- general ---")
		fmt.Printf("  matchId: %v (type: %T)\n", general["matchId"], general["matchId"])
		fmt.Printf("  leagueId: %v\n", general["leagueId"])
		fmt.Printf("  leagueName: %v\n", general["leagueName"])
		if ht, ok := general["homeTeam"].(map[string]any); ok {
			fmt.Printf("  homeTeam: id=%v, name=%v\n", ht["id"], ht["name"])
		}
		if at, ok := general["awayTeam"].(map[string]any); ok {
			fmt.Printf("  awayTeam: id=%v, name=%v\n", at["id"], at["name"])
		}
		fmt.Printf("  matchRound: %v\n", general["matchRound"])
		fmt.Printf("  parentLeagueId: %v\n", general["parentLeagueId"])
	}

	// Check 'content' structure
	if content, ok := pageProps["content"].(map[string]any); ok {
		fmt.Println("\n--- content ---")
		contentKeys := make([]string, 0, len(content))
		for k := range content {
			contentKeys = append(contentKeys, k)
		}
		fmt.Printf("  keys: %v\n", contentKeys)

		if mf, ok := content["matchFacts"].(map[string]any); ok {
			fmt.Println("  matchFacts present")
			mfKeys := make([]string, 0, len(mf))
			for k := range mf {
				mfKeys = append(mfKeys, k)
			}
			fmt.Printf("    keys: %v\n", mfKeys)

			if events, ok := mf["events"].(map[string]any); ok {
				if evList, ok := events["events"].([]any); ok {
					fmt.Printf("    events count: %d\n", len(evList))
					if len(evList) > 0 {
						ev0, _ := json.MarshalIndent(evList[0], "      ", "  ")
						fmt.Printf("    first event: %s\n", string(ev0))
					}
				}
			}
			if hl, ok := mf["highlights"]; ok && hl != nil {
				fmt.Printf("    highlights: %v\n", hl)
			}
			if ib, ok := mf["infoBox"].(map[string]any); ok {
				fmt.Printf("    infoBox keys: ")
				for k := range ib {
					fmt.Printf("%s ", k)
				}
				fmt.Println()
				if stadium, ok := ib["Stadium"].(map[string]any); ok {
					fmt.Printf("    stadium: %v\n", stadium["name"])
				}
				if ref, ok := ib["Referee"].(map[string]any); ok {
					fmt.Printf("    referee: %v\n", ref["text"])
				}
			}
		}

		if stats, ok := content["stats"].(map[string]any); ok {
			fmt.Println("  stats present")
			if periods, ok := stats["Ede"].(map[string]any); ok {
				fmt.Printf("    periods keys: %v\n", periods)
			}
			statsJSON, _ := json.MarshalIndent(stats, "    ", "  ")
			sj := string(statsJSON)
			if len(sj) > 1000 {
				sj = sj[:1000] + "..."
			}
			fmt.Printf("    stats structure: %s\n", sj)
		}

		if lineup, ok := content["lineup"].(map[string]any); ok {
			fmt.Println("  lineup present")
			lineupKeys := make([]string, 0, len(lineup))
			for k := range lineup {
				lineupKeys = append(lineupKeys, k)
			}
			fmt.Printf("    keys: %v\n", lineupKeys)
		}
	}

	fmt.Println("\n=== Also check scoreStr field in leagues match listing ===")
	for _, m := range leagueResp.Fixtures.AllMatches {
		if m.Status.Finished && m.Status.ScoreStr != "" {
			fmt.Printf("Match %s: scoreStr=%q\n", m.ID, m.Status.ScoreStr)
			break
		}
	}

	fmt.Println("\n=== SUMMARY ===")
	fmt.Println("The pageProps from __NEXT_DATA__ / Next.js data route contains")
	fmt.Println("the SAME structure as the old /api/matchDetails endpoint:")
	fmt.Println("  - header.teams[].name, header.teams[].score")
	fmt.Println("  - header.status (finished, started, utcTime, liveTime)")
	fmt.Println("  - general (matchId, leagueId, leagueName, homeTeam, awayTeam)")
	fmt.Println("  - content.matchFacts (events, highlights, infoBox)")
	fmt.Println("  - content.stats, content.lineup")
}
