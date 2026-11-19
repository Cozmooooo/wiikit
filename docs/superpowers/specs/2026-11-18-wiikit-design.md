# WiiKit — Design Spec

## Intent & Why

WiiKit is a cross-platform desktop app (macOS / Windows / Linux) that turns a blank USB drive and SD card into a fully modded, online-capable Wii in 2–3 clicks. The operator plugs in their drives, selects games from their library, and WiiKit handles everything: formatting, game copying, full homebrew stack layout, and WiiLink online patching. The goal is to run this pipeline repeatedly on cheap sourced Wiis and flip them on eBay / Facebook Marketplace.

## Goals

- Format USB and SD card correctly for Wii (FAT32 partitions, correct block sizes)
- Copy game library (ISO / WBFS) to USB in USBLoader GX–compatible structure
- Install full homebrew stack to SD card: Homebrew Channel bootstrap files, WiiFlow Lite, USBLoader GX, WiiLink patcher output, Nintendont, Not64, WiiStation, SNES9x GX, custom themes
- Run WiiLink patcher non-interactively against SD
- Verify every downloaded binary by SHA-256 against a pinned manifest before writing to drives
- Cross-platform: macOS, Windows, Linux — single codebase, three release binaries
- 2–3 click UX for the operator

## Non-Goals

- Not a general-purpose Wii emulator or ROM manager
- Not responsible for sourcing or legally distributing game ISOs (operator provides library)
- Not a Wii softmod tool — assumes Homebrew Channel is already installed on the Wii (or operator uses a separate one-time Letterbomb / str2hax step; this is out of scope for v1)
- Azurabox deal-finder integration deferred to Plan 2

## Anti-Goals

- Must NOT silently download unsigned or unverified binaries
- Must NOT phone home or send any telemetry
- Must NOT hardcode credentials or API keys in the binary
- Must NOT break if WiiLink or any upstream URL changes — fail loudly with the manifest URL

## Constraints

- Language: Go 1.22+
- UI: Wails v2 with Svelte frontend
- Drive formatting: use OS-native tools via exec (diskutil on macOS, diskpart/format on Windows, parted/mkfs on Linux) — no CGO dependency for this
- Homebrew manifest: versioned JSON file, SHA-256 for every asset, pinned by default, user can point to custom manifest
- All third-party binaries fetched at runtime, never bundled (legal + size reasons)
- Minimum drive sizes: SD ≥ 2 GB, USB ≥ 8 GB recommended
- Target Wii format: FAT32 for both USB and SD (WBFS optional for USB, but FAT32 is safer with WiiFlow/USBLoader GX both working)
- Game folder structure on USB: `/wbfs/<Title> [GAMEID]/GAMEID.wbfs` (USBLoader GX standard)

## Approaches Considered

| Approach | Decision |
|---|---|
| **A: Tauri (Rust + web UI)** | Good performance, but Rust slower to iterate; eBay/HTTP work easier in Go |
| **B: Go + Wails (chosen)** | Go's concurrency ideal for parallel game copying; great HTTP libs for deal-finder; Wails gives polished cross-platform UI; single codebase |
| **C: Pure Rust + egui** | Fastest binary, but egui not polished enough for a product UX |

**Chosen: Go + Wails v2 + Svelte**

## Homebrew Stack (all verified against manifest)

| Tool | Purpose | Source |
|---|---|---|
| WiiFlow Lite | Game loader (primary) | GitHub: Fledge68/WiiFlow_Lite |
| USBLoader GX | Game loader (fallback) | GitHub: usbloadergx/usbloadergx |
| WiiLink patcher | Online services | WiiLink GitHub org |
| Nintendont | GameCube games | GitHub: FIX94/Nintendont |
| Not64 | N64 emulator | GitHub: Extrems/Not64 |
| WiiStation | PS1 emulator | GitHub: PS2ClassicsDev/WiiStation |
| SNES9x GX | SNES emulator | GitHub: dborth/snes9xgx |
| FCE Ultra GX | NES emulator | GitHub: dborth/fceugx |
| Visual Boy Advance GX | GBA/GB | GitHub: dborth/vbagx |
| Homebrew Browser (optional) | App discovery | (manifest only, not auto-installed) |

## Security Model

- Manifest file is SHA-256 signed; app verifies manifest integrity before any download
- Each asset's SHA-256 is checked post-download before it touches the drive
- WiiLink patcher is run as a subprocess — its network calls go to WiiLink's own servers (RC24); operator is informed of this in UI
- No binary is ever executed on the host machine — only written to drives
- Wii U Downloader / NUSDownloader: DO NOT integrate in v1 pending safety audit result

## UI Flow (3 steps)

```
Step 1: Library        Step 2: Drives          Step 3: Install
─────────────────      ─────────────────       ─────────────────
[Scan for games]       [Select USB drive]      [Start]
 ↓ game list           [Select SD card]         ↓ progress log
[+Add folder]          [Format options]        [Done — eject]
```

## File Structure

```
wiikit/
├── main.go                        # Wails entry point
├── app.go                         # Wails App struct, bound methods
├── internal/
│   ├── library/
│   │   ├── scanner.go             # Find ISOs/WBFS on local paths
│   │   └── metadata.go            # Parse game ID, title from ISO header
│   ├── drive/
│   │   ├── detector.go            # List connected drives (cross-platform)
│   │   ├── formatter.go           # Format USB/SD to FAT32
│   │   └── writer.go              # Copy games + homebrew to drives
│   ├── homebrew/
│   │   ├── manifest.go            # Load + verify manifest JSON
│   │   ├── downloader.go          # Fetch assets, verify SHA-256
│   │   └── layout.go              # SD card directory structure
│   ├── wiilink/
│   │   └── patcher.go             # Download + run WiiLink patcher subprocess
│   └── audit/
│       └── verify.go              # SHA-256 file verification helpers
├── manifest.json                  # Pinned homebrew versions + checksums
├── frontend/
│   ├── src/
│   │   ├── App.svelte
│   │   ├── stores/
│   │   │   └── state.ts           # Svelte store: library, drives, progress
│   │   └── components/
│   │       ├── LibraryPanel.svelte
│   │       ├── DrivePanel.svelte
│   │       ├── InstallPanel.svelte
│   │       └── ProgressLog.svelte
│   └── wailsjs/                   # Auto-generated Wails bindings
├── docs/
│   └── superpowers/
│       ├── specs/2026-11-18-wiikit-design.md
│       └── plans/2026-11-18-wiikit-core-pipeline.md
└── go.mod
```
