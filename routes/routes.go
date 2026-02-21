package routes

import (
	"time"

	"firebase.google.com/go/v4/auth"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"

	"github.com/rakshitg600/notakto-solo/handlers"
	"github.com/rakshitg600/notakto-solo/middleware"
)

func SetupRoutes(e *echo.Echo, pool *pgxpool.Pool, authClient *auth.Client, valkeyClient *redis.Client) {

	ipRateLimit := middleware.IPRateLimitMiddleware(valkeyClient, 120)
	firebaseAuth := middleware.FirebaseAuthMiddleware(authClient)
	uidRateLimit := middleware.UIDRateLimitMiddleware(valkeyClient, 60)
	uidLock := middleware.UIDLockMiddleware(valkeyClient)
	healthCooldown := middleware.CooldownMiddleware(5 * time.Second)

	handler := handlers.NewHandler(pool, authClient)

	// ── Health (cooldown-protected, no auth, no external deps) ──
	e.HEAD("/v1/health-head", handler.HealthHeadHandler, healthCooldown)
	e.GET("/v1/health-get", handler.HealthGetHandler, healthCooldown)

	// ── Authenticated routes ──
	e.POST("/v1/sign-in", handler.SignInHandler, ipRateLimit, firebaseAuth, uidRateLimit, uidLock)
	e.POST("/v1/create-game", handler.CreateGameHandler, ipRateLimit, firebaseAuth, uidRateLimit, uidLock)
	e.POST("/v1/make-move", handler.MakeMoveHandler, ipRateLimit, firebaseAuth, uidRateLimit, uidLock)
	e.POST("/v1/skip-move", handler.SkipMoveHandler, ipRateLimit, firebaseAuth, uidRateLimit, uidLock)
	e.POST("/v1/undo-move", handler.UndoMoveHandler, ipRateLimit, firebaseAuth, uidRateLimit, uidLock)
	e.POST("/v1/quit-game", handler.QuitGameHandler, ipRateLimit, firebaseAuth, uidRateLimit, uidLock)
	e.GET("/v1/get-wallet", handler.GetWalletHandler, ipRateLimit, firebaseAuth, uidRateLimit, uidLock)
	e.POST("/v1/update-name", handler.UpdateNameHandler, ipRateLimit, firebaseAuth, uidRateLimit, uidLock)
}
