package homebrew

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// makeZip creates a temporary zip file with the given files (name → content).
func makeZip(t *testing.T, files map[string]string) string {
	t.Helper()
	f, err := os.CreateTemp("", "wiikit-test-*.zip")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	w := zip.NewWriter(f)
	for name, content := range files {
		fw, err := w.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		fw.Write([]byte(content))
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return f.Name()
}

// walkFiles returns all file paths under root, relative to root.
func walkFiles(t *testing.T, root string) []string {
	t.Helper()
	var paths []string
	filepath.Walk(root, func(p string, info os.FileInfo, _ error) error {
		if !info.IsDir() {
			rel, _ := filepath.Rel(root, p)
			paths = append(paths, filepath.ToSlash(rel))
		}
		return nil
	})
	return paths
}

// --- resolveDestPath ---

func TestResolveDestPath(t *testing.T) {
	sd := t.TempDir()

	cases := []struct{ dest, suffix string }{
		{"sd://", ""},
		{"sd://apps/", "apps"},
		{"sd://boot.dol", "boot.dol"},
		{"sd://apps/Wii Mini Ethernet Enable/", "apps/Wii Mini Ethernet Enable"},
	}
	for _, c := range cases {
		got := resolveDestPath(c.dest, sd)
		want := sd
		if c.suffix != "" {
			want = filepath.Join(sd, filepath.FromSlash(c.suffix))
		}
		if got != want {
			t.Errorf("resolveDestPath(%q) = %q, want %q", c.dest, got, want)
		}
	}
}

// --- extractZip ---

// Standard HBC app ZIP: internal root is apps/<name>/
// dest "sd://" → extractDir = sdRoot
// Expected: sdRoot/apps/<name>/boot.dol
func TestExtractZip_StandardApp(t *testing.T) {
	zipPath := makeZip(t, map[string]string{
		"apps/priiloader/boot.dol":       "fake dol",
		"apps/priiloader/hacks_hash.ini": "{}",
		"apps/priiloader/icon.png":       "png",
		"LICENSE":                        "GPL",
	})
	defer os.Remove(zipPath)

	sdRoot := t.TempDir()

	if err := extractZip(zipPath, sdRoot); err != nil {
		t.Fatalf("extractZip: %v", err)
	}

	got := walkFiles(t, sdRoot)
	t.Logf("extracted files: %v", got)

	want := "apps/priiloader/boot.dol"
	found := false
	for _, p := range got {
		if p == want {
			found = true
		}
	}
	if !found {
		t.Errorf("missing %s; got %v", want, got)
	}

	// Confirm the double-nesting bug is gone
	bad := "apps/apps/priiloader/boot.dol"
	for _, p := range got {
		if strings.Contains(p, "apps/apps") {
			t.Errorf("double-nesting bug still present: found %s", p)
		}
		_ = bad
	}
}

// Ethernet Enabler ZIP: internal root is "Wii Mini Ethernet Enable/"
// dest "sd://apps/" → extractDir = sdRoot/apps
// Expected: sdRoot/apps/Wii Mini Ethernet Enable/boot.dol
func TestExtractZip_EthernetEnabler(t *testing.T) {
	zipPath := makeZip(t, map[string]string{
		"Wii Mini Ethernet Enable/boot.dol": "fake eth enabler",
		"Wii Mini Ethernet Enable/meta.xml": "<app/>",
		"Wii Mini Ethernet Enable/icon.png": "png",
	})
	defer os.Remove(zipPath)

	sdRoot := t.TempDir()
	appsDir := filepath.Join(sdRoot, "apps")
	os.MkdirAll(appsDir, 0755)

	if err := extractZip(zipPath, appsDir); err != nil {
		t.Fatalf("extractZip: %v", err)
	}

	got := walkFiles(t, sdRoot)
	t.Logf("extracted files: %v", got)

	want := "apps/Wii Mini Ethernet Enable/boot.dol"
	found := false
	for _, p := range got {
		if p == want {
			found = true
		}
	}
	if !found {
		t.Errorf("missing %s; got %v", want, got)
	}
}

// Non-ZIP asset (a .dol file) goes to the exact dest path.
func TestCopyFile_ExactDest(t *testing.T) {
	src, err := os.CreateTemp("", "wiikit-src-*.dol")
	if err != nil {
		t.Fatal(err)
	}
	src.WriteString("dol content")
	src.Close()
	defer os.Remove(src.Name())

	sdRoot := t.TempDir()
	dst := filepath.Join(sdRoot, "boot.dol")

	if err := copyFile(src.Name(), dst); err != nil {
		t.Fatalf("copyFile: %v", err)
	}

	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read dst: %v", err)
	}
	if string(data) != "dol content" {
		t.Errorf("content mismatch: %q", data)
	}
}

// RegressionTest: documents the OLD bug explicitly.
// With the old code, extractDir = filepath.Dir(destPath) where
// destPath = sdRoot/apps/priiloader → extractDir = sdRoot/apps.
// Extracting a zip with apps/priiloader/boot.dol into sdRoot/apps
// produced sdRoot/apps/apps/priiloader/boot.dol — double nested.
func TestRegression_NoDoubleNesting(t *testing.T) {
	zipPath := makeZip(t, map[string]string{
		"apps/priiloader/boot.dol": "dol",
	})
	defer os.Remove(zipPath)

	sdRoot := t.TempDir()

	// New behaviour: extract to sdRoot (dest="sd://")
	extractZip(zipPath, sdRoot)

	correct := filepath.Join(sdRoot, "apps", "priiloader", "boot.dol")
	doubled := filepath.Join(sdRoot, "apps", "apps", "priiloader", "boot.dol")

	if _, err := os.Stat(correct); err != nil {
		t.Error("correct path missing — extraction bug present")
	}
	if _, err := os.Stat(doubled); err == nil {
		t.Error("double-nested path exists — old bug not fixed")
	}
}
