package store

import (
	"context"
	"log"
	"time"

	db "github.com/rakshitg600/notakto-solo/db/generated"
)

func UpdateSessionState(ctx context.Context, q *db.Queries, sessionID string, boards []int32) (
	err error,
) {
	start := time.Now()
	err = q.UpdateSessionState(ctx, db.UpdateSessionStateParams{
		SessionID: sessionID,
		Boards:    boards,
	})
	if time.Since(start) > 2*time.Second {
		//logging slow DB calls
		log.Printf("Update session state took %v, err: %v", time.Since(start), err)
	}
	return err
}
