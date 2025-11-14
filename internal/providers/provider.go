package providers

import (
	"context"

	"hostaggr/internal/models"
)

// Provider defines the interface that all hotel search providers must implement
type Provider interface {
	// Search performs a hotel search based on the provided request
	Search(ctx context.Context, req models.SearchRequest) ([]models.ProviderHotel, error)

	// Name returns the unique identifier/name of the provider
	Name() string
}
