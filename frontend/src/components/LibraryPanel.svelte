<script lang="ts">
  import { library, step } from '../stores/state'
  import type { Game, GameType } from '../stores/state'

  const SD_CAPACITY  = 32  * 1024 * 1024 * 1024  // 32 GB
  const USB_CAPACITY = 64  * 1024 * 1024 * 1024  // 64 GB

  const TYPE_LABELS: Record<GameType, string> = {
    wii:  'Wii',
    gcn:  'GameCube',
    n64:  'Nintendo 64',
    ps1:  'PlayStation 1',
    snes: 'Super Nintendo',
    nes:  'NES',
    gba:  'Game Boy Advance',
  }

  const USB_TYPES  = new Set<GameType>(['wii', 'gcn'])
  const TYPE_ORDER: GameType[] = ['wii', 'gcn', 'n64', 'ps1', 'snes', 'nes', 'gba']

  let scanning = false
  let error    = ''

  async function pickFolder() {
    error = ''
    // @ts-ignore — Wails binding injected at runtime
    const folder = await window.go.main.App.OpenDirectoryDialog()
    if (!folder) return
    scanning = true
    try {
      // @ts-ignore
      const games: Game[] = await window.go.main.App.ScanLibrary([folder])
      library.set(games.map(g => ({ ...g, selected: true })))
    } catch (e: any) {
      error = String(e)
    } finally {
      scanning = false
    }
  }

  // Group games by type, in TYPE_ORDER order
  $: groups = TYPE_ORDER
    .map(type => ({ type, games: $library.filter(g => g.type === type) }))
    .filter(g => g.games.length > 0)

  function groupAllSelected(games: Game[]) {
    return games.every(g => g.selected)
  }

  function toggleGroup(type: GameType) {
    library.update(gs => {
      const allOn = gs.filter(g => g.type === type).every(g => g.selected)
      return gs.map(g => g.type === type ? { ...g, selected: !allOn } : g)
    })
  }

  function toggleGame(path: string) {
    library.update(gs => gs.map(g => g.path === path ? { ...g, selected: !g.selected } : g))
  }

  $: selected     = $library.filter(g => g.selected)
  $: sdSelected   = selected.filter(g => !USB_TYPES.has(g.type))
  $: usbSelected  = selected.filter(g =>  USB_TYPES.has(g.type))
  $: sdBytes      = sdSelected.reduce((s, g) => s + g.sizeBytes, 0)
  $: usbBytes     = usbSelected.reduce((s, g) => s + g.sizeBytes, 0)

  function fmtGB(bytes: number) {
    return (bytes / 1024 / 1024 / 1024).toFixed(1)
  }

  function pct(used: number, cap: number) {
    return Math.min(100, (used / cap) * 100)
  }

  function fmtSize(bytes: number): string {
    if (bytes >= 1024 * 1024 * 1024) return fmtGB(bytes) + ' GB'
    if (bytes >= 1024 * 1024)        return (bytes / 1024 / 1024).toFixed(0) + ' MB'
    return (bytes / 1024).toFixed(0) + ' KB'
  }

  function back() { step.set(0) }
  function next() { step.set(2) }
</script>

