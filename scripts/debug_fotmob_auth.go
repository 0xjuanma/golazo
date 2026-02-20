package main

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// Three Lions (Football's Coming Home) lyrics - the actual secret from FotMob's JS
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

func generateXMas(apiPath string) string {
	now := time.Now()
	code := now.UnixMilli()

	// Build the body exactly as FotMob's JS does
	type bodyStruct struct {
		URL  string `json:"url"`
		Code int64  `json:"code"`
		Foo  string `json:"foo"`
	}

	body := bodyStruct{
		URL:  apiPath,
		Code: code,
		Foo:  "production:33324f727a7a2706a154eab6f683920b1df36aee",
	}

	// g=(e,t)=>l(`${JSON.stringify(e)}${t}`)
	// l=e=>o()(e).toUpperCase()  (o is MD5)
	bodyJSON, _ := json.Marshal(body)
	combined := string(bodyJSON) + secret
	hash := md5.Sum([]byte(combined))
	signature := strings.ToUpper(fmt.Sprintf("%x", hash))

	// Build the full token
	type tokenStruct struct {
		Body      bodyStruct `json:"body"`
		Signature string     `json:"signature"`
	}

	token := tokenStruct{
		Body:      body,
		Signature: signature,
	}

	tokenJSON, _ := json.Marshal(token)
	return base64.StdEncoding.EncodeToString(tokenJSON)
}

func main() {
	matchID := "4813684"
	if len(os.Args) > 1 {
		matchID = os.Args[1]
	}

	apiPath := fmt.Sprintf("/api/matchDetails?matchId=%s", matchID)
	url := fmt.Sprintf("https://www.fotmob.com%s", apiPath)

	fmt.Printf("Testing x-mas header for match ID: %s\n", matchID)
	fmt.Printf("API path: %s\n\n", apiPath)

	token := generateXMas(apiPath)
	fmt.Printf("Generated x-mas token (first 80 chars): %s...\n\n", token[:80])

	// Decode to verify
	decoded, _ := base64.StdEncoding.DecodeString(token)
	decodedStr := string(decoded)
	if len(decodedStr) > 200 {
		decodedStr = decodedStr[:200]
	}
	fmt.Printf("Decoded token: %s\n\n", decodedStr)

	// Test the request
	fmt.Println("=== Test: matchDetails with x-mas header ===")
	client := &http.Client{Timeout: 15 * time.Second}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
	req.Header.Set("x-mas", token)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body2, _ := io.ReadAll(resp.Body)
	fmt.Printf("Status: %d %s\n", resp.StatusCode, resp.Status)
	fmt.Printf("Content-Type: %s\n", resp.Header.Get("Content-Type"))
	fmt.Printf("Response size: %d bytes\n\n", len(body2))

	if resp.StatusCode == http.StatusOK {
		var raw map[string]any
		if err := json.Unmarshal(body2, &raw); err == nil {
			keys := make([]string, 0, len(raw))
			for k := range raw {
				keys = append(keys, k)
			}
			fmt.Printf("Response keys: %v\n", keys)

			// Check for key fields
			if header, ok := raw["header"]; ok {
				headerMap := header.(map[string]any)
				if teams, ok := headerMap["teams"]; ok {
					teamsArr := teams.([]any)
					if len(teamsArr) >= 2 {
						t0 := teamsArr[0].(map[string]any)
						t1 := teamsArr[1].(map[string]any)
						fmt.Printf("\nMatch: %s vs %s\n", t0["name"], t1["name"])
						fmt.Printf("Score: %v - %v\n", t0["score"], t1["score"])
					}
				}
			}
			if general, ok := raw["general"]; ok {
				gMap := general.(map[string]any)
				fmt.Printf("League: %s\n", gMap["leagueName"])
			}
			fmt.Println("\nSUCCESS! matchDetails returned valid data!")
		}
	} else {
		fmt.Printf("Response body: %s\n", string(body2))
	}

	// Also test leagues endpoint for comparison
	fmt.Println("\n=== Test: leagues endpoint with x-mas header ===")
	leaguePath := "/api/leagues?id=47&tab=fixtures"
	leagueURL := "https://www.fotmob.com" + leaguePath
	leagueToken := generateXMas(leaguePath)

	req2, _ := http.NewRequest("GET", leagueURL, nil)
	req2.Header.Set("User-Agent", "Mozilla/5.0")
	req2.Header.Set("x-mas", leagueToken)

	resp2, err := client.Do(req2)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp2.Body.Close()
	fmt.Printf("Status: %d %s\n", resp2.StatusCode, resp2.Status)
}
