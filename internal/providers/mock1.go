package providers

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"hostaggr/internal/models"
)

type Mock1 struct{}

func NewMock1() *Mock1 {
	return &Mock1{}
}

func (m *Mock1) Name() string {
	return "Mock1"
}

// Search performs a hotel search with simulated latency and random failures
func (m *Mock1) Search(ctx context.Context, req models.SearchRequest) ([]models.ProviderHotel, error) {
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
		return nil, errors.New("Mock1: random provider failure")
	}

	// Use inconsistent city casing
	cityCasings := []string{
		req.City,
		toTitle(req.City),
		toUpper(req.City),
		toLower(req.City),
	}
	randomCity := cityCasings[rand.Intn(len(cityCasings))]

	hotels := []models.ProviderHotel{
		{
			HotelID:  "H123",
			Name:     "Hotel Atlas",
			City:     randomCity,
			Currency: "EUR",
			Price:    129.90,
			Nights:   req.Nights,
		},
		{
			HotelID:  "H456",
			Name:     "Riad Zitoun",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    89.50,
			Nights:   req.Nights,
		},
		{
			HotelID:  "H789",
			Name:     "Le Meridien",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    199.00,
			Nights:   req.Nights,
		},
	}

	return hotels, nil
}

// Helper functions for string casing
func toTitle(s string) string {
	if len(s) == 0 {
		return s
	}
	runes := []rune(s)
	result := make([]rune, len(runes))
	makeUpper := true
	for i, r := range runes {
		if makeUpper && r >= 'a' && r <= 'z' {
			result[i] = r - 32
			makeUpper = false
		} else if !makeUpper && r >= 'A' && r <= 'Z' {
			result[i] = r + 32
		} else {
			result[i] = r
		}
		if r == ' ' {
			makeUpper = true
		}
	}
	return string(result)
}

func toUpper(s string) string {
	runes := []rune(s)
	for i, r := range runes {
		if r >= 'a' && r <= 'z' {
			runes[i] = r - 32
		}
	}
	return string(runes)
}

func toLower(s string) string {
	runes := []rune(s)
	for i, r := range runes {
		if r >= 'A' && r <= 'Z' {
			runes[i] = r + 32
		}
	}
	return string(runes)
}
