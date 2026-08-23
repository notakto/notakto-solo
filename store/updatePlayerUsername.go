package store

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/rakshitg600/notakto-solo/contextkey"
	db "github.com/rakshitg600/notakto-solo/db/generated"
)

func UpdatePlayerUsername(ctx context.Context, q *db.Queries, username string) (player db.Player, err error) {
	uid, ok := contextkey.UIDFromContext(ctx)
	if !ok || uid == "" {
		return db.Player{}, errors.New("missing or invalid uid in context")
	}
	start := time.Now()
	player, err = q.UpdatePlayerUsername(ctx, db.UpdatePlayerUsernameParams{
		Uid:      uid,
		Username: username,
	})
	if time.Since(start) > 2*time.Second {
		log.Printf("Update player username took %v, err: %v", time.Since(start), err)
	}
	return player, err
}
