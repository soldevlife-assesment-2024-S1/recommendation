package router

import (
	"recommendation-service/internal/module/recommendation/handler"
	"recommendation-service/internal/pkg/middleware"

	"github.com/gofiber/fiber/v2"
)

func Initialize(app *fiber.App, handlerTicket *handler.RecommendationHandler, m *middleware.Middleware) *fiber.App {

	// health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).SendString("OK")
	})

	Api := app.Group("/api")

	// public routes
	v1 := Api.Group("/v1")
	v1.Get("/recommendation", m.ValidateToken, handlerTicket.GetRecommendation)

	_ = Api.Group("/private")

	return app

}
