package handlers

import (
	db "github.com/rakshitg600/notakto-solo/db/generated"
)

type Handler struct {
	Queries *db.Queries
}

func NewHandler(q *db.Queries) *Handler {
	return &Handler{Queries: q}
}
