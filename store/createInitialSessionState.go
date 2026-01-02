package store

import (
	"context"
	"log"
	"time"

	db "github.com/rakshitg600/notakto-solo/db/generated"
)

func CreateInitialSessionState(
	ctx context.Context,
	q *db.Queries,
	newSessionID string) (
	err error,
) {
	start := time.Now()
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	err = q.CreateInitialSessionState(ctx, db.CreateInitialSessionStateParams{
		SessionID: newSessionID,
		Boards:    []int32{}, // empty initial boards
	})
	if time.Since(start) > 2*time.Second {
		//logging slow DB calls
		log.Printf("CreateInitialSessionState took %v, err: %v", time.Since(start), err)
	}
	return err
}
