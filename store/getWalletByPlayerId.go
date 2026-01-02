package store

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	db "github.com/rakshitg600/notakto-solo/db/generated"
)

func GetWalletByPlayerId(ctx context.Context, q *db.Queries, uid string) (
	wallet db.Wallet,
	err error,
) {
	start := time.Now()
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	wallet, err = q.GetWalletByPlayerId(ctx, uid)
	if time.Since(start) > 2*time.Second {
		//logging slow DB calls
		log.Printf("GetWalletByPlayerId took %v, err: %v", time.Since(start), err)
	}
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.Wallet{}, sql.ErrNoRows
		}
		return db.Wallet{}, err
	}
	return wallet, nil
}
