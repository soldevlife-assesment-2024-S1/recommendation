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
	repo                  repositories.Repositories
	discountedTicketRules zen.Decision
}

// GetOnlineTicket implements Usecases.
func (u *usecases) GetOnlineTicket(ctx context.Context, regionName string) (response.OnlineTicket, error) {
	var response response.OnlineTicket

	venue, err := u.repo.FindVenueByName(ctx, regionName)
	if err != nil {
		return response, err
	}

	if venue.ID == 0 {
		venue.IsSoldOut = false
		venue.IsFirstSoldOut = false
	}

	response.IsSoldOut = venue.IsSoldOut
	response.IsFirstSoldOut = venue.IsFirstSoldOut

	return response, nil
}

// UpdateTicketSoldOut implements Usecases.
func (u *usecases) UpdateTicketSoldOut(ctx context.Context, payload *request.TicketSoldOut) error {

	venue, err := u.repo.FindVenueByName(ctx, payload.VenueName)
	if err != nil {
		return err
	}

	venue.IsSoldOut = payload.IsSoldOut

	venues, err := u.repo.FindVenues(ctx)
	if err != nil {
		return err
	}

	for _, v := range venues {
		// check if all venue is sold out first
		if v.IsFirstSoldOut {
			venue.IsFirstSoldOut = false
			break
		} else {
			venue.IsFirstSoldOut = true
		}
	}

	err = u.repo.UpsertVenue(ctx, venue)
	if err != nil {
		return err
	}

	return nil
}

// GetRecommendation implements Usecases.
func (u *usecases) GetRecommendation(ctx context.Context, userID int64) ([]response.Recomendation, error) {

	userProfile, err := u.repo.FindUserProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	venues, err := u.repo.FindVenueByName(ctx, userProfile.Region)
	if err != nil {
		return nil, err
	}

	if venues.ID == 0 {
		venues.IsSoldOut = false
		venues.IsFirstSoldOut = false
	}

	tickets, err := u.repo.FindTicketByRegionName(ctx, userProfile.Region)
	if err != nil {
		return nil, err
	}

	var responses []response.Recomendation

	for _, ticket := range tickets {

		result, err := u.discountedTicketRules.Evaluate(
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

		result.Result.UnmarshalJSON(byteRes)

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
	UpdateTicketSoldOut(ctx context.Context, payload *request.TicketSoldOut) error
	GetRecommendation(ctx context.Context, userID int64) ([]response.Recomendation, error)
	GetOnlineTicket(ctx context.Context, regionName string) (response.OnlineTicket, error)
}

func New(repo repositories.Repositories, discountedTicketRules zen.Decision) Usecases {
	return &usecases{
		repo:                  repo,
		discountedTicketRules: discountedTicketRules,
	}
}
