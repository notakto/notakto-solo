package functions

import (
	"context"
	"time"

	db "github.com/rakshitg600/notakto-solo/db/generated"
)

func EnsureUpdateName(ctx context.Context, q *db.Queries, name string, uid string) (string, error) {
	updateNameCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	player, err := q.UpdatePlayerName(updateNameCtx, db.UpdatePlayerNameParams{
		Uid:  uid,
		Name: name,
	})
	if err != nil {
		return "", err
	}
	return player.Name, nil
}
