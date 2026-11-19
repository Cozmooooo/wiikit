# WiiKit Core Pipeline — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a cross-platform (macOS/Windows/Linux) Go + Wails desktop app that formats USB/SD drives and deploys a full modded Wii stack (games + homebrew + WiiLink) in 2–3 clicks.

**Architecture:** Go backend exposes drive detection, formatting, game copying, and homebrew download/install as Wails-bound methods; Svelte frontend presents a 3-step wizard (Library → Drives → Install); all third-party binaries are fetched at runtime and verified by SHA-256 against a pinned manifest before writing to any drive.

**Tech Stack:** Go 1.22+, Wails v2, Svelte 4, Node 20 (frontend build only), `diskutil`/`diskpart`/`parted` via exec for formatting, Go standard library only for file I/O.

**Spec:** `docs/superpowers/specs/2026-11-18-wiikit-design.md`

## Global Constraints

- Go 1.22+ minimum; no CGO anywhere in the codebase
- Wails v2 (not v3 alpha); Svelte frontend, TypeScript enabled
- All third-party binaries: fetched at runtime, never bundled; SHA-256 verified against `manifest.json` before touching any drive
- Game folder structure on USB: `/wbfs/<Title> [GAMEID]/GAMEID.wbfs` — USBLoader GX standard
- SD card root must contain `/apps/` directory with one subdirectory per homebrew app
- FAT32 for both USB and SD card (no WBFS partition; FAT32 works with both WiiFlow and USBLoader GX)
- No outbound network calls except: GitHub releases (manifest-pinned URLs), WiiLink patcher (explicitly disclosed to user before run)
- No telemetry, no analytics, no crash reporting without explicit opt-in
- Must compile and run on macOS 12+, Windows 10+, Ubuntu 22.04+
- WiiUDownloader: NOT integrated in this plan — legal gray area; deferred to separate decision

## Review Focus

- **Drive mis-identification:** operator selects the wrong physical drive and formats their system disk. Confirm drive path, label, and size in a summary dialog before any destructive operation; require explicit "Yes, format this drive" confirmation with device path shown.
- **Partial install on failure:** a crash mid-copy leaves a half-written USB that boots to a broken Wii. All writes are staged to a temp directory first; atomic rename/move as final step, with rollback on error.
- **SHA-256 mismatch after download:** corrupted or MITM'd asset gets written to SD card. Verification happens before any file is moved to the drive; mismatch aborts with a clear error naming the file and the expected vs. actual hash.
- **Platform-specific formatting failures:** `diskpart` on Windows requires admin, `diskutil` on macOS requires sudo. Detect privilege level on startup and prompt for elevation if missing, rather than failing silently mid-install.
- **WiiLink patcher subprocess hanging:** patcher may prompt for input on unexpected paths. Run with a timeout (60s), capture stdout/stderr to the progress log, kill and report if it exceeds timeout.

---

### Task 1: Project Scaffold

**Files:**
- Create: `main.go`
- Create: `app.go`
- Create: `go.mod`
- Create: `frontend/src/App.svelte`
- Create: `frontend/src/stores/state.ts`
- Create: `wails.json`
- Create: `.github/workflows/build.yml`

**Interfaces:**
- Produces: `App` struct in `app.go` with Wails `startup(ctx)` method; Wails runtime bound via `wails.Run()`

- [ ] **Step 1: Install Wails CLI and scaffold**
```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
wails init -n wiikit -t svelte-ts
```
Expected: `wiikit/` directory with Go + Svelte structure

- [ ] **Step 2: Verify dev server runs on all three platforms**
```bash
wails dev
```
Expected: browser window opens with default Svelte template

- [ ] **Step 3: Add GitHub Actions matrix build**

`.github/workflows/build.yml` — matrix: `[ubuntu-latest, macos-latest, windows-latest]`; each job runs `wails build`; artifacts uploaded per platform.

- [ ] **Step 4: Confirm CI passes on all three**

Expected: three build artifacts in GitHub Actions run

- [ ] **Step 5: Commit**
```bash
git add .
git commit -m "feat: scaffold Wails + Svelte project with 3-platform CI"
```

---

