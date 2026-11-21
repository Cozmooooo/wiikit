// Package wiiflow configures WiiFlow Lite on the SD card.
// It writes plugin INI files so every installed emulator appears
// in WiiFlow's unified library with the correct boot.dol path.
package wiiflow

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type plugin struct {
	name        string // display name in WiiFlow coverflow
	file        string // INI filename (without .ini)
	bootFile    string // SD-absolute path to emulator boot.dol
	fileTypes   string // comma-separated ROM extensions (no dots)
	coverFolder string // subdirectory under WiiFlow/covers/
}

// plugins maps every emulator WiiKit installs to its WiiFlow plugin config.
var plugins = []plugin{
	{
		name:        "Snes9xGX",
		file:        "snes9xgx",
		bootFile:    "/apps/snes9xgx/boot.dol",
		fileTypes:   "sfc,smc,fig,swc,bs",
		coverFolder: "snes",
	},
	{
		name:        "FCEUltraGX",
		file:        "fceugx",
		bootFile:    "/apps/fceugx/boot.dol",
		fileTypes:   "nes,fds,nsf,unf,unif",
		coverFolder: "nes",
	},
	{
		name:        "VisualBoyAdvanceGX",
		file:        "vbagx",
		bootFile:    "/apps/vbagx/boot.dol",
		fileTypes:   "gba,gb,gbc,sgb",
		coverFolder: "gba",
	},
	{
		name:        "Not64",
		file:        "not64",
		bootFile:    "/apps/not64/boot.dol",
		fileTypes:   "n64,v64,z64,rom,ndd",
		coverFolder: "n64",
	},
	{
		name:        "WiiStation",
		file:        "wiistation",
		bootFile:    "/apps/wiistation/boot.dol",
		fileTypes:   "bin,cue,img,mdf,pbp,cbn",
		coverFolder: "ps1",
	},
}

// pluginINI returns the contents of a WiiFlow Lite plugin INI file.
func pluginINI(p plugin) string {
	var b strings.Builder
	b.WriteString("[PLUGIN]\n")
	fmt.Fprintf(&b, "name = %s\n", p.name)
	b.WriteString("magic = 0x0\n")
	fmt.Fprintf(&b, "coverFolder = %s\n", p.coverFolder)
	fmt.Fprintf(&b, "bootFile = %s\n", p.bootFile)
	fmt.Fprintf(&b, "fileTypes = %s\n", p.fileTypes)
	return b.String()
}

// WritePlugins writes SD:/WiiFlow/plugins/*.ini for every emulator WiiKit
// installed, creates cover cache directories, and writes a minimal
// SD:/WiiFlow/wiiflow.ini wiring Nintendont (skipped if file already exists).
func WritePlugins(sdRoot string) error {
	pluginDir := filepath.Join(sdRoot, "WiiFlow", "plugins")
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return fmt.Errorf("mkdir WiiFlow/plugins: %w", err)
	}

	for _, p := range plugins {
		ini := pluginINI(p)
		dest := filepath.Join(pluginDir, p.file+".ini")
		if err := os.WriteFile(dest, []byte(ini), 0644); err != nil {
			return fmt.Errorf("write plugin %s: %w", p.name, err)
		}
		// Pre-create covers directory so WiiFlow finds it on first scan.
		coversDir := filepath.Join(sdRoot, "WiiFlow", "covers", p.coverFolder)
		if err := os.MkdirAll(coversDir, 0755); err != nil {
			return fmt.Errorf("mkdir covers/%s: %w", p.coverFolder, err)
		}
	}

	// Standard WiiFlow directories used by the app.
	for _, dir := range []string{"cache", "music", "themes"} {
		os.MkdirAll(filepath.Join(sdRoot, "WiiFlow", dir), 0755)
	}
	// Wii and GameCube cover folders (WiiFlow built-in, not plugins).
	for _, dir := range []string{"wii", "gc"} {
		os.MkdirAll(filepath.Join(sdRoot, "WiiFlow", "covers", dir), 0755)
	}

	// Write minimal wiiflow.ini only if it does not exist yet.
	iniPath := filepath.Join(sdRoot, "WiiFlow", "wiiflow.ini")
	if _, err := os.Stat(iniPath); os.IsNotExist(err) {
		ini := "[GENERAL]\nnintendont_path = /apps/Nintendont/boot.dol\n"
		if err := os.WriteFile(iniPath, []byte(ini), 0644); err != nil {
			return fmt.Errorf("write wiiflow.ini: %w", err)
		}
	}

	return nil
}
