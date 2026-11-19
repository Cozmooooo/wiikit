package library

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type Game struct {
	Path      string `json:"path"`
	Title     string `json:"title"`
	GameID    string `json:"gameId"`
	Region    string `json:"region"`
	SizeBytes int64  `json:"sizeBytes"`
}

var wiiExtensions = map[string]bool{
	".iso":  true,
	".wbfs": true,
}

func ScanLibrary(roots []string) ([]Game, error) {
	var games []Game

	for _, root := range roots {
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil // skip unreadable dirs
			}
			if d.IsDir() {
				return nil
			}

			ext := strings.ToLower(filepath.Ext(path))
			if !wiiExtensions[ext] {
				return nil
			}

			gameID, title, ok := ParseISOHeader(path)
			if !ok {
				return nil
			}

			info, err := os.Stat(path)
			if err != nil {
				return nil
			}

			games = append(games, Game{
				Path:      path,
				Title:     title,
				GameID:    gameID,
				Region:    regionFromID(gameID),
				SizeBytes: info.Size(),
			})
			return nil
		})
		if err != nil {
			return games, err
		}
	}

	return games, nil
}

func regionFromID(gameID string) string {
	if len(gameID) < 4 {
		return "Unknown"
	}
	switch gameID[3] {
	case 'E':
		return "NTSC-U"
	case 'P':
		return "PAL"
	case 'J':
		return "NTSC-J"
	case 'K':
		return "NTSC-K"
	default:
		return "Unknown"
	}
}
