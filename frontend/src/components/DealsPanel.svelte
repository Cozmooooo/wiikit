<script lang="ts">
  let query = 'Nintendo Wii console'
  let maxPrice = 60
  let minPrice = 5
  let loading = false
  let listings: any[] = []
  let ebayError = ''
  let fbError = ''
  let searched = false

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

  function scoreColor(score: number): string {
    if (score >= 75) return '#22c55e'
    if (score >= 50) return '#f59e0b'
    return '#ef4444'
  }

  function sourceLabel(source: string): string {
    return source === 'facebook' ? 'FB Marketplace' : 'eBay'
  }
</script>

<div class="deals-panel">
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
    <div class="source-error">Facebook: {fbError}</div>
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
      <span>{listings.length} deals found · sorted by flip potential</span>
    </div>
    <div class="listings-grid">
      {#each listings as l}
        <a class="listing-card" href={l.url} target="_blank" rel="noopener noreferrer">
          <div class="card-img">
            {#if l.imageUrl}
              <img src={l.imageUrl} alt={l.title} loading="lazy" />
            {:else}
              <div class="no-img">📦</div>
            {/if}
            <div class="source-badge" class:fb={l.source === 'facebook'}>
              {sourceLabel(l.source)}
            </div>
          </div>
          <div class="card-body">
            <div class="card-title">{l.title}</div>
            <div class="card-meta">
              <span class="price">${l.price > 0 ? l.price.toFixed(2) : '?'}</span>
              {#if l.condition}<span class="condition">{l.condition}</span>{/if}
              {#if l.location}<span class="location">📍 {l.location}</span>{/if}
            </div>
            <div class="score-row">
              <div class="score-bar-bg">
                <div
                  class="score-bar-fill"
                  style="width:{l.score}%; background:{scoreColor(l.score)}"
                ></div>
              </div>
              <span class="score-label" style="color:{scoreColor(l.score)}">
                {Math.round(l.score)} / 100
              </span>
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
    gap: 16px;
    height: 100%;
    overflow: hidden;
  }

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
  }

  .search-btn:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .source-error {
    background: #2d1515;
    border: 1px solid #7f1d1d;
    border-radius: 6px;
    padding: 8px 12px;
    color: #fca5a5;
    font-size: 13px;
  }

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
    color: #64748b;
    font-size: 13px;
  }

  .listings-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
    gap: 14px;
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

  .listing-card:hover {
    border-color: #f97316;
    transform: translateY(-2px);
  }

  .card-img {
    position: relative;
    height: 160px;
    background: #1a1a24;
    overflow: hidden;
  }

  .card-img img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

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

  .source-badge.fb {
    background: #1877f2;
  }

  .card-body {
    padding: 12px;
    display: flex;
    flex-direction: column;
    gap: 8px;
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

  .price {
    font-size: 15px;
    font-weight: 700;
    color: #e2e8f0;
  }

  .condition {
    padding: 1px 6px;
    background: #1e293b;
    border-radius: 3px;
  }

  .score-row {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-top: auto;
  }

  .score-bar-bg {
    flex: 1;
    height: 4px;
    background: #1e293b;
    border-radius: 2px;
    overflow: hidden;
  }

  .score-bar-fill {
    height: 100%;
    border-radius: 2px;
    transition: width 0.3s;
  }

  .score-label {
    font-size: 11px;
    font-weight: 600;
    white-space: nowrap;
  }
</style>
