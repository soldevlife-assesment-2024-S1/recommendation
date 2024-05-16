package middleware

import (
	"errors"
	"fmt"
	"go/token"
	"recommendation-service/internal/module/recommendation/repositories"
	"recommendation-service/internal/pkg/helpers"

	"github.com/gofiber/fiber/v2"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
)

type Middleware struct {
	Log  *otelzap.Logger
	Repo repositories.Repositories
}

func (m *Middleware) ValidateToken(ctx *fiber.Ctx) error {
	// get token from header
	auth := ctx.Get("Authorization")
	if auth == "" {
		m.Log.Ctx(ctx.UserContext()).Error("error validate token")
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	// grab token (Bearer token) from header 7 is the length of "Bearer "
	token := auth[7:token.Pos(len(auth))]

	// check repostipories if token is valid
	resp, err := m.Repo.ValidateToken(ctx.Context(), token)
	if err != nil {
		m.Log.Ctx(ctx.UserContext()).Error(fmt.Sprintf("error validate token: %v", err))
		return helpers.RespError(ctx, m.Log, err)
	}

	if !resp.IsValid {
		m.Log.Ctx(ctx.UserContext()).Error("error validate token")
		return helpers.RespError(ctx, m.Log, errors.New("error validate token"))
	}

	ctx.Locals("user_id", resp.UserID)

	return ctx.Next()
}
