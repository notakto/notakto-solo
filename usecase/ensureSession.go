package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/rakshitg600/notakto-solo/db/generated"
	"github.com/rakshitg600/notakto-solo/contextkey"
	"github.com/rakshitg600/notakto-solo/store"
)

func EnsureSession(ctx context.Context, pool *pgxpool.Pool, numberOfBoards int32, boardSize int32, difficulty int32) (
	sessionID string,
	uidOut string,
	boards []int32,
	isAiMoveOut []bool,
	winner bool,
	boardSizeOut int32,
	numberOfBoardsOut int32,
	difficultyOut int32,
	gameover bool,
	createdAt time.Time,
	err error,
) {
	uid, ok := contextkey.UIDFromContext(ctx)
	if !ok || uid == "" {
		return "", "", nil, nil, false, 0, 0, 0, false, time.Time{}, errors.New("missing or invalid uid in context")
	}
	queries := db.New(pool)
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.Serializable,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		return "", "", nil, nil, false, 0, 0, 0, false, time.Time{}, err
	}
	defer tx.Rollback(ctx)

	qtx := queries.WithTx(tx)
	// STEP 1: Try existing session
	existing, err := store.GetLatestSessionStateByPlayerIdWithLock(ctx, qtx)
	if err == nil && existing.SessionID != "" {
		isGameOver := existing.Gameover.Valid && existing.Gameover.Bool
		if !isGameOver {
			sessionID = existing.SessionID
			uidOut = existing.Uid
			boards = existing.Boards
			if existing.Winner.Valid {
				winner = existing.Winner.Bool
			} else {
				winner = false
			}
			if existing.BoardSize.Valid {
				boardSizeOut = existing.BoardSize.Int32
			} else {
				boardSizeOut = 0
			}
			if existing.NumberOfBoards.Valid {
				numberOfBoardsOut = existing.NumberOfBoards.Int32
			} else {
				numberOfBoardsOut = 0
			}
			if existing.Difficulty.Valid {
				difficultyOut = existing.Difficulty.Int32
			} else {
				difficultyOut = 0
			}
			if existing.Gameover.Valid {
				gameover = existing.Gameover.Bool
			} else {
				gameover = false
			}
			if existing.CreatedAt.Valid {
				createdAt = existing.CreatedAt.Time
			} else {
				createdAt = time.Time{}
			}
			return sessionID, uidOut, boards, existing.IsAiMove, winner, boardSizeOut, numberOfBoardsOut, difficultyOut, gameover, createdAt, nil
		}
	}

	// STEP 2: Create a new session
	newSessionID := uuid.New().String()

	// a) Insert into session

	if err = store.CreateSession(ctx, qtx, boardSize, numberOfBoards, difficulty, newSessionID); err != nil {
		return "", "", nil, nil, false, 0, 0, 0, false, time.Time{}, err
	}

	// b) Insert initial session state
	if err = store.CreateInitialSessionState(ctx, qtx, newSessionID); err != nil {
		return "", "", nil, nil, false, 0, 0, 0, false, time.Time{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return "", "", nil, nil, false, 0, 0, 0, false, time.Time{}, err
	}
	// STEP 3: Return newly created session state values
	return newSessionID, uid, []int32{}, []bool{}, false, boardSize, numberOfBoards, difficulty, false, time.Now(), nil
}
