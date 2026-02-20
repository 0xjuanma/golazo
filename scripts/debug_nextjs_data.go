package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

func main() {
	matchID := "4813684"
	if len(os.Args) > 1 {
		matchID = os.Args[1]
	}

	client := &http.Client{Timeout: 30 * time.Second}
	ua := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"

	// Step 1: Get a valid match pageUrl from the leagues endpoint
	fmt.Println("=== Step 1: Get match pageUrl from leagues endpoint ===")
	leagueURL := "https://www.fotmob.com/api/leagues?id=47&tab=results"
	req, _ := http.NewRequest("GET", leagueURL, nil)
	req.Header.Set("User-Agent", ua)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	leagueBody, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	var leagueResp struct {
		Fixtures struct {
			AllMatches []struct {
				ID      string `json:"id"`
				PageURL string `json:"pageUrl"`
				Home    struct {
					Name string `json:"name"`
				} `json:"home"`
				Away struct {
					Name string `json:"name"`
				} `json:"away"`
				Status struct {
					Finished bool   `json:"finished"`
					ScoreStr string `json:"scoreStr"`
				} `json:"status"`
			} `json:"allMatches"`
		} `json:"fixtures"`
	}
	json.Unmarshal(leagueBody, &leagueResp)

	// Find a finished match
	var matchPageURL string
	var targetMatchID string
	for _, m := range leagueResp.Fixtures.AllMatches {
		if m.Status.Finished && m.Status.ScoreStr != "" {
			matchPageURL = m.PageURL
			targetMatchID = m.ID
			fmt.Printf("Found finished match: %s vs %s (ID: %s, Score: %s)\n", m.Home.Name, m.Away.Name, m.ID, m.Status.ScoreStr)
			fmt.Printf("Page URL: %s\n", m.PageURL)
			break
		}
	}

	if matchPageURL == "" {
		fmt.Println("No finished matches found")
		return
	}

	_ = matchID // use targetMatchID from leagues

	// Step 2: Fetch the match page HTML using the pageUrl
	fmt.Println("\n=== Step 2: Fetch match page HTML ===")
	fullPageURL := "https://www.fotmob.com" + matchPageURL
	// Remove the fragment (#matchID)
	if idx := strings.Index(fullPageURL, "#"); idx != -1 {
		fullPageURL = fullPageURL[:idx]
	}
	fmt.Printf("Fetching: %s\n", fullPageURL)

	req2, _ := http.NewRequest("GET", fullPageURL, nil)
	req2.Header.Set("User-Agent", ua)
	req2.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	resp2, err := client.Do(req2)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp2.Body.Close()

	pageBody, _ := io.ReadAll(resp2.Body)
	pageStr := string(pageBody)
	fmt.Printf("Status: %d, Size: %d bytes\n", resp2.StatusCode, len(pageBody))

	// Look for __NEXT_DATA__ which contains server-rendered data
	nextDataIdx := strings.Index(pageStr, `__NEXT_DATA__`)
	if nextDataIdx != -1 {
		// Extract the JSON content between > and </script>
		startIdx := strings.Index(pageStr[nextDataIdx:], ">")
		if startIdx != -1 {
			startIdx += nextDataIdx + 1
			endIdx := strings.Index(pageStr[startIdx:], "</script>")
			if endIdx != -1 {
				nextDataJSON := pageStr[startIdx : startIdx+endIdx]
				fmt.Printf("Found __NEXT_DATA__ (%d bytes)\n", len(nextDataJSON))

				var nextData map[string]any
				if err := json.Unmarshal([]byte(nextDataJSON), &nextData); err == nil {
					// Check if match data is in props.pageProps
					if props, ok := nextData["props"].(map[string]any); ok {
						if pageProps, ok := props["pageProps"].(map[string]any); ok {
							keys := make([]string, 0, len(pageProps))
							for k := range pageProps {
								keys = append(keys, k)
							}
							fmt.Printf("pageProps keys: %v\n", keys)

							// Print some match data if available
							prettyJSON, _ := json.MarshalIndent(pageProps, "", "  ")
							ppStr := string(prettyJSON)
							if len(ppStr) > 2000 {
								fmt.Printf("pageProps (first 2000 chars): %s...\n", ppStr[:2000])
							} else {
								fmt.Printf("pageProps: %s\n", ppStr)
							}
						}
					}
				} else {
					fmt.Printf("Error parsing __NEXT_DATA__: %v\n", err)
				}
			}
		}
	} else {
		fmt.Println("No __NEXT_DATA__ found in HTML")
	}

	// Step 3: Try Next.js data route
	fmt.Println("\n=== Step 3: Try Next.js data route ===")
	// Find the build ID from the HTML
	re := regexp.MustCompile(`"buildId":"([^"]+)"`)
	buildMatches := re.FindStringSubmatch(pageStr)
	if len(buildMatches) >= 2 {
		buildID := buildMatches[1]
		fmt.Printf("Found buildId: %s\n", buildID)

		// Try /_next/data/{buildId}/matches/{slug}.json
		// Extract slug from pageUrl: /matches/liverpool-vs-afc-bournemouth/2he69q
		slug := strings.TrimPrefix(matchPageURL, "/")
		if idx := strings.Index(slug, "#"); idx != -1 {
			slug = slug[:idx]
		}
		dataURL := fmt.Sprintf("https://www.fotmob.com/_next/data/%s/%s.json", buildID, slug)
		fmt.Printf("Data URL: %s\n", dataURL)

		req3, _ := http.NewRequest("GET", dataURL, nil)
		req3.Header.Set("User-Agent", ua)
		req3.Header.Set("Accept", "application/json")
		resp3, err := client.Do(req3)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		defer resp3.Body.Close()

		dataBody, _ := io.ReadAll(resp3.Body)
		fmt.Printf("Status: %d, Size: %d bytes\n", resp3.StatusCode, len(dataBody))
		if resp3.StatusCode == 200 {
			var dataResp map[string]any
			if err := json.Unmarshal(dataBody, &dataResp); err == nil {
				if pageProps, ok := dataResp["pageProps"].(map[string]any); ok {
					keys := make([]string, 0, len(pageProps))
					for k := range pageProps {
						keys = append(keys, k)
					}
					fmt.Printf("pageProps keys: %v\n", keys)
					fmt.Println("SUCCESS via Next.js data route!")
				}
			}
		}
	} else {
		fmt.Println("Could not find buildId in HTML")
	}

	// Step 4: Try the API endpoint with the correct match ID from the leagues
	fmt.Println("\n=== Step 4: Direct API with targetMatchID from leagues ===")
	apiURL := fmt.Sprintf("https://www.fotmob.com/api/matchDetails?matchId=%s", targetMatchID)
	req4, _ := http.NewRequest("GET", apiURL, nil)
	req4.Header.Set("User-Agent", ua)
	resp4, err := client.Do(req4)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp4.Body.Close()
	body4, _ := io.ReadAll(resp4.Body)
	fmt.Printf("Status: %d, Body: %s\n", resp4.StatusCode, string(body4))
}
