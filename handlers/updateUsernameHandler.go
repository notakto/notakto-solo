package handlers

import (
	"errors"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/rakshitg600/notakto-solo/contextkey"
	"github.com/rakshitg600/notakto-solo/usecase"
)

type UpdateUsernameRequest struct {
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

	updatedUsername, err := usecase.EnsureUpdateUsername(c.Request().Context(), h.Pool, req.Username)
	if err != nil {
		if errors.Is(err, usecase.ErrUsernameExists) {
			return echo.NewHTTPError(http.StatusConflict, "username already exists")
		}
		log.Printf("UpdateUsernameHandler error for uid %s: %v", uid, err)
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	log.Printf("Updated username for uid %s to %s", uid, updatedUsername)
	return c.JSON(http.StatusOK, map[string]string{"username": updatedUsername})
}
