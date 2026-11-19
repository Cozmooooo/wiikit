package main

import (
	"fmt"
	"wiikit/internal/homebrew"
)

func main() {
	m, err := homebrew.LoadManifest("manifest.json")
	if err != nil {
		fmt.Println("ERROR:", err)
		return
	}
	fmt.Printf("OK — %d assets loaded\n", len(m.Assets))
	for _, a := range m.Assets {
		fmt.Printf("  %-12s  %s...\n", a.Name, a.SHA256[:16])
	}
}
