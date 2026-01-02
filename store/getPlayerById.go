package store

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	db "github.com/rakshitg600/notakto-solo/db/generated"
)

func GetPlayerById(ctx context.Context, q *db.Queries, uid string) (
	player db.Player,
	err error,
) {
	start := time.Now()
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	player, err = q.GetPlayerById(ctx, uid)
	if time.Since(start) > 2*time.Second {
		//logging slow DB calls
		log.Printf("GetPlayerById took %v, err: %v", time.Since(start), err)
	}
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.Player{}, sql.ErrNoRows
		}
		return db.Player{}, err
	}
	return player, nil
}
