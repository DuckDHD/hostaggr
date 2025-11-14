package models

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
