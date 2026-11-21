package drive

import (
	"os"
	"path/filepath"
)

const howToAddGames = `HOW TO ADD YOUR OWN GAMES TO THIS WII
======================================

Your KRUNK Wii is set up with WiiFlow as the main game launcher.
It auto-detects games from both your USB drive and SD card.

USB DRIVE (Wii & GameCube games)
---------------------------------
  Wii games:       USB:\wbfs\<Title> [GAMEID]\<Title> [GAMEID].wbfs
                   Example: USB:\wbfs\Super Mario Galaxy [RMGE01]\RMGE01.wbfs

  GameCube games:  USB:\games\<Title> [GAMEID]\game.iso
                   Example: USB:\games\Super Smash Bros Melee [GALE01]\game.iso

  Tip: Use Wii Backup Manager (Windows) or Witgui (Mac) to convert
  and copy disc images. They handle the folder names automatically.

SD CARD (Emulator ROMs)
------------------------
  SNES roms:       SD:\roms\snes\<game>.sfc  or  .smc
  NES roms:        SD:\roms\nes\<game>.nes   or  .fds
  GBA roms:        SD:\roms\gba\<game>.gba
  N64 roms:        SD:\roms\n64\<game>.n64   or  .z64  or  .v64
  PS1 roms:        SD:\roms\ps1\<game>.bin + .cue  (or  .img  .pbp)

  After copying, open WiiFlow on the Wii and press + to scan for games.

COVERS (optional but looks great)
-----------------------------------
  WiiFlow downloads covers automatically over WiiLink when connected.
  You can also drop your own artwork into SD:\WiiFlow\covers\<system>\

NEED HELP?
-----------
  krunk.site/support
`

// WriteReadme writes HOW_TO_ADD_GAMES.txt to the USB drive root.
// Non-fatal: if the drive is read-only or missing, it returns nil.
func WriteReadme(usbRoot string) error {
	p := filepath.Join(usbRoot, "HOW_TO_ADD_GAMES.txt")
	return os.WriteFile(p, []byte(howToAddGames), 0644)
}
