package handlers

import (
	"firebase.google.com/go/v4/auth"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rakshitg600/notakto-solo/nowpayments"
)

type Handler struct {
	Pool              *pgxpool.Pool
	AuthClient        *auth.Client
	NowpaymentsClient *nowpayments.Client
	IPNSecret         string
}

func NewHandler(pool *pgxpool.Pool, authClient *auth.Client, npClient *nowpayments.Client, ipnSecret string) *Handler {
	return &Handler{
		Pool:              pool,
		AuthClient:        authClient,
		NowpaymentsClient: npClient,
		IPNSecret:         ipnSecret,
	}
}
