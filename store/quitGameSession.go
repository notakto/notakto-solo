package store

import (
	"context"
	"log"
	"time"

	db "github.com/rakshitg600/notakto-solo/db/generated"
)

func QuitGameSession(ctx context.Context, q *db.Queries, sessionID string) (
	err error,
) {
	start := time.Now()
	err = q.QuitGameSession(ctx, sessionID)
	if time.Since(start) > 2*time.Second {
		//logging slow DB calls
		log.Printf("QuitGameSession took %v, err: %v", time.Since(start), err)
	}
	return err
}
