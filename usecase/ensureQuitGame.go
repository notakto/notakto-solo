package usecase

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rakshitg600/notakto-solo/store"
)

// EnsureQuitGame verifies that the provided sessionID matches the player's latest session and marks that session as quit in the database.
// It returns true if the session was already marked game over or was successfully updated to quit.
// It returns false and a non-nil error when the session does not match (session expired or not found) or when a database operation fails.
func EnsureQuitGame(ctx context.Context, pool *pgxpool.Pool, uid string, sessionID string) (
	success bool,
	err error,
) {
	// STEP 1: Validate sessionId
	existing, err := store.GetLatestSessionStateByPlayerId(ctx, q, uid)
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
	err = store.QuitGameSession(ctx, q, sessionID)
	if err != nil {
		return false, err
	}
	return true, nil
}
