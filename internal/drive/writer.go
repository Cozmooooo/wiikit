package drive

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"wiikit/internal/library"
)

type Progress struct {
	GameID     string `json:"gameId"`
	Title      string `json:"title"`
	BytesDone  int64  `json:"bytesDone"`
	BytesTotal int64  `json:"bytesTotal"`
	Stage      string `json:"stage"`
}

func WriteGamesToUSB(ctx context.Context, games []library.Game, usbRoot string, progress chan<- Progress) error {
	sem := make(chan struct{}, 2) // max 2 concurrent copies
	var wg sync.WaitGroup
	errs := make(chan error, len(games))

	for _, g := range games {
		g := g
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			if err := copyGame(ctx, g, usbRoot, progress); err != nil {
				errs <- fmt.Errorf("game %s: %w", g.GameID, err)
			}
		}()
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}

func copyGame(ctx context.Context, g library.Game, usbRoot string, progress chan<- Progress) error {
	dirName := fmt.Sprintf("%s [%s]", g.Title, g.GameID)
	destDir := filepath.Join(usbRoot, "wbfs", dirName)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	destFile := filepath.Join(destDir, g.GameID+".wbfs")
	tmpFile := destFile + ".tmp"

	src, err := os.Open(g.Path)
	if err != nil {
		return fmt.Errorf("open source: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(tmpFile)
	if err != nil {
		return fmt.Errorf("create dest: %w", err)
	}

	buf := make([]byte, 1024*1024) // 1 MB chunks
	var bytesDone int64
	for {
		select {
		case <-ctx.Done():
			dst.Close()
			os.Remove(tmpFile)
			return ctx.Err()
		default:
		}

		n, err := src.Read(buf)
		if n > 0 {
			if _, werr := dst.Write(buf[:n]); werr != nil {
				dst.Close()
				os.Remove(tmpFile)
				return fmt.Errorf("write: %w", werr)
			}
			bytesDone += int64(n)
			if progress != nil {
				select {
				case progress <- Progress{GameID: g.GameID, Title: g.Title, BytesDone: bytesDone, BytesTotal: g.SizeBytes, Stage: "copying"}:
				default:
				}
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			dst.Close()
			os.Remove(tmpFile)
			return fmt.Errorf("read: %w", err)
		}
	}

	dst.Close()
	if err := os.Rename(tmpFile, destFile); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("rename: %w", err)
	}
	return nil
}
