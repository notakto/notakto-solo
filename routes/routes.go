package routes

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rakshitg600/notakto-solo/concurrency"
	"github.com/rakshitg600/notakto-solo/handlers"
	"github.com/rakshitg600/notakto-solo/middleware"
	"github.com/rakshitg600/notakto-solo/ratelimit"
	"github.com/rakshitg600/notakto-solo/valkey"
)

func RegisterRoutes(e *echo.Echo, h *handlers.Handler, valkeyClient *valkey.Client) {

	public := e.Group("/v1")
	public.HEAD("/v1/health-head", h.HealthHeadHandler)
	public.GET("/v1/health-get", h.HealthGetHandler)

	ipLimiter := ratelimit.NewTokenBucket(valkeyClient, 300, 10, "rl:ip:")
	uidLimiter := ratelimit.NewTokenBucket(valkeyClient, 100, 5, "rl:uid:")
	uidGuard := concurrency.NewRedisGuard(
		valkeyClient,
		5*time.Second, // lock TTL
		0,             // wait = 0 → reject mode
		"lock:uid:",
	)
	frontend := e.Group(
		"/v1",
		// 1 CORS handling
		middleware.CORSMiddleware,

		// 2 IP based (before auth)
		middleware.RateLimit(ipLimiter, middleware.IPKey),

		// 3 Auth
		middleware.FirebaseAuthMiddleware,

		// 4 UID based (after auth)
		middleware.RateLimit(uidLimiter, middleware.UIDKey),

		// 5 UID serial middleware
		middleware.UIDSerial(uidGuard),
	)

	frontend.POST("/create-game", h.CreateGameHandler)
	frontend.POST("/sign-in", h.SignInHandler)
	frontend.POST("/update-name", h.UpdateNameHandler)
	frontend.POST("/make-move", h.MakeMoveHandler)
	frontend.POST("/quit-game", h.QuitGameHandler)
	frontend.GET("/get-wallet", h.GetWalletHandler)
	frontend.POST("/skip-move", h.SkipMoveHandler)
	frontend.POST("/undo-move", h.UndoMoveHandler)
}
