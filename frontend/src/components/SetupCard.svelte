<script lang="ts">
  import { onMount } from 'svelte'
  import QRCode from 'qrcode'
  import { library, consoleType, ethernetEnabled } from '../stores/state'

  let wiiLinkQR = ''
  let blueBombQR = ''
  let tiramisuQR = ''
  let printing = false

  const WIILINK_URL = 'https://wiilink24.com/guide'
  const BLUEBOMB_URL = 'https://github.com/Fullmetal5/bluebomb'
  const TIRAMISU_URL = 'https://wiiu.hacks.guide'

  onMount(async () => {
    wiiLinkQR   = await QRCode.toDataURL(WIILINK_URL,   { width: 200, margin: 1, color: { dark: '#000', light: '#fff' } })
    blueBombQR  = await QRCode.toDataURL(BLUEBOMB_URL,  { width: 200, margin: 1, color: { dark: '#000', light: '#fff' } })
    tiramisuQR  = await QRCode.toDataURL(TIRAMISU_URL,  { width: 200, margin: 1, color: { dark: '#000', light: '#fff' } })
  })

  $: games = $library.filter(g => g.selected)
  $: isWii        = $consoleType === 'wii'
  $: isWiiU       = $consoleType === 'wiiu'
  $: isWiiMini    = $consoleType === 'wiimini'
  $: isWiiMiniEth = $consoleType === 'wiimini' && $ethernetEnabled

  function print() {
    printing = true
    setTimeout(() => {
      window.print()
      printing = false
    }, 100)
  }
</script>