### Task 2: Drive Detection

**Files:**
- Create: `internal/drive/detector.go`
- Test: `internal/drive/detector_test.go`

**Interfaces:**
- Produces: `ListDrives() ([]DriveInfo, error)` where `DriveInfo` is `{Path string, Label string, SizeBytes int64, IsRemovable bool}`
- Consumed by: Task 6 (DrivePanel UI), Task 4 (Formatter)

- [ ] **Step 1: Write failing test**
```go
func TestListDrivesReturnsAtLeastOne(t *testing.T) {
    drives, err := ListDrives()
    require.NoError(t, err)
    assert.NotEmpty(t, drives)
}
```

- [ ] **Step 2: Run test to verify it fails**
```bash
go test ./internal/drive/... -run TestListDrivesReturnsAtLeastOne -v
```
Expected: FAIL — `ListDrives` undefined

- [ ] **Step 3: Implement `ListDrives() ([]DriveInfo, error)` in `internal/drive/detector.go`**

Use `runtime.GOOS` to branch:
- macOS: parse `diskutil list -plist` XML output
- Windows: parse `wmic logicaldisk get DeviceID,Size,VolumeName /format:csv`
- Linux: parse `/proc/partitions` + `lsblk -J` JSON output

Filter to removable drives only where the OS reports that flag; include size and label.

- [ ] **Step 4: Run test to verify it passes**
```bash
go test ./internal/drive/... -run TestListDrivesReturnsAtLeastOne -v
```
Expected: PASS

- [ ] **Step 5: Add Review Focus test — drive path in result**
```go
func TestDriveInfoIncludesPath(t *testing.T) {
    drives, _ := ListDrives()
    for _, d := range drives {
        assert.NotEmpty(t, d.Path, "drive path must not be empty")
        assert.Positive(t, d.SizeBytes)
    }
}
```

- [ ] **Step 6: Commit**
```bash
git commit -m "feat: cross-platform drive detection"
```

---

### Task 3: SHA-256 Verification Helper

**Files:**
- Create: `internal/audit/verify.go`
- Test: `internal/audit/verify_test.go`

**Interfaces:**
- Produces: `VerifyFile(path string, expectedSHA256 string) error`
- Consumed by: Task 5 (Downloader)

- [ ] **Step 1: Write failing test**
```go
func TestVerifyFileMatchingHash(t *testing.T) {
    f := writeTempFile(t, []byte("hello"))
    hash := sha256hex([]byte("hello")) // "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
    assert.NoError(t, VerifyFile(f, hash))
}

func TestVerifyFileMismatchReturnsError(t *testing.T) {
    f := writeTempFile(t, []byte("hello"))
    err := VerifyFile(f, "deadbeef")
    assert.ErrorContains(t, err, "sha256 mismatch")
}
```

- [ ] **Step 2: Run test to verify it fails**
```bash
go test ./internal/audit/... -v
```

- [ ] **Step 3: Implement `VerifyFile` in `internal/audit/verify.go`**

Open file, stream through `crypto/sha256`, hex-encode result, compare to `expectedSHA256` (case-insensitive); return `fmt.Errorf("sha256 mismatch for %s: got %s, want %s", path, got, want)` on failure.

- [ ] **Step 4: Run tests to verify they pass**
```bash
go test ./internal/audit/... -v
```

- [ ] **Step 5: Commit**
```bash
git commit -m "feat: sha256 file verification"
```

---

### Task 4: Homebrew Manifest

**Files:**
- Create: `internal/homebrew/manifest.go`
- Create: `manifest.json`
- Test: `internal/homebrew/manifest_test.go`

**Interfaces:**
- Produces: `LoadManifest(path string) (*Manifest, error)`; `Manifest` struct contains `[]Asset{Name, URL, SHA256, DestPath string}`
- Consumed by: Task 5 (Downloader), Task 7 (Stack Installer)

- [ ] **Step 1: Write manifest.json with pinned entries**

