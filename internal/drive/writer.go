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

// WriteGamesToUSB copies Wii/GCN titles to the USB drive and ROMs to the SD card.
func WriteGamesToUSB(ctx context.Context, games []library.Game, usbRoot string, progress chan<- Progress) error {
	return writeGames(ctx, games, usbRoot, "", progress)
}

// WriteGames routes each game to USB or SD based on its type.
func WriteGames(ctx context.Context, games []library.Game, usbRoot, sdRoot string, progress chan<- Progress) error {
	return writeGames(ctx, games, usbRoot, sdRoot, progress)
}

func writeGames(ctx context.Context, games []library.Game, usbRoot, sdRoot string, progress chan<- Progress) error {
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

			var err error
			if g.GoesOnUSB() {
				err = copyToUSB(ctx, g, usbRoot, progress)
			} else if sdRoot != "" {
				err = copyToSD(ctx, g, sdRoot, progress)
			}
			if err != nil {
				errs <- fmt.Errorf("%s (%s): %w", g.Title, g.Type, err)
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

// copyToUSB places a Wii title in /wbfs/<Title> [<ID>]/<ID>.wbfs
// and a GCN title in /games/<Title> [<ID>]/<ID>.iso (Nintendont layout).
func copyToUSB(ctx context.Context, g library.Game, usbRoot string, progress chan<- Progress) error {
	var destDir, destName string
	if g.Type == library.TypeGCN {
		dirName := fmt.Sprintf("%s [%s]", g.Title, g.GameID)
		destDir = filepath.Join(usbRoot, "games", dirName)
		destName = g.GameID + ".iso"
	} else {
		dirName := fmt.Sprintf("%s [%s]", g.Title, g.GameID)
		destDir = filepath.Join(usbRoot, "wbfs", dirName)
		destName = g.GameID + ".wbfs"
	}
	return copyFile(ctx, g, filepath.Join(destDir, destName), progress)
}

// copyToSD places a ROM in the emulator's standard SD subfolder.
func copyToSD(ctx context.Context, g library.Game, sdRoot string, progress chan<- Progress) error {
	destDir := filepath.Join(sdRoot, filepath.FromSlash(g.SDDestDir()))
	destPath := filepath.Join(destDir, filepath.Base(g.Path))
	return copyFile(ctx, g, destPath, progress)
}

func copyFile(ctx context.Context, g library.Game, destPath string, progress chan<- Progress) error {
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	tmpPath := destPath + ".tmp"

	src, err := os.Open(g.Path)
	if err != nil {
		return fmt.Errorf("open source: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("create dest: %w", err)
	}

	buf := make([]byte, 1024*1024)
	var bytesDone int64
	for {
		select {
		case <-ctx.Done():
			dst.Close()
			os.Remove(tmpPath)
			return ctx.Err()
		default:
		}

		n, err := src.Read(buf)
		if n > 0 {
			if _, werr := dst.Write(buf[:n]); werr != nil {
				dst.Close()
				os.Remove(tmpPath)
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
			os.Remove(tmpPath)
			return fmt.Errorf("read: %w", err)
		}
	}

	dst.Close()
	if err := os.Rename(tmpPath, destPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("rename: %w", err)
	}
	return nil
}
