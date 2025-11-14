package providers

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"hostaggr/internal/models"
)

type Mock2 struct{}

func NewMock2() *Mock2 {
	return &Mock2{}
}

func (m *Mock2) Name() string {
	return "Mock2"
}

// Search performs a hotel search with simulated latency and random failures
func (m *Mock2) Search(ctx context.Context, req models.SearchRequest) ([]models.ProviderHotel, error) {
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
		return nil, errors.New("Mock2: random provider failure")
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
			HotelID:  "H123",
			Name:     "Hotel Atlas",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    135.00,
			Nights:   req.Nights,
		},
		{
			HotelID:  "H999",
			Name:     "Sofitel Palais",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    250.00,
			Nights:   req.Nights,
		},
		{
			HotelID:  "H111",
			Name:     "Dar Soukkar",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    75.00,
			Nights:   req.Nights,
		},
		{
			HotelID:  "H222",
			Name:     "Kech Boutique",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    110.00,
			Nights:   req.Nights,
		},
	}

	return hotels, nil
}
