<script lang="ts">
  import { onDestroy } from 'svelte'
  import { library, usbDrive, sdDrive, installing, log, progress, step, wiilinkConfirmed, consoleType, ethernetEnabled } from '../stores/state'

  const CONSOLE_LABELS: Record<string, string> = {
    wii:     'Wii',
    wiiu:    'Wii U',
    wiimini: 'Wii Mini',
  }

  let done = false
  let errorMsg = ''
  let wiilinkOk = false
  let elapsedSecs = 0
  let elapsedTimer: ReturnType<typeof setInterval> | null = null
  let logEl: HTMLDivElement

  $: wiilinkConfirmed.set(wiilinkOk)

  $: if ($installing) {
    scrollLog()
  }

  function scrollLog() {
    if (logEl) setTimeout(() => { logEl.scrollTop = logEl.scrollHeight }, 0)
  }

  $: $log, scrollLog()

  async function startInstall() {
    const usb = $usbDrive
    const sd = $sdDrive
    if (!usb || !sd) return

    installing.set(true)
    log.set([])
    done = false
    errorMsg = ''
    elapsedSecs = 0
    elapsedTimer = setInterval(() => { elapsedSecs++ }, 1000)

    const selectedGames = $library.filter(g => g.selected)

    window.runtime?.EventsOn('log', (msg: unknown) => {
      log.update(l => [...l, String(msg)])
    })
    window.runtime?.EventsOn('progress', (p: unknown) => {
      progress.set(p as typeof $progress)
    })
    window.runtime?.EventsOn('error', (msg: unknown) => {
      errorMsg = String(msg)
      installing.set(false)
      stopTimer()
    })
    window.runtime?.EventsOn('done', () => {
      done = true
      installing.set(false)
      stopTimer()
    })

    try {
      // reuseSD=false, skipWiiLink=false — wizard always does a full fresh install
      await window.go.main.App.StartInstall(selectedGames, usb, sd, $consoleType, $ethernetEnabled, false, false)
    } catch (e: unknown) {
      errorMsg = e instanceof Error ? e.message : String(e)
      installing.set(false)
      stopTimer()
    }
  }

  function stopTimer() {
    if (elapsedTimer) { clearInterval(elapsedTimer); elapsedTimer = null }
  }

  function fmtElapsed(s: number): string {
    return Math.floor(s / 60) + ':' + String(s % 60).padStart(2, '0')
  }

  function back() { step.set(2) }

  function restart() {
    stopTimer()
    step.set(0)
    library.set([])
    usbDrive.set(null)
    sdDrive.set(null)
    log.set([])
    progress.set({})
    done = false
    errorMsg = ''
    elapsedSecs = 0
    wiilinkOk = false
    ethernetEnabled.set(false)
  }

  onDestroy(stopTimer)
</script>

