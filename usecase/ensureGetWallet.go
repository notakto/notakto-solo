package usecase

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/rakshitg600/notakto-solo/db/generated"
	"github.com/rakshitg600/notakto-solo/contextkey"
	"github.com/rakshitg600/notakto-solo/store"
)

func EnsureGetWallet(ctx context.Context, pool *pgxpool.Pool) (
	coins int32,
	xp int32,
	err error,
) {
	uid, ok := contextkey.UIDFromContext(ctx)
	if !ok || uid == "" {
		return 0, 0, errors.New("missing or invalid uid in context")
	}
	queries := db.New(pool)
	wallet, err := store.GetWalletByPlayerId(ctx, queries)
	if err != nil {
		return 0, 0, err
	}
	if wallet.Coins.Valid == false || wallet.Xp.Valid == false {
		return 0, 0, errors.New("invalid wallet response from db")
	}
	return wallet.Coins.Int32, wallet.Xp.Int32, nil
}
