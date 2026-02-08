package store

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/rakshitg600/notakto-solo/db/generated"
	"github.com/rakshitg600/notakto-solo/contextkey"
)

func UpdateWalletXpReward(ctx context.Context, q *db.Queries, xpReward int32) (err error) {
	uid, ok := contextkey.UIDFromContext(ctx)
	if !ok || uid == "" {
		return errors.New("missing or invalid uid in context")
	}
	start := time.Now()
	err = q.UpdateWalletXpReward(ctx, db.UpdateWalletXpRewardParams{
		Uid: uid,
		Xp:  pgtype.Int4{Int32: xpReward, Valid: true},
	})
	if time.Since(start) > 2*time.Second {
		//logging slow DB calls
		log.Printf("Update wallet xp reward took %v, err: %v", time.Since(start), err)
	}
	return err
}
