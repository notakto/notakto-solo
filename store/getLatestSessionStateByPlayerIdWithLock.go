package store

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	db "github.com/rakshitg600/notakto-solo/db/generated"
)

func GetLatestSessionStateByPlayerIdWithLock(ctx context.Context, q *db.Queries, uid string) (
	latestSessionState db.GetLatestSessionStateByPlayerIdWithLockRow,
	err error,
) {
	start := time.Now()
	latestSessionState, err = q.GetLatestSessionStateByPlayerIdWithLock(ctx, uid)
	if time.Since(start) > 2*time.Second {
		//logging slow DB calls
		log.Printf("GetLatestSessionStateByPlayerIdWithLock took %v, err: %v", time.Since(start), err)
	}
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.GetLatestSessionStateByPlayerIdWithLockRow{}, pgx.ErrNoRows
		}
		return db.GetLatestSessionStateByPlayerIdWithLockRow{}, err
	}
	return latestSessionState, nil
}
