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

// EnsureProcessWebhook maps a NOWPayments IPN payment_status to an internal
// Payment row status and credits coins when the payment is finalized.
//
// Status semantics (from NOWPayments docs):
//   - waiting / confirming / confirmed / sending → payment in-flight, mark pending
//   - finished → funds settled in our wallet, credit coins
//   - failed / refunded / expired → terminal failure
//   - partially_paid → user underpaid; terminal, logged for manual review
func EnsureProcessWebhook(ctx context.Context, pool *pgxpool.Pool, paymentStatus string, orderID string) error {
	switch paymentStatus {
	case "waiting", "confirming", "confirmed", "sending":
		return processPaymentPending(ctx, pool, orderID)
	case "finished":
		return processPaymentFinished(ctx, pool, orderID)
	case "failed", "refunded", "expired":
		return processPaymentFailed(ctx, pool, orderID, paymentStatus)
	case "partially_paid":
		log.Printf("webhook: order %s partially_paid — marking failed, needs manual review", orderID)
		return processPaymentFailed(ctx, pool, orderID, paymentStatus)
	default:
		log.Printf("ignoring unhandled nowpayments payment_status: %s", paymentStatus)
		return nil
	}
}

func processPaymentPending(ctx context.Context, pool *pgxpool.Pool, orderID string) error {
	queries := db.New(pool)
	rowsAffected, err := store.UpdatePaymentStatusIfNotConfirmed(ctx, queries, orderID, "pending")
	if err != nil {
		return fmt.Errorf("failed to update payment to pending: %w", err)
	}
	if rowsAffected == 0 {
		log.Printf("order %s already confirmed or not found, skipping pending update", orderID)
	}
	return nil
}

func processPaymentFinished(ctx context.Context, pool *pgxpool.Pool, orderID string) error {
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

	payment, err := store.GetPaymentByIdWithLock(ctx, qtx, orderID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("payment not found for order: %s", orderID)
		}
		return fmt.Errorf("failed to lock payment row: %w", err)
	}

	if payment.Status == "confirmed" {
		log.Printf("order %s already confirmed, skipping", orderID)
		return nil
	}

	rowsAffected, err := store.UpdatePaymentStatusIfNotConfirmed(ctx, qtx, orderID, "confirmed")
	if err != nil {
		return fmt.Errorf("failed to update payment to confirmed: %w", err)
	}
	if rowsAffected == 0 {
		log.Printf("order %s status update returned 0 rows, already confirmed", orderID)
		return nil
	}

	err = store.CreditWalletCoins(ctx, qtx, payment.Uid, payment.Coins)
	if err != nil {
		return fmt.Errorf("failed to credit wallet coins: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("order %s finished: credited %d coins to uid %s", orderID, payment.Coins, payment.Uid)
	return nil
}

func processPaymentFailed(ctx context.Context, pool *pgxpool.Pool, orderID string, reason string) error {
	queries := db.New(pool)
	rowsAffected, err := store.UpdatePaymentStatusIfNotConfirmed(ctx, queries, orderID, "failed")
	if err != nil {
		return fmt.Errorf("failed to update payment to failed (%s): %w", reason, err)
	}
	if rowsAffected == 0 {
		log.Printf("order %s already confirmed or not found, skipping failed update (reason: %s)", orderID, reason)
	}
	return nil
}
