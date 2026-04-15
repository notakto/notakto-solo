package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rakshitg600/notakto-solo/config"
	"github.com/rakshitg600/notakto-solo/contextkey"
	db "github.com/rakshitg600/notakto-solo/db/generated"
	"github.com/rakshitg600/notakto-solo/nowpayments"
	"github.com/rakshitg600/notakto-solo/store"
)

func EnsureCreateCharge(ctx context.Context, pool *pgxpool.Pool, npClient *nowpayments.Client, packageID string) (
	chargeID string,
	hostedURL string,
	err error,
) {
	uid, ok := contextkey.UIDFromContext(ctx)
	if !ok || uid == "" {
		return "", "", errors.New("missing or invalid uid in context")
	}

	pkg, ok := config.CoinPackages[packageID]
	if !ok {
		return "", "", fmt.Errorf("invalid package ID: %s", packageID)
	}

	// order_id is our internal identifier echoed back in IPN callbacks, so we
	// key the Payment row on it (not on NOWPayments' invoice id). This keeps
	// lookups provider-agnostic.
	orderID := uuid.NewString()

	invoice, err := npClient.CreateInvoice(ctx, nowpayments.InvoiceRequest{
		PriceAmount:      float64(pkg.AmountCents) / 100.0,
		PriceCurrency:    "usd",
		OrderID:          orderID,
		OrderDescription: fmt.Sprintf("Notakto %d coins (%s)", pkg.Coins, pkg.ID),
	})
	if err != nil {
		return "", "", fmt.Errorf("nowpayments create invoice failed: %w", err)
	}

	queries := db.New(pool)
	err = store.CreatePayment(ctx, queries,
		orderID,
		uid,
		pkg.ID,
		pkg.Coins,
		pkg.AmountCents,
		"created",
		invoice.InvoiceURL,
	)
	if err != nil {
		return "", "", fmt.Errorf("failed to save payment record: %w", err)
	}

	return orderID, invoice.InvoiceURL, nil
}