Include: WiiFlow Lite, USBLoader GX, WiiLink Patcher, Nintendont, Not64, WiiStation, SNES9x GX, FCE Ultra GX, Visual Boy Advance GX. Each entry:
```json
{
  "name": "wiiflow",
  "version": "5.5.4-1",
  "url": "https://github.com/Fledge68/WiiFlow_Lite/releases/download/v5.5.4-1/WiiFlow_Lite-5.5.4-1.zip",
  "sha256": "<actual hash>",
  "dest": "sd://apps/wiiflow/"
}
```

Populate actual SHA-256 values by downloading each release and running `sha256sum`.

- [ ] **Step 2: Write failing test**
```go
func TestLoadManifestParsesAssets(t *testing.T) {
    m, err := LoadManifest("../../manifest.json")
    require.NoError(t, err)
    assert.Greater(t, len(m.Assets), 5)
    for _, a := range m.Assets {
        assert.NotEmpty(t, a.Name)
        assert.NotEmpty(t, a.SHA256)
        assert.Regexp(t, `^[0-9a-f]{64}$`, a.SHA256)
    }
}
```

- [ ] **Step 3: Run test to verify it fails**
```bash
go test ./internal/homebrew/... -run TestLoadManifestParsesAssets -v
```

- [ ] **Step 4: Implement `LoadManifest` in `internal/homebrew/manifest.go`**

Read JSON file, unmarshal into `Manifest` struct; validate each Asset has non-empty Name, URL, SHA256 (64-char hex), DestPath; return error if any field is empty.

- [ ] **Step 5: Run test to verify it passes**

- [ ] **Step 6: Commit**
```bash
git commit -m "feat: homebrew manifest loader with validation"
```

---

### Task 5: Homebrew Downloader

**Files:**
- Create: `internal/homebrew/downloader.go`
- Test: `internal/homebrew/downloader_test.go`

**Interfaces:**
- Consumes: `VerifyFile(path, sha256) error` from `internal/audit/verify.go`; `Asset` from Task 4
- Produces: `DownloadAsset(ctx context.Context, asset Asset, destDir string) (localPath string, error)`

- [ ] **Step 1: Write failing test (uses httptest)**
```go
func TestDownloadAssetVerifiesHash(t *testing.T) {
    content := []byte("fake-zip-content")
    hash := sha256hex(content)
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Write(content)
    }))
    defer srv.Close()
    asset := Asset{Name: "test", URL: srv.URL, SHA256: hash}
    path, err := DownloadAsset(context.Background(), asset, t.TempDir())
    require.NoError(t, err)
    assert.FileExists(t, path)
}

func TestDownloadAssetRejectsWrongHash(t *testing.T) {
    // same setup, wrong hash
    _, err := DownloadAsset(...)
    assert.ErrorContains(t, err, "sha256 mismatch")
}
```

- [ ] **Step 2: Run tests to verify they fail**

- [ ] **Step 3: Implement `DownloadAsset` in `internal/homebrew/downloader.go`**

Use `net/http` with the provided context; stream response body to a temp file; call `VerifyFile`; on success return temp file path; on mismatch delete the temp file and return the error. Set a 5-minute timeout via context if none set.

- [ ] **Step 4: Run tests to verify they pass**

- [ ] **Step 5: Add Review Focus test — SHA-256 mismatch aborts before any file lands in destDir**
```go
func TestDownloadAssetMismatchLeavesNoFile(t *testing.T) {
    dir := t.TempDir()
    // serve valid content but pass wrong hash
    _, err := DownloadAsset(ctx, Asset{..., SHA256: "badhash"}, dir)
    assert.Error(t, err)
    entries, _ := os.ReadDir(dir)
    assert.Empty(t, entries, "no files should remain on hash failure")
}
```

- [ ] **Step 6: Commit**
```bash
git commit -m "feat: asset downloader with sha256 verification"
```

---

### Task 6: Drive Formatter

**Files:**
- Create: `internal/drive/formatter.go`
- Test: `internal/drive/formatter_test.go`

**Interfaces:**
- Consumes: `DriveInfo` from Task 2
- Produces: `FormatFAT32(ctx context.Context, drive DriveInfo, label string) error`

- [ ] **Step 1: Write failing test (uses a temp loop device on Linux in CI, skip on Windows/macOS)**
```go
func TestFormatFAT32RejectsNonRemovable(t *testing.T) {
    err := FormatFAT32(context.Background(), DriveInfo{IsRemovable: false}, "WIISD")
    assert.ErrorContains(t, err, "not removable")
}
```

