package library

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// GameType identifies where a file belongs and which emulator loads it.
type GameType string

const (
	TypeWii  GameType = "wii"  // ISO/WBFS → USB /wbfs/
	TypeGCN  GameType = "gcn"  // ISO       → USB /games/ (Nintendont)
	TypeN64  GameType = "n64"  // ROM       → SD  /roms/N64/
	TypePS1  GameType = "ps1"  // ROM/ISO   → SD  /roms/PSX/
	TypeSNES GameType = "snes" // ROM       → SD  /roms/snes/
	TypeNES  GameType = "nes"  // ROM       → SD  /roms/nes/
	TypeGBA  GameType = "gba"  // ROM       → SD  /roms/gba/
)

type Game struct {
	Path      string   `json:"path"`
	Title     string   `json:"title"`
	GameID    string   `json:"gameId"`
	Region    string   `json:"region"`
	SizeBytes int64    `json:"sizeBytes"`
	Type      GameType `json:"type"`
}

// SDDestDir returns the SD card subdirectory for ROM types. Empty for USB types.
func (g Game) SDDestDir() string {
	switch g.Type {
	case TypeN64:
		return "roms/N64"
	case TypePS1:
		return "roms/PSX"
	case TypeSNES:
		return "roms/snes"
	case TypeNES:
		return "roms/nes"
	case TypeGBA:
		return "roms/gba"
	default:
		return ""
	}
}

// GoesOnUSB reports whether this game belongs on the USB drive.
func (g Game) GoesOnUSB() bool {
	return g.Type == TypeWii || g.Type == TypeGCN
}

// extToType maps file extensions to game types.
var extToType = map[string]GameType{
	// Wii
	".iso":  TypeWii,
	".wbfs": TypeWii,
	// N64
	".z64": TypeN64,
	".n64": TypeN64,
	".v64": TypeN64,
	// PS1
	".bin": TypePS1,
	".cue": TypePS1,
	".img": TypePS1,
	// SNES
	".smc": TypeSNES,
	".sfc": TypeSNES,
	".fig": TypeSNES,
	// NES
	".nes": TypeNES,
	".fds": TypeNES,
	// GBA
	".gba": TypeGBA,
	".gbc": TypeGBA,
	".gb":  TypeGBA,
}

func ScanLibrary(roots []string) ([]Game, error) {
	var games []Game

	for _, root := range roots {
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				return nil
			}

			ext := strings.ToLower(filepath.Ext(path))
			gType, ok := extToType[ext]
			if !ok {
				return nil
			}

			info, err := os.Stat(path)
			if err != nil {
				return nil
			}

			// Wii ISOs: parse header for title + game ID.
			// Falls back to PS1 when the header doesn't match — PS1 games are
			// commonly distributed as .iso and would otherwise be silently dropped.
			if gType == TypeWii {
				gameID, title, ok := ParseISOHeader(path)
				if !ok {
					// Not a Wii/GCN disc image — treat as PS1
					base := filepath.Base(path)
					ps1Title := strings.TrimSuffix(base, filepath.Ext(base))
					games = append(games, Game{
						Path:      path,
						Title:     ps1Title,
						SizeBytes: info.Size(),
						Type:      TypePS1,
					})
					return nil
				}
				// GCN ISOs have a different magic — recheck type
				magic2, _ := readMagic(path, 0x1C)
				if magic2 == gcnMagic {
					gType = TypeGCN
				}
				games = append(games, Game{
					Path:      path,
					Title:     title,
					GameID:    gameID,
					Region:    regionFromID(gameID),
					SizeBytes: info.Size(),
					Type:      gType,
				})
				return nil
			}

			// ROMs: use filename as title
			base := filepath.Base(path)
			title := strings.TrimSuffix(base, filepath.Ext(base))
			games = append(games, Game{
				Path:      path,
				Title:     title,
				SizeBytes: info.Size(),
				Type:      gType,
			})
			return nil
		})
		if err != nil {
			return games, err
		}
	}

	games = deduplicateAndSort(games)
	return games, nil
}

// deduplicateAndSort removes duplicate Wii/GCN games by GameID (keeps largest file),
// then sorts all games by title within each type.
func deduplicateAndSort(games []Game) []Game {
	seen := make(map[string]int) // GameID → index in result
	result := make([]Game, 0, len(games))

	for _, g := range games {
		if g.GameID == "" {
			result = append(result, g)
			continue
		}
		key := strings.ToUpper(g.GameID)
		if idx, exists := seen[key]; exists {
			if g.SizeBytes > result[idx].SizeBytes {
				result[idx] = g
			}
		} else {
			seen[key] = len(result)
			result = append(result, g)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].Type != result[j].Type {
			return result[i].Type < result[j].Type
		}
		return strings.ToLower(result[i].Title) < strings.ToLower(result[j].Title)
	})
	return result
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
