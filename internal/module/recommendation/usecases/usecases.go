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
	venue, err := u.repo.FindVenueByName(ctx, payload.VenueName)
	if err != nil {
		return err
	}

	venue.IsSoldOut = payload.IsSoldOut
	venue.IsFirstSoldOut = payload.IsFirstSoldOut

	err = u.repo.UpsertVenue(ctx, venue)
	if err != nil {
		return err
	}

	return nil
}

type Usecases interface {
	UpdateVenueStatus(ctx context.Context, payload *request.UpdateVenueStatus) error
}

func New(repo repositories.Repositories) Usecases {
	return &usecases{
		repo: repo,
	}
}
