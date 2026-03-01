package store

import (
	"context"
	"log"
	"time"

	db "github.com/rakshitg600/notakto-solo/db/generated"
)

func CreatePayment(ctx context.Context, q *db.Queries, id string, uid string, packageID string, coins int32, amountCents int32, status string, hostedURL string) error {
	start := time.Now()
	err := q.CreatePayment(ctx, db.CreatePaymentParams{
		ID:          id,
		Uid:         uid,
		PackageID:   packageID,
		Coins:       coins,
		AmountCents: amountCents,
		Status:      status,
		HostedUrl:   hostedURL,
	})
	if time.Since(start) > 2*time.Second {
		log.Printf("CreatePayment took %v, err: %v", time.Since(start), err)
	}
	return err
}
