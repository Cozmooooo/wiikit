package main

import (
	"context"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"wiikit/internal/deals"
	"wiikit/internal/drive"
	"wiikit/internal/homebrew"
	"wiikit/internal/library"
	"wiikit/internal/wiiflow"
	"wiikit/internal/wiilink"
)

type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	go func() {
		for event := range drive.WatchDrives(ctx, 2*time.Second) {
			runtime.EventsEmit(a.ctx, "drive-"+event.Type, event.Drive)
		}
	}()
}

func (a *App) ClearCache() (int, error) {
	return homebrew.ClearCache()
}

func (a *App) ListDrives() ([]drive.DriveInfo, error) {
	return drive.ListDrives()
}

func (a *App) PreviewFormat(d drive.DriveInfo, label string) string {
	return drive.PreviewFormat(d, label)
}

func (a *App) ScanLibrary(roots []string) ([]library.Game, error) {
	return library.ScanLibrary(roots)
}

func (a *App) OpenDirectoryDialog() (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Game Library Folder",
	})
}

func (a *App) SearchDeals(query string, maxPrice float64, minPrice float64) deals.SearchResult {
	params := deals.SearchParams{
		Query:    query,
		MaxPrice: maxPrice,
		MinPrice: minPrice,
		Limit:    25,
	}
	if params.Query == "" {
		params = deals.DefaultWiiParams()
	}
	return deals.Search(a.ctx, params)
}

// ConsoleType mirrors the frontend store value.
type ConsoleType string

const (
	ConsoleWii     ConsoleType = "wii"
	ConsoleWiiU    ConsoleType = "wiiu"
	ConsoleWiiMini ConsoleType = "wiimini"
)

// StartInstall runs the full Wii prep pipeline.
// reuseSD: skip SD format, homebrew install, and WiiFlow config (SD already prepped).
// skipWiiLink: skip the interactive WiiLink patcher (useful for mass production).
func (a *App) StartInstall(games []library.Game, usbDrive drive.DriveInfo, sdDrive drive.DriveInfo, consoleType ConsoleType, ethernetEnabled bool, reuseSD bool, skipWiiLink bool) error {
	go func() {
		progress := make(chan drive.Progress, 32)

		// Compute total bytes that will be copied so we can emit cumulative progress.
		var totalBytes int64
		for _, g := range games {
			if reuseSD && !g.GoesOnUSB() {
				continue // SD not being written in reuseSD mode
			}
			totalBytes += g.SizeBytes
		}

		// Fan-in: accumulate per-game deltas into one smooth overall percentage.
		go func() {
			lastSeen := make(map[string]int64)
			var cumBytes int64
			for p := range progress {
				delta := p.BytesDone - lastSeen[p.GameID]
				if delta > 0 {
					cumBytes += delta
					lastSeen[p.GameID] = p.BytesDone
				}
				runtime.EventsEmit(a.ctx, "progress", drive.Progress{
					Stage:      p.Title,
					BytesDone:  cumBytes,
					BytesTotal: totalBytes,
				})
			}
		}()

		emit := func(msg string) {
			runtime.EventsEmit(a.ctx, "log", msg)
		}

		emit("Formatting USB drive...")
		if err := drive.FormatFAT32(a.ctx, usbDrive, "WIIGAMES"); err != nil {
			runtime.EventsEmit(a.ctx, "error", err.Error())
			return
		}

		if reuseSD {
			emit("SD card: reusing existing setup (skipping format + homebrew reinstall).")
			emit("Copying Wii/GCN games to USB...")
			if err := drive.WriteGamesToUSB(a.ctx, games, usbDrive.Path, progress); err != nil {
				runtime.EventsEmit(a.ctx, "error", err.Error())
				return
			}
		} else {
			emit("Formatting SD card...")
			if err := drive.FormatFAT32(a.ctx, sdDrive, "WIISD"); err != nil {
				runtime.EventsEmit(a.ctx, "error", err.Error())
				return
			}

			emit("Copying games to USB/SD...")
			if err := drive.WriteGames(a.ctx, games, usbDrive.Path, sdDrive.Path, progress); err != nil {
				runtime.EventsEmit(a.ctx, "error", err.Error())
				return
			}

			emit("Loading homebrew manifest...")
			manifest, err := homebrew.LoadManifest("manifest.json")
			if err != nil {
				runtime.EventsEmit(a.ctx, "error", err.Error())
				return
			}

			emit("Downloading and installing homebrew stack...")
			if err := homebrew.InstallStack(a.ctx, manifest, sdDrive.Path, string(consoleType), progress); err != nil {
				runtime.EventsEmit(a.ctx, "error", err.Error())
				return
			}

			emit("Writing WiiFlow plugin configs...")
			if err := wiiflow.WritePlugins(sdDrive.Path); err != nil {
				emit("Warning: wiiflow plugins: " + err.Error())
			}

			// Splash screen + Priiloader (Wii only)
			if consoleType == ConsoleWii {
				emit("Downloading Krunk splash screen...")
				if err := homebrew.WriteSplashDOL(a.ctx, sdDrive.Path, emit); err != nil {
					emit("Warning: splash screen skipped: " + err.Error())
				}
				emit("Writing Priiloader auto-boot config...")
				if err := homebrew.WritePriiloaderConfig(sdDrive.Path); err != nil {
					emit("Warning: could not write priiloader.cfg: " + err.Error())
				}
			}
		}

		if err := drive.WriteReadme(usbDrive.Path); err != nil {
			emit("Warning: could not write HOW_TO_ADD_GAMES.txt: " + err.Error())
		}

		// WiiLink: skip when reuseSD (already patched) or skipWiiLink flag set
		wiiLinkNeeded := !skipWiiLink && !reuseSD &&
			(consoleType == ConsoleWii || (consoleType == ConsoleWiiMini && ethernetEnabled))
		if wiiLinkNeeded {
			emit("Running WiiLink patcher...")
			if err := wiilink.RunPatcher(a.ctx, sdDrive.Path, nil); err != nil {
				runtime.EventsEmit(a.ctx, "error", err.Error())
				return
			}
		} else if skipWiiLink {
			emit("WiiLink: skipped.")
		} else {
			emit("Skipping WiiLink (not supported on " + string(consoleType) + ").")
		}

		switch consoleType {
		case ConsoleWii:
			emit("SD ready. LetterBomb exploit: visit please.hackmii.com on another device,")
			emit("  enter your Wii's MAC address (Settings › Internet › Console Info), download the ZIP,")
			emit("  copy 'private' folder and 'boot.elf' to the root of this SD card, then follow on-screen steps.")
		case ConsoleWiiU:
			emit("Wii U note: install Tiramisu or Aroma CFW first, then access vWii.")
			emit("  Games load via USBLoader GX in vWii mode. WiiLink is not available on vWii.")
		case ConsoleWiiMini:
			emit("Wii Mini note: no built-in WiFi — LetterBomb won't work.")
			emit("  Use BlueBomb (requires a Bluetooth-capable Linux device nearby).")
			emit("  See: github.com/Fullmetal5/bluebomb for setup instructions.")
			if ethernetEnabled {
				emit("Ethernet Enabler is on your SD card (apps/Wii Mini Ethernet Enable/).")
				emit("  Run it from HBC once — it makes a one-time NAND change to enable USB Ethernet.")
				emit("  Recommended adapter: ASIX AX88179 USB Ethernet (≤$8 on Amazon).")
				emit("  After running it, plug in your USB adapter and WiiLink will be available.")
			}
		}

		emit("Done! Eject your drives and power on the console.")
		runtime.EventsEmit(a.ctx, "done", nil)
	}()
	return nil
}
