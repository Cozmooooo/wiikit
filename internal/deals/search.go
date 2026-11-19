package deals

import (
	"context"
	"sort"
	"sync"
)

type SearchResult struct {
	Listings []Listing `json:"listings"`
	EbayErr  string    `json:"ebayError,omitempty"`
	FbErr    string    `json:"facebookError,omitempty"`
}

// Search runs eBay and Facebook searches in parallel and merges results.
func Search(ctx context.Context, params SearchParams) SearchResult {
	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		result  SearchResult
	)

	wg.Add(2)

	go func() {
		defer wg.Done()
		listings, err := SearchEbay(ctx, params)
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			result.EbayErr = err.Error()
		} else {
			result.Listings = append(result.Listings, listings...)
		}
	}()

	go func() {
		defer wg.Done()
		listings, err := SearchFacebook(ctx, params)
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			result.FbErr = err.Error()
		} else {
			result.Listings = append(result.Listings, listings...)
		}
	}()

	wg.Wait()

	// Sort merged results by score descending
	sort.Slice(result.Listings, func(i, j int) bool {
		return result.Listings[i].Score > result.Listings[j].Score
	})

	return result
}

// DefaultWiiParams returns sensible search defaults for Wii console hunting.
func DefaultWiiParams() SearchParams {
	return SearchParams{
		Query:    "Nintendo Wii console",
		MaxPrice: 60,
		MinPrice: 5,
		Limit:    25,
	}
}
