package store

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	db "github.com/rakshitg600/notakto-solo/db/generated"
)

func GetPlayerById(ctx context.Context, q *db.Queries, uid string) (
	player db.Player,
	err error,
) {
	start := time.Now()
	player, err = q.GetPlayerById(ctx, uid)
	if time.Since(start) > 2*time.Second {
		//logging slow DB calls
		log.Printf("GetPlayerById took %v, err: %v", time.Since(start), err)
	}
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Player{}, pgx.ErrNoRows
		}
		return db.Player{}, err
	}
	return player, nil
}
