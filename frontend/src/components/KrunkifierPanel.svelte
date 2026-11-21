<script lang="ts">
  import { onMount, onDestroy } from 'svelte'
  import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'
  import { StartInstall, ListDrives } from '../../wailsjs/go/main/App'
  import {
    krunkPack, krunkUsbDrive, krunkSdDrive, krunkDetectedDrives,
    consoleType, library, ethernetEnabled,
    type PackTier, type DriveInfo,
  } from '../stores/state'
  import { get } from 'svelte/store'

  const PACKS: Record<PackTier, { label: string; desc: string; wiiLimit: number; perSystem: number }> = {
    essentials: { label: 'Essentials', desc: 'Top 20 Wii + best-of each system  (~32 GB)', wiiLimit: 20, perSystem: 25 },
    pro:        { label: 'Pro',        desc: 'Top 50 Wii + full emulator libs (~128 GB)', wiiLimit: 50, perSystem: 100 },
    complete:   { label: 'Complete',   desc: 'Everything in your library (~256 GB+)',   wiiLimit: Infinity, perSystem: Infinity },
  }

  let installing = false
  let done = false
  let log: string[] = []
  let progress = 0
  let errorMsg = ''
  let startTime = 0
  let elapsed = ''
  let elapsedInterval: ReturnType<typeof setInterval>

  // Drive auto-assignment: largest removable → USB, next → SD
  function autoAssign(drives: DriveInfo[]) {
    const removable = [...drives].filter(d => d.isRemovable).sort((a, b) => b.sizeBytes - a.sizeBytes)
    if (removable.length >= 1 && !get(krunkUsbDrive)) krunkUsbDrive.set(removable[0])
    if (removable.length >= 2 && !get(krunkSdDrive))  krunkSdDrive.set(removable[1])
  }

  function handleInserted(drive: DriveInfo) {
    krunkDetectedDrives.update(ds => [...ds.filter(d => d.path !== drive.path), drive])
    // Auto-fill empty slots
    if (!get(krunkUsbDrive) && drive.isRemovable) { krunkUsbDrive.set(drive); return }
    if (!get(krunkSdDrive)  && drive.isRemovable) { krunkSdDrive.set(drive) }
  }

  function handleRemoved(drive: DriveInfo) {
    krunkDetectedDrives.update(ds => ds.filter(d => d.path !== drive.path))
    if (get(krunkUsbDrive)?.path === drive.path) krunkUsbDrive.set(null)
    if (get(krunkSdDrive)?.path  === drive.path) krunkSdDrive.set(null)
  }

  onMount(async () => {
    // Seed with currently connected drives
    const initial = await ListDrives()
    krunkDetectedDrives.set(initial)
    autoAssign(initial)

    EventsOn('drive-inserted', handleInserted)
    EventsOn('drive-removed',  handleRemoved)
    EventsOn('log',      (msg: string) => { log = [...log, msg] })
    EventsOn('progress', (p: { stage?: string }) => { if (p.stage) log = [...log, p.stage] })
    EventsOn('error',    (msg: string) => { errorMsg = msg; installing = false })
    EventsOn('done',     () => {
      installing = false
      done = true
      clearInterval(elapsedInterval)
    })
  })

  onDestroy(() => {
    EventsOff('drive-inserted', 'drive-removed', 'log', 'progress', 'error', 'done')
    clearInterval(elapsedInterval)
  })

  function fmt(bytes: number) {
    if (bytes >= 1e9) return (bytes / 1e9).toFixed(0) + ' GB'
    return (bytes / 1e6).toFixed(0) + ' MB'
  }

  function filterGames(pack: PackTier) {
    const cfg = PACKS[pack]
    const games = get(library)
    if (cfg.wiiLimit === Infinity) return games
    const byType: Record<string, typeof games> = {}
    for (const g of games) {
      if (!byType[g.type]) byType[g.type] = []
      byType[g.type].push(g)
    }
    const result: typeof games = []
    for (const [type, list] of Object.entries(byType)) {
      const limit = type === 'wii' ? cfg.wiiLimit : cfg.perSystem
      result.push(...list.slice(0, limit))
    }
    return result
  }

  async function krunkify() {
    const usb = get(krunkUsbDrive)
    const sd  = get(krunkSdDrive)
    if (!usb || !sd) { errorMsg = 'Plug in both a USB drive and an SD card first.'; return }

    installing = true
    done = false
    errorMsg = ''
    log = []
    progress = 0
    startTime = Date.now()
    elapsed = '0:00'
    elapsedInterval = setInterval(() => {
      const s = Math.floor((Date.now() - startTime) / 1000)
      elapsed = `${Math.floor(s / 60)}:${String(s % 60).padStart(2, '0')}`
    }, 1000)

    const games = filterGames(get(krunkPack))
    await StartInstall(games, usb, sd, get(consoleType), get(ethernetEnabled))
  }

  function reset() {
    done = false
    log = []
    errorMsg = ''
    progress = 0
    krunkUsbDrive.set(null)
    krunkSdDrive.set(null)
    // Re-auto-assign from detected drives
    autoAssign(get(krunkDetectedDrives))
  }
