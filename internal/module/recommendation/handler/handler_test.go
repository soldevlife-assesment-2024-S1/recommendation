package handler_test

import (
	"context"
	"net/http/httptest"
	"recommendation-service/internal/module/recommendation/handler"
	"recommendation-service/internal/module/recommendation/mocks"
	"recommendation-service/internal/module/recommendation/models/request"
	"recommendation-service/internal/module/recommendation/models/response"
	log_internal "recommendation-service/internal/pkg/log"
	"testing"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/go-playground/validator/v10"
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

var (
	ucMock  *mocks.Usecases
	h       handler.RecommendationHandler
	vld     *validator.Validate
	p       message.Publisher
	ctx     context.Context
	app     *fiber.App
	logMock log_internal.Logger
)

type mockPublisher struct{}

// Close implements message.Publisher.
func (m *mockPublisher) Close() error {
	return nil
}

// Publish implements message.Publisher.
func (m *mockPublisher) Publish(topic string, messages ...*message.Message) error {
	return nil
}

func NewMockPublisher() message.Publisher {
	return &mockPublisher{}
}

func setup() {
	ucMock = new(mocks.Usecases)
	vld = validator.New()
	p = NewMockPublisher()
	logZap := log_internal.SetupLogger()
	log_internal.Init(logZap)
	logMock := log_internal.GetLogger()
	h = handler.RecommendationHandler{
		Usecase:   ucMock,
		Validator: vld,
		Publish:   p,
		Log:       logMock,
	}
	ctx = context.Background()
	app = fiber.New()
}

func teardown() {
	ucMock = nil
	vld = nil
	p = nil
	h = handler.RecommendationHandler{}
	ctx = nil
	app = nil
}

func TestUpdateVenueStatus(t *testing.T) {
	setup()
	defer teardown()

	t.Run("success", func(t *testing.T) {
		// mock data
		payloadMock := request.UpdateVenueStatus{
			VenueName:      "Jakarta",
			IsSoldOut:      true,
			IsFirstSoldOut: true,
		}

		jsonPayload, _ := json.Marshal(payloadMock)

		msg := message.NewMessage(watermill.NewUUID(), jsonPayload)
		ucMock.On("UpdateVenueStatus", ctx, &payloadMock).Return(nil)

		err := h.UpdateVenueStatus(msg)
		if err != nil {
			t.Error(err)
		}
	})
}

func TestUpdateTicketSoldOut(t *testing.T) {
	setup()
	defer teardown()

	t.Run("success", func(t *testing.T) {
		// mock data
		payloadMock := request.TicketSoldOut{
			VenueName: "Jakarta",
			IsSoldOut: true,
		}

		jsonPayload, _ := json.Marshal(payloadMock)

		msg := message.NewMessage(watermill.NewUUID(), jsonPayload)
		ucMock.On("UpdateTicketSoldOut", ctx, &payloadMock).Return(nil)

		err := h.UpdateTicketSoldOut(msg)
		if err != nil {
			t.Error(err)
		}
	})
}

func TestGetRecommendation(t *testing.T) {
	setup()
	defer teardown()

	t.Run("success", func(t *testing.T) {
		// mock data
		httpReq := httptest.NewRequest("GET", "/recommendation", nil)
		httpReq.Header.Set("Content-Type", "application/json")

		ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
		ctx.Request().SetRequestURI("/recommendation")
		ctx.Request().Header.SetMethod("GET")
		ctx.Request().Header.SetContentType("application/json")
		ctx.Locals("user_id", int64(1))

		// mock usecase
		ucMock.On("GetRecommendation", ctx.Context(), int64(1)).Return(nil, nil)

		// test
		err := h.GetRecommendation(ctx)

		// assert
		assert.NoError(t, err)
	})
}

func TestGetOnlineTicket(t *testing.T) {
	setup()
	defer teardown()

	t.Run("success", func(t *testing.T) {
		// mock data
		mockResponse := response.OnlineTicket{
			IsSoldOut:      false,
			IsFirstSoldOut: false,
		}

		httpReq := httptest.NewRequest("GET", "/recommendation", nil)
		httpReq.Header.Set("Content-Type", "application/json")

		ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
		ctx.Request().SetRequestURI("/recommendation")
		ctx.Request().Header.SetMethod("GET")
		ctx.Request().Header.SetContentType("application/json")
		ctx.Request().URI().QueryArgs().Add("region_name", "Jakarta")

		// mock usecase
		ucMock.On("GetOnlineTicket", ctx.Context(), "Jakarta").Return(mockResponse, nil)

		// test
		err := h.GetOnlineTicket(ctx)

		// assert
		assert.NoError(t, err)
	})
}
