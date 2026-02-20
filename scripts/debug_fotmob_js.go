package main

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

func main() {
	// Step 1: Fetch a FotMob match page to find the JavaScript bundles
	fmt.Println("=== Step 1: Fetching FotMob match page to find JS bundles ===")

	client := &http.Client{Timeout: 30 * time.Second}
	req, _ := http.NewRequest("GET", "https://www.fotmob.com/matches/4813684", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	html := string(body)
	fmt.Printf("Page status: %d, size: %d bytes\n", resp.StatusCode, len(body))

	// Find all JavaScript bundle URLs
	re := regexp.MustCompile(`src="(/[^"]*\.js)"`)
	matches := re.FindAllStringSubmatch(html, -1)
	fmt.Printf("Found %d JS bundles\n\n", len(matches))

	// Step 2: Search each JS bundle for x-fm-req or signature-related code
	for i, m := range matches {
		jsPath := m[1]
		jsURL := "https://www.fotmob.com" + jsPath

		fmt.Printf("=== Bundle %d: %s ===\n", i+1, jsPath)

		req2, _ := http.NewRequest("GET", jsURL, nil)
		req2.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")

		resp2, err := client.Do(req2)
		if err != nil {
			fmt.Printf("  Error fetching: %v\n", err)
			continue
		}

		jsBody, _ := io.ReadAll(resp2.Body)
		resp2.Body.Close()
		jsCode := string(jsBody)
		fmt.Printf("  Size: %d bytes\n", len(jsCode))

		// Search for relevant patterns
		patterns := []string{
			"x-fm-req",
			"X-Fm-Req",
			"turnstile",
			"TURNSTILE",
			"signature",
			"md5",
			"MD5",
			"strangers",
			"never gonna",
			"Never gonna",
			"give you up",
			"generateToken",
			"fm-req",
		}

		found := false
		for _, pattern := range patterns {
			if idx := strings.Index(strings.ToLower(jsCode), strings.ToLower(pattern)); idx != -1 {
				fmt.Printf("  FOUND '%s' at offset %d\n", pattern, idx)
				// Print surrounding context (200 chars before and after)
				start := idx - 200
				if start < 0 {
					start = 0
				}
				end := idx + len(pattern) + 300
				if end > len(jsCode) {
					end = len(jsCode)
				}
				context := jsCode[start:end]
				// Clean up for display
				context = strings.ReplaceAll(context, "\n", " ")
				fmt.Printf("  Context: ...%s...\n\n", context)
				found = true
			}
		}
		if !found {
			fmt.Println("  No relevant patterns found")
		}
		fmt.Println()
	}
}
