package wiilink

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"time"

	"wiikit/internal/audit"
	"wiikit/internal/homebrew"
)

const (
	patcherURLWin   = "https://github.com/WiiLink24/WiiLink-Patcher/releases/latest/download/WiiLink-Patcher-Windows.exe"
	patcherURLMac   = "https://github.com/WiiLink24/WiiLink-Patcher/releases/latest/download/WiiLink-Patcher-macOS"
	patcherURLLinux = "https://github.com/WiiLink24/WiiLink-Patcher/releases/latest/download/WiiLink-Patcher-Linux"
)

func RunPatcher(ctx context.Context, sdRoot string, logWriter io.Writer) error {
	if logWriter == nil {
		logWriter = io.Discard
	}

	url, sha256 := patcherURLForOS()

	cacheDir, err := os.MkdirTemp("", "wiikit-wiilink-*")
	if err != nil {
		return fmt.Errorf("create cache dir: %w", err)
	}
	defer os.RemoveAll(cacheDir)

	asset := homebrew.Asset{
		Name:   "wiilink-patcher",
		URL:    url,
		SHA256: sha256,
	}

	patcherPath, err := homebrew.DownloadAsset(ctx, asset, cacheDir)
	if err != nil {
		return fmt.Errorf("download wiilink patcher: %w", err)
	}

	if runtime.GOOS != "windows" {
		if err := os.Chmod(patcherPath, 0755); err != nil {
			return fmt.Errorf("chmod patcher: %w", err)
		}
	}

	runCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(runCtx, patcherPath, "--sd", sdRoot, "--unattended")
	cmd.Stdout = logWriter
	cmd.Stderr = logWriter

	if err := cmd.Run(); err != nil {
		if runCtx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("wiilink patcher timed out after 60s: %w", context.DeadlineExceeded)
		}
		return fmt.Errorf("wiilink patcher: %w", err)
	}
	return nil
}

func patcherURLForOS() (url, sha256 string) {
	// SHA-256 values must be updated when a new patcher version is pinned in manifest.json
	// These are placeholder values — populate from manifest.json in production
	switch runtime.GOOS {
	case "darwin":
		return patcherURLMac, "" // populated from manifest
	case "windows":
		return patcherURLWin, "" // populated from manifest
	default:
		return patcherURLLinux, "" // populated from manifest
	}
}

// VerifyPatcher checks the already-downloaded patcher binary against a known hash.
func VerifyPatcher(path, expectedHash string) error {
	return audit.VerifyFile(path, expectedHash)
}
