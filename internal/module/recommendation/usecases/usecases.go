package usecases

import (
	"context"
	"recommendation-service/internal/module/recommendation/models/request"
	"recommendation-service/internal/module/recommendation/repositories"
)

type usecases struct {
	repo repositories.Repositories
}

// UpdateVenueStatus implements Usecases.
func (u *usecases) UpdateVenueStatus(ctx context.Context, payload *request.UpdateVenueStatus) error {
	panic("unimplemented")
}

type Usecases interface {
	UpdateVenueStatus(ctx context.Context, payload *request.UpdateVenueStatus) error
}

func New(repo repositories.Repositories) Usecases {
	return &usecases{
		repo: repo,
	}
}
