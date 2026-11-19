<script lang="ts">
  import { onMount } from 'svelte'
  import { drives, usbDrive, sdDrive, step, formatConfirmed, library } from '../stores/state'
  import type { DriveInfo } from '../stores/state'

  let loading = true
  let selectedUsb: DriveInfo | null = null
  let selectedSd: DriveInfo | null = null
  let confirmed = false

  $: usbDrive.set(selectedUsb)
  $: sdDrive.set(selectedSd)
  $: formatConfirmed.set(confirmed)

  onMount(async () => {
    await refreshDrives()
  })

  async function refreshDrives() {
    loading = true
    try {
      const list: DriveInfo[] = await window.go.main.App.ListDrives()
      drives.set(list)
    } catch (e) {
      drives.set([])
    }
    loading = false
  }

  $: usbPreview = selectedUsb
    ? `Will erase ${selectedUsb.path} (${fmtSize(selectedUsb.sizeBytes)}) and format FAT32 "WIIGAMES". CANNOT BE UNDONE.`
    : ''
  $: sdPreview = selectedSd
    ? `Will erase ${selectedSd.path} (${fmtSize(selectedSd.sizeBytes)}) and format FAT32 "WIISD". CANNOT BE UNDONE.`
    : ''
  $: canProceed = selectedUsb && selectedSd && confirmed && selectedUsb.path !== selectedSd.path

  function fmtSize(bytes: number) {
    if (bytes < 1024 * 1024 * 1024) return (bytes / 1024 / 1024).toFixed(0) + ' MB'
    return (bytes / 1024 / 1024 / 1024).toFixed(1) + ' GB'
  }

  const USB_TYPES = ['wii', 'gcn']
  const SD_ROM_TYPES = ['n64', 'ps1', 'snes', 'nes', 'gba']

  $: selectedGames = $library.filter(g => g.selected)
  $: usbBytes  = selectedGames.filter(g => USB_TYPES.includes(g.type)).reduce((s, g) => s + g.sizeBytes, 0)
  $: sdRomBytes = selectedGames.filter(g => SD_ROM_TYPES.includes(g.type)).reduce((s, g) => s + g.sizeBytes, 0)
  $: wiiCount  = selectedGames.filter(g => g.type === 'wii').length
  $: gcnCount  = selectedGames.filter(g => g.type === 'gcn').length
  $: romCount  = selectedGames.filter(g => SD_ROM_TYPES.includes(g.type)).length

  $: usbFit = !selectedUsb || usbBytes <= selectedUsb.sizeBytes * 0.95
  $: sdFit  = !selectedSd  || sdRomBytes <= selectedSd.sizeBytes * 0.90 // homebrew also lives on SD

  function next() { step.set(3) }
  function back() { step.set(1) }
</script>

