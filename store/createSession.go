package store

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/rakshitg600/notakto-solo/db/generated"
)

func CreateSession(
	ctx context.Context,
	q *db.Queries,
	uid string,
	boardSize int32,
	numberOfBoards int32,
	difficulty int32,
	newSessionID string) (
	err error,
) {
	start := time.Now()
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	err = q.CreateSession(ctx, db.CreateSessionParams{
		SessionID:      newSessionID,
		Uid:            uid,
		BoardSize:      pgtype.Int4{Int32: boardSize, Valid: true},
		NumberOfBoards: pgtype.Int4{Int32: numberOfBoards, Valid: true},
		Difficulty:     pgtype.Int4{Int32: difficulty, Valid: true},
	})
	if time.Since(start) > 2*time.Second {
		//logging slow DB calls
		log.Printf("CreateSession took %v, err: %v", time.Since(start), err)
	}
	return err
}
