package handlers

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/rakshitg600/notakto-solo/functions"
	"github.com/rakshitg600/notakto-solo/types"
)

func (h *Handler) CreateGameHandler(c echo.Context) error {
	// ✅ Get UID
	uid, ok := c.Get("uid").(string)
	if !ok || uid == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized: missing or invalid uid")
	}
	log.Printf("CreateGameHandler called for uid: %s", uid)
	// ✅ Try binding the body
	var req types.CreateGameRequest
	if err := c.Bind(&req); err != nil {
		req = types.CreateGameRequest{} // reset if malformed JSON
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
	sessionID, uidOut, boards, winner, boardSize, numberOfBoards, difficulty, gameover, createdAt, err := functions.EnsureSession(
		c.Request().Context(),
		h.Queries,
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

	resp := map[string]interface{}{
		"session_id":       sessionID,
		"uid":              uidOut,
		"boards":           boards,
		"winner":           winner,
		"board_size":       boardSize,
		"number_of_boards": numberOfBoards,
		"difficulty":       difficulty,
		"gameover":         gameover,
		"created_at":       createdAt,
	}
	log.Printf("Created new game session for user %s: sessionID=%s, boards=%v, boardSize=%d, numberOfBoards=%d, difficulty=%d", uid, sessionID, boards, boardSize, numberOfBoards, difficulty)
	return c.JSON(http.StatusOK, resp)
}
