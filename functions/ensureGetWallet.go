package functions

import (
	"context"
	"errors"
	"time"

	db "github.com/rakshitg600/notakto-solo/db/generated"
)

func EnsureGetWallet(ctx context.Context, q *db.Queries, uid string) (
	coins int32,
	xp int32,
	err error,
) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	wallet, err := q.GetWalletByPlayerId(ctx, uid)
	if err != nil {
		return 0, 0, err
	}
	if wallet.Coins.Valid == false || wallet.Xp.Valid == false {
		return 0, 0, errors.New("invalid wallet response from db")
	}
	return wallet.Coins.Int32, wallet.Xp.Int32, nil
}
