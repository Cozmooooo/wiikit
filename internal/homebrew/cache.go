package homebrew

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// persistentCacheDir returns (and creates) %LOCALAPPDATA%\WiiKit on Windows,
// ~/Library/Caches/WiiKit on macOS, ~/.cache/WiiKit on Linux.
func persistentCacheDir() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("user cache dir: %w", err)
	}
	dir := filepath.Join(base, "WiiKit")
	return dir, os.MkdirAll(dir, 0755)
}

// cacheFilename returns "<sha256><ext>" for an asset.
func cacheFilename(asset Asset) string {
	ext := filepath.Ext(asset.URL)
	if i := strings.IndexByte(ext, '?'); i >= 0 {
		ext = ext[:i] // strip query string from extension
	}
	if ext == "" {
		ext = ".bin"
	}
	return asset.SHA256 + ext
}

// cachedPath returns the full path to a verified cached asset, or "".
// Assets with no SHA256 are never cached.
func cachedPath(asset Asset) string {
	if asset.SHA256 == "" {
		return ""
	}
	dir, err := persistentCacheDir()
	if err != nil {
		return ""
	}
	p := filepath.Join(dir, cacheFilename(asset))
	info, err := os.Stat(p)
	if err != nil || info.Size() == 0 {
		return ""
	}
	return p // filename IS the hash — written only after sha256 verification
}

// storeInCache moves tmpPath into the persistent cache for asset.
// Returns the final cache path. Non-fatal: returns tmpPath on any error.
func storeInCache(asset Asset, tmpPath string) string {
	if asset.SHA256 == "" {
		return tmpPath
	}
	dir, err := persistentCacheDir()
	if err != nil {
		return tmpPath
	}
	dest := filepath.Join(dir, cacheFilename(asset))
	if err := os.Rename(tmpPath, dest); err != nil {
		// Cross-device: fall back to copy + delete
		if err2 := copyFile(tmpPath, dest); err2 != nil {
			return tmpPath
		}
		os.Remove(tmpPath)
	}
	return dest
}

// ClearCache removes all cached downloads. Returns the number of files removed.
func ClearCache() (int, error) {
	dir, err := persistentCacheDir()
	if err != nil {
		return 0, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, e := range entries {
		if !e.IsDir() {
			os.Remove(filepath.Join(dir, e.Name()))
			n++
		}
	}
	return n, nil
}
