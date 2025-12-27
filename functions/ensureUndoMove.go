package functions

import (
	"context"
	"database/sql"
	"errors"

	db "github.com/rakshitg600/notakto-solo/db/generated"
)

func EnsureUndoMove(ctx context.Context, q *db.Queries, uid string, sessionID string) (
	boards []int32,
	err error,
) {
	// STEP 1: Validate sessionId
	existing, err := q.GetLatestSessionStateByPlayerId(ctx, uid)
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
	existing.Gameover = sql.NullBool{Bool: true, Valid: true}
	for i := int32(0); i < existing.NumberOfBoards.Int32; i++ {
		if !IsBoardDead(i, existing.Boards, existing.BoardSize.Int32) {
			existing.Gameover = sql.NullBool{Bool: false, Valid: true}
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
	err = q.UpdateWalletReduceCoins(ctx, db.UpdateWalletReduceCoinsParams{
		Uid:   uid,
		Coins: sql.NullInt32{Int32: undoMoveCost, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	// STEP 7: Pop last two elements (player move + AI move)
	existing.Boards = existing.Boards[:len(existing.Boards)-2]
	// Update session state after AI move
	err = q.UpdateSessionState(ctx, db.UpdateSessionStateParams{
		SessionID: sessionID,
		Boards:    existing.Boards,
	})
	if err != nil {
		return nil, err
	}
	return existing.Boards, nil
}