package usecase

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/rakshitg600/notakto-solo/db/generated"
	"github.com/rakshitg600/notakto-solo/logic"
	"github.com/rakshitg600/notakto-solo/store"
)

// EnsureUndoMove validates the session and wallet, charges the undo cost, removes the last two moves (player + AI) from the session boards, persists changes, and returns the updated boards.
//
// It checks that the provided sessionID matches the latest session for uid, verifies the game is not over, ensures at least two moves exist, deducts 100 coins from the wallet, updates the session state in the database, and returns the new boards slice.
//
// The function returns an error if the session is missing or expired, the game is already over, there are fewer than two moves to undo, the wallet has insufficient coins, or any database operation fails.
func EnsureUndoMove(ctx context.Context, q *db.Queries, uid string, sessionID string) (
	boards []int32,
	err error,
) {
	// STEP 1: Validate sessionId
	existing, err := store.GetLatestSessionStateByPlayerId(ctx, q, uid)
	if err != nil {
		return nil, err
	}
	if existing.SessionID != sessionID {
		return nil, errors.New("session expired or not found")
	}
	// STEP 2: Validate gameover
	if existing.Gameover.Valid && existing.Gameover.Bool {
		return nil, errors.New("game is already over")
	}
	// STEP 3: Verify if game is over before undoing move
	existing.Gameover = pgtype.Bool{Bool: true, Valid: true}
	for i := int32(0); i < existing.NumberOfBoards.Int32; i++ {
		if !logic.IsBoardDead(i, existing.Boards, existing.BoardSize.Int32) {
			existing.Gameover = pgtype.Bool{Bool: false, Valid: true}
			break
		}
	}
	if existing.Gameover.Valid && existing.Gameover.Bool {
		//TODO: Update session state in DB to reflect gameover
		return nil, errors.New("game is already over")
	}

	// STEP 4: Check wallet for sufficient coins
	coins, _, err := EnsureGetWallet(ctx, q, uid)
	if err != nil {
		return nil, err
	}
	if coins < 100 {
		return nil, errors.New("insufficient coins to undo move")
	}

	// STEP 5: Verify there are moves to undo
	if len(existing.Boards) < 2 {
		return nil, errors.New("no moves to undo")
	}

	// STEP 6: Deduct coins
	const undoMoveCost = 100
	err = store.UpdateWalletReduceCoins(ctx, q, uid, undoMoveCost)
	if err != nil {
		return nil, err
	}

	// STEP 7: Pop last two elements (player move + AI move)
	existing.Boards = existing.Boards[:len(existing.Boards)-2]
	// Update session state after AI move
	err = store.UpdateSessionState(ctx, q, sessionID, existing.Boards)
	if err != nil {
		return nil, err
	}
	return existing.Boards, nil
}