<div class="drives">
  <h2>Select Drives</h2>
  <p class="sub">Plug in your USB drive and SD card, then select each below.</p>

  <div class="refresh-row">
    <button class="ghost" on:click={refreshDrives} disabled={loading}>
      {loading ? 'Scanning…' : '↻ Refresh'}
    </button>
  </div>

  {#if loading}
    <div class="empty">Detecting drives…</div>
  {:else if $drives.length === 0}
    <div class="empty">No removable drives detected. Plug in your USB and SD card, then refresh.</div>
  {:else}
    <div class="drive-row">
      <label for="usb-select" class="drive-label">USB Drive (games)</label>
      <select id="usb-select" bind:value={selectedUsb}>
        <option value={null}>— Select USB —</option>
        {#each $drives as d}
          <option value={d}>{d.path} · {d.label || 'Unlabeled'} · {fmtSize(d.sizeBytes)}</option>
        {/each}
      </select>
    </div>

    <div class="drive-row">
      <label for="sd-select" class="drive-label">SD Card (homebrew)</label>
      <select id="sd-select" bind:value={selectedSd}>
        <option value={null}>— Select SD Card —</option>
        {#each $drives as d}
          <option value={d}>{d.path} · {d.label || 'Unlabeled'} · {fmtSize(d.sizeBytes)}</option>
        {/each}
      </select>
    </div>

    {#if selectedGames.length > 0}
      <div class="storage-estimate">
        <div class="est-row">
          <span class="est-label">USB (Wii{gcnCount > 0 ? ' + GCN' : ''} games)</span>
          <span class="est-val" class:over={!usbFit}>
            {fmtSize(usbBytes)} needed
            {#if selectedUsb}· {fmtSize(selectedUsb.sizeBytes)} available{/if}
            {#if wiiCount > 0}<span class="count">{wiiCount} Wii</span>{/if}
            {#if gcnCount > 0}<span class="count">{gcnCount} GCN</span>{/if}
          </span>
        </div>
        {#if romCount > 0}
          <div class="est-row">
            <span class="est-label">SD card (emulator ROMs)</span>
            <span class="est-val" class:over={!sdFit}>
              {fmtSize(sdRomBytes)} needed
              {#if selectedSd}· {fmtSize(selectedSd.sizeBytes)} available{/if}
              <span class="count">{romCount} ROMs</span>
            </span>
          </div>
        {/if}
        {#if !usbFit}
          <p class="over-warn">⚠ USB drive may be too small for selected games.</p>
        {/if}
        {#if !sdFit}
          <p class="over-warn">⚠ SD card may not have enough space for ROMs + homebrew.</p>
        {/if}
      </div>
    {/if}

    {#if selectedUsb || selectedSd}
      <div class="warning-box">
        {#if usbPreview}<p>{usbPreview}</p>{/if}
        {#if sdPreview}<p>{sdPreview}</p>{/if}
      </div>

      <label class="confirm-row">
        <input type="checkbox" bind:checked={confirmed} />
        I understand this will permanently erase the selected drives.
      </label>
    {/if}
  {/if}

  <div class="footer">
    <button class="ghost" on:click={back}>← Back</button>
    {#if canProceed}
      <button class="primary" on:click={next}>Next: Install →</button>
    {/if}
  </div>
</div>

<style>
  h2 { font-size: 20px; margin-bottom: 6px; }
  .sub { color: #888; margin-bottom: 20px; }
  .refresh-row { margin-bottom: 20px; }

  button {
    padding: 8px 16px; border-radius: 6px; border: 1px solid #333;
    background: #1a1a28; color: #e8e8f0; cursor: pointer; font-size: 13px;
  }
  button:hover:not(:disabled) { background: #24243a; }
  button:disabled { opacity: 0.4; }
  button.ghost { background: transparent; }
  button.primary { background: #7c6fef; border-color: #7c6fef; font-weight: 600; }

  .empty { color: #555; padding: 40px 0; text-align: center; }

  .drive-row { display: flex; flex-direction: column; gap: 6px; margin-bottom: 20px; }
  .drive-label { font-weight: 500; font-size: 13px; color: #aaa; }
  select {
    background: #12121a; border: 1px solid #2a2a3e; color: #e8e8f0;
    padding: 10px 12px; border-radius: 6px; font-size: 13px; width: 100%; max-width: 480px;
  }

  .storage-estimate {
    background: #0d0d18; border: 1px solid #252535; border-radius: 8px;
    padding: 12px 16px; margin-bottom: 16px; display: flex; flex-direction: column; gap: 8px;
  }
  .est-row { display: flex; justify-content: space-between; align-items: baseline; gap: 12px; font-size: 12px; }
  .est-label { color: #888; white-space: nowrap; }
  .est-val { color: #ccc; display: flex; align-items: center; gap: 8px; flex-wrap: wrap; justify-content: flex-end; }
  .est-val.over { color: #ff7777; }
  .count { background: #1e1e2e; border: 1px solid #2a2a3e; border-radius: 10px; padding: 1px 7px; font-size: 11px; color: #888; }
  .over-warn { font-size: 12px; color: #ffb347; margin-top: 2px; }

  .warning-box {
    background: #1a1200; border: 1px solid #5a3800; border-radius: 6px;
    padding: 12px 16px; margin: 16px 0; color: #ffb347;
    display: flex; flex-direction: column; gap: 6px;
  }

  .confirm-row {
    display: flex; align-items: center; gap: 10px;
    color: #ccc; font-size: 13px; cursor: pointer; margin-bottom: 80px;
  }

  .footer {
    position: fixed; bottom: 0; left: 0; right: 0;
    display: flex; justify-content: space-between; align-items: center;
    padding: 16px 24px; background: #0a0a10; border-top: 1px solid #1e1e2e;
  }
</style>
