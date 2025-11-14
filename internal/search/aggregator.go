package search

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"hostaggr/internal/models"
	"hostaggr/internal/providers"
)

// Aggregator coordinates searches across multiple providers
type Aggregator struct {
	providers []providers.Provider
	cache     *Cache
}

// NewAggregator creates a new Aggregator instance
func NewAggregator(provs []providers.Provider, cache *Cache) *Aggregator {
	return &Aggregator{
		providers: provs,
		cache:     cache,
	}
}

// Search performs an aggregated search across all providers
func (a *Aggregator) Search(ctx context.Context, req models.SearchRequest) (models.SearchResponse, error) {
	startTime := time.Now()

	// Check cache first
	if a.cache != nil {
		if cachedHotels, hit := a.cache.Get(req); hit {
			// Build response from cached hotels
			response := models.SearchResponse{
				Search: models.SearchInfo{
					City:    req.City,
					CheckIn: req.CheckIn,
					Nights:  req.Nights,
					Adults:  req.Adults,
				},
				Stats: models.Stats{
					ProvidersTotal:     len(a.providers),
					ProvidersSucceeded: 0,
					ProvidersFailed:    0,
					Cache:              "hit",
					DurationMs:         time.Since(startTime).Milliseconds(),
				},
				Hotels: cachedHotels,
			}
			return response, nil
		}
	}

	// Query all providers concurrently
	providerHotels, succeeded, failed := a.queryProviders(ctx, req)

	// Validate hotels
	validHotels := make([]models.ProviderHotel, 0)
	for _, hotel := range providerHotels {
		if a.isValidHotel(hotel, req) {
			validHotels = append(validHotels, hotel)
		}
	}

	// Deduplicate and select best prices
	deduplicatedHotels := a.deduplicateHotels(validHotels)

	// Sort by price ascending
	sort.Slice(deduplicatedHotels, func(i, j int) bool {
		return deduplicatedHotels[i].Price < deduplicatedHotels[j].Price
	})

	// Build response
	response := models.SearchResponse{
		Search: models.SearchInfo{
			City:    req.City,
			CheckIn: req.CheckIn,
			Nights:  req.Nights,
			Adults:  req.Adults,
		},
		Stats: models.Stats{
			ProvidersTotal:     len(a.providers),
			ProvidersSucceeded: succeeded,
			ProvidersFailed:    failed,
			Cache:              "miss",
			DurationMs:         time.Since(startTime).Milliseconds(),
		},
		Hotels: deduplicatedHotels,
	}

	// Cache the result
	if a.cache != nil {
		a.cache.Set(req, deduplicatedHotels)
	}

	return response, nil
}

// queryProviders queries all providers concurrently with timeout and error handling
func (a *Aggregator) queryProviders(ctx context.Context, req models.SearchRequest) ([]models.ProviderHotel, int, int) {
	// Create context with 2-second timeout
	queryCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	g, gCtx := errgroup.WithContext(queryCtx)

	var mu sync.Mutex
	var allHotels []models.ProviderHotel
	succeeded := 0
	failed := 0

	for _, provider := range a.providers {
		p := provider
		g.Go(func() error {
			hotels, err := p.Search(gCtx, req)

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				failed++
				return nil
			}

			succeeded++
			allHotels = append(allHotels, hotels...)
			return nil
		})
	}

	_ = g.Wait()

	return allHotels, succeeded, failed
}

// isValidHotel validates a hotel against the search request
func (a *Aggregator) isValidHotel(h models.ProviderHotel, req models.SearchRequest) bool {

	if h.HotelID == "" || h.Name == "" || h.City == "" || h.Currency == "" {
		return false
	}

	if h.Price <= 0 {
		return false
	}

	if !strings.EqualFold(h.City, req.City) {
		return false
	}

	return true
}

// deduplicateHotels removes duplicates by hotel_id, keeping the lowest price
func (a *Aggregator) deduplicateHotels(hotels []models.ProviderHotel) []models.Hotel {
	bestPrices := make(map[string]models.Hotel)

	for _, ph := range hotels {
		existing, exists := bestPrices[ph.HotelID]

		hotel := models.Hotel{
			HotelID:  ph.HotelID,
			Name:     ph.Name,
			Currency: ph.Currency,
			Price:    ph.Price,
		}

		if !exists || hotel.Price < existing.Price {
			bestPrices[ph.HotelID] = hotel
		}
	}

	result := make([]models.Hotel, 0, len(bestPrices))
	for _, hotel := range bestPrices {
		result = append(result, hotel)
	}

	return result
}
