package library

import (
	"encoding/binary"
	"os"
)

const wiiMagic uint32 = 0x5D1C9EA3

func ParseISOHeader(path string) (gameID, title string, ok bool) {
	f, err := os.Open(path)
	if err != nil {
		return "", "", false
	}
	defer f.Close()

	buf := make([]byte, 96)
	if _, err := f.Read(buf); err != nil {
		return "", "", false
	}

	magic := binary.BigEndian.Uint32(buf[0x18:0x1C])
	if magic != wiiMagic {
		return "", "", false
	}

	gameID = string(buf[0:6])

	// Title at offset 0x20, null-terminated, max 64 bytes
	titleBytes := buf[0x20:0x60]
	end := 0
	for end < len(titleBytes) && titleBytes[end] != 0 {
		end++
	}
	title = string(titleBytes[:end])
	if title == "" {
		title = gameID
	}

	return gameID, title, true
}
