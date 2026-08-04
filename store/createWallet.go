package store

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rakshitg600/notakto-solo/config"
	"github.com/rakshitg600/notakto-solo/contextkey"
	db "github.com/rakshitg600/notakto-solo/db/generated"
)

func CreateWallet(ctx context.Context, q *db.Queries, signUp config.SignUpConfig) (err error) {
	uid, ok := contextkey.UIDFromContext(ctx)
	if !ok || uid == "" {
		return errors.New("missing or invalid uid in context")
	}
	start := time.Now()
	err = q.CreateWallet(ctx, db.CreateWalletParams{
		Uid: uid,
		Coins: pgtype.Int4{
			Int32: signUp.InitialCoins,
			Valid: true,
		},
		Xp: pgtype.Int4{
			Int32: signUp.InitialXP,
			Valid: true,
		},
	})
	if time.Since(start) > 2*time.Second {
		//logging slow DB calls
		log.Printf("CreateWallet took %v, err: %v", time.Since(start), err)
	}
	return err
}
