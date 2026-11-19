<script lang="ts">
  import { drives, usbDrive, sdDrive, step, formatConfirmed } from '../stores/state'
  import type { DriveInfo } from '../stores/state'
  import { onMount } from 'svelte'

  let loading = true

  onMount(async () => {
    // @ts-ignore
    const list: DriveInfo[] = await window.go.main.App.ListDrives()
    drives.set(list)
    loading = false
  })

  async function refreshDrives() {
    loading = true
    // @ts-ignore
    const list: DriveInfo[] = await window.go.main.App.ListDrives()
    drives.set(list)
    loading = false
  }

  function getPreview(d: DriveInfo | null, label: string) {
    if (!d) return ''
    // @ts-ignore
    return window.go?.main?.App?.PreviewFormat(d, label) ?? ''
  }

  $: usbPreview = $usbDrive ? `Will erase ${$usbDrive.path} (${fmtSize($usbDrive.sizeBytes)}) and format as FAT32 "WIIGAMES". THIS CANNOT BE UNDONE.` : ''
  $: sdPreview = $sdDrive ? `Will erase ${$sdDrive.path} (${fmtSize($sdDrive.sizeBytes)}) and format as FAT32 "WIISD". THIS CANNOT BE UNDONE.` : ''

  $: canProceed = $usbDrive && $sdDrive && $formatConfirmed && $usbDrive.path !== $sdDrive.path

  function fmtSize(bytes: number) {
    return (bytes / 1024 / 1024 / 1024).toFixed(1) + ' GB'
  }

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
      <label class="drive-label">USB Drive (games)</label>
      <select bind:value={$usbDrive}>
        <option value={null}>— Select USB —</option>
        {#each $drives as d}
          <option value={d}>{d.path} · {d.label || 'Unlabeled'} · {fmtSize(d.sizeBytes)}</option>
        {/each}
      </select>
    </div>

    <div class="drive-row">
      <label class="drive-label">SD Card (homebrew)</label>
      <select bind:value={$sdDrive}>
        <option value={null}>— Select SD Card —</option>
        {#each $drives as d}
          <option value={d}>{d.path} · {d.label || 'Unlabeled'} · {fmtSize(d.sizeBytes)}</option>
        {/each}
      </select>
    </div>

    {#if $usbDrive || $sdDrive}
      <div class="warning-box">
        {#if usbPreview}<p>{usbPreview}</p>{/if}
        {#if sdPreview}<p>{sdPreview}</p>{/if}
      </div>

      <label class="confirm-row">
        <input type="checkbox" bind:checked={$formatConfirmed} />
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
