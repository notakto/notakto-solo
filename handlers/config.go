package handlers

import (
	"firebase.google.com/go/v4/auth"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rakshitg600/notakto-solo/nowpayments"
	"github.com/redis/go-redis/v9"
)

type Handler struct {
	Pool              *pgxpool.Pool
	AuthClient        *auth.Client
	ValkeyClient      *redis.Client
	NowpaymentsClient *nowpayments.Client
	IPNSecret         string
}

func NewHandler(pool *pgxpool.Pool, authClient *auth.Client, valkeyClient *redis.Client, npClient *nowpayments.Client, ipnSecret string) *Handler {
	return &Handler{
		Pool:              pool,
		AuthClient:        authClient,
		ValkeyClient:      valkeyClient,
		NowpaymentsClient: npClient,
		IPNSecret:         ipnSecret,
	}
}
