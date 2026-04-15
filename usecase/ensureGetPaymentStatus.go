package usecase

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rakshitg600/notakto-solo/contextkey"
	db "github.com/rakshitg600/notakto-solo/db/generated"
	"github.com/rakshitg600/notakto-solo/store"
)

var ErrPaymentForbidden = errors.New("payment access denied")

func EnsureGetPaymentStatus(ctx context.Context, pool *pgxpool.Pool, chargeID string) (db.Payment, error) {
	uid, ok := contextkey.UIDFromContext(ctx)
	if !ok || uid == "" {
		return db.Payment{}, errors.New("missing or invalid uid in context")
	}

	queries := db.New(pool)
	payment, err := store.GetPaymentById(ctx, queries, chargeID)
	if err != nil {
		return db.Payment{}, err
	}

	if payment.Uid != uid {
		return db.Payment{}, ErrPaymentForbidden
	}

	return payment, nil
}
