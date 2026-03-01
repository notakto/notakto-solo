package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/rakshitg600/notakto-solo/db/generated"
	"github.com/rakshitg600/notakto-solo/store"
)

func EnsureProcessWebhook(ctx context.Context, pool *pgxpool.Pool, eventType string, chargeID string) error {
	switch eventType {
	case "charge:pending":
		return processChargePending(ctx, pool, chargeID)
	case "charge:confirmed":
		return processChargeConfirmed(ctx, pool, chargeID)
	case "charge:failed":
		return processChargeFailed(ctx, pool, chargeID)
	default:
		log.Printf("ignoring unhandled webhook event type: %s", eventType)
		return nil
	}
}

func processChargePending(ctx context.Context, pool *pgxpool.Pool, chargeID string) error {
	queries := db.New(pool)
	rowsAffected, err := store.UpdatePaymentStatusIfNotConfirmed(ctx, queries, chargeID, "pending")
	if err != nil {
		return fmt.Errorf("failed to update payment to pending: %w", err)
	}
	if rowsAffected == 0 {
		log.Printf("charge %s already confirmed or not found, skipping pending update", chargeID)
	}
	return nil
}

func processChargeConfirmed(ctx context.Context, pool *pgxpool.Pool, chargeID string) error {
	queries := db.New(pool)

	tx, err := pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.Serializable,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := queries.WithTx(tx)

	payment, err := store.GetPaymentByIdWithLock(ctx, qtx, chargeID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("payment not found for charge: %s", chargeID)
		}
		return fmt.Errorf("failed to lock payment row: %w", err)
	}

	if payment.Status == "confirmed" {
		log.Printf("charge %s already confirmed, skipping", chargeID)
		return nil
	}

	rowsAffected, err := store.UpdatePaymentStatusIfNotConfirmed(ctx, qtx, chargeID, "confirmed")
	if err != nil {
		return fmt.Errorf("failed to update payment to confirmed: %w", err)
	}
	if rowsAffected == 0 {
		log.Printf("charge %s status update returned 0 rows, already confirmed", chargeID)
		return nil
	}

	err = store.CreditWalletCoins(ctx, qtx, payment.Uid, payment.Coins)
	if err != nil {
		return fmt.Errorf("failed to credit wallet coins: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("charge %s confirmed: credited %d coins to uid %s", chargeID, payment.Coins, payment.Uid)
	return nil
}

func processChargeFailed(ctx context.Context, pool *pgxpool.Pool, chargeID string) error {
	queries := db.New(pool)
	rowsAffected, err := store.UpdatePaymentStatusIfNotConfirmed(ctx, queries, chargeID, "failed")
	if err != nil {
		return fmt.Errorf("failed to update payment to failed: %w", err)
	}
	if rowsAffected == 0 {
		log.Printf("charge %s already confirmed or not found, skipping failed update", chargeID)
	}
	return nil
}
