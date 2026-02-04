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
func EnsureSkipMove(ctx context.Context, pool *pgxpool.Pool, uid string, sessionID string) (
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
	// STEP 2: Validate gameover
	if existing.Gameover.Valid && existing.Gameover.Bool {
		return nil, true, existing.Winner.Bool, 0, 0, errors.New("game is already over")
	}
	// STEP 3: Verify if game is over before skipping move
	existing.Gameover = pgtype.Bool{Bool: true, Valid: true}
	for i := int32(0); i < existing.NumberOfBoards.Int32; i++ {
		if !logic.IsBoardDead(i, existing.Boards, existing.BoardSize.Int32) {
			existing.Gameover = pgtype.Bool{Bool: false, Valid: true}
			break
		}
	}
	if existing.Gameover.Valid && existing.Gameover.Bool {
		//TODO: Update session state in DB to reflect gameover
		return nil, true, existing.Winner.Bool, 0, 0, errors.New("game is already over")
	}

	// STEP 4: Check wallet for sufficient coins
	const skipMoveCost = 200
	wallet, err := store.GetWalletByPlayerIdWithLock(ctx, qtx, uid)
	if err != nil {
		return nil, false, false, 0, 0, err
	}
	if wallet.Coins.Valid == false || wallet.Xp.Valid == false {
		return nil, false, false, 0, 0, errors.New("invalid wallet response from db")
	}
	if wallet.Coins.Int32 < skipMoveCost {
		return nil, false, false, 0, 0, errors.New("insufficient coins to skip move")
	}

	// STEP 5: Deduct coins
	err = store.UpdateWalletReduceCoins(ctx, qtx, uid, skipMoveCost)
	if err != nil {
		return nil, existing.Gameover.Valid && existing.Gameover.Bool, existing.Winner.Valid && existing.Winner.Bool, 0, 0, err
	}

	// STEP 6: AI makes a move
	aiMoveIndex := logic.GetAIMove(existing.Boards, existing.BoardSize.Int32, existing.NumberOfBoards.Int32, existing.Difficulty.Int32)
	if aiMoveIndex == -1 {
		// No valid moves for AI - this shouldn't happen if game is not over
		return existing.Boards, false, false, 0, 0, errors.New("AI could not find a valid move")
	}

	existing.Boards = append(existing.Boards, aiMoveIndex)
	existing.IsAiMove = append(existing.IsAiMove, true) // AI move (skip move only adds AI move)
	// Check for gameover after AI move
	existing.Gameover = pgtype.Bool{Bool: true, Valid: true}
	for i := int32(0); i < existing.NumberOfBoards.Int32; i++ {
		if !logic.IsBoardDead(i, existing.Boards, existing.BoardSize.Int32) {
			existing.Gameover = pgtype.Bool{Bool: false, Valid: true}
			break
		}
	}
	if existing.Gameover.Valid && existing.Gameover.Bool {
		existing.Winner = pgtype.Bool{Bool: true, Valid: true}
	} else if existing.Gameover.Valid && !existing.Gameover.Bool {
		existing.Winner = pgtype.Bool{Bool: false, Valid: false}
	}
	// Update session state after AI move
	err = store.UpdateSessionState(ctx, qtx, sessionID, existing.Boards, existing.IsAiMove)
	if err != nil {
		return nil, existing.Gameover.Valid && existing.Gameover.Bool, existing.Winner.Valid && existing.Winner.Bool, 0, 0, err
	}
	if existing.Gameover.Valid && !existing.Gameover.Bool {
		if err := tx.Commit(ctx); err != nil {
			return nil, false, false, 0, 0, err
		}
		return existing.Boards,
			false,
			false,
			0,
			0,
			nil
	}
	// If gameover after AI move, update session
	if existing.Gameover.Valid && existing.Gameover.Bool {
		err = store.UpdateSessionAfterGameover(ctx, qtx, sessionID, existing.Winner)
		if err != nil {
			return nil, existing.Gameover.Valid && existing.Gameover.Bool, existing.Winner.Valid && existing.Winner.Bool, 0, 0, err
		}
		coinsReward, xpReward := logic.CalculateRewards(existing.NumberOfBoards.Int32, existing.BoardSize.Int32, existing.Difficulty.Int32, existing.Winner.Valid && existing.Winner.Bool)
		err = store.UpdateWalletCoinsAndXpReward(ctx, qtx, uid, coinsReward, xpReward)
		if err != nil {
			return nil, existing.Gameover.Valid && existing.Gameover.Bool, existing.Winner.Valid && existing.Winner.Bool, 0, 0, err
		}
		if err := tx.Commit(ctx); err != nil {
			return nil, false, false, 0, 0, err
		}
		return existing.Boards, existing.Gameover.Valid && existing.Gameover.Bool, existing.Winner.Valid && existing.Winner.Bool, coinsReward, xpReward, nil
	}
	return nil, false, false, 0, 0, errors.New("invalid gameover state obtained from db")
}
