package store

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	db "github.com/rakshitg600/notakto-solo/db/generated"
)

func GetPaymentById(ctx context.Context, q *db.Queries, id string) (db.Payment, error) {
	start := time.Now()
	payment, err := q.GetPaymentById(ctx, id)
	if time.Since(start) > 2*time.Second {
		log.Printf("GetPaymentById took %v, err: %v", time.Since(start), err)
	}
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Payment{}, pgx.ErrNoRows
		}
		return db.Payment{}, err
	}
	return payment, nil
}

func GetPaymentByIdWithLock(ctx context.Context, q *db.Queries, id string) (db.Payment, error) {
	start := time.Now()
	payment, err := q.GetPaymentByIdWithLock(ctx, id)
	if time.Since(start) > 2*time.Second {
		log.Printf("GetPaymentByIdWithLock took %v, err: %v", time.Since(start), err)
	}
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Payment{}, pgx.ErrNoRows
		}
		return db.Payment{}, err
	}
	return payment, nil
}
