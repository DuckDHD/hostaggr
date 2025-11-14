package providers

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"hostaggr/internal/models"
)

type Mock3 struct{}

func NewMock3() *Mock3 {
	return &Mock3{}
}

func (m *Mock3) Name() string {
	return "Mock3"
}

// Search performs a hotel search with simulated latency and random failures
func (m *Mock3) Search(ctx context.Context, req models.SearchRequest) ([]models.ProviderHotel, error) {
	// Random latency between 50-500ms
	latency := time.Duration(50+rand.Intn(451)) * time.Millisecond

	timer := time.NewTimer(latency)
	defer timer.Stop()

	// Respect context cancellation during sleep
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-timer.C:
		// Continue after sleep
	}

	// 20% random failure rate
	if rand.Float32() < 0.2 {
		return nil, errors.New("Mock3: random provider failure")
	}

	// Use inconsistent city casing
	cityCasings := []string{
		req.City,
		toTitle(req.City),
		toUpper(req.City),
		toLower(req.City),
	}

	hotels := []models.ProviderHotel{
		{
			HotelID:  "H789",
			Name:     "Le Meridien",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    195.00,
			Nights:   req.Nights,
		},
		{
			HotelID:  "H333",
			Name:     "Royal Mansour",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    450.00,
			Nights:   req.Nights,
		},
		{
			HotelID:  "H444",
			Name:     "La Mamounia",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    380.00,
			Nights:   req.Nights,
		},
	}

	return hotels, nil
}
