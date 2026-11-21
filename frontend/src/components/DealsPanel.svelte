<script lang="ts">
  // Rough sell prices after KRUNK prep (Essentials pack)
  const SELL_PRICES: Record<string, number> = {
    wii:     75,
    wiimini: 55,
    wiiu:    95,
  }
  const DEFAULT_SELL_PRICE = 75

  const PRESETS = [
    { label: 'Wii',       query: 'Nintendo Wii console',      key: 'wii'     },
    { label: 'Wii Mini',  query: 'Nintendo Wii Mini console', key: 'wiimini' },
    { label: 'Wii U',     query: 'Nintendo Wii U console',    key: 'wiiu'    },
    { label: 'Bundle',    query: 'Nintendo Wii console bundle with games', key: 'wii' },
  ]

  let query = 'Nintendo Wii console'
  let activePresetKey = 'wii'
  let maxPrice = 60
  let minPrice = 5
  let loading = false
  let listings: any[] = []
  let ebayError = ''
  let fbError = ''
  let searched = false

  $: sellPrice = SELL_PRICES[activePresetKey] ?? DEFAULT_SELL_PRICE
  $: avgPrice = listings.length
    ? listings.reduce((s, l) => s + (l.price || 0), 0) / listings.length
    : 0
  $: bestMargin = listings.length
    ? Math.max(...listings.map(l => sellPrice - (l.price || 0)))
    : 0

  async function search() {
    loading = true
    ebayError = ''
    fbError = ''
    listings = []
    searched = false
    try {
      const result = await window.go.main.App.SearchDeals(query, maxPrice, minPrice)
      listings = result.listings || []
      ebayError = result.ebayError || ''
      fbError = result.facebookError || ''
    } catch (e: any) {
      ebayError = String(e)
    } finally {
      loading = false
      searched = true
    }
  }

  function applyPreset(p: typeof PRESETS[0]) {
    query = p.query
    activePresetKey = p.key
    search()
  }

  function flipTier(score: number): { label: string; color: string; bg: string } {
    if (score >= 75) return { label: 'HIGH',  color: '#22c55e', bg: '#052e16' }
    if (score >= 50) return { label: 'MED',   color: '#f59e0b', bg: '#1c1407' }
    return               { label: 'LOW',   color: '#ef4444', bg: '#1c0707' }
  }

  function margin(price: number): number {
    return sellPrice - price
  }

  function marginColor(m: number): string {
    if (m >= 40) return '#22c55e'
    if (m >= 20) return '#f59e0b'
    return '#ef4444'
  }

  function sourceLabel(source: string): string {
    return source === 'facebook' ? 'FB Marketplace' : 'eBay'
  }

  function timeAgo(listedAt: string): string {
    if (!listedAt) return ''
    const d = new Date(listedAt)
    if (isNaN(d.getTime())) return ''
    const mins = Math.floor((Date.now() - d.getTime()) / 60000)
    if (mins < 60)  return `${mins}m ago`
    if (mins < 1440) return `${Math.floor(mins/60)}h ago`
    return `${Math.floor(mins/1440)}d ago`
  }
</script>

