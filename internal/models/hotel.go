package models

// ProviderHotel represents raw data from a provider
type ProviderHotel struct {
	HotelID  string  `json:"hotel_id"`
	Name     string  `json:"name"`
	City     string  `json:"city"`
	Currency string  `json:"currency"`
	Price    float64 `json:"price"`
	Nights   int     `json:"nights"`
}

// Hotel represents normalized hotel data
type Hotel struct {
	HotelID  string  `json:"hotel_id"`
	Name     string  `json:"name"`
	Currency string  `json:"currency"`
	Price    float64 `json:"price"`
}
