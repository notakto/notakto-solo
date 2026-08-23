package usecase

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rakshitg600/notakto-solo/contextkey"
	db "github.com/rakshitg600/notakto-solo/db/generated"
	"github.com/rakshitg600/notakto-solo/logic"
	"github.com/rakshitg600/notakto-solo/store"
)

var ErrUsernameExists = errors.New("username already exists")

func EnsureUpdateUsername(ctx context.Context, pool *pgxpool.Pool, username string) (string, int, error) {
	uid, ok := contextkey.UIDFromContext(ctx)
	if !ok || uid == "" {
		return "", http.StatusUnauthorized, errors.New("missing or invalid uid in context")
	}
	trimmed := strings.TrimSpace(username)
	if trimmed == "" {
		return "", http.StatusBadRequest, errors.New("username is required")
	}

	if err := logic.ValidateUsername(trimmed); err != nil {
		return "", http.StatusBadRequest, err
	}

	queries := db.New(pool)
	exists, err := store.CheckUsernameExists(ctx, queries, trimmed)
	if err != nil {
		return "", http.StatusInternalServerError, err
	}
	if exists {
		return "", http.StatusConflict, ErrUsernameExists
	}

	player, err := store.UpdatePlayerUsername(ctx, queries, trimmed)
	if err != nil {
		return "", http.StatusInternalServerError, err
	}
	return player.Username, http.StatusOK, nil
}
