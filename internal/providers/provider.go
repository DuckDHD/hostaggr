package providers

import (
	"context"

	"hostaggr/internal/models"
)

type Provider interface {
	Search(ctx context.Context, req models.SearchRequest) ([]models.ProviderHotel, error)
	Name() string
}
