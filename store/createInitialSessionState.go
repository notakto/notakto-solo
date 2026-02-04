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
	err = q.CreateInitialSessionState(ctx, db.CreateInitialSessionStateParams{
		SessionID: newSessionID,
		Boards:    []int32{},  // empty initial boards
		IsAiMove:  []bool{},   // empty initial is_ai_move
	})
	if time.Since(start) > 2*time.Second {
		//logging slow DB calls
		log.Printf("CreateInitialSessionState took %v, err: %v", time.Since(start), err)
	}
	return err
}
