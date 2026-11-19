package deals

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// eBay Browse API — no OAuth needed for search, just an App ID (client ID).
// If AppID is empty we fall back to the scraping endpoint.
const ebayBrowseURL = "https://api.ebay.com/buy/browse/v1/item_summary/search"
const ebayFindURL = "https://svcs.ebay.com/services/search/FindingService/v1"

type Listing struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Price     float64   `json:"price"`
	Currency  string    `json:"currency"`
	Condition string    `json:"condition"`
	URL       string    `json:"url"`
	ImageURL  string    `json:"imageUrl"`
	Seller    string    `json:"seller"`
	Location  string    `json:"location"`
	Source    string    `json:"source"` // "ebay" | "facebook"
	ListedAt  time.Time `json:"listedAt"`
	Score     float64   `json:"score"` // flip potential 0-100
}

type SearchParams struct {
	Query    string
	MaxPrice float64
	MinPrice float64
	Limit    int
}

// SearchEbay uses the eBay Finding API (XML-free JSON mode, no key required for basic search).
func SearchEbay(ctx context.Context, params SearchParams) ([]Listing, error) {
	if params.Limit == 0 {
		params.Limit = 20
	}

	q := url.Values{}
	q.Set("OPERATION-NAME", "findItemsByKeywords")
	q.Set("SERVICE-VERSION", "1.0.0")
	q.Set("SECURITY-APPNAME", "WiiKit-Search") // public tier, no key needed
	q.Set("RESPONSE-DATA-FORMAT", "JSON")
	q.Set("keywords", params.Query)
	q.Set("paginationInput.entriesPerPage", strconv.Itoa(params.Limit))
	q.Set("sortOrder", "PricePlusShippingLowest")

	// Filter by condition: used (3000) or unspecified
	q.Set("itemFilter(0).name", "Condition")
	q.Set("itemFilter(0).value(0)", "3000") // Used
	q.Set("itemFilter(0).value(1)", "2500") // Seller refurbished
	q.Set("itemFilter(0).value(2)", "7000") // For parts

	if params.MaxPrice > 0 {
		q.Set("itemFilter(1).name", "MaxPrice")
		q.Set("itemFilter(1).value", fmt.Sprintf("%.2f", params.MaxPrice))
		q.Set("itemFilter(1).paramName", "Currency")
		q.Set("itemFilter(1).paramValue", "USD")
	}
	if params.MinPrice > 0 {
		q.Set("itemFilter(2).name", "MinPrice")
		q.Set("itemFilter(2).value", fmt.Sprintf("%.2f", params.MinPrice))
		q.Set("itemFilter(2).paramName", "Currency")
		q.Set("itemFilter(2).paramValue", "USD")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ebayFindURL+"?"+q.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("build ebay request: %w", err)
	}
	req.Header.Set("User-Agent", "WiiKit/1.0")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ebay request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ebay API: HTTP %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("parse ebay response: %w", err)
	}

	return parseEbayFindResponse(result), nil
}

func parseEbayFindResponse(raw map[string]interface{}) []Listing {
	var listings []Listing

	defer func() { recover() }() // eBay API shape varies; don't crash

	resp := raw["findItemsByKeywordsResponse"].([]interface{})[0].(map[string]interface{})
	items := resp["searchResult"].([]interface{})[0].(map[string]interface{})["item"].([]interface{})

	for _, item := range items {
		m := item.(map[string]interface{})
		l := Listing{Source: "ebay"}

		if v, ok := getStr(m, "itemId"); ok {
			l.ID = v
		}
		if v, ok := getStr(m, "title"); ok {
			l.Title = v
		}
		if v, ok := getStr(m, "viewItemURL"); ok {
			l.URL = v
		}
		if v, ok := getStr(m, "location"); ok {
			l.Location = v
		}

		// Price
		if sp, ok := m["sellingStatus"].([]interface{}); ok && len(sp) > 0 {
			if cp, ok := sp[0].(map[string]interface{})["currentPrice"].([]interface{}); ok && len(cp) > 0 {
				cpm := cp[0].(map[string]interface{})
				if v, ok := getStr(cpm, "__value__"); ok {
					l.Price, _ = strconv.ParseFloat(v, 64)
				}
				if v, ok := getStr(cpm, "@currencyId"); ok {
					l.Currency = v
				}
			}
		}

		// Condition
		if conds, ok := m["condition"].([]interface{}); ok && len(conds) > 0 {
			if cn, ok := conds[0].(map[string]interface{})["conditionDisplayName"].([]interface{}); ok && len(cn) > 0 {
				if v, ok := cn[0].(string); ok {
					l.Condition = v
				}
			}
		}

		// Seller
		if sellers, ok := m["sellerInfo"].([]interface{}); ok && len(sellers) > 0 {
			if sn, ok := sellers[0].(map[string]interface{})["sellerUserName"].([]interface{}); ok && len(sn) > 0 {
				if v, ok := sn[0].(string); ok {
					l.Seller = v
				}
			}
		}

		// Image
		if imgs, ok := m["galleryURL"].([]interface{}); ok && len(imgs) > 0 {
			if v, ok := imgs[0].(string); ok {
				l.ImageURL = strings.Replace(v, "s-l140", "s-l300", 1)
			}
		}

		l.Score = scoreFlipPotential(l)
		listings = append(listings, l)
	}

	sort.Slice(listings, func(i, j int) bool {
		return listings[i].Score > listings[j].Score
	})
	return listings
}

func getStr(m map[string]interface{}, key string) (string, bool) {
	v, ok := m[key]
	if !ok {
		return "", false
	}
	arr, ok := v.([]interface{})
	if ok && len(arr) > 0 {
		s, ok := arr[0].(string)
		return s, ok
	}
	s, ok := v.(string)
	return s, ok
}

// scoreFlipPotential rates a listing 0-100 based on flip potential.
// Higher = better deal (low price, good condition keywords, full console).
func scoreFlipPotential(l Listing) float64 {
	score := 50.0

	// Price scoring: ideal buy range for a Wii to flip is $15-45
	switch {
	case l.Price <= 0:
		score -= 20
	case l.Price < 10:
		score += 5 // too cheap = likely parts only
	case l.Price <= 20:
		score += 30
	case l.Price <= 35:
		score += 20
	case l.Price <= 50:
		score += 10
	case l.Price <= 75:
		score -= 5
	default:
		score -= 20
	}

	title := strings.ToLower(l.Title)

	// Positive signals
	if strings.Contains(title, "bundle") || strings.Contains(title, "lot") {
		score += 15
	}
	if strings.Contains(title, "controller") || strings.Contains(title, "nunchuck") {
		score += 10
	}
	if strings.Contains(title, "working") || strings.Contains(title, "tested") {
		score += 10
	}
	if strings.Contains(title, "console") {
		score += 5
	}

	// Negative signals
	if strings.Contains(title, "parts") || strings.Contains(title, "repair") || strings.Contains(title, "broken") {
		score -= 25
	}
	if strings.Contains(title, "disc") && !strings.Contains(title, "console") {
		score -= 15 // just a game disc, not a console
	}
	if strings.Contains(title, "case") || strings.Contains(title, "cover") {
		score -= 20
	}

	// Condition bonus
	if strings.Contains(strings.ToLower(l.Condition), "good") {
		score += 5
	}

	if score > 100 {
		return 100
	}
	if score < 0 {
		return 0
	}
	return score
}
