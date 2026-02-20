package main

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strings"
	"time"
)

var secret = `[Spoken Intro: Alan Hansen & Trevor Brooking]
I think it's bad news for the English game
We're not creative enough, and we're not positive enough

[Refrain: Ian Broudie & Jimmy Hill]
It's coming home, it's coming home, it's coming
Football's coming home (We'll go on getting bad results)
It's coming home, it's coming home, it's coming
Football's coming home
It's coming home, it's coming home, it's coming
Football's coming home
It's coming home, it's coming home, it's coming
Football's coming home

[Verse 1: Frank Skinner]
Everyone seems to know the score, they've seen it all before
They just know, they're so sure
That England's gonna throw it away, gonna blow it away
But I know they can play, 'cause I remember

[Chorus: All]
Three lions on a shirt
Jules Rimet still gleaming
Thirty years of hurt
Never stopped me dreaming

[Verse 2: David Baddiel]
So many jokes, so many sneers
But all those "Oh, so near"s wear you down through the years
But I still see that tackle by Moore and when Lineker scored
Bobby belting the ball, and Nobby dancing

[Chorus: All]
Three lions on a shirt
Jules Rimet still gleaming
Thirty years of hurt
Never stopped me dreaming

[Bridge]
England have done it, in the last minute of extra time!
What a save, Gordon Banks!
Good old England, England that couldn't play football!
England have got it in the bag!
I know that was then, but it could be again

[Refrain: Ian Broudie]
It's coming home, it's coming
Football's coming home
It's coming home, it's coming home, it's coming
Football's coming home
(England have done it!)
It's coming home, it's coming home, it's coming
Football's coming home
It's coming home, it's coming home, it's coming
Football's coming home
[Chorus: All]
(It's coming home) Three lions on a shirt
(It's coming home, it's coming) Jules Rimet still gleaming
(Football's coming home
It's coming home) Thirty years of hurt
(It's coming home, it's coming) Never stopped me dreaming
(Football's coming home
It's coming home) Three lions on a shirt
(It's coming home, it's coming) Jules Rimet still gleaming
(Football's coming home
It's coming home) Thirty years of hurt
(It's coming home, it's coming) Never stopped me dreaming
(Football's coming home
It's coming home) Three lions on a shirt
(It's coming home, it's coming) Jules Rimet still gleaming
(Football's coming home
It's coming home) Thirty years of hurt
(It's coming home, it's coming) Never stopped me dreaming
(Football's coming home)`

type body struct {
	URL  string `json:"url"`
	Code int64  `json:"code"`
	Foo  string `json:"foo"`
}

func generateXMas(apiPath string) string {
	b := body{
		URL:  apiPath,
		Code: time.Now().UnixMilli(),
		Foo:  "production:33324f727a7a2706a154eab6f683920b1df36aee",
	}

	bodyJSON, _ := json.Marshal(b)
	hash := md5.Sum([]byte(string(bodyJSON) + secret))
	sig := strings.ToUpper(fmt.Sprintf("%x", hash))

	type token struct {
		Body      body   `json:"body"`
		Signature string `json:"signature"`
	}

	tokenJSON, _ := json.Marshal(token{Body: b, Signature: sig})
	return base64.StdEncoding.EncodeToString(tokenJSON)
}

func main() {
	matchID := "4813684"
	if len(os.Args) > 1 {
		matchID = os.Args[1]
	}

	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Timeout: 30 * time.Second,
		Jar:     jar,
	}

	ua := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"

	// Step 1: Visit the main page to get session cookies
	fmt.Println("=== Step 1: Visit main page to get cookies ===")
	req, _ := http.NewRequest("GET", "https://www.fotmob.com/", nil)
	req.Header.Set("User-Agent", ua)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	io.ReadAll(resp.Body)
	resp.Body.Close()
	fmt.Printf("Main page status: %d\n", resp.StatusCode)
	fmt.Printf("Cookies received: %d\n", len(jar.Cookies(req.URL)))
	for _, c := range jar.Cookies(req.URL) {
		fmt.Printf("  %s = %s...\n", c.Name, truncate(c.Value, 40))
	}

	// Step 2: Try matchDetails with cookies + x-mas
	fmt.Println("\n=== Step 2: matchDetails with cookies + x-mas ===")
	apiPath := fmt.Sprintf("/api/matchDetails?matchId=%s", matchID)
	url := fmt.Sprintf("https://www.fotmob.com%s", apiPath)
	token := generateXMas(apiPath)

	req2, _ := http.NewRequest("GET", url, nil)
	req2.Header.Set("User-Agent", ua)
	req2.Header.Set("x-mas", token)
	req2.Header.Set("Accept", "application/json")
	req2.Header.Set("Referer", "https://www.fotmob.com/")

	resp2, err := client.Do(req2)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp2.Body.Close()

	body2, _ := io.ReadAll(resp2.Body)
	fmt.Printf("Status: %d\n", resp2.StatusCode)
	fmt.Printf("Size: %d bytes\n", len(body2))

	if resp2.StatusCode == http.StatusOK {
		var raw map[string]any
		if err := json.Unmarshal(body2, &raw); err == nil {
			keys := make([]string, 0, len(raw))
			for k := range raw {
				keys = append(keys, k)
			}
			fmt.Printf("Response keys: %v\n", keys)
			fmt.Println("SUCCESS!")
		}
	} else {
		fmt.Printf("Response: %s\n", string(body2))
	}

	// Step 3: Try without x-mas but with cookies
	fmt.Println("\n=== Step 3: matchDetails with cookies only (no x-mas) ===")
	req3, _ := http.NewRequest("GET", url, nil)
	req3.Header.Set("User-Agent", ua)
	req3.Header.Set("Accept", "application/json")

	resp3, err := client.Do(req3)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp3.Body.Close()

	body3, _ := io.ReadAll(resp3.Body)
	fmt.Printf("Status: %d\n", resp3.StatusCode)
	fmt.Printf("Response: %s\n", string(body3))

	// Step 4: Try match page directly (not API)
	fmt.Println("\n=== Step 4: Match page HTML (look for embedded data) ===")
	pageURL := fmt.Sprintf("https://www.fotmob.com/matches/%s", matchID)
	req4, _ := http.NewRequest("GET", pageURL, nil)
	req4.Header.Set("User-Agent", ua)
	req4.Header.Set("Accept", "text/html")

	resp4, err := client.Do(req4)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp4.Body.Close()

	body4, _ := io.ReadAll(resp4.Body)
	pageStr := string(body4)
	fmt.Printf("Status: %d\n", resp4.StatusCode)
	fmt.Printf("Size: %d bytes\n", len(body4))

	// Check for __NEXT_DATA__ which often contains pre-rendered data in Next.js apps
	if idx := strings.Index(pageStr, `__NEXT_DATA__`); idx != -1 {
		start := idx
		end := idx + 500
		if end > len(pageStr) {
			end = len(pageStr)
		}
		fmt.Printf("Found __NEXT_DATA__ at offset %d\n", idx)
		fmt.Printf("Content: %s...\n", pageStr[start:end])
	}

	// Look for initial match data in the HTML
	if idx := strings.Index(pageStr, `"matchId"`); idx != -1 {
		start := idx - 50
		if start < 0 {
			start = 0
		}
		end := idx + 200
		if end > len(pageStr) {
			end = len(pageStr)
		}
		fmt.Printf("Found matchId at offset %d: %s\n", idx, pageStr[start:end])
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
