package store

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rakshitg600/notakto-solo/config"
	db "github.com/rakshitg600/notakto-solo/db/generated"
)

func CreateWallet(ctx context.Context, q *db.Queries, uid string) (
	err error,
) {
	start := time.Now()
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	err = q.CreateWallet(ctx, db.CreateWalletParams{
		Uid: uid,
		Coins: pgtype.Int4{
			Int32: config.Wallet.InitialCoins,
			Valid: true,
		},
		Xp: pgtype.Int4{
			Int32: config.Wallet.InitialXP,
			Valid: true,
		},
	})
	if time.Since(start) > 2*time.Second {
		//logging slow DB calls
		log.Printf("CreateWallet took %v, err: %v", time.Since(start), err)
	}
	return err
}