- [ ] **Step 2: Run test to verify it fails**

- [ ] **Step 3: Implement `FormatFAT32` in `internal/drive/formatter.go`**

Guard: return error if `drive.IsRemovable == false`. Then branch on `runtime.GOOS`:
- macOS: `diskutil eraseDisk FAT32 <label> <path>`
- Windows: build a `diskpart` script string, run via `cmd /C diskpart /s <scriptfile>`
- Linux: `mkfs.fat -F 32 -n <label> <path>`

Capture stderr; if exit code != 0 return an error wrapping stderr output.

- [ ] **Step 4: Run test to verify it passes**

- [ ] **Step 5: Add Review Focus test — confirmation gate**
```go
func TestFormatFAT32SummaryContainsPath(t *testing.T) {
    // FormatFAT32 is called only after UI confirms; this test verifies
    // the summary string returned by a companion PreviewFormat() func
    summary := PreviewFormat(DriveInfo{Path: "/dev/disk4", Label: "MY_USB", SizeBytes: 32 * 1e9})
    assert.Contains(t, summary, "/dev/disk4")
    assert.Contains(t, summary, "32")
}
```

- [ ] **Step 6: Commit**
```bash
git commit -m "feat: cross-platform FAT32 formatter"
```

---

### Task 7: Library Scanner

**Files:**
- Create: `internal/library/scanner.go`
- Create: `internal/library/metadata.go`
- Test: `internal/library/scanner_test.go`

**Interfaces:**
- Produces: `ScanLibrary(roots []string) ([]Game, error)`; `Game` struct: `{Path, Title, GameID, Region string, SizeBytes int64}`

- [ ] **Step 1: Write failing test**
```go
func TestScanLibraryFindsISO(t *testing.T) {
    dir := t.TempDir()
    // create a fake ISO with Wii header magic at offset 0x18 (0x5D1C9EA3)
    createFakeWiiISO(t, filepath.Join(dir, "game.iso"), "RMCE01", "Mario Kart Wii")
    games, err := ScanLibrary([]string{dir})
    require.NoError(t, err)
    require.Len(t, games, 1)
    assert.Equal(t, "RMCE01", games[0].GameID)
}
```

- [ ] **Step 2: Run test to verify it fails**

- [ ] **Step 3: Implement `ScanLibrary` and `ParseISOHeader` in scanner.go / metadata.go**

Walk each root with `filepath.WalkDir`; collect files with `.iso` or `.wbfs` extension; for each, read first 96 bytes and parse: GameID at offset 0x00 (6 bytes), title at offset 0x20 (64 bytes, null-terminated), verify Wii magic `0x5D1C9EA3` at offset 0x18. Skip files that fail magic check (not a Wii disc).

- [ ] **Step 4: Run test to verify it passes**

- [ ] **Step 5: Commit**
```bash
git commit -m "feat: Wii ISO library scanner with header parsing"
```

---

### Task 8: Game Writer (USB)

**Files:**
- Create: `internal/drive/writer.go`
- Test: `internal/drive/writer_test.go`

**Interfaces:**
- Consumes: `Game` from Task 7; `DriveInfo` from Task 2
- Produces: `WriteGamesToUSB(ctx context.Context, games []Game, usbRoot string, progress chan<- Progress) error`; `Progress` struct: `{GameID, Title string, BytesDone, BytesTotal int64}`

- [ ] **Step 1: Write failing test**
```go
func TestWriteGamesCreatesCorrectStructure(t *testing.T) {
    usbRoot := t.TempDir()
    games := []Game{{Path: "testdata/fake.iso", GameID: "RMCE01", Title: "Mario Kart Wii"}}
    err := WriteGamesToUSB(context.Background(), games, usbRoot, nil)
    require.NoError(t, err)
    expected := filepath.Join(usbRoot, "wbfs", "Mario Kart Wii [RMCE01]", "RMCE01.wbfs")
    assert.FileExists(t, expected)
}
```

- [ ] **Step 2: Run test to verify it fails**

