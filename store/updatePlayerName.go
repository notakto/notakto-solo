package store

import (
	"context"
	"log"
	"time"

	db "github.com/rakshitg600/notakto-solo/db/generated"
)

func UpdatePlayerName(ctx context.Context, q *db.Queries, uid string, name string) (
	player db.Player,
	err error,
) {
	start := time.Now()
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
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
