package usecases

import (
	"context"
	"encoding/json"
	"recommendation-service/internal/module/recommendation/models/request"
	"recommendation-service/internal/module/recommendation/models/response"
	"recommendation-service/internal/module/recommendation/repositories"

	"github.com/gorules/zen-go"
)

type usecases struct {
	repo    repositories.Repositories
	gorules zen.Decision
}

// GetRecommendation implements Usecases.
func (u *usecases) GetRecommendation(ctx context.Context, userID int64) ([]response.Recomendation, error) {
	// TODO: find user profile

	userProfile, err := u.repo.FindUserProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	// TODO: find venue by region name

	venues, err := u.repo.FindVenueByName(ctx, userProfile.Region)
	if err != nil {
		return nil, err
	}

	// TODO: find ticket by region name

	tickets, err := u.repo.FindTicketByRegionName(ctx, userProfile.Region)
	if err != nil {
		return nil, err
	}

	// TODO: calculate recommendation by bre

	var responses []response.Recomendation

	for _, ticket := range tickets {

		result, err := u.gorules.Evaluate(
			map[string]any{
				"price":              ticket.Price,
				"region":             userProfile.Region,
				"is_ticket_sold_out": venues.IsSoldOut,
			})

		if err != nil {
			return nil, err
		}

		// decode json response
		var responseBre response.BreTicketDiscounted

		byteRes, err := result.Result.MarshalJSON()
		if err != nil {
			return nil, err
		}

		err = result.Result.UnmarshalJSON(byteRes)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(byteRes, &responseBre)
		if err != nil {
			return nil, err
		}

		responses = append(responses, response.Recomendation{
			RegionName:         userProfile.Region,
			Level:              ticket.Level,
			DiscountPrice:      responseBre.DiscountedPrice,
			OriginalPrice:      ticket.Price,
			DiscountPercentage: float64(responseBre.DiscountAmount) / float64(ticket.Price) * 100,
		})

	}

	// TODO: return recommendation response
	return responses, nil
}

// UpdateVenueStatus implements Usecases
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
	GetRecommendation(ctx context.Context, userID int64) ([]response.Recomendation, error)
}

func New(repo repositories.Repositories, gorules zen.Decision) Usecases {
	return &usecases{
		repo:    repo,
		gorules: gorules,
	}
}
