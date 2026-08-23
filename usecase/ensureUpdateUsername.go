package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rakshitg600/notakto-solo/contextkey"
	db "github.com/rakshitg600/notakto-solo/db/generated"
	"github.com/rakshitg600/notakto-solo/store"
)

var ErrUsernameExists = errors.New("username already exists")

func EnsureUpdateUsername(ctx context.Context, pool *pgxpool.Pool, username string) (string, error) {
	uid, ok := contextkey.UIDFromContext(ctx)
	if !ok || uid == "" {
		return "", errors.New("missing or invalid uid in context")
	}
	trimmed := strings.TrimSpace(username)
	if trimmed == "" {
		return "", errors.New("username is required")
	}

	queries := db.New(pool)
	exists, err := store.CheckUsernameExists(ctx, queries, trimmed)
	if err != nil {
		return "", err
	}
	if exists {
		return "", ErrUsernameExists
	}

	player, err := store.UpdatePlayerUsername(ctx, queries, trimmed)
	if err != nil {
		return "", err
	}
	return player.Username, nil
}