- [ ] **Step 3: Implement `WriteGamesToUSB` in `internal/drive/writer.go`**

For each game: create directory `<usbRoot>/wbfs/<Title> [<GameID>]/`; copy source file to `<GameID>.wbfs` (rename extension regardless of source); send `Progress` updates every 1 MB if channel non-nil; use goroutines limited to 2 concurrent copies (avoid overwhelming USB bus). Write to a `.tmp` file first; rename on completion; remove `.tmp` on any error.

- [ ] **Step 4: Run test to verify it passes**

- [ ] **Step 5: Add Review Focus test — partial write cleanup**
```go
func TestWriteGamesCleansTmpOnContextCancel(t *testing.T) {
    ctx, cancel := context.WithCancel(context.Background())
    usbRoot := t.TempDir()
    // large fake game so copy is in-progress when we cancel
    go func() { time.Sleep(10*time.Millisecond); cancel() }()
    WriteGamesToUSB(ctx, []Game{{...}}, usbRoot, nil)
    // no .tmp files should remain
    var tmps []string
    filepath.WalkDir(usbRoot, func(p string, _ fs.DirEntry, _ error) error {
        if strings.HasSuffix(p, ".tmp") { tmps = append(tmps, p) }
        return nil
    })
    assert.Empty(t, tmps)
}
```

- [ ] **Step 6: Commit**
```bash
git commit -m "feat: game writer — USB WBFS directory structure"
```

---

### Task 9: Homebrew Stack Layout (SD Card)

**Files:**
- Create: `internal/homebrew/layout.go`
- Test: `internal/homebrew/layout_test.go`

**Interfaces:**
- Consumes: `Manifest` from Task 4; downloaded asset paths from Task 5
- Produces: `InstallStack(ctx context.Context, manifest *Manifest, sdRoot string, progress chan<- Progress) error`

- [ ] **Step 1: Write failing test**
```go
func TestInstallStackCreatesAppsDir(t *testing.T) {
    sdRoot := t.TempDir()
    m := &Manifest{Assets: []Asset{{Name: "wiiflow", DestPath: "sd://apps/wiiflow/boot.dol"}}}
    // provide a fake already-downloaded zip
    err := InstallStack(context.Background(), m, sdRoot, nil)
    require.NoError(t, err)
    assert.DirExists(t, filepath.Join(sdRoot, "apps", "wiiflow"))
}
```

- [ ] **Step 2: Run test to verify it fails**

- [ ] **Step 3: Implement `InstallStack` in `internal/homebrew/layout.go`**

For each asset in manifest: resolve `DestPath` (replace `sd://` prefix with `sdRoot`); if asset is a `.zip`, extract to resolved path using `archive/zip`; if `.dol` or other single file, copy directly. Create all parent directories. Report progress per asset.

- [ ] **Step 4: Run test to verify it passes**

- [ ] **Step 5: Commit**
```bash
git commit -m "feat: homebrew SD card layout installer"
```

---

### Task 10: WiiLink Patcher Integration

**Files:**
- Create: `internal/wiilink/patcher.go`
- Test: `internal/wiilink/patcher_test.go`

**Interfaces:**
- Produces: `RunPatcher(ctx context.Context, sdRoot string, logWriter io.Writer) error`

- [ ] **Step 1: Write failing test**
```go
func TestRunPatcherTimesOutOnHang(t *testing.T) {
    // patcher that hangs (sleep 999)
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    err := runSubprocess(ctx, "sleep", []string{"999"}, io.Discard)
    assert.ErrorIs(t, err, context.DeadlineExceeded)
}
```

- [ ] **Step 2: Run test to verify it fails**

- [ ] **Step 3: Implement `RunPatcher` in `internal/wiilink/patcher.go`**

Download WiiLink patcher binary from manifest URL; verify SHA-256; on macOS/Linux `chmod +x`; run as subprocess with the SD root path as argument; pipe stdout/stderr to `logWriter`; wrap in a 60-second context timeout; on timeout kill the process and return `context.DeadlineExceeded`. Disclose to user via UI before running (see Task 11).

- [ ] **Step 4: Run test to verify it passes**

