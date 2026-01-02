package functions

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	db "github.com/rakshitg600/notakto-solo/db/generated"
)

// EnsureSession returns the latest existing session for a user, or creates a new one if none exists.
// It returns typed values for the handler to compose the JSON response.
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
	getLatestSessionStateByPlayerIdCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	existing, err := q.GetLatestSessionStateByPlayerId(getLatestSessionStateByPlayerIdCtx, uid)
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
	createSessionCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err = q.CreateSession(createSessionCtx, db.CreateSessionParams{
		SessionID:      newSessionID,
		Uid:            uid,
		BoardSize:      sql.NullInt32{Int32: boardSize, Valid: true},
		NumberOfBoards: sql.NullInt32{Int32: numberOfBoards, Valid: true},
		Difficulty:     sql.NullInt32{Int32: difficulty, Valid: true},
	}); err != nil {
		return "", "", nil, false, 0, 0, 0, false, time.Time{}, err
	}

	// b) Insert initial session state
	createInitialSessionStateCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err = q.CreateInitialSessionState(createInitialSessionStateCtx, db.CreateInitialSessionStateParams{
		SessionID: newSessionID,
		Boards:    []int32{}, // empty initial boards
	}); err != nil {
		return "", "", nil, false, 0, 0, 0, false, time.Time{}, err
	}

	// STEP 3: Return newly created session state values
	return newSessionID, uid, []int32{}, false, boardSize, numberOfBoards, difficulty, false, time.Now(), nil
}
