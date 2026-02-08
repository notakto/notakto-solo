package usecase

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/rakshitg600/notakto-solo/db/generated"
	"github.com/rakshitg600/notakto-solo/contextkey"
	"github.com/rakshitg600/notakto-solo/store"
)

func EnsureUpdateName(ctx context.Context, pool *pgxpool.Pool, name string) (string, error) {
	uid, ok := contextkey.UIDFromContext(ctx)
	if !ok || uid == "" {
		return "", errors.New("missing or invalid uid in context")
	}
	queries := db.New(pool)
	player, err := store.UpdatePlayerName(ctx, queries, name)
	if err != nil {
		return "", err
	}
	return player.Name, nil
}
