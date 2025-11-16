package tests

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"hostaggr/internal/models"
	"hostaggr/internal/providers"
	"hostaggr/internal/search"
)

// mockProvider implements the Provider interface for testing
type mockProvider struct {
	name    string
	hotels  []models.ProviderHotel
	err     error
	delay   time.Duration
}

func (m *mockProvider) Search(ctx context.Context, req models.SearchRequest) ([]models.ProviderHotel, error) {
	if m.delay > 0 {
		select {
		case <-time.After(m.delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	return m.hotels, m.err
}

func (m *mockProvider) Name() string {
	return m.name
}

func TestAggregator_Deduplication(t *testing.T) {
	tests := []struct {
		name           string
		providers      []providers.Provider
		expectedHotels []models.Hotel
		description    string
	}{
		{
			name: "duplicate hotel_ids keep lowest price",
			providers: []providers.Provider{
				&mockProvider{
					name: "provider1",
					hotels: []models.ProviderHotel{
						{HotelID: "H001", Name: "Hotel Alpha", City: "Tokyo", Currency: "USD", Price: 150.00, Nights: 2},
						{HotelID: "H002", Name: "Hotel Beta", City: "Tokyo", Currency: "USD", Price: 200.00, Nights: 2},
					},
				},
				&mockProvider{
					name: "provider2",
					hotels: []models.ProviderHotel{
						{HotelID: "H001", Name: "Hotel Alpha", City: "Tokyo", Currency: "USD", Price: 120.00, Nights: 2}, // Lower price
						{HotelID: "H003", Name: "Hotel Gamma", City: "Tokyo", Currency: "USD", Price: 180.00, Nights: 2},
					},
				},
			},
			expectedHotels: []models.Hotel{
				{HotelID: "H001", Name: "Hotel Alpha", Currency: "USD", Price: 120.00},
				{HotelID: "H003", Name: "Hotel Gamma", Currency: "USD", Price: 180.00},
				{HotelID: "H002", Name: "Hotel Beta", Currency: "USD", Price: 200.00},
			},
			description: "should deduplicate and keep lowest price for H001",
		},
		{
			name: "no duplicates",
			providers: []providers.Provider{
				&mockProvider{
					name: "provider1",
					hotels: []models.ProviderHotel{
						{HotelID: "H001", Name: "Hotel A", City: "Paris", Currency: "EUR", Price: 100.00, Nights: 1},
					},
				},
				&mockProvider{
					name: "provider2",
					hotels: []models.ProviderHotel{
						{HotelID: "H002", Name: "Hotel B", City: "Paris", Currency: "EUR", Price: 200.00, Nights: 1},
					},
				},
			},
			expectedHotels: []models.Hotel{
				{HotelID: "H001", Name: "Hotel A", Currency: "EUR", Price: 100.00},
				{HotelID: "H002", Name: "Hotel B", Currency: "EUR", Price: 200.00},
			},
			description: "should return all hotels when no duplicates",
		},
		{
			name: "multiple duplicates across three providers",
			providers: []providers.Provider{
				&mockProvider{
					name: "provider1",
					hotels: []models.ProviderHotel{
						{HotelID: "H001", Name: "Hotel One", City: "London", Currency: "GBP", Price: 300.00, Nights: 3},
					},
				},
				&mockProvider{
					name: "provider2",
					hotels: []models.ProviderHotel{
						{HotelID: "H001", Name: "Hotel One", City: "London", Currency: "GBP", Price: 250.00, Nights: 3},
					},
				},
				&mockProvider{
					name: "provider3",
					hotels: []models.ProviderHotel{
						{HotelID: "H001", Name: "Hotel One", City: "London", Currency: "GBP", Price: 280.00, Nights: 3},
					},
				},
			},
			expectedHotels: []models.Hotel{
				{HotelID: "H001", Name: "Hotel One", Currency: "GBP", Price: 250.00},
			},
			description: "should keep lowest price among three providers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agg := search.NewAggregator(tt.providers, nil, 2*time.Second, nil, nil)
			req := models.SearchRequest{
				City:    "Tokyo",
				CheckIn: "2025-01-15",
				Nights:  2,
				Adults:  2,
			}

			if tt.name == "no duplicates" {
				req.City = "Paris"
				req.Nights = 1
			} else if tt.name == "multiple duplicates across three providers" {
				req.City = "London"
				req.Nights = 3
			}

			result, err := agg.Search(context.Background(), req)

			require.NoError(t, err)
			assert.Len(t, result.Hotels, len(tt.expectedHotels), tt.description)

			// Create a map for easier comparison
			resultMap := make(map[string]models.Hotel)
			for _, h := range result.Hotels {
				resultMap[h.HotelID] = h
			}

			for _, expected := range tt.expectedHotels {
				actual, exists := resultMap[expected.HotelID]
				require.True(t, exists, "Hotel %s should exist", expected.HotelID)
				assert.Equal(t, expected.Price, actual.Price, "Price for %s should match", expected.HotelID)
				assert.Equal(t, expected.Name, actual.Name, "Name for %s should match", expected.HotelID)
				assert.Equal(t, expected.Currency, actual.Currency, "Currency for %s should match", expected.HotelID)
			}
		})
	}
}

func TestAggregator_CityMatching(t *testing.T) {
	tests := []struct {
		name            string
		requestCity     string
		providerHotels  []models.ProviderHotel
		expectedCount   int
		description     string
	}{
		{
			name:        "exact case match",
			requestCity: "Tokyo",
			providerHotels: []models.ProviderHotel{
				{HotelID: "H001", Name: "Hotel A", City: "Tokyo", Currency: "USD", Price: 100.00},
			},
			expectedCount: 1,
			description:   "should match exact case",
		},
		{
			name:        "case insensitive - lowercase request",
			requestCity: "tokyo",
			providerHotels: []models.ProviderHotel{
				{HotelID: "H001", Name: "Hotel A", City: "Tokyo", Currency: "USD", Price: 100.00},
			},
			expectedCount: 1,
			description:   "should match case insensitively with lowercase request",
		},
		{
			name:        "case insensitive - uppercase request",
			requestCity: "TOKYO",
			providerHotels: []models.ProviderHotel{
				{HotelID: "H001", Name: "Hotel A", City: "Tokyo", Currency: "USD", Price: 100.00},
			},
			expectedCount: 1,
			description:   "should match case insensitively with uppercase request",
		},
		{
			name:        "case insensitive - mixed case",
			requestCity: "ToKyO",
			providerHotels: []models.ProviderHotel{
				{HotelID: "H001", Name: "Hotel A", City: "tOkYo", Currency: "USD", Price: 100.00},
			},
			expectedCount: 1,
			description:   "should match case insensitively with mixed case",
		},
		{
			name:        "city mismatch",
			requestCity: "Tokyo",
			providerHotels: []models.ProviderHotel{
				{HotelID: "H001", Name: "Hotel A", City: "Osaka", Currency: "USD", Price: 100.00},
			},
			expectedCount: 0,
			description:   "should not match different city",
		},
		{
			name:        "mixed cities - filter only matching",
			requestCity: "Paris",
			providerHotels: []models.ProviderHotel{
				{HotelID: "H001", Name: "Hotel A", City: "Paris", Currency: "EUR", Price: 100.00},
				{HotelID: "H002", Name: "Hotel B", City: "London", Currency: "GBP", Price: 150.00},
				{HotelID: "H003", Name: "Hotel C", City: "paris", Currency: "EUR", Price: 120.00},
			},
			expectedCount: 2,
			description:   "should filter and return only Paris hotels",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &mockProvider{
				name:   "test-provider",
				hotels: tt.providerHotels,
			}
			agg := search.NewAggregator([]providers.Provider{provider}, nil, 2*time.Second, nil, nil)

			req := models.SearchRequest{
				City:    tt.requestCity,
				CheckIn: "2025-01-15",
				Nights:  2,
				Adults:  2,
			}

			result, err := agg.Search(context.Background(), req)

			require.NoError(t, err)
			assert.Len(t, result.Hotels, tt.expectedCount, tt.description)
		})
	}
}

func TestAggregator_InvalidDataFiltering(t *testing.T) {
	tests := []struct {
		name          string
		hotels        []models.ProviderHotel
		expectedCount int
		description   string
	}{
		{
			name: "missing hotel_id",
			hotels: []models.ProviderHotel{
				{HotelID: "", Name: "Hotel A", City: "Tokyo", Currency: "USD", Price: 100.00},
			},
			expectedCount: 0,
			description:   "should drop hotel with missing hotel_id",
		},
		{
			name: "missing name",
			hotels: []models.ProviderHotel{
				{HotelID: "H001", Name: "", City: "Tokyo", Currency: "USD", Price: 100.00},
			},
			expectedCount: 0,
			description:   "should drop hotel with missing name",
		},
		{
			name: "missing city",
			hotels: []models.ProviderHotel{
				{HotelID: "H001", Name: "Hotel A", City: "", Currency: "USD", Price: 100.00},
			},
			expectedCount: 0,
			description:   "should drop hotel with missing city",
		},
		{
			name: "missing currency",
			hotels: []models.ProviderHotel{
				{HotelID: "H001", Name: "Hotel A", City: "Tokyo", Currency: "", Price: 100.00},
			},
			expectedCount: 0,
			description:   "should drop hotel with missing currency",
		},
		{
			name: "zero price",
			hotels: []models.ProviderHotel{
				{HotelID: "H001", Name: "Hotel A", City: "Tokyo", Currency: "USD", Price: 0},
			},
			expectedCount: 0,
			description:   "should drop hotel with zero price",
		},
		{
			name: "negative price",
			hotels: []models.ProviderHotel{
				{HotelID: "H001", Name: "Hotel A", City: "Tokyo", Currency: "USD", Price: -50.00},
			},
			expectedCount: 0,
			description:   "should drop hotel with negative price",
		},
		{
			name: "mixed valid and invalid",
			hotels: []models.ProviderHotel{
				{HotelID: "H001", Name: "Hotel A", City: "Tokyo", Currency: "USD", Price: 100.00},
				{HotelID: "", Name: "Hotel B", City: "Tokyo", Currency: "USD", Price: 150.00},
				{HotelID: "H003", Name: "Hotel C", City: "Tokyo", Currency: "USD", Price: 0},
				{HotelID: "H004", Name: "Hotel D", City: "Tokyo", Currency: "USD", Price: 200.00},
			},
			expectedCount: 2,
			description:   "should keep only valid hotels",
		},
		{
			name: "all fields valid",
			hotels: []models.ProviderHotel{
				{HotelID: "H001", Name: "Hotel A", City: "Tokyo", Currency: "USD", Price: 100.00},
				{HotelID: "H002", Name: "Hotel B", City: "Tokyo", Currency: "USD", Price: 150.00},
			},
			expectedCount: 2,
			description:   "should keep all valid hotels",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &mockProvider{
				name:   "test-provider",
				hotels: tt.hotels,
			}
			agg := search.NewAggregator([]providers.Provider{provider}, nil, 2*time.Second, nil, nil)

			req := models.SearchRequest{
				City:    "Tokyo",
				CheckIn: "2025-01-15",
				Nights:  2,
				Adults:  2,
			}

			result, err := agg.Search(context.Background(), req)

			require.NoError(t, err)
			assert.Len(t, result.Hotels, tt.expectedCount, tt.description)
		})
	}
}

