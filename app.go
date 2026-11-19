package main

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"wiikit/internal/drive"
	"wiikit/internal/library"
	"wiikit/internal/homebrew"
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

func (a *App) StartInstall(games []library.Game, usbDrive drive.DriveInfo, sdDrive drive.DriveInfo) error {
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

		emit("Copying games to USB...")
		if err := drive.WriteGamesToUSB(a.ctx, games, usbDrive.Path, progress); err != nil {
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
		if err := homebrew.InstallStack(a.ctx, manifest, sdDrive.Path, progress); err != nil {
			runtime.EventsEmit(a.ctx, "error", err.Error())
			return
		}

		emit("Running WiiLink patcher...")
		if err := wiilink.RunPatcher(a.ctx, sdDrive.Path, nil); err != nil {
			runtime.EventsEmit(a.ctx, "error", err.Error())
			return
		}

		emit("Done! Eject your drives and power on the Wii.")
		runtime.EventsEmit(a.ctx, "done", nil)
	}()
	return nil
}
