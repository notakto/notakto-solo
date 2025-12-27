package functions

import (
	"context"
	"errors"

	db "github.com/rakshitg600/notakto-solo/db/generated"
)

func EnsureQuitGame(ctx context.Context, q *db.Queries, uid string, sessionID string) (
	success bool,
	err error,
) {
	// STEP 1: Validate sessionId
	existing, err := q.GetLatestSessionStateByPlayerId(ctx, uid)
	if err != nil {
		return false, err
	}
	if existing.SessionID != sessionID {
		return false, errors.New("session expired or not found")
	}
	// STEP 2: Validate gameover
	if existing.Gameover.Valid && existing.Gameover.Bool {
		return true, nil
	}
	// STEP 3: Update gameover to true
	err = q.QuitGameSession(ctx, sessionID)
	if err != nil {
		return false, err
	}
	return true, nil
}
