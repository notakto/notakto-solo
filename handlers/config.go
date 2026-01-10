package handlers

import "github.com/jackc/pgx/v5/pgxpool"

type Handler struct {
	Pool *pgxpool.Pool
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{Pool: pool}
}
