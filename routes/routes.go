package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/rakshitg600/notakto-solo/handlers"
	"github.com/rakshitg600/notakto-solo/middleware"
)

func RegisterRoutes(e *echo.Echo, h *handlers.Handler) {
	e.POST("/v1/create-game", h.CreateGameHandler, middleware.FirebaseAuthMiddleware)
	e.POST("/v1/sign-in", h.SignInHandler, middleware.FirebaseAuthMiddleware)
	e.POST("/v1/update-name", h.UpdateNameHandler, middleware.FirebaseAuthMiddleware)
	e.HEAD("/v1/health-head", h.HealthHeadHandler)
	e.GET("/v1/health-get", h.HealthGetHandler)
	e.POST("/v1/make-move", h.MakeMoveHandler, middleware.FirebaseAuthMiddleware)
}
