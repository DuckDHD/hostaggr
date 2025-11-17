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

	hotels := []models.ProviderHotel{
		// Shared hotel IDs with other mocks
		{
			HotelID:  "H123",
			Name:     "Hotel Atlas",
			City:     cityCasings[rand.Intn(len(cityCasings))],
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
		{
			HotelID:  "H999",
			Name:     "Sofitel Palais Imperial",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    285.00,
			Nights:   req.Nights,
		},
		{
			HotelID:  "H111",
			Name:     "Dar Soukkar",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    95.00,
			Nights:   req.Nights,
		},
		// Mock1 unique hotels (Budget tier)
		{
			HotelID:  "H001",
			Name:     "City Hostel",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    45.00,
			Nights:   req.Nights,
		},
		{
			HotelID:  "H002",
			Name:     "Backpacker Inn",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    52.50,
			Nights:   req.Nights,
		},
		{
			HotelID:  "H003",
			Name:     "Hotel Central",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    68.00,
			Nights:   req.Nights,
		},
		{
			HotelID:  "H004",
			Name:     "Medina Budget Stay",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    75.99,
			Nights:   req.Nights,
		},
		{
			HotelID:  "H005",
			Name:     "Kasbah Guesthouse",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    82.00,
			Nights:   req.Nights,
		},
		// Mid-range tier
		{
			HotelID:  "H006",
			Name:     "Riad Yasmine",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    105.00,
			Nights:   req.Nights,
		},
		{
			HotelID:  "H007",
			Name:     "Atlas Garden Hotel",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    118.50,
			Nights:   req.Nights,
		},
		{
			HotelID:  "H008",
			Name:     "Kech Boutique Hotel",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    135.00,
			Nights:   req.Nights,
		},
		{
			HotelID:  "H009",
			Name:     "Palmera Resort",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    149.99,
			Nights:   req.Nights,
		},
		{
			HotelID:  "H010",
			Name:     "Ibis Styles Downtown",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    165.00,
			Nights:   req.Nights,
		},
		{
			HotelID:  "H011",
			Name:     "Novotel City Center",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    172.50,
			Nights:   req.Nights,
		},
		// Upscale tier
		{
			HotelID:  "H012",
			Name:     "Marriott Palace",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    215.00,
			Nights:   req.Nights,
		},
		{
			HotelID:  "H013",
			Name:     "Hilton Downtown",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    245.00,
			Nights:   req.Nights,
		},
		{
			HotelID:  "H014",
			Name:     "Riad Fes Luxury",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    289.95,
			Nights:   req.Nights,
		},
		{
			HotelID:  "H015",
			Name:     "Movenpick Grand",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    325.00,
			Nights:   req.Nights,
		},
		// Luxury tier
		{
			HotelID:  "H016",
			Name:     "Four Seasons Resort",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    425.00,
			Nights:   req.Nights,
		},
		{
			HotelID:  "H017",
			Name:     "Royal Mansour",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    550.00,
			Nights:   req.Nights,
		},
		{
			HotelID:  "H018",
			Name:     "La Mamounia",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    495.99,
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
