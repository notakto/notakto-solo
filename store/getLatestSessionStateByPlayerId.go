package store

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	db "github.com/rakshitg600/notakto-solo/db/generated"
)

func GetLatestSessionStateByPlayerId(ctx context.Context, q *db.Queries, uid string) (
	latestSessionState db.GetLatestSessionStateByPlayerIdRow,
	err error,
) {
	start := time.Now()
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	latestSessionState, err = q.GetLatestSessionStateByPlayerId(ctx, uid)
	if time.Since(start) > 2*time.Second {
		//logging slow DB calls
		log.Printf("GetLatestSessionStateByPlayerId took %v, err: %v", time.Since(start), err)
	}
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.GetLatestSessionStateByPlayerIdRow{}, pgx.ErrNoRows
		}
		return db.GetLatestSessionStateByPlayerIdRow{}, err
	}
	return latestSessionState, nil
}
