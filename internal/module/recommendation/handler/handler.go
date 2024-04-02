package handler

import (
	"context"
	"recommendation-service/internal/module/recommendation/models/request"
	"recommendation-service/internal/module/recommendation/usecases"
	"recommendation-service/internal/pkg/helpers"
	"recommendation-service/internal/pkg/log"
	"strconv"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/go-playground/validator/v10"
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
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

func (h *RecommendationHandler) GetRecommendation(ctx *fiber.Ctx) error {
	userID := ctx.Locals("userID").(string)

	// convert userID to int64
	userIDInt, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return helpers.RespError(ctx, h.Log, err)
	}

	resp, err := h.Usecase.GetRecommendation(ctx.Context(), userIDInt)

	if err != nil {
		return helpers.RespError(ctx, h.Log, err)
	}

	return helpers.RespSuccess(ctx, h.Log, resp, "Success get recommendation")
}
