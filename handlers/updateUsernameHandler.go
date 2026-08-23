package handlers

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/rakshitg600/notakto-solo/contextkey"
	"github.com/rakshitg600/notakto-solo/usecase"
)

type UpdateUsernameRequest struct {
	Username string `json:"username"`
}

type UpdateUsernameResponse struct {
	Username string `json:"username"`
}

func (h *Handler) UpdateUsernameHandler(c echo.Context) error {
	uid, ok := contextkey.UIDFromContext(c.Request().Context())
	if !ok || uid == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized: missing or invalid uid")
	}
	log.Printf("UpdateUsernameHandler called for uid: %s", uid)

	var req UpdateUsernameRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if req.Username == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "username is required")
	}

	updatedUsername, statusCode, err := usecase.EnsureUpdateUsername(c.Request().Context(), h.Pool, req.Username)
	if err != nil {
		if statusCode >= 500 {
			log.Printf("UpdateUsernameHandler error for uid %s: %v", uid, err)
		}
		return echo.NewHTTPError(statusCode, err.Error())
	}

	log.Printf("Updated username for uid %s to %s", uid, updatedUsername)
	return c.JSON(http.StatusOK, UpdateUsernameResponse{Username: updatedUsername})
}
