package store

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/rakshitg600/notakto-solo/db/generated"
)

func CreditWalletCoins(ctx context.Context, q *db.Queries, uid string, coins int32) error {
	start := time.Now()
	err := q.UpdateWalletCoinsAndXpReward(ctx, db.UpdateWalletCoinsAndXpRewardParams{
		Uid:   uid,
		Coins: pgtype.Int4{Int32: coins, Valid: true},
		Xp:    pgtype.Int4{Int32: 0, Valid: true},
	})
	if time.Since(start) > 2*time.Second {
		log.Printf("CreditWalletCoins took %v, err: %v", time.Since(start), err)
	}
	return err
}
