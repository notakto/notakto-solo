package store

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/rakshitg600/notakto-solo/db/generated"
)

func UpdateWalletReduceCoins(ctx context.Context, q *db.Queries, uid string, coins int32) (
	err error,
) {
	start := time.Now()
	err = q.UpdateWalletReduceCoins(ctx, db.UpdateWalletReduceCoinsParams{
		Uid:   uid,
		Coins: pgtype.Int4{Int32: coins, Valid: true},
	})
	if time.Since(start) > 2*time.Second {
		//logging slow DB calls
		log.Printf("Update wallet reduce coins took %v, err: %v", time.Since(start), err)
	}
	return err
}
