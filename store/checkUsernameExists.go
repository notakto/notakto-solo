package store

import (
	"context"
	"log"
	"time"

	"github.com/rakshitg600/notakto-solo/contextkey"
	db "github.com/rakshitg600/notakto-solo/db/generated"
)

func CheckUsernameExists(ctx context.Context, q *db.Queries, username string) (bool, error) {
	uid, _ := contextkey.UIDFromContext(ctx)
	start := time.Now()
	exists, err := q.CheckUsernameExists(ctx, db.CheckUsernameExistsParams{
		Username: username,
		Uid:      uid,
	})
	if time.Since(start) > 2*time.Second {
		log.Printf("Check username exists took %v, err: %v", time.Since(start), err)
	}
	return exists, err
}
