package homebrew

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"wiikit/internal/drive"
)

func InstallStack(ctx context.Context, manifest *Manifest, sdRoot string, progress chan<- drive.Progress) error {
	cacheDir, err := os.MkdirTemp("", "wiikit-cache-*")
	if err != nil {
		return fmt.Errorf("create cache dir: %w", err)
	}
	defer os.RemoveAll(cacheDir)

	for _, asset := range manifest.Assets {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if progress != nil {
			select {
			case progress <- drive.Progress{Stage: "downloading " + asset.Name}:
			default:
			}
		}

		localPath, err := DownloadAsset(ctx, asset, cacheDir)
		if err != nil {
			return fmt.Errorf("download %s: %w", asset.Name, err)
		}

		destPath := resolveDestPath(asset.DestPath, sdRoot)
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return fmt.Errorf("mkdir for %s: %w", asset.Name, err)
		}

		if strings.HasSuffix(strings.ToLower(localPath), ".zip") {
			if err := extractZip(localPath, filepath.Dir(destPath)); err != nil {
				return fmt.Errorf("extract %s: %w", asset.Name, err)
			}
		} else {
			if err := copyFile(localPath, destPath); err != nil {
				return fmt.Errorf("install %s: %w", asset.Name, err)
			}
		}
	}
	return nil
}

func resolveDestPath(destPath, sdRoot string) string {
	p := strings.TrimPrefix(destPath, "sd://")
	return filepath.Join(sdRoot, filepath.FromSlash(p))
}

func extractZip(zipPath, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(destDir, filepath.FromSlash(f.Name))
		if !strings.HasPrefix(fpath, filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("zip path traversal detected: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, 0755)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}
		dst, err := os.Create(fpath)
		if err != nil {
			rc.Close()
			return err
		}
		_, err = io.Copy(dst, rc)
		dst.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}
