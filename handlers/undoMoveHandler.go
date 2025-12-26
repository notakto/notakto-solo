package handlers

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rakshitg600/notakto-solo/functions"
	"github.com/rakshitg600/notakto-solo/types"
)

func (h *Handler) UndoMoveHandler(c echo.Context) error {
	// ✅ Get UID
	uid, ok := c.Get("uid").(string)
	if !ok || uid == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized: missing or invalid uid")
	}

	log.Printf("UndoMoveHandler called for uid: %s", uid)
	// ✅ Try binding the body
	var req types.UndoMoveRequest
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
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to process undo move")
	}

	resp := types.UndoMoveResponse{
		Boards: boards,
	}
	log.Printf("UndoMoveHandler completed for uid: %s, sessionID: %s", uid, req.SessionID)
	return c.JSON(http.StatusOK, resp)
}