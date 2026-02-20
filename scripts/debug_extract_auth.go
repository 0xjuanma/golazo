package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

func main() {
	// Fetch the _app bundle that contains the auth code
	jsURL := "https://www.fotmob.com/_next/static/chunks/pages/_app-22ec5ed3ffb5d17b.js"

	client := &http.Client{Timeout: 30 * time.Second}
	req, _ := http.NewRequest("GET", jsURL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	jsCode := string(body)

	// Extract the auth section around the signature/x-mas code
	// We found 'signature' at offset 278325
	// Let's get a larger context around that area

	fmt.Println("=== Auth code section (around signature/x-mas) ===\n")

	// Find the key variables and functions
	idx := strings.Index(jsCode, `x-mas`)
	if idx == -1 {
		fmt.Println("x-mas not found, searching for x-fm-req...")
		idx = strings.Index(jsCode, `x-fm-req`)
	}

	if idx != -1 {
		start := idx - 1500
		if start < 0 {
			start = 0
		}
		end := idx + 1500
		if end > len(jsCode) {
			end = len(jsCode)
		}
		fmt.Printf("Context around x-mas (offset %d):\n", idx)
		fmt.Println(jsCode[start:end])
	}

	fmt.Println("\n\n=== Searching for hash function (l=) and secret (h=) ===\n")

	// Search for the hash function definition and the secret variable 'h'
	// From context: l(`${JSON.stringify(e)}${t}`) and g=(e,t)=>l(...)
	// We need to find where 'h' is defined (the secret string)
	// and where 'l' is defined (the hash function)

	// Look for md5 or hash-related code near the signature area
	sigIdx := strings.Index(jsCode, "signature:n}")
	if sigIdx != -1 {
		// Go back further to find the variable declarations
		start := sigIdx - 3000
		if start < 0 {
			start = 0
		}
		end := sigIdx + 500
		if end > len(jsCode) {
			end = len(jsCode)
		}
		fmt.Printf("Extended context before signature (offset %d):\n", sigIdx)
		fmt.Println(jsCode[start:end])
	}

	// Also search for the turnstile callback which contains the token verification
	fmt.Println("\n\n=== Turnstile callback section ===\n")
	turnIdx := strings.Index(jsCode, "cf-turnstile")
	if turnIdx != -1 {
		start := turnIdx - 500
		if start < 0 {
			start = 0
		}
		end := turnIdx + 2000
		if end > len(jsCode) {
			end = len(jsCode)
		}
		fmt.Println(jsCode[start:end])
	}
}
