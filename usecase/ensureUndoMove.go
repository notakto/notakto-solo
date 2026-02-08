package usecase

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/rakshitg600/notakto-solo/db/generated"
	"github.com/rakshitg600/notakto-solo/contextkey"
	"github.com/rakshitg600/notakto-solo/logic"
	"github.com/rakshitg600/notakto-solo/store"
)

func EnsureUndoMove(ctx context.Context, pool *pgxpool.Pool, sessionID string) (
	boards []int32,
	err error,
) {
	uid, ok := contextkey.UIDFromContext(ctx)
	if !ok || uid == "" {
		return nil, errors.New("missing or invalid uid in context")
	}
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
	// STEP 1: Validate sessionId
	existing, err := store.GetLatestSessionStateByPlayerIdWithLock(ctx, qtx)
	if err != nil {
		return nil, err
	}
	if existing.SessionID != sessionID {
		return nil, errors.New("session expired or not found")
	}
	// Validate IsAiMove and Boards length alignment
	if len(existing.IsAiMove) != len(existing.Boards) {
		return nil, errors.New("session state corrupted: IsAiMove and Boards length mismatch")
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
	const undoMoveCost = 100
	wallet, err := store.GetWalletByPlayerIdWithLock(ctx, qtx)
	if err != nil {
		return nil, err
	}
	if wallet.Coins.Valid == false || wallet.Xp.Valid == false {
		return nil, errors.New("invalid wallet response from db")
	}
	if wallet.Coins.Int32 < undoMoveCost {
		return nil, errors.New("insufficient coins to undo move")
	}
	// STEP 5: Verify there are moves to undo
	if len(existing.Boards) < 1 {
		return nil, errors.New("no moves to undo")
	}

	// STEP 6: Deduct coins
	err = store.UpdateWalletReduceCoins(ctx, qtx, undoMoveCost)
	if err != nil {
		return nil, err
	}

	// STEP 7: Determine how many moves to undo based on isAiMove
	// If last move was AI and second-to-last was player: delete 2 (regular player+AI turn)
	// If last move was AI and second-to-last was also AI (or doesn't exist): delete 1 (skip move case)
	movesToDelete := 1
	if len(existing.IsAiMove) >= 2 {
		lastIsAi := existing.IsAiMove[len(existing.IsAiMove)-1]
		secondLastIsAi := existing.IsAiMove[len(existing.IsAiMove)-2]
		if lastIsAi && !secondLastIsAi {
			// Regular turn: player move followed by AI move
			movesToDelete = 2
		}
	}

	// Pop moves from boards and isAiMove arrays
	existing.Boards = existing.Boards[:len(existing.Boards)-movesToDelete]
	existing.IsAiMove = existing.IsAiMove[:len(existing.IsAiMove)-movesToDelete]

	// Update session state
	err = store.UpdateSessionState(ctx, qtx, sessionID, existing.Boards, existing.IsAiMove)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return existing.Boards, nil
}
