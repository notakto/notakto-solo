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

// EnsureSkipMove validates the session and processes a player "skip" move by charging the wallet,
// applying an AI move, updating session state, and awarding rewards if the game ends.
//
// EnsureSkipMove verifies that the provided sessionID matches the latest session for the user and
// that the game is not already over. It requires the player to have at least 200 coins, deducts
// that cost, computes and applies an AI move, updates the session state, and if the move ends the
// game it marks the session as finished and credits coins and XP to the player's wallet.
// Errors are returned for session mismatches or expirations, insufficient coins, failure to find an
// AI move, and any database operation failures.
func EnsureSkipMove(
	ctx context.Context,
	pool *pgxpool.Pool,
	uid string,
	sessionID string,
) (
	boards []int32,
	gameOver bool,
	winner bool,
	coinsRewarded int32,
	xpRewarded int32,
	err error,
) {
	queries := db.New(pool)
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.Serializable,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		return nil, false, false, 0, 0, err
	}
	defer tx.Rollback(ctx)

	qtx := queries.WithTx(tx)
	// STEP 1: Validate sessionId
	existing, err := store.GetLatestSessionStateByPlayerIdWithLock(ctx, qtx, uid)
	if err != nil {
		return nil, false, false, 0, 0, err
	}
	if existing.SessionID != sessionID {
		return nil, false, false, 0, 0, errors.New("session expired or not found")
	}

	// STEP 2: Validate gameover flag
	if existing.Gameover.Valid && existing.Gameover.Bool {
		return nil, true, existing.Winner.Bool, 0, 0, errors.New("game is already over")
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
		return nil, true, existing.Winner.Bool, 0, 0, errors.New("game is already over")
	}

	// STEP 4: Check wallet
	const skipMoveCost = 200

	wallet, err := store.GetWalletByPlayerIdWithLock(ctx, qtx, uid)
	if err != nil {
		return nil, false, false, 0, 0, err
	}
	if !wallet.Coins.Valid {
		return nil, false, false, 0, 0, errors.New("invalid wallet response from db")
	}
	if wallet.Coins.Int32 < skipMoveCost {
		return nil, false, false, 0, 0, errors.New("insufficient coins to skip move")
	}

	// STEP 5: Deduct coins
	if err = store.UpdateWalletReduceCoins(ctx, qtx, uid, skipMoveCost); err != nil {
		return nil, false, false, 0, 0, err
	}

	// STEP 6: AI move
	aiMoveIndex := logic.GetAIMove(
		existing.Boards,
		existing.BoardSize.Int32,
		existing.NumberOfBoards.Int32,
		existing.Difficulty.Int32,
	)
	if aiMoveIndex == -1 {
		return existing.Boards, false, false, 0, 0, errors.New("AI could not find a valid move")
	}

	existing.Boards = append(existing.Boards, -1) // skipped move marker
	existing.Boards = append(existing.Boards, aiMoveIndex)

	// STEP 7: Check gameover after AI move
	existing.Gameover = pgtype.Bool{Bool: true, Valid: true}
	for i := int32(0); i < existing.NumberOfBoards.Int32; i++ {
		if !logic.IsBoardDead(i, existing.Boards, existing.BoardSize.Int32) {
			existing.Gameover = pgtype.Bool{Bool: false, Valid: true}
			break
		}
	}

	if existing.Gameover.Bool {
		existing.Winner = pgtype.Bool{Bool: true, Valid: true}
	} else {
		existing.Winner = pgtype.Bool{Valid: false}
	}

	// STEP 8: Persist session state
	if err = store.UpdateSessionState(ctx, qtx, sessionID, existing.Boards); err != nil {
		return nil, existing.Gameover.Bool, existing.Winner.Bool, 0, 0, err
	}

	// STEP 9: If game continues
	if !existing.Gameover.Bool {
		if err = tx.Commit(ctx); err != nil {
			return nil, false, false, 0, 0, err
		}
		return existing.Boards, false, false, 0, 0, nil
	}

	// STEP 10: Gameover handling
	if err = store.UpdateSessionAfterGameover(ctx, qtx, sessionID, existing.Winner); err != nil {
		return nil, true, existing.Winner.Bool, 0, 0, err
	}

	coinsReward, xpReward := logic.CalculateRewards(
		existing.NumberOfBoards.Int32,
		existing.BoardSize.Int32,
		existing.Difficulty.Int32,
		true,
	)

	if err = store.UpdateWalletCoinsAndXpReward(ctx, qtx, uid, coinsReward, xpReward); err != nil {
		return nil, true, existing.Winner.Bool, 0, 0, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, true, existing.Winner.Bool, 0, 0, err
	}

	return existing.Boards, true, true, coinsReward, xpReward, nil
}
