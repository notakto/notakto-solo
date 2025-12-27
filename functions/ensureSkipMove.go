package functions

import (
	"context"
	"database/sql"
	"errors"

	db "github.com/rakshitg600/notakto-solo/db/generated"
)

func EnsureSkipMove(ctx context.Context, q *db.Queries, uid string, sessionID string) (
	boards []int32,
	gameOver bool,
	winner bool,
	coinsRewarded int32,
	xpRewarded int32,
	err error,
) {
	// STEP 1: Validate sessionId
	existing, err := q.GetLatestSessionStateByPlayerId(ctx, uid)
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
	existing.Gameover = sql.NullBool{Bool: true, Valid: true}
	for i := int32(0); i < existing.NumberOfBoards.Int32; i++ {
		if !IsBoardDead(i, existing.Boards, existing.BoardSize.Int32) {
			existing.Gameover = sql.NullBool{Bool: false, Valid: true}
			break
		}
	}
	if existing.Gameover.Valid && existing.Gameover.Bool {
		//TODO: Update session state in DB to reflect gameover
		return nil, true, existing.Winner.Bool, 0, 0, errors.New("game is already over")
	}

	// STEP 4: Check wallet for sufficient coins
	coins, _, err := EnsureGetWallet(ctx, q, uid)
	if err != nil {
		return nil, false, false, 0, 0, err
	}
	if coins < 200 {
		return nil, false, false, 0, 0, errors.New("insufficient coins to skip move")
	}

	// STEP 5: Deduct coins
	const skipMoveCost = 200
	err = q.UpdateWalletReduceCoins(ctx, db.UpdateWalletReduceCoinsParams{
		Uid:   uid,
		Coins: sql.NullInt32{Int32: skipMoveCost, Valid: true},
	})
	if err != nil {
		return nil, existing.Gameover.Valid && existing.Gameover.Bool, existing.Winner.Valid && existing.Winner.Bool, 0, 0, err
	}

	// STEP 6: AI makes a move
	aiMoveIndex := GetAIMove(existing.Boards, existing.BoardSize.Int32, existing.NumberOfBoards.Int32, existing.Difficulty.Int32)
	if aiMoveIndex == -1 {
		// No valid moves for AI - this shouldn't happen if game is not over
		return existing.Boards, false, false, 0, 0, errors.New("AI could not find a valid move")
	}

	existing.Boards = append(existing.Boards, -1) // Placeholder for player's skipped move
	existing.Boards = append(existing.Boards, aiMoveIndex)
	// Check for gameover after AI move
	existing.Gameover = sql.NullBool{Bool: true, Valid: true}
	for i := int32(0); i < existing.NumberOfBoards.Int32; i++ {
		if !IsBoardDead(i, existing.Boards, existing.BoardSize.Int32) {
			existing.Gameover = sql.NullBool{Bool: false, Valid: true}
			break
		}
	}
	if existing.Gameover.Valid && existing.Gameover.Bool {
		existing.Winner = sql.NullBool{Bool: true, Valid: true}
	} else if existing.Gameover.Valid && !existing.Gameover.Bool {
		existing.Winner = sql.NullBool{Bool: false, Valid: false}
	}
	// Update session state after AI move
	err = q.UpdateSessionState(ctx, db.UpdateSessionStateParams{
		SessionID: sessionID,
		Boards:    existing.Boards,
	})
	if err != nil {
		return nil, existing.Gameover.Valid && existing.Gameover.Bool, existing.Winner.Valid && existing.Winner.Bool, 0, 0, err
	}
	if existing.Gameover.Valid && !existing.Gameover.Bool {
		return existing.Boards,
			false,
			false,
			0,
			0,
			nil
	}
	// If gameover after AI move, update session
	if existing.Gameover.Valid && existing.Gameover.Bool {
		err = q.UpdateSessionAfterGameover(ctx, db.UpdateSessionAfterGameoverParams{
			SessionID: sessionID,
			Winner:    existing.Winner,
		})
		if err != nil {
			return nil, existing.Gameover.Valid && existing.Gameover.Bool, existing.Winner.Valid && existing.Winner.Bool, 0, 0, err
		}
		coinsReward, xpReward := calculateRewards(existing.NumberOfBoards.Int32, existing.BoardSize.Int32, existing.Difficulty.Int32, existing.Winner.Valid && existing.Winner.Bool)
		err = q.UpdateWalletCoinsAndXpReward(ctx, db.UpdateWalletCoinsAndXpRewardParams{
			Uid:   uid,
			Coins: sql.NullInt32{Int32: coinsReward, Valid: true},
			Xp:    sql.NullInt32{Int32: xpReward, Valid: true},
		})
		if err != nil {
			return nil, existing.Gameover.Valid && existing.Gameover.Bool, existing.Winner.Valid && existing.Winner.Bool, 0, 0, err
		}
		return existing.Boards, existing.Gameover.Valid && existing.Gameover.Bool, existing.Winner.Valid && existing.Winner.Bool, coinsReward, xpReward, nil
	}
	return existing.Boards, existing.Gameover.Valid && existing.Gameover.Bool, existing.Winner.Bool, 0, 0, nil
}
