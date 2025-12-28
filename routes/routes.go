package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/rakshitg600/notakto-solo/handlers"
	"github.com/rakshitg600/notakto-solo/middleware"
	"github.com/rakshitg600/notakto-solo/types"
	"github.com/rakshitg600/notakto-solo/valkey"
)

func RegisterRoutes(e *echo.Echo, h *handlers.Handler, valkeyClient *valkey.Client, rateLimiterCfg types.RateLimiterConfig) {

	public := e.Group("/v1")
	public.HEAD("/v1/health-head", h.HealthHeadHandler)
	public.GET("/v1/health-get", h.HealthGetHandler)

	frontend := e.Group("/v1", middleware.CORSMiddleware, middleware.RateLimit(valkeyClient, rateLimiterCfg), middleware.FirebaseAuthMiddleware)

	frontend.POST("/v1/create-game", h.CreateGameHandler)
	frontend.POST("/v1/sign-in", h.SignInHandler)
	frontend.POST("/v1/update-name", h.UpdateNameHandler)
	frontend.POST("/v1/make-move", h.MakeMoveHandler)
	frontend.POST("/v1/quit-game", h.QuitGameHandler)
	frontend.GET("/v1/get-wallet", h.GetWalletHandler)
	frontend.POST("/v1/skip-move", h.SkipMoveHandler)
	frontend.POST("/v1/undo-move", h.UndoMoveHandler)
}
