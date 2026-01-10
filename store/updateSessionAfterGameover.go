package store

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/rakshitg600/notakto-solo/db/generated"
)

func UpdateSessionAfterGameover(ctx context.Context, q *db.Queries, sessionID string, winner pgtype.Bool) (
	err error,
) {
	start := time.Now()
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	err = q.UpdateSessionAfterGameover(ctx, db.UpdateSessionAfterGameoverParams{
		SessionID: sessionID,
		Winner:    winner,
	})
	if time.Since(start) > 2*time.Second {
		//logging slow DB calls
		log.Printf("Update session took %v, err: %v", time.Since(start), err)
	}
	return err
}
