package functions

import (
	"context"
	"time"

	db "github.com/rakshitg600/notakto-solo/db/generated"
)

// EnsureUpdateName updates a player's name in the database and returns the updated name.
// The update operation uses a 3-second timeout derived from the provided context.
// It returns the updated name on success or an error if the update fails.
func EnsureUpdateName(ctx context.Context, q *db.Queries, name string, uid string) (string, error) {
	updatePlayerNameCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	player, err := q.UpdatePlayerName(updatePlayerNameCtx, db.UpdatePlayerNameParams{
		Uid:  uid,
		Name: name,
	})
	if err != nil {
		return "", err
	}
	return player.Name, nil
}