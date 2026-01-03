package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	db "github.com/rakshitg600/notakto-solo/db/generated"
	"github.com/rakshitg600/notakto-solo/store"
)

// EnsureSession returns the latest existing session for a user, or creates a new one if none exists.
// EnsureSession retrieves the user's most recent active session if one exists and is not game over; otherwise it creates a new session and initial state and returns the session fields suitable for a JSON response.
// When an existing non-game-over session is found, its stored session ID, user ID, boards, winner, board size, number of boards, difficulty, gameover flag, and creation time are returned. If no active session exists, a new session and an empty initial state are created and the new session's ID, provided inputs, an empty boards slice, `false` winner, `false` gameover, and the current time are returned.
// Database operations use per-call timeouts of 3 seconds; any database error is returned.
func EnsureSession(ctx context.Context, q *db.Queries, uid string, numberOfBoards int32, boardSize int32, difficulty int32) (
	sessionID string,
	uidOut string,
	boards []int32,
	winner bool,
	boardSizeOut int32,
	numberOfBoardsOut int32,
	difficultyOut int32,
	gameover bool,
	createdAt time.Time,
	err error,
) {
	// STEP 1: Try existing session
	existing, err := store.GetLatestSessionStateByPlayerId(ctx, q, uid)
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
			return sessionID, uidOut, boards, winner, boardSizeOut, numberOfBoardsOut, difficultyOut, gameover, createdAt, nil
		}
	}

	// STEP 2: Create a new session
	newSessionID := uuid.New().String()

	// a) Insert into session

	if err = store.CreateSession(ctx, q, uid, boardSize, numberOfBoards, difficulty, newSessionID); err != nil {
		return "", "", nil, false, 0, 0, 0, false, time.Time{}, err
	}

	// b) Insert initial session state
	if err = store.CreateInitialSessionState(ctx, q, newSessionID); err != nil {
		return "", "", nil, false, 0, 0, 0, false, time.Time{}, err
	}

	// STEP 3: Return newly created session state values
	return newSessionID, uid, []int32{}, false, boardSize, numberOfBoards, difficulty, false, time.Now(), nil
}
