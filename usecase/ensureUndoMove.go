package usecase

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/rakshitg600/notakto-solo/db/generated"
	"github.com/rakshitg600/notakto-solo/logic"
	"github.com/rakshitg600/notakto-solo/store"
)

// EnsureUndoMove validates the session and wallet, charges the undo cost, removes the last two moves (player + AI) from the session boards, persists changes, and returns the updated boards.
//
// It checks that the provided sessionID matches the latest session for uid, verifies the game is not over, ensures at least two moves exist, deducts 100 coins from the wallet, updates the session state in the database, and returns the new boards slice.
//
// The function returns an error if the session is missing or expired, the game is already over, there are fewer than two moves to undo, the wallet has insufficient coins, or any database operation fails.
func EnsureUndoMove(
	ctx context.Context,
	pool *pgxpool.Pool,
	uid string,
	sessionID string,
) (
	boards []int32,
	err error,
) {
	queries := db.New(pool)

	tx, err := pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.Serializable,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	qtx := queries.WithTx(tx)

	// STEP 1: Validate session
	existing, err := store.GetLatestSessionStateByPlayerIdWithLock(ctx, qtx, uid)
	if err != nil {
		return nil, err
	}
	if existing.SessionID != sessionID {
		return nil, errors.New("session expired or not found")
	}

	// STEP 2: Validate gameover flag
	if existing.Gameover.Valid && existing.Gameover.Bool {
		return nil, errors.New("game is already over")
	}

	// STEP 3: Re-evaluate gameover from boards
	existing.Gameover = pgtype.Bool{Bool: true, Valid: true}
	for i := int32(0); i < existing.NumberOfBoards.Int32; i++ {
		if !logic.IsBoardDead(i, existing.Boards, existing.BoardSize.Int32) {
			existing.Gameover = pgtype.Bool{Bool: false, Valid: true}
			break
		}
	}
	if existing.Gameover.Bool {
		return nil, errors.New("game is already over")
	}

	// STEP 4: Validate wallet
	const undoMoveCost = 100
	wallet, err := store.GetWalletByPlayerIdWithLock(ctx, qtx, uid)
	if err != nil {
		return nil, err
	}
	if !wallet.Coins.Valid {
		return nil, errors.New("invalid wallet response from db")
	}
	if wallet.Coins.Int32 < undoMoveCost {
		return nil, errors.New("insufficient coins to undo move")
	}

	// STEP 5: Validate moves exist
	if len(existing.Boards) < 2 {
		return nil, errors.New("no moves to undo")
	}

	// STEP 6: Deduct coins
	if err = store.UpdateWalletReduceCoins(ctx, qtx, uid, undoMoveCost); err != nil {
		return nil, err
	}

	// STEP 7: Undo last two moves
	existing.Boards = existing.Boards[:len(existing.Boards)-2]

	// STEP 8: Persist session state
	if err = store.UpdateSessionState(ctx, qtx, sessionID, existing.Boards); err != nil {
		return nil, err
	}

	// STEP 9: Commit
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	return existing.Boards, nil
}
