package handlers

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/rakshitg600/notakto-solo/contextkey"
	"github.com/rakshitg600/notakto-solo/usecase"
)

type QuitGameRequest struct {
	SessionID string `json:"sessionId"`
}
type QuitGameResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

func (h *Handler) QuitGameHandler(c echo.Context) error {
	uid, ok := contextkey.UIDFromContext(c.Request().Context())
	if !ok || uid == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized: missing or invalid uid")
	}
	log.Printf("QuitGameHandler called for uid: %s", uid)
	// ✅ Try binding the body
	var req QuitGameRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	success, err := usecase.EnsureQuitGame(c.Request().Context(), h.Pool, req.SessionID)
	if err != nil {
		c.Logger().Errorf("EnsureQuitGame failed: %v", err)
		return c.JSON(http.StatusOK, QuitGameResponse{
			Success: success,
			Error:   err.Error(),
		})
	}

	resp := QuitGameResponse{
		Success: success,
	}
	log.Printf("QuitGameHandler completed for uid: %s, success: %v", uid, success)
	return c.JSON(http.StatusOK, resp)
}