<div class="install">
  <h2>Install</h2>

  {#if !$installing && !done && !errorMsg}
    <div class="summary">
      <div class="summary-row">
        <span>Console</span>
        <strong>{CONSOLE_LABELS[$consoleType] ?? $consoleType}</strong>
      </div>
      <div class="summary-row">
        <span>Games</span>
        <strong>{$library.filter(g => g.selected).length} selected</strong>
      </div>
      <div class="summary-row">
        <span>USB</span>
        <strong>{$usbDrive?.path} ({$usbDrive?.label || 'Unlabeled'})</strong>
      </div>
      <div class="summary-row">
        <span>SD Card</span>
        <strong>{$sdDrive?.path} ({$sdDrive?.label || 'Unlabeled'})</strong>
      </div>
    </div>

    {#if $consoleType === 'wii'}
      <div class="wiilink-disclosure">
        <strong>WiiLink Online Services</strong>
        <p>
          WiiKit will download and launch the <strong>WiiLink Patcher GUI</strong> (verified SHA-256).
          It's an interactive app — a window will open and you'll point it at your SD card path.
          The patcher connects to <code>wiilink24.com</code> to restore Nintendo's discontinued
          online channels (Wii Room, Food Channel, etc.).
          Your Wii will phone home to WiiLink servers after setup.
        </p>
        <label class="confirm-row">
          <input type="checkbox" bind:checked={wiilinkOk} />
          I'll complete the WiiLink Patcher window when it opens
        </label>
      </div>
    {:else if $consoleType === 'wiiu'}
      <div class="console-note wiiu">
        <strong>Wii U (vWii mode)</strong>
        <p>WiiLink is not supported on vWii. All other homebrew will install normally.</p>
        <p>Requires <strong>Tiramisu or Aroma CFW</strong> already installed on your Wii U before proceeding.</p>
      </div>
    {:else if $consoleType === 'wiimini' && $ethernetEnabled}
      <div class="console-note mini eth">
        <strong>Wii Mini + Ethernet</strong>
        <p>WiiKit will install <strong>Ethernet Enabler</strong> on your SD card and run the <strong>WiiLink Patcher</strong>.</p>
        <p>After setup: plug in your ASIX AX88179 USB adapter, then run Ethernet Enabler from the Homebrew Channel once to activate it.</p>
        <p>BlueBomb exploit required to install homebrew on a stock Wii Mini.</p>
        <label class="confirm-row">
          <input type="checkbox" bind:checked={wiilinkOk} />
          I'll complete the WiiLink Patcher window when it opens
        </label>
      </div>
    {:else if $consoleType === 'wiimini'}
      <div class="console-note mini">
        <strong>Wii Mini</strong>
        <p>No built-in WiFi — WiiLink is skipped. Games load from USB only.</p>
        <p>To install homebrew on a stock Wii Mini, use <strong>BlueBomb</strong> (requires a Bluetooth-capable Linux device nearby). Instructions will appear in the log after setup completes.</p>
      </div>
    {/if}
  {/if}

  {#if $installing}
    <div class="progress-bar-wrap">
      <div class="progress-header">
        {#if $progress.stage}
          <span class="progress-stage">{$progress.stage}</span>
        {:else}
          <span class="progress-stage">Working…</span>
        {/if}
        <span class="elapsed">{fmtElapsed(elapsedSecs)}</span>
      </div>
      {#if $progress.bytesTotal && $progress.bytesTotal > 0}
        <div class="progress-track">
          <div class="progress-fill" style="width: {Math.round(($progress.bytesDone ?? 0) / $progress.bytesTotal * 100)}%"></div>
        </div>
        <div class="progress-pct">{Math.round(($progress.bytesDone ?? 0) / $progress.bytesTotal * 100)}%</div>
      {/if}
    </div>
  {/if}

  {#if $log.length > 0}
    <div class="log-box" bind:this={logEl}>
      {#each $log as line}
        <div class="log-line">{line}</div>
      {/each}
    </div>
  {/if}

  {#if errorMsg}
    <div class="error-box">
      <strong>Error:</strong> {errorMsg}
    </div>
  {/if}

  {#if done}
    <div class="success-box">
      ✅ All done! Safely eject both drives and power on your Wii.
    </div>
  {/if}

  <div class="footer">
    {#if !$installing && !done}
      <button class="ghost" on:click={back}>← Back</button>
      <button
        class="primary big"
        disabled={(($consoleType === 'wii' || ($consoleType === 'wiimini' && $ethernetEnabled)) && !wiilinkOk) || $installing}
        on:click={startInstall}
      >
        {$installing ? 'Installing…' : '🚀 Start Install'}
      </button>
    {:else if done}
      <button class="primary" on:click={restart}>Prep Another Wii</button>
    {/if}
  </div>
</div>

<style>
  h2 { font-size: 20px; margin-bottom: 24px; }

  .summary {
    background: #12121a; border: 1px solid #1e1e2e; border-radius: 8px;
    padding: 16px; margin-bottom: 24px; display: flex; flex-direction: column; gap: 10px;
  }
  .summary-row { display: flex; justify-content: space-between; font-size: 13px; }
  .summary-row span { color: #888; }

  .console-note {
    border-radius: 8px; padding: 16px; margin-bottom: 80px;
  }
  .console-note.wiiu { background: #0d1a2e; border: 1px solid #1a3a5e; }
  .console-note.mini { background: #1a0d00; border: 1px solid #3a1a00; }
  .console-note.mini.eth { background: #0d1a0d; border: 1px solid #1a3a1a; }
  .console-note strong { display: block; margin-bottom: 8px; }
  .console-note.wiiu strong { color: #7cb8ff; }
  .console-note.mini strong { color: #f0a000; }
  .console-note.mini.eth strong { color: #7fff7f; }
  .console-note p { color: #aaa; font-size: 13px; line-height: 1.5; margin-bottom: 6px; }

  .progress-bar-wrap {
    background: #0a0a0f; border: 1px solid #1e1e2e; border-radius: 6px;
    padding: 12px 16px; margin-bottom: 16px;
  }
  .progress-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px; }
  .progress-stage { font-size: 12px; color: #9f9; font-family: monospace; }
  .elapsed { font-size: 12px; color: #666; font-family: monospace; }
  .progress-track {
    height: 6px; background: #1e1e2e; border-radius: 3px; overflow: hidden; margin-bottom: 4px;
  }
  .progress-fill { height: 100%; background: #7c6fef; border-radius: 3px; transition: width 0.2s; }
  .progress-pct { font-size: 11px; color: #666; text-align: right; }

  .wiilink-disclosure {
    background: #0d1a2e; border: 1px solid #1a3a5e; border-radius: 8px;
    padding: 16px; margin-bottom: 80px;
  }
  .wiilink-disclosure strong { display: block; margin-bottom: 8px; color: #7cb8ff; }
  .wiilink-disclosure p { color: #aaa; font-size: 13px; line-height: 1.5; margin-bottom: 12px; }
  code { background: #1a2a40; padding: 2px 6px; border-radius: 3px; font-size: 12px; }

  .confirm-row { display: flex; align-items: center; gap: 10px; font-size: 13px; cursor: pointer; }

  .log-box {
    background: #0a0a0f; border: 1px solid #1e1e2e; border-radius: 6px;
    padding: 12px 16px; font-family: monospace; font-size: 12px; color: #9f9;
    max-height: 300px; overflow-y: auto; margin-bottom: 16px;
  }
  .log-line { padding: 2px 0; }

  .error-box {
    background: #1a0505; border: 1px solid #5a0000; border-radius: 6px;
    padding: 12px 16px; color: #ff7777; margin-bottom: 16px;
  }

  .success-box {
    background: #0a1a0a; border: 1px solid #1a4a1a; border-radius: 6px;
    padding: 16px; color: #7fff7f; font-size: 15px; margin-bottom: 16px;
  }

  button {
    padding: 8px 16px; border-radius: 6px; border: 1px solid #333;
    background: #1a1a28; color: #e8e8f0; cursor: pointer; font-size: 13px;
  }
  button:hover:not(:disabled) { background: #24243a; }
  button:disabled { opacity: 0.4; cursor: not-allowed; }
  button.ghost { background: transparent; }
  button.primary { background: #7c6fef; border-color: #7c6fef; font-weight: 600; }
  button.big { padding: 12px 32px; font-size: 15px; }

  .footer {
    position: fixed; bottom: 0; left: 0; right: 0;
    display: flex; justify-content: space-between; align-items: center;
    padding: 16px 24px; background: #0a0a10; border-top: 1px solid #1e1e2e;
  }
</style>
