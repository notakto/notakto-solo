package handlers

import (
	"firebase.google.com/go/v4/auth"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	Pool       *pgxpool.Pool
	AuthClient *auth.Client
}

func NewHandler(pool *pgxpool.Pool, authClient *auth.Client) *Handler {
	return &Handler{Pool: pool, AuthClient: authClient}
}