<div class="setup-card-wrap">
  <div class="toolbar no-print">
    <h2>Setup Card</h2>
    <p class="sub">Print this card and include it with the console so the buyer knows what they have.</p>
    <button class="primary" on:click={print} disabled={printing}>🖨 Print Setup Card</button>
  </div>

  <!-- PRINTABLE CARD -->
  <div class="card print-card" id="setup-card">
    <div class="card-header">
      <div class="card-header-left">
        <div class="card-title">🎮 Your Modded Console</div>
        <div class="card-sub">Pre-configured by KRUNK · Plug in and play</div>
      </div>
      <div class="card-badge">
        {#if isWii}Wii{:else if isWiiU}Wii U{:else if isWiiMiniEth}Wii Mini + Ethernet{:else}Wii Mini{/if}
      </div>
    </div>

    <div class="card-body">
      <!-- Left column: installed software -->
      <div class="col">
        <section>
          <h3>What's Installed</h3>
          <ul>
            <li>Homebrew Channel (HBC)</li>
            <li>WiiFlow — game launcher with coverflow UI</li>
            <li>USBLoader GX — backup launcher</li>
            <li>Nintendont — GameCube game support</li>
            <li>Not64 — Nintendo 64 emulator</li>
            <li>WiiStation — PlayStation 1 emulator</li>
            <li>Snes9xGX — SNES emulator</li>
            <li>FCEUltraGX — NES emulator</li>
            <li>Visual Boy Advance GX — GBA emulator</li>
            {#if isWii || isWiiMiniEth}
              <li>WiiLink — restored online channels</li>
            {/if}
            {#if isWiiMiniEth}
              <li>Ethernet Enabler — USB internet on Wii Mini</li>
            {/if}
          </ul>
        </section>

        {#if games.length > 0}
          <section>
            <h3>Games on USB ({games.length})</h3>
            <ul class="game-list">
              {#each games.slice(0, 24) as g}
                <li>{g.title}</li>
              {/each}
              {#if games.length > 24}
                <li class="more">+ {games.length - 24} more…</li>
              {/if}
            </ul>
          </section>
        {/if}
      </div>

      <!-- Right column: QR codes + setup steps -->
      <div class="col qr-col">
        {#if isWii}
          <section class="qr-section">
            <img src={wiiLinkQR} alt="WiiLink QR" class="qr" />
            <div class="qr-label">
              <strong>Activate WiiLink Online</strong>
              <span>Scan to follow the guide — restores Wii Room,<br>Food Channel, and more.</span>
            </div>
          </section>
          <section>
            <h3>Quick Start</h3>
            <ol>
              <li>Plug in USB drive (back USB port)</li>
              <li>Insert SD card</li>
              <li>Power on → Homebrew Channel</li>
              <li>Launch <strong>WiiFlow</strong> to play games</li>
              <li>Scan QR code above to activate online channels</li>
            </ol>
          </section>
        {:else if isWiiU}
          <section class="qr-section">
            <img src={tiramisuQR} alt="Tiramisu QR" class="qr" />
            <div class="qr-label">
              <strong>Wii U Hacks Guide</strong>
              <span>Tiramisu CFW is installed. Scan for<br>the full guide and updates.</span>
            </div>
          </section>
          <section>
            <h3>Quick Start</h3>
            <ol>
              <li>Plug in USB drive to Wii U</li>
              <li>Launch vWii from Wii U Menu</li>
              <li>Open Homebrew Channel → WiiFlow to play games</li>
              <li>Note: WiiLink not available in vWii mode</li>
            </ol>
          </section>
        {:else if isWiiMiniEth}
          <section class="qr-section">
            <img src={wiiLinkQR} alt="WiiLink QR" class="qr" />
            <div class="qr-label">
              <strong>Activate WiiLink Online</strong>
              <span>Scan to follow the guide — restores<br>online channels via USB Ethernet.</span>
            </div>
          </section>
          <section>
            <h3>Quick Start</h3>
            <ol>
              <li>Plug in USB drive (only USB port)</li>
              <li>Power on → Homebrew Channel</li>
              <li>Run <strong>Ethernet Enabler</strong> once (makes NAND change)</li>
              <li>Plug in USB Ethernet adapter (ASIX AX88179)</li>
              <li>Launch <strong>WiiFlow</strong> to play games</li>
              <li>Scan QR code above to activate online channels</li>
            </ol>
          </section>
        {:else if isWiiMini}
          <section class="qr-section">
            <img src={blueBombQR} alt="BlueBomb QR" class="qr" />
            <div class="qr-label">
              <strong>BlueBomb Setup</strong>
              <span>If homebrew isn't installed yet,<br>scan for the BlueBomb guide.</span>
            </div>
          </section>
          <section>
            <h3>Quick Start</h3>
            <ol>
              <li>Plug in USB drive (only USB port)</li>
              <li>Power on → Homebrew Channel</li>
              <li>Launch <strong>WiiFlow</strong> to play games</li>
              <li>Note: No WiFi — online features unavailable</li>
            </ol>
          </section>
        {/if}
      </div>
    </div>

    <div class="card-footer">
      <span>Prepped by <strong>KRUNK</strong> · krunk.site/support</span>
      <span class="footer-sep">·</span>
      <span>See <strong>HOW_TO_ADD_GAMES.txt</strong> on your USB drive to add your own games</span>
    </div>
  </div>
</div>

<style>
  .setup-card-wrap { display: flex; flex-direction: column; gap: 20px; }

  .toolbar { margin-bottom: 4px; }
  h2 { font-size: 20px; margin-bottom: 6px; }
  .sub { color: #888; font-size: 13px; margin-bottom: 16px; }

  button {
    padding: 9px 20px; border-radius: 6px; border: 1px solid #333;
    background: #1a1a28; color: #e8e8f0; cursor: pointer; font-size: 13px;
  }
  button.primary { background: #7c6fef; border-color: #7c6fef; font-weight: 600; }
  button:disabled { opacity: 0.5; cursor: not-allowed; }

  /* Card itself — white/print-friendly */
  .print-card {
    background: #fff;
    color: #111;
    border-radius: 10px;
    border: 1px solid #ddd;
    overflow: hidden;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
    font-size: 12px;
    max-width: 760px;
  }

  .card-header {
    background: #1a1230;
    color: #fff;
    padding: 14px 20px;
    display: flex;
    align-items: center;
    justify-content: space-between;
  }
  .card-header-left { display: flex; flex-direction: column; }
  .card-title { font-size: 18px; font-weight: 700; }
  .card-sub   { font-size: 12px; opacity: 0.7; margin-top: 2px; }
  .card-badge {
    padding: 4px 12px;
    background: rgba(255,255,255,0.12);
    border: 1px solid rgba(255,255,255,0.2);
    border-radius: 20px;
    font-size: 11px;
    font-weight: 700;
    letter-spacing: 0.5px;
    white-space: nowrap;
  }

  .card-body {
    display: flex;
    gap: 0;
    padding: 16px 20px;
  }

  .col { flex: 1; padding-right: 20px; }
  .col:last-child { padding-right: 0; }
  .qr-col { flex: 0 0 220px; padding-right: 0; }

  section { margin-bottom: 16px; }
  h3 { font-size: 11px; font-weight: 700; text-transform: uppercase;
       letter-spacing: 0.8px; color: #555; margin-bottom: 6px; border-bottom: 1px solid #eee; padding-bottom: 4px; }

  ul { list-style: none; padding: 0; margin: 0; }
  ul li { padding: 2px 0; color: #222; }
  ul li::before { content: '✓ '; color: #7c6fef; font-weight: 700; }
  .game-list li::before { content: '▸ '; color: #aaa; }
  .more { color: #888 !important; font-style: italic; }

  ol { padding-left: 18px; margin: 0; }
  ol li { padding: 3px 0; color: #222; }

  .qr-section { display: flex; flex-direction: column; align-items: center; gap: 8px; margin-bottom: 16px; }
  .qr { width: 140px; height: 140px; border: 1px solid #ddd; border-radius: 4px; }
  .qr-label { text-align: center; }
  .qr-label strong { display: block; font-size: 12px; margin-bottom: 2px; }
  .qr-label span { font-size: 11px; color: #555; line-height: 1.4; }

  .card-footer {
    background: #f5f5f5;
    color: #888;
    font-size: 10px;
    text-align: center;
    padding: 8px 16px;
    border-top: 1px solid #eee;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    flex-wrap: wrap;
  }
  .card-footer strong { color: #444; }
  .footer-sep { opacity: 0.4; }

  /* Print styles */
  @media print {
    :global(body) { background: #fff !important; }
    .no-print { display: none !important; }
    .print-card {
      border: none;
      border-radius: 0;
      max-width: 100%;
      box-shadow: none;
    }
  }
</style>
