package fotmob

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/0xjuanma/golazo/internal/api"
)

// fetchMatchDetailsFromPage fetches match details by scraping the match page HTML
// and extracting the __NEXT_DATA__ JSON embedded by Next.js server-side rendering.
//
// This is an alternative to the /api/matchDetails endpoint, which now requires
// Cloudflare Turnstile verification and returns 403 for non-browser clients.
//
// The match page URL uses a slug format from the leagues endpoint pageUrl field
// (e.g., "/matches/wolverhampton-wanderers-vs-arsenal/2t3bl7").
// The embedded __NEXT_DATA__ contains the same data structure as the old API.
func fetchMatchDetailsFromPage(ctx context.Context, httpClient *http.Client, pageSlug string) (*api.MatchDetails, error) {
	if pageSlug == "" {
		return nil, fmt.Errorf("empty page slug")
	}

	url := "https://www.fotmob.com" + pageSlug

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create page request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch match page: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("match page returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read match page body: %w", err)
	}

	pageProps, err := extractPageProps(string(body))
	if err != nil {
		return nil, fmt.Errorf("extract page props: %w", err)
	}

	var response fotmobMatchDetails
	if err := json.Unmarshal(pageProps, &response); err != nil {
		return nil, fmt.Errorf("decode match details from page props: %w", err)
	}

	details := response.toAPIMatchDetails()
	return details, nil
}

// extractPageProps extracts the pageProps JSON from a Next.js page's __NEXT_DATA__ script tag.
func extractPageProps(html string) (json.RawMessage, error) {
	const marker = `__NEXT_DATA__`

	markerIdx := strings.Index(html, marker)
	if markerIdx == -1 {
		return nil, fmt.Errorf("__NEXT_DATA__ not found in page")
	}

	// Find the opening > of the script tag content
	startIdx := strings.Index(html[markerIdx:], ">")
	if startIdx == -1 {
		return nil, fmt.Errorf("script tag opening not found")
	}
	startIdx += markerIdx + 1

	// Find the closing </script> tag
	endIdx := strings.Index(html[startIdx:], "</script>")
	if endIdx == -1 {
		return nil, fmt.Errorf("script tag closing not found")
	}
	endIdx += startIdx

	// Parse the wrapper to extract props.pageProps
	var wrapper struct {
		Props struct {
			PageProps json.RawMessage `json:"pageProps"`
		} `json:"props"`
	}

	if err := json.Unmarshal([]byte(html[startIdx:endIdx]), &wrapper); err != nil {
		return nil, fmt.Errorf("parse __NEXT_DATA__ JSON: %w", err)
	}

	if len(wrapper.Props.PageProps) == 0 {
		return nil, fmt.Errorf("pageProps is empty")
	}

	return wrapper.Props.PageProps, nil
}