func TestAggregator_ProviderFailures(t *testing.T) {
	tests := []struct {
		name               string
		providers          []providers.Provider
		expectedHotels     int
		expectedSucceeded  int
		expectedFailed     int
		description        string
	}{
		{
			name: "all providers succeed",
			providers: []providers.Provider{
				&mockProvider{
					name: "provider1",
					hotels: []models.ProviderHotel{
						{HotelID: "H001", Name: "Hotel A", City: "Tokyo", Currency: "USD", Price: 100.00},
					},
				},
				&mockProvider{
					name: "provider2",
					hotels: []models.ProviderHotel{
						{HotelID: "H002", Name: "Hotel B", City: "Tokyo", Currency: "USD", Price: 150.00},
					},
				},
			},
			expectedHotels:    2,
			expectedSucceeded: 2,
			expectedFailed:    0,
			description:       "should return all hotels when all providers succeed",
		},
		{
			name: "one provider fails",
			providers: []providers.Provider{
				&mockProvider{
					name: "provider1",
					hotels: []models.ProviderHotel{
						{HotelID: "H001", Name: "Hotel A", City: "Tokyo", Currency: "USD", Price: 100.00},
					},
				},
				&mockProvider{
					name:   "provider2",
					err:    errors.New("provider unavailable"),
				},
			},
			expectedHotels:    1,
			expectedSucceeded: 1,
			expectedFailed:    1,
			description:       "should return partial results when one provider fails",
		},
		{
			name: "all providers fail",
			providers: []providers.Provider{
				&mockProvider{
					name: "provider1",
					err:  errors.New("provider error 1"),
				},
				&mockProvider{
					name: "provider2",
					err:  errors.New("provider error 2"),
				},
			},
			expectedHotels:    0,
			expectedSucceeded: 0,
			expectedFailed:    2,
			description:       "should return empty results when all providers fail",
		},
		{
			name: "two succeed one fails",
			providers: []providers.Provider{
				&mockProvider{
					name: "provider1",
					hotels: []models.ProviderHotel{
						{HotelID: "H001", Name: "Hotel A", City: "Tokyo", Currency: "USD", Price: 100.00},
					},
				},
				&mockProvider{
					name:   "provider2",
					err:    errors.New("provider error"),
				},
				&mockProvider{
					name: "provider3",
					hotels: []models.ProviderHotel{
						{HotelID: "H002", Name: "Hotel B", City: "Tokyo", Currency: "USD", Price: 150.00},
					},
				},
			},
			expectedHotels:    2,
			expectedSucceeded: 2,
			expectedFailed:    1,
			description:       "should return partial results when some providers fail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agg := search.NewAggregator(tt.providers, nil, 2*time.Second, nil, nil)

			req := models.SearchRequest{
				City:    "Tokyo",
				CheckIn: "2025-01-15",
				Nights:  2,
				Adults:  2,
			}

			result, err := agg.Search(context.Background(), req)

			require.NoError(t, err)
			assert.Len(t, result.Hotels, tt.expectedHotels, tt.description)
			assert.Equal(t, tt.expectedSucceeded, result.Stats.ProvidersSucceeded, "succeeded count should match")
			assert.Equal(t, tt.expectedFailed, result.Stats.ProvidersFailed, "failed count should match")
			assert.Equal(t, len(tt.providers), result.Stats.ProvidersTotal, "total count should match")
		})
	}
}

