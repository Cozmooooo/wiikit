<script lang="ts">
  import { library, step } from '../stores/state'
  import type { Game } from '../stores/state'

  let scanning = false
  let selectedAll = false

  async function pickFolder() {
    // @ts-ignore — Wails binding injected at runtime
    const folder = await window.go.main.App.OpenDirectoryDialog()
    if (!folder) return
    scanning = true
    try {
      // @ts-ignore
      const games: Game[] = await window.go.main.App.ScanLibrary([folder])
      library.set(games.map(g => ({ ...g, selected: true })))
    } finally {
      scanning = false
    }
  }

  function toggleAll() {
    selectedAll = !selectedAll
    library.update(gs => gs.map(g => ({ ...g, selected: selectedAll })))
  }

  $: selectedCount = $library.filter(g => g.selected).length

  function next() { step.set(2) }

  function fmtSize(bytes: number) {
    return (bytes / 1024 / 1024 / 1024).toFixed(1) + ' GB'
  }
</script>

<div class="lib">
  <h2>Game Library</h2>
  <p class="sub">Select the folder where your Wii ISOs are stored.</p>

  <div class="actions">
    <button on:click={pickFolder} disabled={scanning}>
      {scanning ? 'Scanning…' : '+ Add Folder'}
    </button>
    {#if $library.length > 0}
      <button class="ghost" on:click={toggleAll}>
        {selectedAll ? 'Deselect All' : 'Select All'}
      </button>
    {/if}
  </div>

  {#if $library.length === 0 && !scanning}
    <div class="empty">No games loaded. Add a folder to scan.</div>
  {:else}
    <div class="game-list">
      {#each $library as game (game.path)}
        <label class="game-row">
          <input type="checkbox" bind:checked={game.selected} />
          <div class="game-info">
            <span class="title">{game.title}</span>
            <span class="meta">{game.gameId} · {game.region} · {fmtSize(game.sizeBytes)}</span>
          </div>
        </label>
      {/each}
    </div>
  {/if}

  {#if selectedCount > 0}
    <div class="footer">
      <span>{selectedCount} game{selectedCount !== 1 ? 's' : ''} selected</span>
      <button class="primary" on:click={next}>Next: Select Drives →</button>
    </div>
  {/if}
</div>

<style>
  h2 { font-size: 20px; margin-bottom: 6px; }
  .sub { color: #888; margin-bottom: 24px; }
  .actions { display: flex; gap: 10px; margin-bottom: 20px; }

  button {
    padding: 8px 16px; border-radius: 6px; border: 1px solid #333;
    background: #1a1a28; color: #e8e8f0; cursor: pointer; font-size: 13px;
  }
  button:hover:not(:disabled) { background: #24243a; }
  button:disabled { opacity: 0.4; cursor: not-allowed; }
  button.ghost { background: transparent; border-color: #333; }
  button.primary { background: #7c6fef; border-color: #7c6fef; font-weight: 600; }
  button.primary:hover { background: #6a5fdf; }

  .empty { color: #555; padding: 40px 0; text-align: center; }

  .game-list { display: flex; flex-direction: column; gap: 2px; margin-bottom: 80px; }
  .game-row {
    display: flex; align-items: center; gap: 12px;
    padding: 10px 12px; border-radius: 6px; cursor: pointer;
  }
  .game-row:hover { background: #14141e; }
  .game-info { display: flex; flex-direction: column; gap: 2px; }
  .title { font-weight: 500; }
  .meta { font-size: 12px; color: #666; }

  .footer {
    position: fixed; bottom: 0; left: 0; right: 0;
    display: flex; justify-content: space-between; align-items: center;
    padding: 16px 24px; background: #0a0a10; border-top: 1px solid #1e1e2e;
  }
</style>
