package usecase

import (
	"context"

	db "github.com/rakshitg600/notakto-solo/db/generated"
	"github.com/rakshitg600/notakto-solo/store"
)

// EnsureUpdateName updates a player's name in the database and returns the updated name.
// It returns the updated name on success or an error if the update fails.
func EnsureUpdateName(ctx context.Context, q *db.Queries, name string, uid string) (string, error) {
	player, err := store.UpdatePlayerName(ctx, q, uid, name)
	if err != nil {
		return "", err
	}
	return player.Name, nil
}
