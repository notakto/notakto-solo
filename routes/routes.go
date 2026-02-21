package routes

import (
	"firebase.google.com/go/v4/auth"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/rakshitg600/notakto-solo/handlers"
	"github.com/rakshitg600/notakto-solo/middleware"
	"github.com/redis/go-redis/v9"
)

func SetupRoutes(e *echo.Echo, pool *pgxpool.Pool, authClient *auth.Client, valkeyClient *redis.Client) {
	handler := handlers.NewHandler(pool, authClient)
	RegisterRoutes(e, handler, authClient, valkeyClient)
}

func RegisterRoutes(e *echo.Echo, h *handlers.Handler, authClient *auth.Client, valkeyClient *redis.Client) {
	firebaseAuth := middleware.FirebaseAuthMiddleware(authClient)
	uidLock := middleware.UIDLockMiddleware(valkeyClient)

	e.POST("/v1/create-game", h.CreateGameHandler, firebaseAuth, uidLock)
	e.POST("/v1/sign-in", h.SignInHandler, firebaseAuth, uidLock)
	e.POST("/v1/update-name", h.UpdateNameHandler, firebaseAuth, uidLock)
	e.HEAD("/v1/health-head", h.HealthHeadHandler)
	e.GET("/v1/health-get", h.HealthGetHandler)
	e.POST("/v1/make-move", h.MakeMoveHandler, firebaseAuth, uidLock)
	e.POST("/v1/quit-game", h.QuitGameHandler, firebaseAuth, uidLock)
	e.GET("/v1/get-wallet", h.GetWalletHandler, firebaseAuth, uidLock)
	e.POST("/v1/skip-move", h.SkipMoveHandler, firebaseAuth, uidLock)
	e.POST("/v1/undo-move", h.UndoMoveHandler, firebaseAuth, uidLock)
}
