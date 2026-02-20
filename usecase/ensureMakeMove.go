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

func EnsureMakeMove(ctx context.Context, pool *pgxpool.Pool, sessionID string, boardIndex int32, cellIndex int32) (
	boards []int32,
	isAiMove []bool,
	gameOver bool,
	winner bool,
	coinsRewarded int32,
	xpRewarded int32,
	err error,
) {
	uid, ok := contextkey.UIDFromContext(ctx)
	if !ok || uid == "" {
		return nil, nil, false, false, 0, 0, errors.New("missing or invalid uid in context")
	}
	queries := db.New(pool)
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.Serializable,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		return nil, nil, false, false, 0, 0, err
	}
	defer tx.Rollback(ctx)

	qtx := queries.WithTx(tx)

	// STEP 1: Validate sessionId
	existing, err := store.GetLatestSessionStateByPlayerIdWithLock(ctx, qtx)
	if err != nil {
		return nil, nil, false, false, 0, 0, err
	}
	if existing.SessionID != sessionID {
		return nil, nil, false, false, 0, 0, errors.New("session expired or not found")
	}
	// Validate IsAiMove and Boards length alignment
	if len(existing.IsAiMove) != len(existing.Boards) {
		return nil, nil, false, false, 0, 0, errors.New("session state corrupted: IsAiMove and Boards length mismatch")
	}
	// STEP 2: Validate gameover
	if existing.Gameover.Valid && existing.Gameover.Bool {
		return nil, nil, true, existing.Winner.Bool, 0, 0, errors.New("game is already over")
	}
	// STEP 3: Validate BoardIndex
	if boardIndex < 0 || boardIndex >= existing.NumberOfBoards.Int32 {
		return nil, nil, false, false, 0, 0, errors.New("invalid board index")
	}
	// STEP 4: Validate CellIndex
	boardSize := existing.BoardSize.Int32
	if cellIndex < 0 || cellIndex >= (boardSize*boardSize) {
		return nil, nil, false, false, 0, 0, errors.New("invalid cell index")
	}
	// STEP 5: Validate if board is alive
	boardDead := logic.IsBoardDead(boardIndex, existing.Boards, boardSize)
	if boardDead {
		return nil, nil, false, false, 0, 0, errors.New("selected board is already dead")
	}
	// STEP 6: Validate if cell is already marked
	moveIndex := boardIndex*boardSize*boardSize + cellIndex
	for i := 0; i < len(existing.Boards); i++ {
		if existing.Boards[i] == moveIndex {
			return nil, nil, false, false, 0, 0, errors.New("cell is already marked")
		}
	}
	// STEP 7: Make Move
	if len(existing.IsAiMove) != len(existing.Boards) {
		return nil, nil, false, false, 0, 0, errors.New("session move history out of sync")
	}
	existing.Boards = append(existing.Boards, moveIndex)
	existing.IsAiMove = append(existing.IsAiMove, false) // Player move
	// STEP 8: Check for gameover
	existing.Gameover = pgtype.Bool{Bool: true, Valid: true}
	for i := int32(0); i < existing.NumberOfBoards.Int32; i++ {
		if !logic.IsBoardDead(i, existing.Boards, boardSize) {
			existing.Gameover = pgtype.Bool{Bool: false, Valid: true}
			break
		}
	}
	if existing.Gameover.Valid && existing.Gameover.Bool {
		existing.Winner = pgtype.Bool{Bool: false, Valid: true}
	} else if existing.Gameover.Valid && !existing.Gameover.Bool {
		existing.Winner = pgtype.Bool{Bool: false, Valid: false}
	}
	// STEP 9: Update DB state || AI Makes move and Update DB state
	// 9.1 Update session state in db
	err = store.UpdateSessionState(ctx, qtx, sessionID, existing.Boards, existing.IsAiMove)
	if err != nil {
		return nil, nil, existing.Gameover.Valid && existing.Gameover.Bool, existing.Winner.Valid && existing.Winner.Bool, 0, 0, err
	}
	// 9.2 If gameover update session and rewards
	if existing.Gameover.Valid && existing.Gameover.Bool {
		err = store.UpdateSessionAfterGameover(ctx, qtx, sessionID, existing.Winner)
		if err != nil {
			return nil, nil, existing.Gameover.Valid && existing.Gameover.Bool, existing.Winner.Valid && existing.Winner.Bool, 0, 0, err
		}
		_, xpReward := logic.CalculateRewards(existing.NumberOfBoards.Int32, existing.BoardSize.Int32, existing.Difficulty.Int32, existing.Winner.Valid && existing.Winner.Bool)

		err = store.UpdateWalletXpReward(ctx, qtx, xpReward)
		if err != nil {
			return nil, nil, existing.Gameover.Valid && existing.Gameover.Bool, existing.Winner.Valid && existing.Winner.Bool, 0, 0, err
		}
		if err := tx.Commit(ctx); err != nil {
			return nil, nil, false, false, 0, 0, err
		}
		return existing.Boards, existing.IsAiMove, existing.Gameover.Valid && existing.Gameover.Bool, existing.Winner.Valid && existing.Winner.Bool, 0, xpReward, nil
	}
	// 9.3 If not gameover, AI makes a move
	if existing.Gameover.Valid && !existing.Gameover.Bool {
		aiMoveIndex := logic.GetAIMove(existing.Boards, boardSize, existing.NumberOfBoards.Int32, existing.Difficulty.Int32)
		if aiMoveIndex == -1 {
			// No valid moves for AI - this shouldn't happen if game is not over
			return existing.Boards, existing.IsAiMove, false, false, 0, 0, errors.New("AI could not find a valid move")
		}
		existing.Boards = append(existing.Boards, aiMoveIndex)
		existing.IsAiMove = append(existing.IsAiMove, true) // AI move
		// Check for gameover after AI move
		existing.Gameover = pgtype.Bool{Bool: true, Valid: true}
		for i := int32(0); i < existing.NumberOfBoards.Int32; i++ {
			if !logic.IsBoardDead(i, existing.Boards, boardSize) {
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
			return nil, nil, existing.Gameover.Valid && existing.Gameover.Bool, existing.Winner.Valid && existing.Winner.Bool, 0, 0, err
		}
		if existing.Gameover.Valid && !existing.Gameover.Bool {
			if err := tx.Commit(ctx); err != nil {
				return nil, nil, false, false, 0, 0, err
			}
			return existing.Boards,
				existing.IsAiMove,
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
				return nil, nil, existing.Gameover.Valid && existing.Gameover.Bool, existing.Winner.Valid && existing.Winner.Bool, 0, 0, err
			}
			coinsReward, xpReward := logic.CalculateRewards(existing.NumberOfBoards.Int32, existing.BoardSize.Int32, existing.Difficulty.Int32, existing.Winner.Valid && existing.Winner.Bool)
			err = store.UpdateWalletCoinsAndXpReward(ctx, qtx, coinsReward, xpReward)
			if err != nil {
				return nil, nil, existing.Gameover.Valid && existing.Gameover.Bool, existing.Winner.Valid && existing.Winner.Bool, 0, 0, err
			}
			if err := tx.Commit(ctx); err != nil {
				return nil, nil, false, false, 0, 0, err
			}
			return existing.Boards, existing.IsAiMove, existing.Gameover.Valid && existing.Gameover.Bool, existing.Winner.Valid && existing.Winner.Bool, coinsReward, xpReward, nil
		}
	}
	return nil, nil, existing.Gameover.Valid && existing.Gameover.Bool, existing.Winner.Valid && existing.Winner.Bool, 0, 0, errors.New("unexpected behaviour")
}
