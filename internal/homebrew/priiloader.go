package homebrew

import (
	"fmt"
	"os"
	"path/filepath"
)

// WritePriiloaderConfig writes a priiloader.cfg to the SD root that auto-boots
// USBLoader GX on console power-on, skipping the Wii Menu entirely.
//
// Priiloader reads this file from SD:/priiloader.cfg.
// The autoboot value "3" means "boot to a specific title"; combined with
// autoboot_title pointing at USBLoader GX's title ID (48415558-4c554458),
// the console launches straight into the game loader.
func WritePriiloaderConfig(sdRoot string) error {
	cfg := `# WiiKit Priiloader Configuration
# Generated automatically — edit with care.

# autoboot: 0=disabled, 1=WiiMenu, 2=priiloader, 3=installed_file, 4=BootMii
autoboot=3

# Title to boot: USBLoader GX (HAXLGX in hex = 48415857-47584758)
# This launches USBLoader GX from the Wii Menu channel.
autoboot_title=0001000148415857

# Skip disc check on boot
skip_title_intro=1

# Allow booting into Wii Menu by holding RESET
hold_reset_to_wiiMenu=1
`
	path := filepath.Join(sdRoot, "priiloader.cfg")
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(cfg), 0644); err != nil {
		return fmt.Errorf("write priiloader.cfg: %w", err)
	}
	return os.Rename(tmp, path)
}
