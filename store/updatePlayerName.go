package store

import (
	"context"
	"errors"
	"log"
	"time"

	db "github.com/rakshitg600/notakto-solo/db/generated"
	"github.com/rakshitg600/notakto-solo/contextkey"
)

func UpdatePlayerName(ctx context.Context, q *db.Queries, name string) (player db.Player, err error) {
	uid, ok := contextkey.UIDFromContext(ctx)
	if !ok || uid == "" {
		return db.Player{}, errors.New("missing or invalid uid in context")
	}
	start := time.Now()
	player, err = q.UpdatePlayerName(ctx, db.UpdatePlayerNameParams{
		Uid:  uid,
		Name: name,
	})
	if time.Since(start) > 2*time.Second {
		//logging slow DB calls
		log.Printf("Update player name took %v, err: %v", time.Since(start), err)
	}
	return player, err
}