- [ ] **Step 5: Commit**
```bash
git commit -m "feat: WiiLink patcher subprocess with timeout"
```

---

### Task 11: Wails Backend Bindings

**Files:**
- Modify: `app.go`

**Interfaces:**
- Consumes: all `internal/` packages
- Produces bound methods (callable from Svelte): `ListDrives`, `ScanLibrary`, `PreviewFormat`, `StartInstall`, `GetProgress`

- [ ] **Step 1: Wire all internal packages into `App` struct methods**

Each method: validate inputs, delegate to internal package, return a typed struct (not `interface{}`). `StartInstall` runs the full pipeline in a goroutine; progress events emitted via `runtime.EventsEmit(ctx, "progress", p)`.

- [ ] **Step 2: Verify Wails auto-generates TypeScript bindings**
```bash
wails dev
# check frontend/wailsjs/go/ for generated TS
```

- [ ] **Step 3: Commit**
```bash
git commit -m "feat: Wails backend bindings for all pipeline methods"
```

---

### Task 12: Svelte UI — 3-Step Wizard

**Files:**
- Modify: `frontend/src/App.svelte`
- Create: `frontend/src/components/LibraryPanel.svelte`
- Create: `frontend/src/components/DrivePanel.svelte`
- Create: `frontend/src/components/InstallPanel.svelte`
- Create: `frontend/src/components/ProgressLog.svelte`
- Modify: `frontend/src/stores/state.ts`

**Interfaces:**
- Consumes: Wails-generated bindings from Task 11

- [ ] **Step 1: Implement `state.ts` store**

Svelte writable stores: `library: Game[]`, `usbDrive: DriveInfo | null`, `sdDrive: DriveInfo | null`, `step: 1 | 2 | 3`, `installing: boolean`, `log: string[]`.

- [ ] **Step 2: LibraryPanel — scan button, game list with checkboxes**

Calls `ScanLibrary([selectedPath])`; lists games with title, game ID, size; "Select All" toggle. File picker via Wails `runtime.OpenDirectoryDialog`.

- [ ] **Step 3: DrivePanel — drive selectors + format warning**

Two dropdowns (USB, SD) populated from `ListDrives()`; show drive path + size; display `PreviewFormat()` summary in a yellow warning box; require checkbox "I understand this will erase the selected drives" before Next is enabled.

- [ ] **Step 4: InstallPanel — start button + progress log**

"Start Install" button (disabled during install); real-time log via `runtime.EventsOn("progress", ...)`. Show WiiLink disclosure: "WiiLink patcher will connect to wiilink24.com to restore online services." Confirm checkbox required.

- [ ] **Step 5: Smoke-test the full UI flow manually on each platform**

Navigate all three steps; verify no console errors; confirm drive list populates; confirm format warning shows correct drive path.

- [ ] **Step 6: Commit**
```bash
git commit -m "feat: 3-step wizard UI — library, drives, install"
```

---

### Task 13: Release Build + Code Signing Prep

**Files:**
- Modify: `.github/workflows/build.yml`
- Create: `build/appicon.png` (1024×1024)

- [ ] **Step 1: Add `wails build` matrix to CI for all three platforms**

macOS: `wails build -platform darwin/amd64,darwin/arm64`
Windows: `wails build -platform windows/amd64`
Linux: `wails build -platform linux/amd64`

- [ ] **Step 2: Upload artifacts**

Each platform uploads its binary/app bundle as a GitHub Actions artifact.

- [ ] **Step 3: Document code signing steps (not automated)**

README note: macOS requires notarization; Windows Defender SmartScreen requires EV cert or sufficient download count. Add `docs/signing.md` with steps.

- [ ] **Step 4: Commit**
```bash
git commit -m "chore: release build matrix + signing docs"
```

---

## Deferred Plans

- **Plan 2: Deal Finder** — eBay Browse API + Facebook Marketplace scraper for sourcing cheap Wiis. Independent module, integrates via a tab in the same Wails UI.
- **Plan 3: Azurabox Integration** — pending receipt of Azurabox project files.
- **Plan 4: WiiUDownloader** — pending legal review decision; would add an optional "Download Missing Games" flow.
