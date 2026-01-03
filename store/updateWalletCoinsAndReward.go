package store

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/rakshitg600/notakto-solo/db/generated"
)

func UpdateWalletCoinsAndXpReward(ctx context.Context, q *db.Queries, uid string, coinsReward int32, xpReward int32) (
	err error,
) {
	start := time.Now()
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	err = q.UpdateWalletCoinsAndXpReward(ctx, db.UpdateWalletCoinsAndXpRewardParams{
		Uid:   uid,
		Coins: pgtype.Int4{Int32: coinsReward, Valid: true},
		Xp:    pgtype.Int4{Int32: xpReward, Valid: true},
	})
	if time.Since(start) > 2*time.Second {
		//logging slow DB calls
		log.Printf("Update wallet coins and xp reward took %v, err: %v", time.Since(start), err)
	}
	return err
}
