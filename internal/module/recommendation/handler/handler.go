package handler

import (
	"context"
	"recommendation-service/internal/module/recommendation/models/request"
	"recommendation-service/internal/module/recommendation/usecases"
	"recommendation-service/internal/pkg/log"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/go-playground/validator/v10"
	"github.com/goccy/go-json"
)

type RecommendationHandler struct {
	Log       log.Logger
	Validator *validator.Validate
	Usecase   usecases.Usecases
	Publish   message.Publisher
}

func (h *RecommendationHandler) UpdateVenueStatus(msg *message.Message) error {

	msg.Ack()

	req := new(request.UpdateVenueStatus)

	if err := json.Unmarshal(msg.Payload, req); err != nil {
		return err
	}

	if err := h.Validator.Struct(req); err != nil {
		return err
	}
	ctx := context.Background()

	if err := h.Usecase.UpdateVenueStatus(ctx, req); err != nil {
		return err
	}

	return nil
}
