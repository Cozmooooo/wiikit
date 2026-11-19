package deals

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Facebook Marketplace has no public API.
// We use the internal graph endpoint that the mobile site calls.
// This is best-effort — FB can change it at any time.
const fbMarketplaceURL = "https://www.facebook.com/api/graphql/"

type fbSearchResult struct {
	Data struct {
		MarketplaceSearch struct {
			FeedUnits struct {
				Edges []struct {
					Node struct {
						Listing struct {
							ID    string `json:"id"`
							Title string `json:"marketplace_listing_title"`
							Price struct {
								Amount   string `json:"amount"`
								Currency string `json:"currency"`
							} `json:"listing_price"`
							Location struct {
								ReverseGeocode struct {
									City  string `json:"city"`
									State string `json:"state"`
								} `json:"reverse_geocode"`
							} `json:"location"`
							PrimaryPhoto struct {
								Image struct {
									URI string `json:"uri"`
								} `json:"image"`
							} `json:"primary_listing_photo"`
						} `json:"listing"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"feed_units"`
		} `json:"marketplace_search"`
	} `json:"data"`
}

var priceRe = regexp.MustCompile(`[\d.]+`)

// SearchFacebook scrapes the Facebook Marketplace internal GraphQL endpoint.
// Returns best-effort results; may return empty if FB changes the API.
func SearchFacebook(ctx context.Context, params SearchParams) ([]Listing, error) {
	if params.Limit == 0 {
		params.Limit = 20
	}

	variables := map[string]interface{}{
		"count":          params.Limit,
		"cursor":         nil,
		"params": map[string]interface{}{
			"bqf": map[string]interface{}{
				"cal_lat": 40.7128,
				"cal_lng": -74.0060,
				"query":   params.Query,
			},
			"browse_request_params": map[string]interface{}{
				"commerce_enable_local_pickup": true,
				"commerce_enable_shipping":     true,
				"filter_location_latitude":     40.7128,
				"filter_location_longitude":    -74.0060,
				"filter_price_min":             int(params.MinPrice * 100),
				"filter_price_max":             int(params.MaxPrice * 100),
			},
		},
	}

	varJSON, _ := json.Marshal(variables)

	form := url.Values{}
	form.Set("variables", string(varJSON))
	form.Set("doc_id", "7111574595556396") // MarketplaceSearchContentContainerQuery
	form.Set("server_timestamps", "true")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fbMarketplaceURL,
		strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("build fb request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Origin", "https://www.facebook.com")
	req.Header.Set("Referer", "https://www.facebook.com/marketplace/")
	req.Header.Set("X-FB-Friendly-Name", "MarketplaceSearchContentContainerQuery")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fb request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 302 || resp.StatusCode == 403 {
		return nil, fmt.Errorf("facebook requires login — open marketplace.facebook.com in your browser and search there")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("facebook: HTTP %d", resp.StatusCode)
	}

	var result fbSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		// FB might return login wall HTML
		return nil, fmt.Errorf("facebook response unparseable (login wall?): %w", err)
	}

	var listings []Listing
	for _, edge := range result.Data.MarketplaceSearch.FeedUnits.Edges {
		node := edge.Node.Listing
		priceStr := priceRe.FindString(node.Price.Amount)
		price, _ := strconv.ParseFloat(priceStr, 64)

		city := node.Location.ReverseGeocode.City
		state := node.Location.ReverseGeocode.State
		loc := city
		if state != "" {
			loc = city + ", " + state
		}

		l := Listing{
			ID:       node.ID,
			Title:    node.Title,
			Price:    price,
			Currency: node.Price.Currency,
			URL:      "https://www.facebook.com/marketplace/item/" + node.ID,
			ImageURL: node.PrimaryPhoto.Image.URI,
			Location: loc,
			Source:   "facebook",
		}
		l.Score = scoreFlipPotential(l)
		listings = append(listings, l)
	}

	return listings, nil
}
