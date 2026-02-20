package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// Test different request configurations against the matchDetails endpoint
func main() {
	matchID := "4813684"
	if len(os.Args) > 1 {
		matchID = os.Args[1]
	}

	url := fmt.Sprintf("https://www.fotmob.com/api/matchDetails?matchId=%s", matchID)
	fmt.Printf("Testing matchDetails endpoint for match ID: %s\n\n", matchID)

	// Test 1: Current app behavior (minimal User-Agent)
	fmt.Println("=== Test 1: Current app headers (User-Agent only) ===")
	testRequest(url, map[string]string{
		"User-Agent": "Mozilla/5.0",
	})

	// Test 2: Full browser-like headers
	fmt.Println("\n=== Test 2: Full browser headers ===")
	testRequest(url, map[string]string{
		"User-Agent":      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
		"Accept":          "application/json, text/plain, */*",
		"Accept-Language":  "en-US,en;q=0.9",
		"Referer":         "https://www.fotmob.com/",
		"Origin":          "https://www.fotmob.com",
		"Sec-Fetch-Dest":  "empty",
		"Sec-Fetch-Mode":  "cors",
		"Sec-Fetch-Site":  "same-origin",
		"Cache-Control":   "no-cache",
	})

	// Test 3: Just Referer + User-Agent
	fmt.Println("\n=== Test 3: User-Agent + Referer ===")
	testRequest(url, map[string]string{
		"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
		"Referer":    "https://www.fotmob.com/",
	})

	// Test 4: Check if leagues endpoint still works (control test)
	fmt.Println("\n=== Test 4: Control - leagues endpoint with same minimal headers ===")
	leagueURL := "https://www.fotmob.com/api/leagues?id=47&tab=fixtures"
	testRequest(leagueURL, map[string]string{
		"User-Agent": "Mozilla/5.0",
	})

	// Test 5: Check response body on 403 to see what FotMob returns
	fmt.Println("\n=== Test 5: 403 response body analysis ===")
	testRequestWithBody(url, map[string]string{
		"User-Agent": "Mozilla/5.0",
	})
}

func testRequest(url string, headers map[string]string) {
	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("  Error creating request: %v\n", err)
		return
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("  Status: %d %s\n", resp.StatusCode, resp.Status)
	fmt.Printf("  Content-Type: %s\n", resp.Header.Get("Content-Type"))
	fmt.Printf("  Content-Length: %s\n", resp.Header.Get("Content-Length"))

	if resp.StatusCode == http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		var raw map[string]any
		if err := json.Unmarshal(body, &raw); err == nil {
			// Check if it has the expected structure
			if _, ok := raw["header"]; ok {
				fmt.Println("  Result: Valid matchDetails response with 'header' field")
			} else if _, ok := raw["general"]; ok {
				fmt.Println("  Result: Valid matchDetails response with 'general' field")
			} else {
				keys := make([]string, 0, len(raw))
				for k := range raw {
					keys = append(keys, k)
				}
				fmt.Printf("  Result: JSON response with keys: %v\n", keys)
			}
			fmt.Printf("  Response size: %d bytes\n", len(body))
		}
	}
}

func testRequestWithBody(url string, headers map[string]string) {
	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("  Error creating request: %v\n", err)
		return
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("  Status: %d\n", resp.StatusCode)
	fmt.Println("  Response headers:")
	for k, v := range resp.Header {
		fmt.Printf("    %s: %v\n", k, v)
	}

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("  Response body (%d bytes):\n", len(body))
	if len(body) > 2000 {
		fmt.Printf("  %s...\n", string(body[:2000]))
	} else {
		fmt.Printf("  %s\n", string(body))
	}
}
