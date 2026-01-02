package functions

import (
	"context"
	"errors"
	"time"

	db "github.com/rakshitg600/notakto-solo/db/generated"
)

// EnsureGetWallet retrieves the wallet for the given player ID and returns its coins and XP.
// 
// It enforces a 3-second deadline for the database call. Returns an error if the query fails
// or if the wallet's coins or XP fields are NULL/invalid.
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