package drive

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

type DriveInfo struct {
	Path        string `json:"path"`
	Label       string `json:"label"`
	SizeBytes   int64  `json:"sizeBytes"`
	IsRemovable bool   `json:"isRemovable"`
}

func ListDrives() ([]DriveInfo, error) {
	switch runtime.GOOS {
	case "darwin":
		return listDrivesMac()
	case "windows":
		return listDrivesWindows()
	default:
		return listDrivesLinux()
	}
}

func listDrivesMac() ([]DriveInfo, error) {
	out, err := exec.Command("diskutil", "list", "-plist").Output()
	if err != nil {
		return nil, fmt.Errorf("diskutil list: %w", err)
	}

	var plist struct {
		AllDisks              []string `plist:"AllDisks"`
		AllDisksAndPartitions []struct {
			DeviceIdentifier string `plist:"DeviceIdentifier"`
			Size             int64  `plist:"Size"`
		} `plist:"AllDisksAndPartitions"`
	}
	_ = plist
	_ = out

	// Parse via diskutil info per disk for removable flag
	infoOut, err := exec.Command("diskutil", "list", "-plist", "external").Output()
	if err != nil {
		return nil, fmt.Errorf("diskutil list external: %w", err)
	}

	// Simple extraction — parse device identifiers from plist text
	var drives []DriveInfo
	lines := strings.Split(string(infoOut), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "<string>disk") && strings.HasSuffix(line, "</string>") {
			dev := strings.TrimPrefix(line, "<string>")
			dev = strings.TrimSuffix(dev, "</string>")
			if !strings.Contains(dev, "s") { // whole disk only, not partitions
				info, err := getDiskInfoMac("/dev/" + dev)
				if err == nil {
					drives = append(drives, info)
				}
			}
		}
	}
	return drives, nil
}

func getDiskInfoMac(path string) (DriveInfo, error) {
	out, err := exec.Command("diskutil", "info", "-plist", path).Output()
	if err != nil {
		return DriveInfo{}, err
	}
	s := string(out)
	info := DriveInfo{Path: path, IsRemovable: true}

	if idx := strings.Index(s, "<key>VolumeName</key>"); idx >= 0 {
		rest := s[idx+len("<key>VolumeName</key>"):]
		if start := strings.Index(rest, "<string>"); start >= 0 {
			end := strings.Index(rest[start+8:], "</string>")
			if end >= 0 {
				info.Label = rest[start+8 : start+8+end]
			}
		}
	}
	if idx := strings.Index(s, "<key>TotalSize</key>"); idx >= 0 {
		rest := s[idx+len("<key>TotalSize</key>"):]
		if start := strings.Index(rest, "<integer>"); start >= 0 {
			end := strings.Index(rest[start+9:], "</integer>")
			if end >= 0 {
				info.SizeBytes, _ = strconv.ParseInt(rest[start+9:start+9+end], 10, 64)
			}
		}
	}
	return info, nil
}

func listDrivesWindows() ([]DriveInfo, error) {
	out, err := exec.Command("powershell", "-NoProfile", "-Command",
		"Get-WmiObject Win32_LogicalDisk | Where-Object {$_.DriveType -eq 2} | Select-Object DeviceID,VolumeName,Size | ConvertTo-Json").Output()
	if err != nil {
		// fallback: list all drives
		out, err = exec.Command("powershell", "-NoProfile", "-Command",
			"Get-WmiObject Win32_LogicalDisk | Select-Object DeviceID,VolumeName,Size,DriveType | ConvertTo-Json").Output()
		if err != nil {
			return nil, fmt.Errorf("get drives: %w", err)
		}
	}

	var raw interface{}
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, fmt.Errorf("parse drives json: %w", err)
	}

	var drives []DriveInfo
	parse := func(obj map[string]interface{}) {
		d := DriveInfo{IsRemovable: true}
		if v, ok := obj["DeviceID"].(string); ok {
			d.Path = v + "\\"
		}
		if v, ok := obj["VolumeName"].(string); ok {
			d.Label = v
		}
		if v, ok := obj["Size"].(float64); ok {
			d.SizeBytes = int64(v)
		}
		if d.Path != "" {
			drives = append(drives, d)
		}
	}

	switch v := raw.(type) {
	case []interface{}:
		for _, item := range v {
			if obj, ok := item.(map[string]interface{}); ok {
				parse(obj)
			}
		}
	case map[string]interface{}:
		parse(v)
	}
	return drives, nil
}

type lsblkOutput struct {
	Blockdevices []lsblkDevice `json:"blockdevices"`
}

type lsblkDevice struct {
	Name      string `json:"name"`
	Size      string `json:"size"`
	Label     string `json:"label"`
	Hotplug   bool   `json:"hotplug"`
	Mountpoint string `json:"mountpoint"`
}

func listDrivesLinux() ([]DriveInfo, error) {
	out, err := exec.Command("lsblk", "-J", "-o", "NAME,SIZE,LABEL,HOTPLUG,MOUNTPOINT").Output()
	if err != nil {
		return nil, fmt.Errorf("lsblk: %w", err)
	}

	var result lsblkOutput
	if err := json.Unmarshal(out, &result); err != nil {
		return nil, fmt.Errorf("parse lsblk: %w", err)
	}

	var drives []DriveInfo
	for _, dev := range result.Blockdevices {
		if !dev.Hotplug {
			continue
		}
		d := DriveInfo{
			Path:        "/dev/" + dev.Name,
			Label:       dev.Label,
			IsRemovable: dev.Hotplug,
		}
		drives = append(drives, d)
	}
	return drives, nil
}

func PreviewFormat(d DriveInfo, label string) string {
	gb := float64(d.SizeBytes) / 1e9
	return fmt.Sprintf("Will erase %s (%s, %.1f GB) and format as FAT32 labeled %q. THIS CANNOT BE UNDONE.", d.Path, d.Label, gb, label)
}
