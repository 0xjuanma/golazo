//go:build ignore

package main

import (
	"fmt"

	"github.com/0xjuanma/golazo/internal/ui/logo"
)

func main() {
	opts := logo.DefaultOpts()

	fmt.Println("=== Compact ===")
	fmt.Println(logo.Render("v0.14.0", true, opts))
	fmt.Println()

	fmt.Println("=== Wide ===")
	opts.Width = 80
	fmt.Println(logo.Render("v0.14.0", false, opts))
	fmt.Println()

	fmt.Println("=== Inline ===")
	fmt.Println(logo.RenderCompact(60))
}
