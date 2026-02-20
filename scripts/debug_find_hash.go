package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

func main() {
	jsURL := "https://www.fotmob.com/_next/static/chunks/pages/_app-22ec5ed3ffb5d17b.js"
	client := &http.Client{Timeout: 30 * time.Second}
	req, _ := http.NewRequest("GET", jsURL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)")
	resp, _ := client.Do(req)
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	jsCode := string(body)

	// Find the hash function 'l' - it's defined somewhere before the signature code
	// The pattern is: l(`${JSON.stringify(e)}${t}`)
	// 'l' is likely an MD5 function. Let's find its definition.

	// Search for the function/variable 'l' definition near the auth code
	// Look for the module that defines l, g, m, h variables

	// The auth code is in a module. Let's find the module boundary
	sigIdx := strings.Index(jsCode, `signature:n}`)
	if sigIdx == -1 {
		fmt.Println("signature not found")
		return
	}

	// Go back much further to find the module start and the 'l' definition
	start := sigIdx - 6000
	if start < 0 {
		start = 0
	}
	section := jsCode[start:sigIdx]

	// Look for common MD5 patterns
	patterns := []string{
		"md5",
		"MD5",
		"createHash",
		"charCodeAt",
		"0x",
		"digest",
		"hex",
		"function l(",
		"l=function",
		"const l=",
		"let l=",
		"var l=",
	}

	for _, p := range patterns {
		idx := strings.LastIndex(section, p)
		if idx != -1 {
			s := idx - 100
			if s < 0 {
				s = 0
			}
			e := idx + len(p) + 200
			if e > len(section) {
				e = len(section)
			}
			fmt.Printf("Found '%s' at offset %d:\n", p, start+idx)
			fmt.Printf("  ...%s...\n\n", section[s:e])
		}
	}

	// Also look for the 'fotmob-client' header and the 'c' function
	fmt.Println("\n=== Looking for fotmob-client header setup ===")
	fcIdx := strings.Index(jsCode, `fotmob-client`)
	if fcIdx != -1 {
		// Go back to find module start
		mstart := fcIdx - 2000
		if mstart < 0 {
			mstart = 0
		}
		fmt.Println(jsCode[mstart : fcIdx+200])
	}
}
