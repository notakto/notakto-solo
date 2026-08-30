package store

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/rakshitg600/notakto-solo/contextkey"
	db "github.com/rakshitg600/notakto-solo/db/generated"
)

func GetPlayerByIdWithLock(ctx context.Context, q *db.Queries) (player db.Player, err error) {
	uid, ok := contextkey.UIDFromContext(ctx)
	if !ok || uid == "" {
		return db.Player{}, errors.New("missing or invalid uid in context")
	}

	start := time.Now()
	player, err = q.GetPlayerByIdWithLock(ctx, uid)
	if time.Since(start) > 2*time.Second {
		log.Printf("GetPlayerByIdWithLock took %v, err: %v", time.Since(start), err)
	}
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Player{}, pgx.ErrNoRows
		}
		return db.Player{}, err
	}
	return player, nil
}
