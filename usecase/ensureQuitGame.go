package usecase

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/rakshitg600/notakto-solo/db/generated"
	"github.com/rakshitg600/notakto-solo/contextkey"
	"github.com/rakshitg600/notakto-solo/store"
)

func EnsureQuitGame(ctx context.Context, pool *pgxpool.Pool, sessionID string) (
	success bool,
	err error,
) {
	uid, ok := contextkey.UIDFromContext(ctx)
	if !ok || uid == "" {
		return false, errors.New("missing or invalid uid in context")
	}
	queries := db.New(pool)
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.Serializable,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		return false, err
	}
	defer tx.Rollback(ctx)

	qtx := queries.WithTx(tx)
	// STEP 1: Validate sessionId
	existing, err := store.GetLatestSessionStateByPlayerIdWithLock(ctx, qtx)
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
	err = store.QuitGameSession(ctx, qtx, sessionID)
	if err != nil {
		return false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return false, err
	}
	return true, nil
}
