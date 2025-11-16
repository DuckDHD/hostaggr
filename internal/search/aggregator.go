package search

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/singleflight"

	"hostaggr/internal/middleware"
	"hostaggr/internal/models"
	"hostaggr/internal/obs"
	"hostaggr/internal/providers"
)

type Aggregator struct {
	providers       []providers.Provider
	cache           *Cache
	providerTimeout time.Duration
	metrics         *obs.Metrics
	logger          *slog.Logger
	sf              singleflight.Group
}

func NewAggregator(provs []providers.Provider, cache *Cache, providerTimeout time.Duration, metrics *obs.Metrics, logger *slog.Logger) *Aggregator {
	return &Aggregator{
		providers:       provs,
		cache:           cache,
		providerTimeout: providerTimeout,
		metrics:         metrics,
		logger:          logger,
	}
}

func (a *Aggregator) Search(ctx context.Context, req models.SearchRequest) (models.SearchResponse, error) {
	startTime := time.Now()
	requestID := middleware.GetRequestID(ctx)

	if a.cache != nil {
		if cachedHotels, hit := a.cache.Get(req); hit {
			if a.logger != nil {
				a.logger.Info("cache hit",
					"request_id", requestID,
					"city", req.City,
					"hotel_count", len(cachedHotels),
				)
			}
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

	if a.logger != nil {
		a.logger.Info("cache miss",
			"request_id", requestID,
			"city", req.City,
		)
	}

	key := fmt.Sprintf("%s-%s-%d-%d", req.City, req.CheckIn, req.Nights, req.Adults)

	result, err, _ := a.sf.Do(key, func() (interface{}, error) {
		providerHotels, succeeded, failed := a.queryProviders(ctx, req)

		validHotels := make([]models.ProviderHotel, 0)
		for _, hotel := range providerHotels {
			if a.isValidHotel(hotel, req) {
				validHotels = append(validHotels, hotel)
			}
		}

		deduplicatedHotels := a.deduplicateHotels(validHotels)

		sort.Slice(deduplicatedHotels, func(i, j int) bool {
			return deduplicatedHotels[i].Price < deduplicatedHotels[j].Price
		})

		if a.cache != nil {
			a.cache.Set(req, deduplicatedHotels)
		}

		return searchResult{
			hotels:    deduplicatedHotels,
			succeeded: succeeded,
			failed:    failed,
		}, nil
	})

	if err != nil {
		return models.SearchResponse{}, err
	}

	sr := result.(searchResult)

	response := models.SearchResponse{
		Search: models.SearchInfo{
			City:    req.City,
			CheckIn: req.CheckIn,
			Nights:  req.Nights,
			Adults:  req.Adults,
		},
		Stats: models.Stats{
			ProvidersTotal:     len(a.providers),
			ProvidersSucceeded: sr.succeeded,
			ProvidersFailed:    sr.failed,
			Cache:              "miss",
			DurationMs:         time.Since(startTime).Milliseconds(),
		},
		Hotels: sr.hotels,
	}

	return response, nil
}

type searchResult struct {
	hotels    []models.Hotel
	succeeded int
	failed    int
}

func (a *Aggregator) queryProviders(ctx context.Context, req models.SearchRequest) ([]models.ProviderHotel, int, int) {
	queryCtx, cancel := context.WithTimeout(ctx, a.providerTimeout)
	defer cancel()

	requestID := middleware.GetRequestID(ctx)
	g, gCtx := errgroup.WithContext(queryCtx)

	var mu sync.Mutex
	var allHotels []models.ProviderHotel
	succeeded := 0
	failed := 0

	for _, provider := range a.providers {
		p := provider
		g.Go(func() error {
			providerStart := time.Now()
			hotels, err := p.Search(gCtx, req)
			duration := time.Since(providerStart).Milliseconds()

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				failed++
				if a.metrics != nil {
					a.metrics.IncrementProviderError(p.Name())
				}
				if a.logger != nil {
					a.logger.Warn("provider query failed",
						"request_id", requestID,
						"provider_name", p.Name(),
						"duration_ms", duration,
						"error", err.Error(),
					)
				}
				return nil
			}

			succeeded++
			if a.metrics != nil {
				a.metrics.IncrementProviderSuccess(p.Name())
			}
			if a.logger != nil {
				a.logger.Info("provider query succeeded",
					"request_id", requestID,
					"provider_name", p.Name(),
					"duration_ms", duration,
					"hotel_count", len(hotels),
				)
			}
			allHotels = append(allHotels, hotels...)
			return nil
		})
	}

	_ = g.Wait()

	return allHotels, succeeded, failed
}

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
