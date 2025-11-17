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
		// Shared hotel IDs - some with better prices, some with worse
		{
			HotelID:  "H123", // WORSE price than Mock1 (129.90)
			Name:     "Hotel Atlas",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    135.00,
			Nights:   req.Nights,
		},
		{
			HotelID:  "H456", // BETTER price than Mock1 (89.50)
			Name:     "Riad Zitoun",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    79.99,
			Nights:   req.Nights,
		},
		{
			HotelID:  "H789", // WORSE price than Mock1 (199.00)
			Name:     "Le Meridien",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    215.00,
			Nights:   req.Nights,
		},
		{
			HotelID:  "H999", // BETTER price than Mock1 (285.00)
			Name:     "Sofitel Palais Imperial",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    250.00,
			Nights:   req.Nights,
		},
		{
			HotelID:  "H111", // BETTER price than Mock1 (95.00)
			Name:     "Dar Soukkar",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    75.00,
			Nights:   req.Nights,
		},
		// Mock2 unique hotels (Budget tier)
		{
			HotelID:  "H019",
			Name:     "Budget Hostel Medina",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    48.00,
			Nights:   req.Nights,
		},
		{
			HotelID:  "H020",
			Name:     "Traveler's Rest",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    55.50,
			Nights:   req.Nights,
		},
		{
			HotelID:  "H021",
			Name:     "Souk Guesthouse",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    63.00,
			Nights:   req.Nights,
		},
		{
			HotelID:  "H022",
			Name:     "Plaza Budget Hotel",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    71.50,
			Nights:   req.Nights,
		},
		{
			HotelID:  "H023",
			Name:     "Oasis Inn",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    85.00,
			Nights:   req.Nights,
		},
		// Mid-range tier
		{
			HotelID:  "H024",
			Name:     "Riad Marrakech",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    98.00,
			Nights:   req.Nights,
		},
		{
			HotelID:  "H025",
			Name:     "Kech Boutique",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    110.00,
			Nights:   req.Nights,
		},
		{
			HotelID:  "H026",
			Name:     "Garden Paradise Hotel",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    125.50,
			Nights:   req.Nights,
		},
		{
			HotelID:  "H027",
			Name:     "Medina Palace",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    142.00,
			Nights:   req.Nights,
		},
		{
			HotelID:  "H028",
			Name:     "Accor Hotel Downtown",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    158.99,
			Nights:   req.Nights,
		},
		{
			HotelID:  "H029",
			Name:     "Best Western Plaza",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    176.00,
			Nights:   req.Nights,
		},
		// Upscale tier
		{
			HotelID:  "H030",
			Name:     "Hyatt Regency",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    195.00,
			Nights:   req.Nights,
		},
		{
			HotelID:  "H031",
			Name:     "InterContinental",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    228.50,
			Nights:   req.Nights,
		},
		{
			HotelID:  "H032",
			Name:     "Sheraton Grand",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    265.00,
			Nights:   req.Nights,
		},
		{
			HotelID:  "H033",
			Name:     "Renaissance Hotel",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    298.00,
			Nights:   req.Nights,
		},
		// Luxury tier
		{
			HotelID:  "H034",
			Name:     "Ritz-Carlton",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    385.00,
			Nights:   req.Nights,
		},
		{
			HotelID:  "H035",
			Name:     "Mandarin Oriental",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    465.50,
			Nights:   req.Nights,
		},
		{
			HotelID:  "H036",
			Name:     "Park Hyatt",
			City:     cityCasings[rand.Intn(len(cityCasings))],
			Currency: "EUR",
			Price:    520.00,
			Nights:   req.Nights,
		},
	}

	return hotels, nil
}
