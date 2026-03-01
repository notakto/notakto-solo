package handlers

import (
	commerce "github.com/coinbase-samples/commerce-sdk-go"
	"firebase.google.com/go/v4/auth"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	Pool            *pgxpool.Pool
	AuthClient      *auth.Client
	CommerceClient  *commerce.Client
	WebhookSecret   string
}

func NewHandler(pool *pgxpool.Pool, authClient *auth.Client, commerceClient *commerce.Client, webhookSecret string) *Handler {
	return &Handler{
		Pool:           pool,
		AuthClient:     authClient,
		CommerceClient: commerceClient,
		WebhookSecret:  webhookSecret,
	}
}
