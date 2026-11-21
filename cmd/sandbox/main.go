// sandbox: download real homebrew ZIPs and verify they extract to the
// correct SD card paths. Run with: go run ./cmd/sandbox [asset-name ...]
// If no names given, tests priiloader + ethernet-enabler by default.
// Use -all to test every non-tool, non-pending asset in the manifest.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"wiikit/internal/homebrew"
)

var (
	all     = flag.Bool("all", false, "test every installable asset (downloads all ZIPs)")
	timeout = flag.Duration("timeout", 10*time.Minute, "download timeout")
)

// expectedPaths maps asset name → slice of paths relative to SD root that MUST exist.
var expectedPaths = map[string][]string{
	"priiloader":      {"apps/priiloader/boot.dol"},
	"wiiflow":         {"apps/wiiflow/boot.dol"},
	"usbloadergx":     {"apps/usbloadergx/boot.dol"},
	"not64":           {"apps/not64/boot.dol"},
	"wiistation":      {"apps/wiistation/boot.dol"},
	"snes9xgx":        {"apps/snes9xgx/boot.dol"},
	"fceugx":          {"apps/fceugx/boot.dol"},
	"vbagx":           {"apps/vbagx/boot.dol"},
	"nintendont":      {"apps/Nintendont/boot.dol"},
	"ethernet-enabler": {"apps/Wii Mini Ethernet Enable/boot.dol"},
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: sandbox [-all] [-timeout 10m] [asset ...]\n\nDownloads real homebrew ZIPs and verifies they extract to expected SD paths.\n\nFlags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	manifest, err := homebrew.LoadManifest("manifest.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "load manifest: %v\n", err)
		os.Exit(1)
	}

	// Build target list
	var targets []homebrew.Asset
	for _, a := range manifest.Assets {
		if a.Group == "tool" || a.Group == "pending" {
			continue
		}
		if *all {
			targets = append(targets, a)
			continue
		}
		// Default: priiloader + ethernet-enabler; or whatever the user named
		names := flag.Args()
		if len(names) == 0 {
			names = []string{"priiloader", "ethernet-enabler"}
		}
		for _, n := range names {
			if a.Name == n {
				targets = append(targets, a)
			}
		}
	}

	if len(targets) == 0 {
		fmt.Println("no matching assets found")
		os.Exit(1)
	}

	sdRoot, err := os.MkdirTemp("", "wiikit-sandbox-sd-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "mktemp: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		fmt.Printf("\ncleaning up sandbox at %s\n", sdRoot)
		os.RemoveAll(sdRoot)
	}()

	fmt.Printf("sandbox SD root: %s\n\n", sdRoot)

	pass, fail := 0, 0
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	cacheDir, err := os.MkdirTemp("", "wiikit-sandbox-cache-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "mktemp cache: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(cacheDir)

	for _, asset := range targets {
		fmt.Printf("── %s (%s)\n", asset.Name, asset.Version)
		fmt.Printf("   downloading %s ...\n", asset.URL)

		localPath, dlErr := homebrew.DownloadAsset(ctx, asset, cacheDir)
		if dlErr != nil {
			fmt.Printf("   FAIL download: %v\n\n", dlErr)
			fail++
			continue
		}
		fmt.Printf("   downloaded → %s\n", filepath.Base(localPath))

		destPath := homebrew.ResolveDestPath(asset.DestPath, sdRoot)

		if strings.HasSuffix(strings.ToLower(localPath), ".zip") {
			os.MkdirAll(destPath, 0755)
			if err := homebrew.ExtractZip(localPath, destPath); err != nil {
				fmt.Printf("   FAIL extract: %v\n\n", err)
				fail++
				continue
			}
			fmt.Printf("   extracted to %s\n", destPath)
		} else {
			os.MkdirAll(filepath.Dir(destPath), 0755)
			if err := homebrew.CopyFile(localPath, destPath); err != nil {
				fmt.Printf("   FAIL copy: %v\n\n", err)
				fail++
				continue
			}
			fmt.Printf("   copied to %s\n", destPath)
		}

		// Walk and print what landed
		fmt.Printf("   files on SD:\n")
		filepath.Walk(sdRoot, func(p string, info os.FileInfo, _ error) error {
			if !info.IsDir() {
				rel, _ := filepath.Rel(sdRoot, p)
				fmt.Printf("     %s\n", filepath.ToSlash(rel))
			}
			return nil
		})

		// Verify expected paths
		want, hasExpected := expectedPaths[asset.Name]
		assetOk := true
		if hasExpected {
			for _, rel := range want {
				check := filepath.Join(sdRoot, filepath.FromSlash(rel))
				if _, err := os.Stat(check); os.IsNotExist(err) {
					fmt.Printf("   ✗ MISSING expected path: %s\n", rel)
					assetOk = false
				} else {
					fmt.Printf("   ✓ %s\n", rel)
				}
			}
		}

		// Also check for double-nesting
		filepath.Walk(sdRoot, func(p string, info os.FileInfo, _ error) error {
			rel, _ := filepath.Rel(sdRoot, p)
			slash := filepath.ToSlash(rel)
			if strings.Contains(slash, "apps/apps/") {
				fmt.Printf("   ✗ DOUBLE-NESTING BUG: %s\n", slash)
				assetOk = false
			}
			return nil
		})

		if assetOk {
			fmt.Printf("   PASS\n")
			pass++
		} else {
			fail++
		}
		fmt.Println()

		// Clean between assets so paths don't accumulate
		os.RemoveAll(sdRoot)
		os.MkdirAll(sdRoot, 0755)
	}

	fmt.Printf("results: %d passed, %d failed\n", pass, fail)
	if fail > 0 {
		os.Exit(1)
	}
}