func TestAggregator_ContextTimeout(t *testing.T) {
	tests := []struct {
		name               string
		providers          []providers.Provider
		contextTimeout     time.Duration
		expectedMaxHotels  int
		description        string
	}{
		{
			name: "slow provider times out",
			providers: []providers.Provider{
				&mockProvider{
					name: "fast-provider",
					hotels: []models.ProviderHotel{
						{HotelID: "H001", Name: "Hotel A", City: "Tokyo", Currency: "USD", Price: 100.00},
					},
					delay: 100 * time.Millisecond,
				},
				&mockProvider{
					name: "slow-provider",
					hotels: []models.ProviderHotel{
						{HotelID: "H002", Name: "Hotel B", City: "Tokyo", Currency: "USD", Price: 150.00},
					},
					delay: 3 * time.Second, // Exceeds 2-second timeout
				},
			},
			contextTimeout:    5 * time.Second,
			expectedMaxHotels: 1, // Only fast provider should complete
			description:       "should timeout slow providers",
		},
		{
			name: "all providers complete within timeout",
			providers: []providers.Provider{
				&mockProvider{
					name: "provider1",
					hotels: []models.ProviderHotel{
						{HotelID: "H001", Name: "Hotel A", City: "Tokyo", Currency: "USD", Price: 100.00},
					},
					delay: 100 * time.Millisecond,
				},
				&mockProvider{
					name: "provider2",
					hotels: []models.ProviderHotel{
						{HotelID: "H002", Name: "Hotel B", City: "Tokyo", Currency: "USD", Price: 150.00},
					},
					delay: 200 * time.Millisecond,
				},
			},
			contextTimeout:    5 * time.Second,
			expectedMaxHotels: 2,
			description:       "should return all results when within timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agg := search.NewAggregator(tt.providers, nil, 2*time.Second, nil, nil)

			req := models.SearchRequest{
				City:    "Tokyo",
				CheckIn: "2025-01-15",
				Nights:  2,
				Adults:  2,
			}

			ctx, cancel := context.WithTimeout(context.Background(), tt.contextTimeout)
			defer cancel()

			result, err := agg.Search(ctx, req)

			require.NoError(t, err)
			assert.LessOrEqual(t, len(result.Hotels), tt.expectedMaxHotels, tt.description)

			// Verify stats are consistent
			assert.Equal(t, len(tt.providers), result.Stats.ProvidersTotal)
			assert.Equal(t, result.Stats.ProvidersSucceeded+result.Stats.ProvidersFailed, len(tt.providers))
		})
	}
}

