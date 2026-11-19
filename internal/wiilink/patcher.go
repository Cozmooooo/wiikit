package wiilink

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"wiikit/internal/audit"
	"wiikit/internal/homebrew"
)

// RunPatcher downloads the WiiLink Patcher GUI, verifies its SHA-256, and launches it.
//
// WiiLink Patcher v1.5.3+ is a GUI application and does not support --unattended flags.
// WiiKit handles download + verification; the user completes the interactive patcher setup.
// The SD card is already formatted and pre-populated by the time this runs, so the user
// only needs to point the patcher at the SD root.
func RunPatcher(ctx context.Context, sdRoot string, logWriter io.Writer) error {
	if logWriter == nil {
		logWriter = io.Discard
	}

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
		return fmt.Errorf("wiilink patcher not found in manifest (expected %q) — add it to manifest.json", assetName)
	}

	cacheDir, err := os.MkdirTemp("", "wiikit-wiilink-*")
	if err != nil {
		return fmt.Errorf("create cache dir: %w", err)
	}
	// Don't defer cleanup — user may still be running the patcher

	downloadedPath, err := homebrew.DownloadAsset(ctx, asset, cacheDir)
	if err != nil {
		os.RemoveAll(cacheDir)
		return fmt.Errorf("download wiilink patcher: %w", err)
	}

	// Extract tar.gz archives (macOS and Linux builds)
	patcherPath := downloadedPath
	if strings.HasSuffix(downloadedPath, ".tar.gz") || strings.HasSuffix(downloadedPath, ".tgz") {
		extracted, err := extractTarGz(downloadedPath, cacheDir)
		if err != nil {
			os.RemoveAll(cacheDir)
			return fmt.Errorf("extract wiilink patcher: %w", err)
		}
		patcherPath = extracted
	}

	if runtime.GOOS != "windows" {
		if err := os.Chmod(patcherPath, 0755); err != nil {
			os.RemoveAll(cacheDir)
			return fmt.Errorf("chmod patcher: %w", err)
		}
	}

	fmt.Fprintf(logWriter, "WiiLink Patcher downloaded to %s\n", patcherPath)
	fmt.Fprintf(logWriter, "Launching WiiLink Patcher GUI — point it at your SD card: %s\n", sdRoot)
	fmt.Fprintf(logWriter, "Complete the setup in the patcher window, then return here.\n")

	// Launch detached — this is a GUI app, user interacts with it directly.
	cmd := exec.CommandContext(ctx, patcherPath)
	cmd.Stdout = logWriter
	cmd.Stderr = logWriter

	if err := cmd.Start(); err != nil {
		os.RemoveAll(cacheDir)
		return fmt.Errorf("launch wiilink patcher: %w", err)
	}

	// Wait for the user to finish — the patcher process will exit when done.
	if err := cmd.Wait(); err != nil {
		// Non-zero exit is common for GUI apps on some platforms; don't treat as fatal.
		fmt.Fprintf(logWriter, "WiiLink Patcher exited: %v\n", err)
	}

	os.RemoveAll(cacheDir)
	return nil
}

// extractTarGz extracts a .tar.gz archive and returns the path to the first executable found.
func extractTarGz(archivePath, destDir string) (string, error) {
	f, err := os.Open(archivePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return "", err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	var firstExe string

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}

		// Guard against path traversal
		clean := filepath.Clean(hdr.Name)
		if strings.HasPrefix(clean, "..") {
			continue
		}

		destPath := filepath.Join(destDir, clean)
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return "", err
		}

		out, err := os.Create(destPath)
		if err != nil {
			return "", err
		}
		if _, err := io.Copy(out, tr); err != nil {
			out.Close()
			return "", err
		}
		out.Close()

		// Pick the first file that looks like the patcher executable
		base := strings.ToLower(filepath.Base(destPath))
		if firstExe == "" && !strings.Contains(base, ".") ||
			strings.Contains(base, "patcher") || strings.Contains(base, "wiilink") {
			firstExe = destPath
		}
	}

	if firstExe == "" {
		return "", fmt.Errorf("no executable found in patcher archive")
	}
	return firstExe, nil
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
