package store

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/rakshitg600/notakto-solo/contextkey"
	db "github.com/rakshitg600/notakto-solo/db/generated"
)

func GetPaymentsByUid(ctx context.Context, q *db.Queries) ([]db.Payment, error) {
	uid, ok := contextkey.UIDFromContext(ctx)
	if !ok || uid == "" {
		return nil, errors.New("missing or invalid uid in context")
	}
	start := time.Now()
	payments, err := q.GetPaymentsByUid(ctx, uid)
	if time.Since(start) > 2*time.Second {
		log.Printf("GetPaymentsByUid took %v, err: %v", time.Since(start), err)
	}
	if err != nil {
		return nil, err
	}
	return payments, nil
}
