package homebrew

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"wiikit/internal/audit"
)

// DownloadAsset returns a local path to asset — from persistent cache when
// available, otherwise downloaded fresh. destDir is the fallback temp directory
// for assets that cannot be cached (SHA256 == "").
func DownloadAsset(ctx context.Context, asset Asset, destDir string) (string, error) {
	// Persistent cache hit: no download, no verification needed
	// (the filename IS the sha256; written only after verification)
	if p := cachedPath(asset); p != "" {
		return p, nil
	}

	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 5*time.Minute)
		defer cancel()
	}

	// Download to a temp file inside the cache dir when possible (same device
	// as the final destination → rename is instant, no cross-device copy).
	var tmpDir string
	if cacheDir, err := persistentCacheDir(); err == nil && asset.SHA256 != "" {
		tmpDir = cacheDir
	} else {
		tmpDir = destDir
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, asset.URL, nil)
	if err != nil {
		return "", fmt.Errorf("build request for %s: %w", asset.Name, err)
	}
	req.Header.Set("User-Agent", "WiiKit/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("download %s: %w", asset.Name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download %s: HTTP %d", asset.Name, resp.StatusCode)
	}

	tmpFile, err := os.CreateTemp(tmpDir, "wiikit-dl-*.tmp")
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	_, err = io.Copy(tmpFile, resp.Body)
	tmpFile.Close()
	if err != nil {
		os.Remove(tmpPath)
		return "", fmt.Errorf("write %s: %w", asset.Name, err)
	}

	if asset.SHA256 != "" {
		if err := audit.VerifyFile(tmpPath, asset.SHA256); err != nil {
			os.Remove(tmpPath)
			return "", err
		}
		// Move verified file into persistent cache; filename becomes the hash
		return storeInCache(asset, tmpPath), nil
	}

	// No hash → leave in destDir with a stable name
	finalPath := filepath.Join(destDir, asset.Name+filepath.Ext(asset.URL))
	if err := os.Rename(tmpPath, finalPath); err != nil {
		os.Remove(tmpPath)
		return "", fmt.Errorf("rename %s: %w", asset.Name, err)
	}
	return finalPath, nil
}