</script>

<div class="krunk">
  <!-- Header -->
  <div class="krunk-header">
    <span class="lightning">⚡</span>
    <span class="title">KRUNKIFIER</span>
    <span class="subtitle">Mass production mode</span>
  </div>

  {#if done}
    <!-- Done screen -->
    <div class="done-screen">
      <div class="done-icon">✓</div>
      <div class="done-label">KRUNK COMPLETE</div>
      <div class="done-time">{elapsed} prep time</div>
      <button class="btn-primary" on:click={reset}>Prep Another</button>
    </div>
  {:else if installing}
    <!-- Progress screen -->
    <div class="progress-screen">
      <div class="progress-label">Krunkifying... {elapsed}</div>
      <div class="progress-bar-track"><div class="progress-bar-fill" style="width:{progress}%"></div></div>
      <div class="log-scroll">
        {#each log.slice(-40) as line}
          <div class="log-line">{line}</div>
        {/each}
      </div>
    </div>
  {:else}
    <!-- Setup screen -->
    <div class="setup">

      <!-- Drive slots -->
      <section class="section">
        <div class="section-label">DRIVES</div>
        <div class="drive-row">
          <div class="drive-slot" class:filled={!!$krunkUsbDrive} class:empty={!$krunkUsbDrive}>
            <div class="drive-icon">📦</div>
            <div class="drive-info">
              <div class="drive-role">USB DRIVE</div>
              {#if $krunkUsbDrive}
                <div class="drive-path">{$krunkUsbDrive.path} · {fmt($krunkUsbDrive.sizeBytes)}</div>
                <div class="drive-label">{$krunkUsbDrive.label || 'unlabeled'}</div>
              {:else}
                <div class="drive-waiting">waiting for drive...</div>
              {/if}
            </div>
            {#if $krunkUsbDrive}
              <button class="clear-btn" on:click={() => krunkUsbDrive.set(null)}>✕</button>
            {/if}
          </div>

          <div class="drive-slot" class:filled={!!$krunkSdDrive} class:empty={!$krunkSdDrive}>
            <div class="drive-icon">💾</div>
            <div class="drive-info">
              <div class="drive-role">SD CARD</div>
              {#if $krunkSdDrive}
                <div class="drive-path">{$krunkSdDrive.path} · {fmt($krunkSdDrive.sizeBytes)}</div>
                <div class="drive-label">{$krunkSdDrive.label || 'unlabeled'}</div>
              {:else}
                <div class="drive-waiting">waiting for card...</div>
              {/if}
            </div>
            {#if $krunkSdDrive}
              <button class="clear-btn" on:click={() => krunkSdDrive.set(null)}>✕</button>
            {/if}
          </div>
        </div>

        <!-- Detected drives available to assign -->
        {#if $krunkDetectedDrives.filter(d => d.isRemovable && d.path !== $krunkUsbDrive?.path && d.path !== $krunkSdDrive?.path).length > 0}
          <div class="detected-label">detected — click to assign:</div>
          <div class="detected-row">
            {#each $krunkDetectedDrives.filter(d => d.isRemovable && d.path !== $krunkUsbDrive?.path && d.path !== $krunkSdDrive?.path) as d}
              <button class="detected-chip"
                on:click={() => { if (!$krunkUsbDrive) krunkUsbDrive.set(d); else if (!$krunkSdDrive) krunkSdDrive.set(d) }}>
                {d.path} {fmt(d.sizeBytes)}
              </button>
            {/each}
          </div>
        {/if}
      </section>

      <!-- Pack tier -->
      <section class="section">
        <div class="section-label">GAME PACK</div>
        <div class="pack-row">
          {#each Object.entries(PACKS) as [key, cfg]}
            <button
              class="pack-btn"
              class:active={$krunkPack === key}
              on:click={() => krunkPack.set(key as PackTier)}
            >
              <span class="pack-name">{cfg.label}</span>
            </button>
          {/each}
        </div>
        <div class="pack-desc">{PACKS[$krunkPack].desc}</div>
      </section>

      <!-- Console (mirrors wizard) -->
      <section class="section">
        <div class="section-label">CONSOLE</div>
        <div class="pack-row">
          {#each [['wii','Wii'], ['wiimini','Wii Mini'], ['wiiu','Wii U']] as [val, label]}
            <button
              class="pack-btn"
              class:active={$consoleType === val}
              on:click={() => consoleType.set(val as any)}
            >
              {label}
            </button>
          {/each}
        </div>
      </section>

      {#if errorMsg}
        <div class="error-msg">{errorMsg}</div>
      {/if}

      <!-- KRUNKIFY button -->
      <button
        class="krunkify-btn"
        class:ready={!!$krunkUsbDrive && !!$krunkSdDrive}
        disabled={!$krunkUsbDrive || !$krunkSdDrive}
        on:click={krunkify}
      >
        ⚡ KRUNKIFY
      </button>

      <div class="hint">
        {#if !$krunkUsbDrive || !$krunkSdDrive}
          Plug in USB + SD card to unlock
        {:else}
          Ready · {filterGames($krunkPack).length} games selected
        {/if}
      </div>
    </div>
  {/if}
</div>

<style>
  .krunk {
    display: flex;
    flex-direction: column;
    gap: 20px;
    max-width: 720px;
    margin: 0 auto;
    width: 100%;
  }

  .krunk-header {
    display: flex;
    align-items: center;
    gap: 10px;
  }
  .lightning { font-size: 28px; }
  .title { font-size: 22px; font-weight: 800; letter-spacing: 2px; color: #f97316; }
  .subtitle { font-size: 12px; color: #64748b; letter-spacing: 1px; text-transform: uppercase; }

  .section { display: flex; flex-direction: column; gap: 8px; }
  .section-label {
    font-size: 10px;
    font-weight: 700;
    letter-spacing: 2px;
    color: #475569;
    text-transform: uppercase;
  }

  /* Drive slots */
  .drive-row { display: flex; gap: 12px; }
  .drive-slot {
    flex: 1;
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 12px 14px;
    border-radius: 10px;
    border: 1px solid #1e1e2e;
    background: #0d0d18;
    min-height: 64px;
    transition: border-color 0.2s;
    position: relative;
  }
  .drive-slot.filled { border-color: #2a2a4a; }
  .drive-slot.empty  { border-color: #1e1e2e; border-style: dashed; opacity: 0.6; }
  .drive-icon { font-size: 22px; }
  .drive-info { flex: 1; min-width: 0; }
  .drive-role { font-size: 10px; font-weight: 700; letter-spacing: 1.5px; color: #64748b; text-transform: uppercase; }
  .drive-path { font-size: 13px; font-weight: 600; color: #e2e8f0; }
  .drive-label { font-size: 11px; color: #94a3b8; }
  .drive-waiting { font-size: 12px; color: #334155; font-style: italic; }
  .clear-btn {
    position: absolute; top: 6px; right: 8px;
    background: none; border: none; color: #475569;
    font-size: 11px; cursor: pointer; padding: 2px 4px;
  }
  .clear-btn:hover { color: #ef4444; }

  .detected-label { font-size: 11px; color: #475569; }
  .detected-row { display: flex; flex-wrap: wrap; gap: 6px; }
  .detected-chip {
    padding: 4px 10px;
    background: #13131c;
    border: 1px solid #252535;
    border-radius: 6px;
    font-size: 12px;
    color: #94a3b8;
    cursor: pointer;
    transition: background 0.15s;
  }
  .detected-chip:hover { background: #1a1a2e; color: #e2e8f0; }

  /* Pack / Console buttons */
  .pack-row { display: flex; gap: 8px; }
  .pack-btn {
    flex: 1;
    padding: 8px 12px;
    background: #0d0d18;
    border: 1px solid #1e1e2e;
    border-radius: 8px;
    color: #64748b;
    font-size: 13px;
    font-weight: 600;
    cursor: pointer;
    transition: background 0.15s, color 0.15s, border-color 0.15s;
    text-align: center;
  }
  .pack-btn.active {
    background: #1a1030;
    border-color: #7c3aed;
    color: #a78bfa;
  }
  .pack-btn:hover:not(.active) { color: #94a3b8; border-color: #252535; }
  .pack-name { display: block; }
  .pack-desc { font-size: 12px; color: #475569; }

  .error-msg {
    padding: 10px 14px;
    background: #200a0a;
    border: 1px solid #450a0a;
    border-radius: 8px;
    color: #fca5a5;
    font-size: 13px;
  }

  /* KRUNKIFY button */
  .krunkify-btn {
    padding: 16px;
    width: 100%;
    background: #1a0a00;
    border: 2px solid #431407;
    border-radius: 12px;
    color: #6b2d02;
    font-size: 18px;
    font-weight: 900;
    letter-spacing: 3px;
    cursor: not-allowed;
    transition: background 0.2s, border-color 0.2s, color 0.2s, transform 0.1s;
  }
  .krunkify-btn.ready {
    background: linear-gradient(135deg, #ea580c 0%, #f97316 100%);
    border-color: #f97316;
    color: #fff;
    cursor: pointer;
    box-shadow: 0 0 20px #f9731640;
  }
  .krunkify-btn.ready:hover {
    background: linear-gradient(135deg, #f97316 0%, #fb923c 100%);
    transform: translateY(-1px);
    box-shadow: 0 0 32px #f9731660;
  }
  .krunkify-btn.ready:active { transform: translateY(0); }

  .hint { text-align: center; font-size: 12px; color: #334155; }

  /* Progress / Done */
  .progress-screen, .done-screen {
    display: flex;
    flex-direction: column;
    gap: 14px;
  }
  .progress-label { font-size: 14px; font-weight: 600; color: #f97316; }
  .progress-bar-track {
    height: 6px;
    background: #1a1a2e;
    border-radius: 3px;
    overflow: hidden;
  }
  .progress-bar-fill {
    height: 100%;
    background: linear-gradient(90deg, #7c3aed, #f97316);
    border-radius: 3px;
    transition: width 0.3s;
  }
  .log-scroll {
    max-height: 320px;
    overflow-y: auto;
    background: #07070e;
    border: 1px solid #1e1e2e;
    border-radius: 8px;
    padding: 10px 12px;
    font-family: 'Consolas', 'Fira Mono', monospace;
    font-size: 12px;
  }
  .log-line { color: #64748b; line-height: 1.6; }
  .log-line:last-child { color: #94a3b8; }

  .done-screen { align-items: center; padding: 40px 0; }
  .done-icon {
    width: 72px; height: 72px;
    background: #0d2b1a;
    border: 2px solid #16a34a;
    border-radius: 50%;
    display: flex; align-items: center; justify-content: center;
    font-size: 32px;
    color: #4ade80;
  }
  .done-label { font-size: 24px; font-weight: 900; letter-spacing: 3px; color: #4ade80; }
  .done-time { font-size: 13px; color: #64748b; }
  .btn-primary {
    padding: 10px 32px;
    background: #1a0a00;
    border: 1px solid #f97316;
    border-radius: 8px;
    color: #f97316;
    font-size: 14px;
    font-weight: 600;
    cursor: pointer;
    transition: background 0.15s;
  }
  .btn-primary:hover { background: #2a1200; }
</style>
