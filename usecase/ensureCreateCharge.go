package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/coinbase-samples/commerce-sdk-go"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rakshitg600/notakto-solo/config"
	"github.com/rakshitg600/notakto-solo/contextkey"
	db "github.com/rakshitg600/notakto-solo/db/generated"
	"github.com/rakshitg600/notakto-solo/store"
)

func EnsureCreateCharge(ctx context.Context, pool *pgxpool.Pool, commerceClient *commerce.Client, packageID string) (
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

	chargeResp, err := commerceClient.CreateCharge(ctx, &commerce.ChargeRequest{
		PricingType: "fixed_price",
		LocalPrice: &commerce.LocalPrice{
			Amount:   pkg.PriceUSD,
			Currency: "USD",
		},
		Metadata: &map[string]interface{}{
			"uid":        uid,
			"package_id": packageID,
		},
	})
	if err != nil {
		return "", "", fmt.Errorf("coinbase create charge failed: %w", err)
	}

	queries := db.New(pool)
	err = store.CreatePayment(ctx, queries,
		chargeResp.Data.Id,
		uid,
		pkg.ID,
		pkg.Coins,
		pkg.AmountCents,
		"created",
		chargeResp.Data.HostedUrl,
	)
	if err != nil {
		return "", "", fmt.Errorf("failed to save payment record: %w", err)
	}

	return chargeResp.Data.Id, chargeResp.Data.HostedUrl, nil
}
