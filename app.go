package main

import (
	"context"

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

func (a *App) StartInstall(games []library.Game, usbDrive drive.DriveInfo, sdDrive drive.DriveInfo, consoleType ConsoleType, ethernetEnabled bool) error {
	go func() {
		progress := make(chan drive.Progress, 32)
		go func() {
			for p := range progress {
				runtime.EventsEmit(a.ctx, "progress", p)
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

		// WiiLink: base Wii always, Wii Mini only when ethernet adapter is connected
		if consoleType == ConsoleWii || (consoleType == ConsoleWiiMini && ethernetEnabled) {
			emit("Running WiiLink patcher...")
			if err := wiilink.RunPatcher(a.ctx, sdDrive.Path, nil); err != nil {
				runtime.EventsEmit(a.ctx, "error", err.Error())
				return
			}
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