func TestAggregator_Sorting(t *testing.T) {
	tests := []struct {
		name           string
		hotels         []models.ProviderHotel
		expectedOrder  []string // HotelIDs in expected order
		description    string
	}{
		{
			name: "hotels sorted by price ascending",
			hotels: []models.ProviderHotel{
				{HotelID: "H003", Name: "Hotel C", City: "Tokyo", Currency: "USD", Price: 300.00},
				{HotelID: "H001", Name: "Hotel A", City: "Tokyo", Currency: "USD", Price: 100.00},
				{HotelID: "H002", Name: "Hotel B", City: "Tokyo", Currency: "USD", Price: 200.00},
			},
			expectedOrder: []string{"H001", "H002", "H003"},
			description:   "should sort hotels by price in ascending order",
		},
		{
			name: "already sorted",
			hotels: []models.ProviderHotel{
				{HotelID: "H001", Name: "Hotel A", City: "Tokyo", Currency: "USD", Price: 100.00},
				{HotelID: "H002", Name: "Hotel B", City: "Tokyo", Currency: "USD", Price: 200.00},
				{HotelID: "H003", Name: "Hotel C", City: "Tokyo", Currency: "USD", Price: 300.00},
			},
			expectedOrder: []string{"H001", "H002", "H003"},
			description:   "should maintain order when already sorted",
		},
		{
			name: "reverse sorted",
			hotels: []models.ProviderHotel{
				{HotelID: "H005", Name: "Hotel E", City: "Tokyo", Currency: "USD", Price: 500.00},
				{HotelID: "H004", Name: "Hotel D", City: "Tokyo", Currency: "USD", Price: 400.00},
				{HotelID: "H003", Name: "Hotel C", City: "Tokyo", Currency: "USD", Price: 300.00},
				{HotelID: "H002", Name: "Hotel B", City: "Tokyo", Currency: "USD", Price: 200.00},
				{HotelID: "H001", Name: "Hotel A", City: "Tokyo", Currency: "USD", Price: 100.00},
			},
			expectedOrder: []string{"H001", "H002", "H003", "H004", "H005"},
			description:   "should sort reverse sorted list",
		},
		{
			name: "same prices",
			hotels: []models.ProviderHotel{
				{HotelID: "H001", Name: "Hotel A", City: "Tokyo", Currency: "USD", Price: 100.00},
				{HotelID: "H002", Name: "Hotel B", City: "Tokyo", Currency: "USD", Price: 100.00},
				{HotelID: "H003", Name: "Hotel C", City: "Tokyo", Currency: "USD", Price: 100.00},
			},
			expectedOrder: nil, // Order undefined for equal prices
			description:   "should handle hotels with same price",
		},
		{
			name: "single hotel",
			hotels: []models.ProviderHotel{
				{HotelID: "H001", Name: "Hotel A", City: "Tokyo", Currency: "USD", Price: 100.00},
			},
			expectedOrder: []string{"H001"},
			description:   "should handle single hotel",
		},
		{
			name: "decimal prices",
			hotels: []models.ProviderHotel{
				{HotelID: "H003", Name: "Hotel C", City: "Tokyo", Currency: "USD", Price: 99.99},
				{HotelID: "H001", Name: "Hotel A", City: "Tokyo", Currency: "USD", Price: 50.50},
				{HotelID: "H002", Name: "Hotel B", City: "Tokyo", Currency: "USD", Price: 75.25},
			},
			expectedOrder: []string{"H001", "H002", "H003"},
			description:   "should correctly sort decimal prices",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &mockProvider{
				name:   "test-provider",
				hotels: tt.hotels,
			}
			agg := search.NewAggregator([]providers.Provider{provider}, nil, 2*time.Second, nil, nil)

			req := models.SearchRequest{
				City:    "Tokyo",
				CheckIn: "2025-01-15",
				Nights:  2,
				Adults:  2,
			}

			result, err := agg.Search(context.Background(), req)

			require.NoError(t, err)
			assert.Len(t, result.Hotels, len(tt.hotels))

			if tt.expectedOrder != nil {
				// Verify order
				actualOrder := make([]string, len(result.Hotels))
				for i, h := range result.Hotels {
					actualOrder[i] = h.HotelID
				}
				assert.Equal(t, tt.expectedOrder, actualOrder, tt.description)

				// Verify prices are in ascending order
				for i := 1; i < len(result.Hotels); i++ {
					assert.LessOrEqual(t, result.Hotels[i-1].Price, result.Hotels[i].Price,
						"prices should be in ascending order")
				}
			} else {
				// For same prices, just verify all have the same price
				if len(result.Hotels) > 0 {
					price := result.Hotels[0].Price
					for _, h := range result.Hotels {
						assert.Equal(t, price, h.Price, "all prices should be equal")
					}
				}
			}
		})
	}
}


