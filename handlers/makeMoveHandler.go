package handlers

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rakshitg600/notakto-solo/functions"
)

type MakeMoveRequest struct {
	SessionID  string `json:"sessionId"`
	BoardIndex int32  `json:"boardIndex"`
	CellIndex  int32  `json:"cellIndex"`
}
type MakeMoveResponse struct {
	Boards        []int32 `json:"boards"`
	Gameover      bool    `json:"gameover"`
	Winner        bool    `json:"winner"`
	CoinsRewarded int32   `json:"coinsRewarded"`
	XpRewarded    int32   `json:"xpRewarded"`
}

func (h *Handler) MakeMoveHandler(c echo.Context) error {
	// ✅ Get UID
	uid, ok := c.Get("uid").(string)
	if !ok || uid == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized: missing or invalid uid")
	}

	log.Printf("MakeMoveHandler called for uid: %s", uid)
	// ✅ Try binding the body
	var req MakeMoveRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	boards, gameOver, winner, coinsRewarded, xpRewarded, err := functions.EnsureMakeMove(
		c.Request().Context(),
		h.Queries,
		uid,
		req.SessionID,
		req.BoardIndex,
		req.CellIndex,
	)
	if err != nil {
		c.Logger().Errorf("MakeMove failed: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	resp := MakeMoveResponse{
		Boards:        boards,
		Gameover:      gameOver,
		Winner:        winner,
		CoinsRewarded: coinsRewarded,
		XpRewarded:    xpRewarded,
	}
	log.Printf("MakeMoveHandler completed for uid: %s, sessionID: %s, boardIndex: %d, cellIndex: %d, gameOver: %v, winner: %v, coinsRewarded: %d, xpRewarded: %d", uid, req.SessionID, req.BoardIndex, req.CellIndex, gameOver, winner, coinsRewarded, xpRewarded)
	return c.JSON(http.StatusOK, resp)
}
