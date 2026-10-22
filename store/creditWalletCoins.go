package store

import (
	"context"
	"log"
	"time"

	db "github.com/rakshitg600/notakto-solo/db/generated"
)

func CreditWalletCoins(ctx context.Context, q *db.Queries, uid string, coins int32) error {
	start := time.Now()
	err := q.UpdateWalletCoinsAndXpReward(ctx, db.UpdateWalletCoinsAndXpRewardParams{
		Uid:   uid,
		Coins: coins,
		Xp:    0,
	})
	if time.Since(start) > 2*time.Second {
		log.Printf("CreditWalletCoins took %v, err: %v", time.Since(start), err)
	}
	return err
}