type countingProvider struct {
	name       string
	hotels     []models.ProviderHotel
	err        error
	delay      time.Duration
	callCount  *int
	mu         *sync.Mutex
}

func (m *countingProvider) Search(ctx context.Context, req models.SearchRequest) ([]models.ProviderHotel, error) {
	if m.mu != nil && m.callCount != nil {
		m.mu.Lock()
		*m.callCount++
		m.mu.Unlock()
	}

	if m.delay > 0 {
		select {
		case <-time.After(m.delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	return m.hotels, m.err
}

func (m *countingProvider) Name() string {
	return m.name
}

func TestAggregator_RequestCollapsing(t *testing.T) {
	var providerCallCount int
	var mu sync.Mutex

	provider := &countingProvider{
		name: "counting-provider",
		hotels: []models.ProviderHotel{
			{HotelID: "H001", Name: "Hotel A", City: "Tokyo", Currency: "USD", Price: 100.00},
			{HotelID: "H002", Name: "Hotel B", City: "Tokyo", Currency: "USD", Price: 150.00},
		},
		callCount: &providerCallCount,
		mu:        &mu,
		delay:     200 * time.Millisecond,
	}

	agg := search.NewAggregator([]providers.Provider{provider}, nil, 2*time.Second, nil, nil)

	req := models.SearchRequest{
		City:    "Tokyo",
		CheckIn: "2025-01-15",
		Nights:  2,
		Adults:  2,
	}

	numConcurrentRequests := 10
	var wg sync.WaitGroup
	results := make([]models.SearchResponse, numConcurrentRequests)
	errors := make([]error, numConcurrentRequests)

	for i := 0; i < numConcurrentRequests; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			result, err := agg.Search(context.Background(), req)
			results[index] = result
			errors[index] = err
		}(i)
	}

	wg.Wait()

	for i := 0; i < numConcurrentRequests; i++ {
		require.NoError(t, errors[i], "request %d should not error", i)
		assert.Len(t, results[i].Hotels, 2, "request %d should return 2 hotels", i)
	}

	mu.Lock()
	actualCallCount := providerCallCount
	mu.Unlock()

	assert.Equal(t, 1, actualCallCount, "provider should be called exactly once for %d concurrent identical requests", numConcurrentRequests)

	for i := 1; i < numConcurrentRequests; i++ {
		assert.Equal(t, results[0].Hotels, results[i].Hotels, "all requests should return identical results")
	}
}

func TestAggregator_RequestCollapsingDifferentRequests(t *testing.T) {
	var providerCallCount int
	var mu sync.Mutex

	provider := &countingProvider{
		name: "counting-provider",
		hotels: []models.ProviderHotel{
			{HotelID: "H001", Name: "Hotel A", City: "Tokyo", Currency: "USD", Price: 100.00},
		},
		callCount: &providerCallCount,
		mu:        &mu,
	}

	agg := search.NewAggregator([]providers.Provider{provider}, nil, 2*time.Second, nil, nil)

	req1 := models.SearchRequest{City: "Tokyo", CheckIn: "2025-01-15", Nights: 2, Adults: 2}
	req2 := models.SearchRequest{City: "Paris", CheckIn: "2025-01-15", Nights: 2, Adults: 2}
	req3 := models.SearchRequest{City: "Tokyo", CheckIn: "2025-02-01", Nights: 2, Adults: 2}

	var wg sync.WaitGroup

	wg.Add(3)
	go func() {
		defer wg.Done()
		agg.Search(context.Background(), req1)
	}()
	go func() {
		defer wg.Done()
		agg.Search(context.Background(), req2)
	}()
	go func() {
		defer wg.Done()
		agg.Search(context.Background(), req3)
	}()

	wg.Wait()

	mu.Lock()
	actualCallCount := providerCallCount
	mu.Unlock()

	assert.Equal(t, 3, actualCallCount, "provider should be called once for each unique request")
}

func TestAggregator_RequestCollapsingWithCache(t *testing.T) {
	var providerCallCount int
	var mu sync.Mutex

	provider := &countingProvider{
		name: "counting-provider",
		hotels: []models.ProviderHotel{
			{HotelID: "H001", Name: "Hotel A", City: "Tokyo", Currency: "USD", Price: 100.00},
		},
		callCount: &providerCallCount,
		mu:        &mu,
		delay:     100 * time.Millisecond,
	}

	cache := search.NewCache(30 * time.Second)
	agg := search.NewAggregator([]providers.Provider{provider}, cache, 2*time.Second, nil, nil)

	req := models.SearchRequest{
		City:    "Tokyo",
		CheckIn: "2025-01-15",
		Nights:  2,
		Adults:  2,
	}

	numConcurrentRequests := 5
	var wg sync.WaitGroup
	results := make([]models.SearchResponse, numConcurrentRequests)

	for i := 0; i < numConcurrentRequests; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			result, _ := agg.Search(context.Background(), req)
			results[index] = result
		}(i)
	}

	wg.Wait()

	mu.Lock()
	firstBatchCalls := providerCallCount
	mu.Unlock()

	assert.Equal(t, 1, firstBatchCalls, "first batch should only call provider once due to request collapsing")

	for i := 0; i < numConcurrentRequests; i++ {
		assert.Equal(t, "miss", results[i].Stats.Cache, "first batch should all be cache misses")
	}

	time.Sleep(200 * time.Millisecond)

	var wg2 sync.WaitGroup
	results2 := make([]models.SearchResponse, numConcurrentRequests)

	for i := 0; i < numConcurrentRequests; i++ {
		wg2.Add(1)
		go func(index int) {
			defer wg2.Done()
			result, _ := agg.Search(context.Background(), req)
			results2[index] = result
		}(i)
	}

	wg2.Wait()

	mu.Lock()
	secondBatchCalls := providerCallCount
	mu.Unlock()

	assert.Equal(t, 1, secondBatchCalls, "second batch should use cache, no additional provider calls")

	for i := 0; i < numConcurrentRequests; i++ {
		assert.Equal(t, "hit", results2[i].Stats.Cache, "second batch should all be cache hits")
	}
}
