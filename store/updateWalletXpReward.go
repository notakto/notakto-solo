package store

import (
	"context"
	"database/sql"
	"log"
	"time"

	db "github.com/rakshitg600/notakto-solo/db/generated"
)

func UpdateWalletXpReward(ctx context.Context, q *db.Queries, uid string, xpReward int32) (
	err error,
) {
	start := time.Now()
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	err = q.UpdateWalletXpReward(ctx, db.UpdateWalletXpRewardParams{
		Uid: uid,
		Xp:  sql.NullInt32{Int32: xpReward, Valid: true},
	})
	if time.Since(start) > 2*time.Second {
		//logging slow DB calls
		log.Printf("Update wallet xp reward took %v, err: %v", time.Since(start), err)
	}
	return err
}
