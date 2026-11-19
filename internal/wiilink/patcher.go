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

func RunPatcher(ctx context.Context, sdRoot string, logWriter io.Writer) error {
	if logWriter == nil {
		logWriter = io.Discard
	}

	// Load patcher asset from manifest so SHA-256 is centrally managed.
	manifest, err := homebrew.LoadManifest("manifest.json")
	if err != nil {
		return fmt.Errorf("load manifest for wiilink patcher: %w", err)
	}

	assetName := patcherAssetName()
	var asset homebrew.Asset
	for _, a := range manifest.Assets {
		if a.Name == assetName {
			asset = a
			break
		}
	}
	if asset.URL == "" {
		return fmt.Errorf("wiilink patcher not found in manifest (expected asset name %q) — add it to manifest.json with a pinned sha256", assetName)
	}

	cacheDir, err := os.MkdirTemp("", "wiikit-wiilink-*")
	if err != nil {
		return fmt.Errorf("create cache dir: %w", err)
	}
	defer os.RemoveAll(cacheDir)

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

// patcherAssetName returns the manifest asset name for the current OS.
func patcherAssetName() string {
	switch runtime.GOOS {
	case "darwin":
		return "wiilink-patcher-macos"
	case "windows":
		return "wiilink-patcher-windows"
	default:
		return "wiilink-patcher-linux"
	}
}

// VerifyPatcher checks the already-downloaded patcher binary against a known hash.
func VerifyPatcher(path, expectedHash string) error {
	return audit.VerifyFile(path, expectedHash)
}
