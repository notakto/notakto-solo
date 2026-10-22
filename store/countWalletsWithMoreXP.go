package store

import (
	"context"
	"log"
	"time"

	db "github.com/rakshitg600/notakto-solo/db/generated"
)

func CountWalletsWithMoreXP(ctx context.Context, q *db.Queries, xp int32) (int64, error) {
	start := time.Now()
	count, err := q.CountWalletsWithMoreXP(ctx, xp)
	if time.Since(start) > 2*time.Second {
		log.Printf("CountWalletsWithMoreXP took %v, err: %v", time.Since(start), err)
	}
	if err != nil {
		return 0, err
	}
	return count, nil
}
