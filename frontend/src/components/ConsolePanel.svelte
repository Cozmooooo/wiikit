<script lang="ts">
  import { step, consoleType, type ConsoleType } from '../stores/state'

  const consoles: { id: ConsoleType; label: string; sub: string; badge?: string }[] = [
    {
      id: 'wii',
      label: 'Nintendo Wii',
      sub: 'Full setup: homebrew, WiiLink online, games, LetterBomb exploit on SD',
    },
    {
      id: 'wiiu',
      label: 'Wii U (vWii mode)',
      sub: 'Wii games via vWii — Tiramisu/Aroma CFW required. WiiLink not supported on vWii.',
      badge: 'Wii U',
    },
    {
      id: 'wiimini',
      label: 'Wii Mini',
      sub: 'USB-only, no WiFi/LAN. WiiLink skipped. BlueBomb exploit required (separate hardware).',
      badge: 'Limited',
    },
  ]

  let selected: ConsoleType = $consoleType

  function next() {
    consoleType.set(selected)
    step.set(1)
  }
</script>

<div class="console-pick">
  <h2>What console are you prepping?</h2>
  <p class="sub">WiiKit will tailor the homebrew stack and exploit method for your hardware.</p>

  <div class="cards">
    {#each consoles as c}
      <button
        class="card"
        class:selected={selected === c.id}
        on:click={() => (selected = c.id)}
      >
        <div class="card-top">
          <span class="card-label">{c.label}</span>
          {#if c.badge}
            <span class="badge">{c.badge}</span>
          {/if}
        </div>
        <p class="card-sub">{c.sub}</p>
      </button>
    {/each}
  </div>

  <div class="footer">
    <button class="primary big" on:click={next}>Next →</button>
  </div>
</div>

<style>
  h2 { font-size: 20px; margin-bottom: 6px; }
  .sub { color: #888; font-size: 13px; margin-bottom: 28px; }

  .cards { display: flex; flex-direction: column; gap: 12px; margin-bottom: 80px; }

  .card {
    text-align: left;
    background: #12121a;
    border: 1px solid #252535;
    border-radius: 10px;
    padding: 16px 20px;
    cursor: pointer;
    transition: border-color 0.15s, background 0.15s;
    color: #e8e8f0;
  }
  .card:hover { border-color: #7c6fef66; background: #15151f; }
  .card.selected { border-color: #7c6fef; background: #1a1830; }

  .card-top { display: flex; align-items: center; gap: 10px; margin-bottom: 6px; }
  .card-label { font-weight: 600; font-size: 15px; }

  .badge {
    font-size: 11px; font-weight: 600;
    background: #2a1f00; color: #f0a000;
    border: 1px solid #3a2a00;
    border-radius: 4px; padding: 2px 7px;
  }

  .card-sub { color: #888; font-size: 12px; line-height: 1.5; }

  button {
    padding: 8px 16px; border-radius: 6px; border: 1px solid #333;
    background: #1a1a28; color: #e8e8f0; cursor: pointer; font-size: 13px;
  }
  button.primary { background: #7c6fef; border-color: #7c6fef; font-weight: 600; }
  button.big { padding: 12px 32px; font-size: 15px; }

  .footer {
    position: fixed; bottom: 0; left: 0; right: 0;
    display: flex; justify-content: flex-end; align-items: center;
    padding: 16px 24px; background: #0a0a10; border-top: 1px solid #1e1e2e;
  }
</style>
