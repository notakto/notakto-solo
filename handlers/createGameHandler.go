package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/rakshitg600/notakto-solo/usecase"
)

type CreateGameRequest struct {
	NumberOfBoards int32 `json:"numberOfBoards"`
	BoardSize      int32 `json:"boardSize"`
	Difficulty     int32 `json:"difficulty"`
}
type CreateGameResponse struct {
	SessionId      string  `json:"sessionId"`
	Uid            string  `json:"uid"`
	Boards         []int32 `json:"boards"`
	Winner         bool    `json:"winner"`
	BoardSize      int32   `json:"boardSize"`
	NumberOfBoards int32   `json:"numberOfBoards"`
	Difficulty     int32   `json:"difficulty"`
	Gameover       bool    `json:"gameover"`
	CreatedAt      string  `json:"createdAt"`
}

func (h *Handler) CreateGameHandler(c echo.Context) error {
	// ✅ Get UID
	uid, ok := c.Get("uid").(string)
	if !ok || uid == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized: missing or invalid uid")
	}
	log.Printf("CreateGameHandler called for uid: %s", uid)
	// ✅ Try binding the body
	var req CreateGameRequest
	if err := c.Bind(&req); err != nil {
		req = CreateGameRequest{} // reset if malformed JSON
	}

	// ✅ Apply defaults if fields are zero or invalid
	if req.NumberOfBoards < 1 || req.NumberOfBoards > 5 {
		req.NumberOfBoards = 3
	}
	if req.BoardSize < 2 || req.BoardSize > 5 {
		req.BoardSize = 3
	}
	if req.Difficulty < 1 || req.Difficulty > 5 {
		req.Difficulty = 1
	}

	// ✅✅ Logic: get typed values from EnsureSession
	sessionID, uidOut, boards, winner, boardSize, numberOfBoards, difficulty, gameover, createdAt, err := usecase.EnsureSession(
		c.Request().Context(),
		h.Pool,
		uid,
		req.NumberOfBoards,
		req.BoardSize,
		req.Difficulty,
	)
	// ✅ Handle errors
	if err != nil {
		c.Logger().Errorf("EnsureSession failed: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	createdAtStr := createdAt.UTC().Format(time.RFC3339)

	resp := CreateGameResponse{
		SessionId:      sessionID,
		Uid:            uidOut,
		Boards:         boards,
		Winner:         winner,
		BoardSize:      boardSize,
		NumberOfBoards: numberOfBoards,
		Difficulty:     difficulty,
		Gameover:       gameover,
		CreatedAt:      createdAtStr,
	}
	log.Printf("Created new game session for user %s: sessionID=%s, boards=%v, boardSize=%d, numberOfBoards=%d, difficulty=%d", uid, sessionID, boards, boardSize, numberOfBoards, difficulty)
	return c.JSON(http.StatusOK, resp)
}
