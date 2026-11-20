package homebrew

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const (
	splashURL     = "https://github.com/Cozmooooo/wiikit-splash/releases/download/v1.0.0/krunk-splash.dol"
	splashSHA256  = "" // pin this after first verified release build
	splashVersion = "1.0.0"
)

// WriteSplashDOL downloads the Krunk splash screen .dol and places it at
// sd:/boot.dol so Priiloader launches it on power-on.
//
// If splashSHA256 is empty the integrity check is skipped and a warning is
// logged — set it once the first release is published and verified.
func WriteSplashDOL(ctx context.Context, sdRoot string, logFn func(string)) error {
	destPath := filepath.Join(sdRoot, "boot.dol")
	tmpPath := destPath + ".tmp"

	logFn("Downloading Krunk splash screen v" + splashVersion + "...")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, splashURL, nil)
	if err != nil {
		return fmt.Errorf("splash: build request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("splash: download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("splash: download returned HTTP %d", resp.StatusCode)
	}

	if err := os.MkdirAll(filepath.Dir(tmpPath), 0755); err != nil {
		return fmt.Errorf("splash: mkdir: %w", err)
	}

	f, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("splash: create tmp: %w", err)
	}

	h := sha256.New()
	if _, err := io.Copy(io.MultiWriter(f, h), resp.Body); err != nil {
		f.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("splash: write: %w", err)
	}
	f.Close()

	if splashSHA256 != "" {
		got := hex.EncodeToString(h.Sum(nil))
		if got != splashSHA256 {
			os.Remove(tmpPath)
			return fmt.Errorf("splash: SHA256 mismatch: got %s want %s", got, splashSHA256)
		}
		logFn("Splash SHA256 verified.")
	} else {
		logFn("Warning: splash SHA256 not pinned — skipping integrity check.")
	}

	if err := os.Rename(tmpPath, destPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("splash: rename: %w", err)
	}

	logFn("Splash screen written to SD:/boot.dol")
	return nil
}
