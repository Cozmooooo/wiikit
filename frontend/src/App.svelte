<script lang="ts">
  import { step } from './stores/state'
  import ConsolePanel from './components/ConsolePanel.svelte'
  import LibraryPanel from './components/LibraryPanel.svelte'
  import DrivePanel from './components/DrivePanel.svelte'
  import InstallPanel from './components/InstallPanel.svelte'
  import DealsPanel from './components/DealsPanel.svelte'
  import SetupCard from './components/SetupCard.svelte'
  import KrunkifierPanel from './components/KrunkifierPanel.svelte'

  let activeTab: 'wizard' | 'deals' | 'card' | 'krunk' = 'wizard'
</script>

<main>
  <header>
    <div class="logo">🎮 WiiKit</div>
    <nav class="nav-tabs">
      <button
        class="nav-tab"
        class:active={activeTab === 'wizard'}
        on:click={() => (activeTab = 'wizard')}
      >
        Setup Wizard
      </button>
      <button
        class="nav-tab"
        class:active={activeTab === 'deals'}
        on:click={() => (activeTab = 'deals')}
      >
        🔍 Deal Finder
      </button>
      <button
        class="nav-tab"
        class:active={activeTab === 'card'}
        on:click={() => (activeTab = 'card')}
      >
        🃏 Setup Card
      </button>
      <button
        class="nav-tab krunk-tab"
        class:active={activeTab === 'krunk'}
        on:click={() => (activeTab = 'krunk')}
      >
        ⚡ Krunkifier
      </button>
    </nav>
    {#if activeTab === 'wizard' && $step >= 0}
      <div class="steps">
        <span class:active={$step >= 0} class:done={$step > 0}>0 · Console</span>
        <span class="sep">›</span>
        <span class:active={$step >= 1} class:done={$step > 1}>1 · Library</span>
        <span class="sep">›</span>
        <span class:active={$step >= 2} class:done={$step > 2}>2 · Drives</span>
        <span class="sep">›</span>
        <span class:active={$step >= 3}>3 · Install</span>
      </div>
    {:else}
      <div class="steps"></div>
    {/if}
  </header>

  <div class="panel">
    {#if activeTab === 'deals'}
      <DealsPanel />
    {:else if activeTab === 'card'}
      <SetupCard />
    {:else if activeTab === 'krunk'}
      <KrunkifierPanel />
    {:else if $step === 0}
      <ConsolePanel />
    {:else if $step === 1}
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
    padding: 10px 24px;
    border-bottom: 1px solid #1e1e2e;
    background: #0a0a10;
    gap: 20px;
  }

  .logo { font-size: 18px; font-weight: 700; letter-spacing: -0.3px; white-space: nowrap; }

  .nav-tabs {
    display: flex;
    gap: 4px;
    background: #13131c;
    border: 1px solid #252535;
    border-radius: 8px;
    padding: 3px;
  }

  .nav-tab {
    padding: 6px 16px;
    background: transparent;
    border: none;
    border-radius: 6px;
    color: #64748b;
    font-size: 13px;
    font-weight: 500;
    cursor: pointer;
    transition: background 0.15s, color 0.15s;
    white-space: nowrap;
  }

  .nav-tab.active {
    background: #1e1e30;
    color: #e2e8f0;
  }

  .nav-tab:hover:not(.active) {
    color: #94a3b8;
  }

  .krunk-tab.active {
    background: #1a0a00;
    color: #f97316;
    border: 1px solid #431407;
  }

  .steps { display: flex; align-items: center; gap: 8px; color: #555; font-size: 13px; min-width: 200px; justify-content: flex-end; }
  .steps span.active { color: #7c6fef; font-weight: 600; }
  .steps span.done { color: #4caf50; }
  .sep { color: #333; }

  .panel { flex: 1; overflow: hidden; padding: 24px; display: flex; flex-direction: column; }
</style>
