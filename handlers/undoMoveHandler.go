package handlers

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rakshitg600/notakto-solo/functions"
)

type UndoMoveRequest struct {
	SessionID string `json:"sessionId"`
}
type UndoMoveResponse struct {
	Boards []int32 `json:"boards"`
}

func (h *Handler) UndoMoveHandler(c echo.Context) error {
	// ✅ Get UID
	uid, ok := c.Get("uid").(string)
	if !ok || uid == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized: missing or invalid uid")
	}

	log.Printf("UndoMoveHandler called for uid: %s", uid)
	// ✅ Try binding the body
	var req UndoMoveRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	boards, err := functions.EnsureUndoMove(
		c.Request().Context(),
		h.Queries,
		uid,
		req.SessionID,
	)
	if err != nil {
		c.Logger().Errorf("UndoMove failed: %v", err)
		// Return appropriate status codes based on error type
		errMsg := err.Error()
		if errMsg == "session expired or not found" || errMsg == "game is already over" ||
			errMsg == "insufficient coins to undo move" || errMsg == "no moves to undo" {
			return echo.NewHTTPError(http.StatusBadRequest, errMsg)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to process undo move")
	}

	resp := UndoMoveResponse{
		Boards: boards,
	}
	log.Printf("UndoMoveHandler completed for uid: %s, sessionID: %s", uid, req.SessionID)
	return c.JSON(http.StatusOK, resp)
}