<div class="lib">
  <h2>Game Library</h2>
  <p class="sub">Select the folder where your ROMs and ISOs are stored.</p>

  <div class="actions">
    <button on:click={pickFolder} disabled={scanning}>
      {scanning ? 'Scanning…' : '+ Add Folder'}
    </button>
    {#if $library.length > 0}
      <span class="count">{$library.length} game{$library.length !== 1 ? 's' : ''} found</span>
      {#if selected.length !== $library.length}
        <span class="count sel-count">{selected.length} selected · {fmtGB(sdBytes + usbBytes)} GB</span>
      {:else}
        <span class="count">{fmtGB(sdBytes + usbBytes)} GB total</span>
      {/if}
    {/if}
  </div>

  {#if error}
    <div class="error">{error}</div>
  {/if}

  {#if $library.length === 0 && !scanning}
    <div class="empty">No games loaded. Add a folder to scan.</div>
  {:else}
    <!-- Storage meters -->
    {#if selected.length > 0}
      <div class="meters">
        <div class="meter-row">
          <span class="meter-label">SD Card (~32 GB)</span>
          <div class="bar-track">
            <div class="bar" class:over={sdBytes > SD_CAPACITY} style="width:{pct(sdBytes, SD_CAPACITY)}%"></div>
          </div>
          <span class="meter-num" class:over={sdBytes > SD_CAPACITY}>
            {fmtGB(sdBytes)} / 32 GB
          </span>
        </div>
        <div class="meter-row">
          <span class="meter-label">USB Drive (~64 GB)</span>
          <div class="bar-track">
            <div class="bar" class:over={usbBytes > USB_CAPACITY} style="width:{pct(usbBytes, USB_CAPACITY)}%"></div>
          </div>
          <span class="meter-num" class:over={usbBytes > USB_CAPACITY}>
            {fmtGB(usbBytes)} / 64 GB
          </span>
        </div>
      </div>
    {/if}

    <!-- Game list grouped by type -->
    <div class="game-list">
      {#each groups as { type, games }}
        <div class="group">
          <div class="group-header" on:click={() => toggleGroup(type)}>
            <input
              type="checkbox"
              checked={groupAllSelected(games)}
              on:click|stopPropagation={() => toggleGroup(type)}
              class="group-check"
            />
            <span class="group-label">{TYPE_LABELS[type]}</span>
            <span class="group-count">{games.filter(g => g.selected).length} / {games.length}</span>
          </div>
          {#each games as game (game.path)}
            <label class="game-row">
              <input type="checkbox" checked={game.selected} on:change={() => toggleGame(game.path)} />
              <div class="game-info">
                <span class="title">{game.title}</span>
                <span class="meta">
                  {#if game.gameId}{game.gameId} · {/if}{#if game.region}{game.region} · {/if}{fmtSize(game.sizeBytes)}
                </span>
              </div>
            </label>
          {/each}
        </div>
      {/each}
    </div>
  {/if}

  <div class="nav-footer">
    <button class="ghost" on:click={back}>← Back</button>
    <button class="primary" on:click={next} disabled={selected.length === 0}>
      Next: Select Drives →
    </button>
  </div>
</div>

<style>
  h2  { font-size: 20px; margin-bottom: 6px; }
  .sub { color: #888; margin-bottom: 20px; }

  .actions { display: flex; align-items: center; gap: 12px; margin-bottom: 16px; }
  .count     { color: #666; font-size: 13px; }
  .sel-count { color: #7c6fef; }
  .error   { color: #e06c75; font-size: 13px; margin-bottom: 12px; }
  .empty   { color: #555; padding: 40px 0; text-align: center; }

  button {
    padding: 8px 16px; border-radius: 6px; border: 1px solid #333;
    background: #1a1a28; color: #e8e8f0; cursor: pointer; font-size: 13px;
  }
  button:hover:not(:disabled) { background: #24243a; }
  button:disabled  { opacity: 0.4; cursor: not-allowed; }
  button.ghost     { background: transparent; border-color: #333; }
  button.primary   { background: #7c6fef; border-color: #7c6fef; font-weight: 600; }
  button.primary:hover:not(:disabled) { background: #6a5fdf; }

  /* Storage meters */
  .meters      { display: flex; flex-direction: column; gap: 8px; margin-bottom: 20px; }
  .meter-row   { display: flex; align-items: center; gap: 10px; }
  .meter-label { font-size: 12px; color: #888; width: 110px; flex-shrink: 0; }
  .bar-track   { flex: 1; height: 6px; background: #1e1e2e; border-radius: 3px; overflow: hidden; }
  .bar         { height: 100%; background: #7c6fef; border-radius: 3px; transition: width 0.2s; }
  .bar.over    { background: #e06c75; }
  .meter-num   { font-size: 12px; color: #666; width: 80px; text-align: right; flex-shrink: 0; }
  .meter-num.over { color: #e06c75; }

  /* Groups */
  .game-list   { display: flex; flex-direction: column; gap: 12px; margin-bottom: 80px; }
  .group       { background: #0d0d18; border: 1px solid #1e1e2e; border-radius: 8px; overflow: hidden; }

  .group-header {
    display: flex; align-items: center; gap: 10px;
    padding: 10px 14px; background: #12121f; cursor: pointer; user-select: none;
  }
  .group-header:hover { background: #16162a; }
  .group-check  { cursor: pointer; }
  .group-label  { font-weight: 600; font-size: 13px; flex: 1; }
  .group-count  { font-size: 12px; color: #555; }

  .game-row {
    display: flex; align-items: center; gap: 12px;
    padding: 8px 14px; cursor: pointer; border-top: 1px solid #1a1a2a;
  }
  .game-row:hover { background: #0f0f1c; }
  .game-info { display: flex; flex-direction: column; gap: 1px; }
  .title     { font-size: 13px; font-weight: 500; }
  .meta      { font-size: 11px; color: #555; }

  /* Nav */
  .nav-footer {
    position: fixed; bottom: 0; left: 0; right: 0;
    display: flex; justify-content: space-between; align-items: center;
    padding: 14px 24px; background: #0a0a10; border-top: 1px solid #1e1e2e;
  }
</style>
