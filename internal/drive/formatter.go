package drive

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func FormatFAT32(ctx context.Context, d DriveInfo, label string) error {
	if !d.IsRemovable {
		return fmt.Errorf("drive %s is not removable — refusing to format", d.Path)
	}
	if d.SizeBytes < 512*1024*1024 {
		return fmt.Errorf("drive %s is too small (%.0f MB) — minimum 512 MB", d.Path, float64(d.SizeBytes)/1024/1024)
	}

	switch runtime.GOOS {
	case "darwin":
		return formatMac(ctx, d.Path, label)
	case "windows":
		return formatWindows(ctx, d.Path, label)
	default:
		return formatLinux(ctx, d.Path, label)
	}
}

func formatMac(ctx context.Context, path, label string) error {
	cmd := exec.CommandContext(ctx, "diskutil", "eraseDisk", "FAT32", label, path)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("diskutil eraseDisk: %w\n%s", err, out)
	}
	return nil
}

func formatWindows(ctx context.Context, path, label string) error {
	// Extract drive letter (e.g. "E:\\" -> "E")
	driveLetter := strings.TrimSuffix(strings.TrimSuffix(path, "\\"), "/")
	if len(driveLetter) > 1 {
		driveLetter = driveLetter[:1]
	}

	script := fmt.Sprintf("select volume %s\nformat fs=fat32 label=%q quick\n", driveLetter, label)
	tmpFile, err := os.CreateTemp("", "wiikit-diskpart-*.txt")
	if err != nil {
		return fmt.Errorf("create diskpart script: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(script); err != nil {
		return fmt.Errorf("write diskpart script: %w", err)
	}
	tmpFile.Close()

	cmd := exec.CommandContext(ctx, "diskpart", "/s", tmpFile.Name())
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("diskpart: %w\n%s", err, out)
	}
	return nil
}

func formatLinux(ctx context.Context, path, label string) error {
	cmd := exec.CommandContext(ctx, "mkfs.fat", "-F", "32", "-n", label, path)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("mkfs.fat: %w\n%s", err, out)
	}
	return nil
}
