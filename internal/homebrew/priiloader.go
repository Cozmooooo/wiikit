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

# autoboot: 0=disabled, 1=WiiMenu, 2=priiloader, 3=installed_file, 4=BootMii, 5=sd:/boot.dol
autoboot=5

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
