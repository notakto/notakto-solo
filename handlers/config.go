package handlers

import (
	"context"
	"time"

	db "github.com/rakshitg600/notakto-solo/db/generated"
)

type Handler struct {
	Queries *db.Queries
}

func NewHandler(q *db.Queries) *Handler {
	return &Handler{Queries: q}
}

const defaultDBTimeout = 3 * time.Second

func WithDBTimeout(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, defaultDBTimeout)
}