<div class="deals-panel">
  <div class="presets-row">
    {#each PRESETS as p}
      <button
        class="preset-btn"
        class:active={activePresetKey === p.key && query === p.query}
        on:click={() => applyPreset(p)}
        disabled={loading}
      >{p.label}</button>
    {/each}
  </div>

  <div class="search-bar">
    <input
      type="text"
      bind:value={query}
      placeholder="Search query"
      on:keydown={(e) => e.key === 'Enter' && search()}
    />
    <div class="price-range">
      <span>$</span>
      <input type="number" bind:value={minPrice} min="0" max="999" placeholder="Min" />
      <span>–</span>
      <input type="number" bind:value={maxPrice} min="0" max="9999" placeholder="Max" />
    </div>
    <button class="search-btn" on:click={search} disabled={loading}>
      {loading ? 'Searching...' : 'Find Deals'}
    </button>
  </div>

  {#if ebayError}
    <div class="source-error">eBay: {ebayError}</div>
  {/if}
  {#if fbError}
    <div class="source-error">
      Facebook: {fbError}
      <a
        class="fb-fallback"
        href="https://www.facebook.com/marketplace/search?query={encodeURIComponent(query)}&minPrice={minPrice}&maxPrice={maxPrice}"
        target="_blank"
        rel="noopener noreferrer"
      >Search on FB Marketplace →</a>
    </div>
  {/if}

  {#if loading}
    <div class="loading-state">
      <div class="spinner"></div>
      <p>Scanning eBay and Facebook Marketplace…</p>
    </div>
  {:else if searched && listings.length === 0}
    <div class="empty-state">
      <p>No listings found. Try adjusting the price range or query.</p>
    </div>
  {:else if listings.length > 0}
    <div class="results-header">
      <span>{listings.length} listings · avg <strong>${avgPrice.toFixed(0)}</strong></span>
      {#if bestMargin > 0}
        <span class="best-margin-pill" style="color:{marginColor(bestMargin)}">
          best flip margin ~${bestMargin.toFixed(0)}
        </span>
      {/if}
      <span class="sell-price-hint">assumes ~${sellPrice} sell</span>
    </div>
    <div class="listings-grid">
      {#each listings as l, i}
        {@const tier = flipTier(l.score)}
        {@const m = margin(l.price || 0)}
        {@const ago = timeAgo(l.listedAt)}
        <a class="listing-card" class:top-deal={i === 0} href={l.url} target="_blank" rel="noopener noreferrer">
          <div class="card-img">
            {#if l.imageUrl}
              <img src={l.imageUrl} alt={l.title} loading="lazy" />
            {:else}
              <div class="no-img">📦</div>
            {/if}
            <div class="source-badge" class:fb={l.source === 'facebook'}>
              {sourceLabel(l.source)}
            </div>
            {#if i === 0}
              <div class="best-badge">🔥 Best Deal</div>
            {/if}
          </div>
          <div class="card-body">
            <div class="card-title">{l.title}</div>
            <div class="card-meta">
              <span class="price">${l.price > 0 ? l.price.toFixed(2) : '?'}</span>
              {#if l.condition}<span class="condition">{l.condition}</span>{/if}
              {#if ago}<span class="ago">{ago}</span>{/if}
            </div>
            {#if l.location}
              <div class="location-row">📍 {l.location}</div>
            {/if}
            <div class="bottom-row">
              <div class="margin-chip" style="color:{marginColor(m)}; background:{marginColor(m)}22">
                {m >= 0 ? '+' : ''}{m.toFixed(0)} margin
              </div>
              <div class="tier-chip" style="color:{tier.color}; background:{tier.bg}">
                FLIP: {tier.label}
              </div>
            </div>
          </div>
        </a>
      {/each}
    </div>
  {/if}
</div>

<style>
  .deals-panel {
    display: flex;
    flex-direction: column;
    gap: 14px;
    height: 100%;
    overflow: hidden;
  }

  /* Quick presets */
  .presets-row {
    display: flex;
    gap: 7px;
    flex-wrap: wrap;
  }
  .preset-btn {
    padding: 5px 14px;
    background: #13131c;
    border: 1px solid #252535;
    border-radius: 20px;
    color: #64748b;
    font-size: 12px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.15s;
    font-family: inherit;
  }
  .preset-btn:hover:not(:disabled) { border-color: #f97316; color: #f97316; }
  .preset-btn.active { background: #1a0a00; border-color: #f97316; color: #f97316; }
  .preset-btn:disabled { opacity: 0.5; cursor: not-allowed; }

  .search-bar {
    display: flex;
    gap: 10px;
    align-items: center;
    flex-wrap: wrap;
  }

  .search-bar input[type="text"] {
    flex: 1;
    min-width: 200px;
    padding: 8px 12px;
    background: #1a1a24;
    border: 1px solid #333;
    border-radius: 6px;
    color: #e2e8f0;
    font-size: 14px;
    font-family: inherit;
  }

  .price-range {
    display: flex;
    align-items: center;
    gap: 6px;
    color: #94a3b8;
    font-size: 14px;
  }

  .price-range input {
    width: 72px;
    padding: 8px 10px;
    background: #1a1a24;
    border: 1px solid #333;
    border-radius: 6px;
    color: #e2e8f0;
    font-size: 14px;
    font-family: inherit;
  }

  .search-btn {
    padding: 8px 20px;
    background: #f97316;
    color: white;
    border: none;
    border-radius: 6px;
    font-size: 14px;
    font-weight: 600;
    cursor: pointer;
    white-space: nowrap;
    font-family: inherit;
  }
  .search-btn:disabled { opacity: 0.6; cursor: not-allowed; }

  .source-error {
    background: #2d1515;
    border: 1px solid #7f1d1d;
    border-radius: 6px;
    padding: 8px 12px;
    color: #fca5a5;
    font-size: 13px;
    display: flex;
    align-items: center;
    gap: 10px;
    flex-wrap: wrap;
  }
  .fb-fallback {
    color: #60a5fa;
    font-size: 12px;
    font-weight: 600;
    text-decoration: none;
    white-space: nowrap;
    margin-left: auto;
  }
  .fb-fallback:hover { text-decoration: underline; }

  .loading-state, .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 12px;
    padding: 60px;
    color: #64748b;
  }

  .spinner {
    width: 32px;
    height: 32px;
    border: 3px solid #333;
    border-top-color: #f97316;
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }

  @keyframes spin { to { transform: rotate(360deg); } }

  .results-header {
    display: flex;
    align-items: center;
    gap: 10px;
    flex-wrap: wrap;
    color: #64748b;
    font-size: 13px;
  }
  .results-header strong { color: #94a3b8; }
  .best-margin-pill {
    padding: 2px 9px;
    border-radius: 20px;
    font-size: 12px;
    font-weight: 600;
    background: rgba(34,197,94,0.1);
  }
  .sell-price-hint { font-size: 11px; color: #334155; margin-left: auto; }

  .listings-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(230px, 1fr));
    gap: 12px;
    overflow-y: auto;
    padding-bottom: 16px;
  }

  .listing-card {
    background: #13131c;
    border: 1px solid #252535;
    border-radius: 10px;
    overflow: hidden;
    text-decoration: none;
    color: inherit;
    transition: border-color 0.15s, transform 0.1s;
    display: flex;
    flex-direction: column;
  }
  .listing-card:hover { border-color: #f97316; transform: translateY(-2px); }
  .listing-card.top-deal { border-color: #78350f; }

  .card-img {
    position: relative;
    height: 140px;
    background: #1a1a24;
    overflow: hidden;
  }
  .card-img img { width: 100%; height: 100%; object-fit: cover; }

  .no-img {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 100%;
    font-size: 36px;
    color: #334155;
  }

  .source-badge {
    position: absolute;
    top: 8px;
    right: 8px;
    padding: 2px 7px;
    border-radius: 4px;
    font-size: 11px;
    font-weight: 700;
    background: #0070c9;
    color: white;
  }
  .source-badge.fb { background: #1877f2; }

  .best-badge {
    position: absolute;
    bottom: 8px;
    left: 8px;
    padding: 2px 8px;
    border-radius: 4px;
    font-size: 11px;
    font-weight: 700;
    background: #f97316;
    color: white;
  }

  .card-body {
    padding: 11px;
    display: flex;
    flex-direction: column;
    gap: 7px;
    flex: 1;
  }

  .card-title {
    font-size: 13px;
    color: #cbd5e1;
    line-height: 1.4;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  .card-meta {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
    align-items: center;
    font-size: 12px;
    color: #64748b;
  }
  .price { font-size: 15px; font-weight: 700; color: #e2e8f0; }
  .condition { padding: 1px 6px; background: #1e293b; border-radius: 3px; }
  .ago { color: #475569; font-size: 11px; }

  .location-row { font-size: 11px; color: #475569; }

  .bottom-row {
    display: flex;
    gap: 7px;
    margin-top: auto;
    flex-wrap: wrap;
  }

  .margin-chip, .tier-chip {
    padding: 2px 8px;
    border-radius: 4px;
    font-size: 11px;
    font-weight: 700;
    letter-spacing: 0.3px;
  }
</style>
