package handlers

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/rakshitg600/notakto-solo/contextkey"
	"github.com/rakshitg600/notakto-solo/usecase"
)

type SkipMoveRequest struct {
	SessionID string `json:"sessionId"`
}
type SkipMoveResponse struct {
	Boards        []int32 `json:"boards"`
	Gameover      bool    `json:"gameover"`
	Winner        bool    `json:"winner"`
	CoinsRewarded int32   `json:"coinsRewarded"`
	XpRewarded    int32   `json:"xpRewarded"`
}

func (h *Handler) SkipMoveHandler(c echo.Context) error {
	uid, ok := contextkey.UIDFromContext(c.Request().Context())
	if !ok || uid == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized: missing or invalid uid")
	}

	log.Printf("SkipMoveHandler called for uid: %s", uid)
	// ✅ Try binding the body
	var req SkipMoveRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	boards, gameOver, winner, coinsRewarded, xpRewarded, err := usecase.EnsureSkipMove(
		c.Request().Context(),
		h.Pool,
		req.SessionID,
	)
	if err != nil {
		c.Logger().Errorf("SkipMove failed: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to process skip move")
	}

	resp := SkipMoveResponse{
		Boards:        boards,
		Gameover:      gameOver,
		Winner:        winner,
		CoinsRewarded: coinsRewarded,
		XpRewarded:    xpRewarded,
	}
	log.Printf("SkipMoveHandler completed for uid: %s, sessionID: %s, gameOver: %v, winner: %v, coinsRewarded: %d, xpRewarded: %d", uid, req.SessionID, gameOver, winner, coinsRewarded, xpRewarded)
	return c.JSON(http.StatusOK, resp)
}
