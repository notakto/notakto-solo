package routes

import (
	"time"

	"firebase.google.com/go/v4/auth"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"

	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/rakshitg600/notakto-solo/handlers"
	"github.com/rakshitg600/notakto-solo/middleware"
)

func SetupRoutes(e *echo.Echo, pool *pgxpool.Pool, authClient *auth.Client, valkeyClient *redis.Client) {
	// CORS must stay global — browsers send OPTIONS preflights that don't match method-specific routes
	e.Use(middleware.CORSMiddleware)

	ipRateLimit := middleware.IPRateLimitMiddleware(valkeyClient, 120)
	ctxTimeout := echoMiddleware.ContextTimeout(5 * time.Second)
	firebaseAuth := middleware.FirebaseAuthMiddleware(authClient)
	uidRateLimit := middleware.UIDRateLimitMiddleware(valkeyClient, 60)
	uidLock := middleware.UIDLockMiddleware(valkeyClient)

	handler := handlers.NewHandler(pool, authClient)

	// ── Health (no auth, no rate limit) ──
	e.HEAD("/v1/health-head", handler.HealthHeadHandler)
	e.GET("/v1/health-get", handler.HealthGetHandler)

	// ── Authenticated routes ──
	e.POST("/v1/sign-in", handler.SignInHandler, ipRateLimit, ctxTimeout, firebaseAuth, uidRateLimit, uidLock)
	e.POST("/v1/create-game", handler.CreateGameHandler, ipRateLimit, ctxTimeout, firebaseAuth, uidRateLimit, uidLock)
	e.POST("/v1/make-move", handler.MakeMoveHandler, ipRateLimit, ctxTimeout, firebaseAuth, uidRateLimit, uidLock)
	e.POST("/v1/skip-move", handler.SkipMoveHandler, ipRateLimit, ctxTimeout, firebaseAuth, uidRateLimit, uidLock)
	e.POST("/v1/undo-move", handler.UndoMoveHandler, ipRateLimit, ctxTimeout, firebaseAuth, uidRateLimit, uidLock)
	e.POST("/v1/quit-game", handler.QuitGameHandler, ipRateLimit, ctxTimeout, firebaseAuth, uidRateLimit, uidLock)
	e.GET("/v1/get-wallet", handler.GetWalletHandler, ipRateLimit, ctxTimeout, firebaseAuth, uidRateLimit, uidLock)
	e.POST("/v1/update-name", handler.UpdateNameHandler, ipRateLimit, ctxTimeout, firebaseAuth, uidRateLimit, uidLock)
}
