package models

// SearchRequest represents the incoming search query
type SearchRequest struct {
	City    string
	CheckIn string
	Nights  int
	Adults  int
}

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

// SearchResponse is the final aggregated response
type SearchResponse struct {
	Search SearchInfo `json:"search"`
	Stats  Stats      `json:"stats"`
	Hotels []Hotel    `json:"hotels"`
}

// SearchInfo contains the search parameters
type SearchInfo struct {
	City    string `json:"city"`
	CheckIn string `json:"checkin"`
	Nights  int    `json:"nights"`
	Adults  int    `json:"adults"`
}

// Stats contains aggregation statistics
type Stats struct {
	ProvidersTotal     int    `json:"providers_total"`
	ProvidersSucceeded int    `json:"providers_succeeded"`
	ProvidersFailed    int    `json:"providers_failed"`
	Cache              string `json:"cache"` // "hit" or "miss"
	DurationMs         int64  `json:"duration_ms"`
}
