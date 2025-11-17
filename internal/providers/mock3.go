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
		// Shared hotel IDs - testing deduplication with different price points
		{
			HotelID:  "H123", // BETTER price than both Mock1 (129.90) and Mock2 (135.00)
			Name:     "Hotel Atlas",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    119.99,
			Nights:   req.Nights,
		},
		{
			HotelID:  "H456", // WORSE price than both Mock1 (89.50) and Mock2 (79.99)
			Name:     "Riad Zitoun",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    92.00,
			Nights:   req.Nights,
		},
		{
			HotelID:  "H789", // BETTER price than both Mock1 (199.00) and Mock2 (215.00)
			Name:     "Le Meridien",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    195.00,
			Nights:   req.Nights,
		},
		{
			HotelID:  "H999", // WORSE price than both Mock1 (285.00) and Mock2 (250.00)
			Name:     "Sofitel Palais Imperial",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    295.50,
			Nights:   req.Nights,
		},
		{
			HotelID:  "H111", // Middle price between Mock1 (95.00) and Mock2 (75.00)
			Name:     "Dar Soukkar",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    85.00,
			Nights:   req.Nights,
		},
		// Mock3 unique hotels with edge cases
		// LOWEST priced hotel overall
		{
			HotelID:  "H037",
			Name:     "Economy Hostel",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    45.00,
			Nights:   req.Nights,
		},
		// Budget tier with decimal edge cases
		{
			HotelID:  "H038",
			Name:     "Cheap Sleep Inn",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    49.95,
			Nights:   req.Nights,
		},
		{
			HotelID:  "H039",
			Name:     "Backpackers Paradise",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    58.99,
			Nights:   req.Nights,
		},
		{
			HotelID:  "H040",
			Name:     "Value Hotel",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    67.50,
			Nights:   req.Nights,
		},
		{
			HotelID:  "H041",
			Name:     "Express Inn",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    78.00,
			Nights:   req.Nights,
		},
		// Mid-range tier
		{
			HotelID:  "H042",
			Name:     "Comfort Suites",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    99.99,
			Nights:   req.Nights,
		},
		{
			HotelID:  "H043",
			Name:     "Holiday Inn Express",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    115.00,
			Nights:   req.Nights,
		},
		{
			HotelID:  "H044",
			Name:     "Courtyard Hotel",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    129.99,
			Nights:   req.Nights,
		},
		{
			HotelID:  "H045",
			Name:     "Radisson Blu",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    145.50,
			Nights:   req.Nights,
		},
		{
			HotelID:  "H046",
			Name:     "Crowne Plaza",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    162.00,
			Nights:   req.Nights,
		},
		{
			HotelID:  "H047",
			Name:     "DoubleTree by Hilton",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    179.95,
			Nights:   req.Nights,
		},
		// Upscale tier
		{
			HotelID:  "H048",
			Name:     "Westin Hotel",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    205.00,
			Nights:   req.Nights,
		},
		{
			HotelID:  "H049",
			Name:     "JW Marriott",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    238.99,
			Nights:   req.Nights,
		},
		{
			HotelID:  "H050",
			Name:     "Conrad Hotel",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    275.00,
			Nights:   req.Nights,
		},
		{
			HotelID:  "H051",
			Name:     "St. Regis",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    315.50,
			Nights:   req.Nights,
		},
		// Luxury tier with edge cases
		{
			HotelID:  "H052",
			Name:     "Royal Mansour",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    450.00,
			Nights:   req.Nights,
		},
		{
			HotelID:  "H053",
			Name:     "La Mamounia",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    380.00,
			Nights:   req.Nights,
		},
		{
			HotelID:  "H054",
			Name:     "Belmond Luxury Resort",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    589.95,
			Nights:   req.Nights,
		},
		// HIGHEST priced hotel overall
		{
			HotelID:  "H055",
			Name:     "Imperial Palace Premium Suite",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    650.00,
			Nights:   req.Nights,
		},
	}

	return hotels, nil
}
