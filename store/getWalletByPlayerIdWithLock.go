package store

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	db "github.com/rakshitg600/notakto-solo/db/generated"
	"github.com/rakshitg600/notakto-solo/contextkey"
)

func GetWalletByPlayerIdWithLock(ctx context.Context, q *db.Queries) (wallet db.Wallet, err error) {
	uid, ok := contextkey.UIDFromContext(ctx)
	if !ok || uid == "" {
		return db.Wallet{}, errors.New("missing or invalid uid in context")
	}
	start := time.Now()
	wallet, err = q.GetWalletByPlayerIdWithLock(ctx, uid)
	if time.Since(start) > 2*time.Second {
		//logging slow DB calls
		log.Printf("GetWalletByPlayerIdWithLock took %v, err: %v", time.Since(start), err)
	}
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Wallet{}, pgx.ErrNoRows
		}
		return db.Wallet{}, err
	}
	return wallet, nil
}
