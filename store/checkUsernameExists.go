package store

import (
	"context"
	"log"
	"time"

	db "github.com/rakshitg600/notakto-solo/db/generated"
)

func CheckUsernameExists(ctx context.Context, q *db.Queries, username string) (bool, error) {
	start := time.Now()
	exists, err := q.CheckUsernameExists(ctx, username)
	if time.Since(start) > 2*time.Second {
		log.Printf("Check username exists took %v, err: %v", time.Since(start), err)
	}
	return exists, err
}
