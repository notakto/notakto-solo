package store

import (
	"context"
	"log"
	"time"

	db "github.com/rakshitg600/notakto-solo/db/generated"
)

func UpdatePaymentStatusIfNotConfirmed(ctx context.Context, q *db.Queries, id string, status string) (int64, error) {
	start := time.Now()
	rowsAffected, err := q.UpdatePaymentStatusIfNotConfirmed(ctx, db.UpdatePaymentStatusIfNotConfirmedParams{
		ID:     id,
		Status: status,
	})
	if time.Since(start) > 2*time.Second {
		log.Printf("UpdatePaymentStatusIfNotConfirmed took %v, err: %v", time.Since(start), err)
	}
	return rowsAffected, err
}
