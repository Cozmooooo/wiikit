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

func DownloadAsset(ctx context.Context, asset Asset, destDir string) (string, error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 5*time.Minute)
		defer cancel()
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

	tmpFile, err := os.CreateTemp(destDir, "wiikit-dl-*.tmp")
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

	if err := audit.VerifyFile(tmpPath, asset.SHA256); err != nil {
		os.Remove(tmpPath)
		return "", err
	}

	finalPath := filepath.Join(destDir, asset.Name+filepath.Ext(asset.URL))
	if err := os.Rename(tmpPath, finalPath); err != nil {
		os.Remove(tmpPath)
		return "", fmt.Errorf("rename %s: %w", asset.Name, err)
	}

	return finalPath, nil
}
