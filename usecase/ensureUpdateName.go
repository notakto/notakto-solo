package usecase

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/rakshitg600/notakto-solo/db/generated"
	"github.com/rakshitg600/notakto-solo/store"
)

// EnsureUpdateName updates a player's name in the database and returns the updated name.
// It returns the updated name on success or an error if the update fails.
func EnsureUpdateName(ctx context.Context, pool *pgxpool.Pool, name string, uid string) (string, error) {
	queries := db.New(pool)
	player, err := store.UpdatePlayerName(ctx, queries, uid, name)
	if err != nil {
		return "", err
	}
	return player.Name, nil
}
