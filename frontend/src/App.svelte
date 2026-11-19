<script lang="ts">
  import { step } from './stores/state'
  import LibraryPanel from './components/LibraryPanel.svelte'
  import DrivePanel from './components/DrivePanel.svelte'
  import InstallPanel from './components/InstallPanel.svelte'
</script>

<main>
  <header>
    <div class="logo">🎮 WiiKit</div>
    <div class="steps">
      <span class:active={$step >= 1} class:done={$step > 1}>1 · Library</span>
      <span class="sep">›</span>
      <span class:active={$step >= 2} class:done={$step > 2}>2 · Drives</span>
      <span class="sep">›</span>
      <span class:active={$step >= 3}>3 · Install</span>
    </div>
  </header>

  <div class="panel">
    {#if $step === 1}
      <LibraryPanel />
    {:else if $step === 2}
      <DrivePanel />
    {:else}
      <InstallPanel />
    {/if}
  </div>
</main>

<style>
  :global(*, *::before, *::after) { box-sizing: border-box; margin: 0; padding: 0; }
  :global(body) {
    background: #0d0d12;
    color: #e8e8f0;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
    font-size: 14px;
  }

  main { display: flex; flex-direction: column; height: 100vh; }

  header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 14px 24px;
    border-bottom: 1px solid #1e1e2e;
    background: #0a0a10;
  }

  .logo { font-size: 18px; font-weight: 700; letter-spacing: -0.3px; }

  .steps { display: flex; align-items: center; gap: 8px; color: #555; font-size: 13px; }
  .steps span.active { color: #7c6fef; font-weight: 600; }
  .steps span.done { color: #4caf50; }
  .sep { color: #333; }

  .panel { flex: 1; overflow-y: auto; padding: 32px 24px; }
</style>
