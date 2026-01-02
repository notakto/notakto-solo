package usecase

import (
	"context"
	"database/sql"
	"errors"
	"time"

	db "github.com/rakshitg600/notakto-solo/db/generated"
	"github.com/rakshitg600/notakto-solo/logic"
)

// EnsureMakeMove validates the session and the requested move, applies the player's move,
// optionally applies an AI response, persists session state changes, and awards rewards when the game ends.
//
// The function performs validation of session ownership, board and cell indices, and move legality;
// it updates the session boards immediately after the player's move, checks for game-over, and if the
// game continues computes and applies an AI move and rechecks game-over. When a game-over occurs the
// session is updated and wallet rewards (coins and XP) are applied to the player's account.
//
// Returns:
//   - boards: the session's boards after applying the player's move and any AI move.
//   - gameOver: true if the game has ended after the applied moves, false otherwise.
//   - winner: true if the AI is the winner, false otherwise (when meaningful).
//   - coinsRewarded: coins awarded to the player as part of the game-over rewards, 0 if none or game not ended.
//   - xpRewarded: XP awarded to the player as part of the game-over rewards, 0 if none or game not ended.
//   - err: non-nil when validation, DB updates, or AI move resolution fail.
func EnsureMakeMove(ctx context.Context, q *db.Queries, uid string, sessionID string, boardIndex int32, cellIndex int32) (
	boards []int32,
	gameOver bool,
	winner bool,
	coinsRewarded int32,
	xpRewarded int32,
	err error,
) {
	// STEP 1: Validate sessionId
	GetLatestSessionStateByPlayerIdCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	existing, err := q.GetLatestSessionStateByPlayerId(GetLatestSessionStateByPlayerIdCtx, uid)
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
	// STEP 3: Validate BoardIndex
	if boardIndex < 0 || boardIndex >= existing.NumberOfBoards.Int32 {
		return nil, false, false, 0, 0, errors.New("invalid board index")
	}
	// STEP 4: Validate CellIndex
	boardSize := existing.BoardSize.Int32
	if cellIndex < 0 || cellIndex >= (boardSize*boardSize) {
		return nil, false, false, 0, 0, errors.New("invalid cell index")
	}
	// STEP 5: Validate if board is alive
	boardDead := logic.IsBoardDead(boardIndex, existing.Boards, boardSize)
	if boardDead {
		return nil, false, false, 0, 0, errors.New("selected board is already dead")
	}
	// STEP 6: Validate if cell is already marked
	moveIndex := boardIndex*boardSize*boardSize + cellIndex
	for i := 0; i < len(existing.Boards); i++ {
		if existing.Boards[i] == moveIndex {
			return nil, false, false, 0, 0, errors.New("cell is already marked")
		}
	}
	// STEP 7: Make Move
	existing.Boards = append(existing.Boards, moveIndex)
	// STEP 8: Check for gameover
	existing.Gameover = sql.NullBool{Bool: true, Valid: true}
	for i := int32(0); i < existing.NumberOfBoards.Int32; i++ {
		if !logic.IsBoardDead(i, existing.Boards, boardSize) {
			existing.Gameover = sql.NullBool{Bool: false, Valid: true}
			break
		}
	}
	if existing.Gameover.Valid && existing.Gameover.Bool {
		existing.Winner = sql.NullBool{Bool: false, Valid: true}
	} else if existing.Gameover.Valid && !existing.Gameover.Bool {
		existing.Winner = sql.NullBool{Bool: false, Valid: false}
	}
	// STEP 9: Update DB state || AI Makes move and Update DB state
	// 9.1 Update session state in db
	updateSessionStateCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	err = q.UpdateSessionState(updateSessionStateCtx, db.UpdateSessionStateParams{
		SessionID: sessionID,
		Boards:    existing.Boards,
	})
	if err != nil {
		return nil, existing.Gameover.Valid && existing.Gameover.Bool, existing.Winner.Valid && existing.Winner.Bool, 0, 0, err
	}
	// 9.2 If gameover update session and rewards
	if existing.Gameover.Valid && existing.Gameover.Bool {
		UpdateSessionAfterGameoverCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
		err = q.UpdateSessionAfterGameover(UpdateSessionAfterGameoverCtx, db.UpdateSessionAfterGameoverParams{
			SessionID: sessionID,
			Winner:    existing.Winner,
		})
		if err != nil {
			return nil, existing.Gameover.Valid && existing.Gameover.Bool, existing.Winner.Valid && existing.Winner.Bool, 0, 0, err
		}
		_, xpReward := logic.CalculateRewards(existing.NumberOfBoards.Int32, existing.BoardSize.Int32, existing.Difficulty.Int32, existing.Winner.Valid && existing.Winner.Bool)
		UpdateWalletXpRewardCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
		err = q.UpdateWalletXpReward(UpdateWalletXpRewardCtx, db.UpdateWalletXpRewardParams{
			Uid: uid,
			Xp:  sql.NullInt32{Int32: xpReward, Valid: true},
		})
		if err != nil {
			return nil, existing.Gameover.Valid && existing.Gameover.Bool, existing.Winner.Valid && existing.Winner.Bool, 0, 0, err
		}
		return existing.Boards, existing.Gameover.Valid && existing.Gameover.Bool, existing.Winner.Valid && existing.Winner.Bool, 0, xpReward, nil
	}
	// 9.3 If not gameover, AI makes a move
	if existing.Gameover.Valid && !existing.Gameover.Bool {
		aiMoveIndex := logic.GetAIMove(existing.Boards, boardSize, existing.NumberOfBoards.Int32, existing.Difficulty.Int32)
		if aiMoveIndex == -1 {
			// No valid moves for AI - this shouldn't happen if game is not over
			return existing.Boards, false, false, 0, 0, errors.New("AI could not find a valid move")
		}
		existing.Boards = append(existing.Boards, aiMoveIndex)
		// Check for gameover after AI move
		existing.Gameover = sql.NullBool{Bool: true, Valid: true}
		for i := int32(0); i < existing.NumberOfBoards.Int32; i++ {
			if !logic.IsBoardDead(i, existing.Boards, boardSize) {
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
		updateSessionStateCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
		err = q.UpdateSessionState(updateSessionStateCtx, db.UpdateSessionStateParams{
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
			UpdateSessionAfterGameoverCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
			defer cancel()
			err = q.UpdateSessionAfterGameover(UpdateSessionAfterGameoverCtx, db.UpdateSessionAfterGameoverParams{
				SessionID: sessionID,
				Winner:    existing.Winner,
			})
			if err != nil {
				return nil, existing.Gameover.Valid && existing.Gameover.Bool, existing.Winner.Valid && existing.Winner.Bool, 0, 0, err
			}
			coinsReward, xpReward := logic.CalculateRewards(existing.NumberOfBoards.Int32, existing.BoardSize.Int32, existing.Difficulty.Int32, existing.Winner.Valid && existing.Winner.Bool)
			updateWalletCoinsAndXpRewardCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
			defer cancel()
			err = q.UpdateWalletCoinsAndXpReward(updateWalletCoinsAndXpRewardCtx, db.UpdateWalletCoinsAndXpRewardParams{
				Uid:   uid,
				Coins: sql.NullInt32{Int32: coinsReward, Valid: true},
				Xp:    sql.NullInt32{Int32: xpReward, Valid: true},
			})
			if err != nil {
				return nil, existing.Gameover.Valid && existing.Gameover.Bool, existing.Winner.Valid && existing.Winner.Bool, 0, 0, err
			}
			return existing.Boards, existing.Gameover.Valid && existing.Gameover.Bool, existing.Winner.Valid && existing.Winner.Bool, coinsReward, xpReward, nil
		}
	}
	return nil, existing.Gameover.Valid && existing.Gameover.Bool, existing.Winner.Valid && existing.Winner.Bool, 0, 0, errors.New("unexpected behaviour")
}
