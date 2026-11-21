package wiiflow

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWritePlugins_AllINIsPresent(t *testing.T) {
	sd := t.TempDir()
	if err := WritePlugins(sd); err != nil {
		t.Fatalf("WritePlugins: %v", err)
	}

	// Every plugin must have an INI file.
	for _, p := range plugins {
		iniPath := filepath.Join(sd, "WiiFlow", "plugins", p.file+".ini")
		data, err := os.ReadFile(iniPath)
		if err != nil {
			t.Errorf("missing plugin INI for %s: %v", p.name, err)
			continue
		}
		content := string(data)
		if !strings.Contains(content, "[PLUGIN]") {
			t.Errorf("%s.ini missing [PLUGIN] header", p.file)
		}
		if !strings.Contains(content, p.bootFile) {
			t.Errorf("%s.ini missing bootFile %s", p.file, p.bootFile)
		}
		if !strings.Contains(content, p.fileTypes) {
			t.Errorf("%s.ini missing fileTypes %s", p.file, p.fileTypes)
		}
	}
}

func TestWritePlugins_CoverDirsCreated(t *testing.T) {
	sd := t.TempDir()
	if err := WritePlugins(sd); err != nil {
		t.Fatal(err)
	}

	// Plugin cover dirs
	for _, p := range plugins {
		dir := filepath.Join(sd, "WiiFlow", "covers", p.coverFolder)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("missing cover dir: WiiFlow/covers/%s", p.coverFolder)
		}
	}
	// Built-in Wii/GC cover dirs
	for _, name := range []string{"wii", "gc"} {
		dir := filepath.Join(sd, "WiiFlow", "covers", name)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("missing cover dir: WiiFlow/covers/%s", name)
		}
	}
}

func TestWritePlugins_WiiflowIni(t *testing.T) {
	sd := t.TempDir()
	if err := WritePlugins(sd); err != nil {
		t.Fatal(err)
	}

	iniPath := filepath.Join(sd, "WiiFlow", "wiiflow.ini")
	data, err := os.ReadFile(iniPath)
	if err != nil {
		t.Fatalf("wiiflow.ini missing: %v", err)
	}
	if !strings.Contains(string(data), "nintendont_path") {
		t.Error("wiiflow.ini missing nintendont_path")
	}
}

func TestWritePlugins_DoesNotOverwriteExistingIni(t *testing.T) {
	sd := t.TempDir()

	// Pre-create a wiiflow.ini with custom content
	wfDir := filepath.Join(sd, "WiiFlow")
	os.MkdirAll(wfDir, 0755)
	iniPath := filepath.Join(wfDir, "wiiflow.ini")
	os.WriteFile(iniPath, []byte("[GENERAL]\ncustom_setting = yes\n"), 0644)

	if err := WritePlugins(sd); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(iniPath)
	if !strings.Contains(string(data), "custom_setting") {
		t.Error("WritePlugins overwrote existing wiiflow.ini")
	}
}

func TestPluginINI_Format(t *testing.T) {
	p := plugin{
		name:        "TestEmu",
		file:        "testemu",
		bootFile:    "/apps/testemu/boot.dol",
		fileTypes:   "rom,bin",
		coverFolder: "test",
	}
	ini := pluginINI(p)

	required := []string{"[PLUGIN]", "name = TestEmu", "bootFile = /apps/testemu/boot.dol", "fileTypes = rom,bin", "coverFolder = test"}
	for _, r := range required {
		if !strings.Contains(ini, r) {
			t.Errorf("pluginINI missing %q", r)
		}
	}
}
