package functions

import (
	"context"
	"time"

	db "github.com/rakshitg600/notakto-solo/db/generated"
)

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
