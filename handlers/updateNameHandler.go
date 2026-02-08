package handlers

import (
	"log"

	"github.com/labstack/echo/v4"

	"github.com/rakshitg600/notakto-solo/contextkey"
	"github.com/rakshitg600/notakto-solo/usecase"
)

type UpdatePlayerNameRequest struct {
	Name string `json:"name"`
}

func (h *Handler) UpdateNameHandler(c echo.Context) error {
	uid, ok := contextkey.UIDFromContext(c.Request().Context())
	if !ok || uid == "" {
		return echo.NewHTTPError(401, "unauthorized: missing or invalid uid")
	}
	log.Printf("UpdateNameHandler called for uid: %s", uid)
	// ✅ Try binding the body
	var req UpdatePlayerNameRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(400, "invalid request body")
	}
	if req.Name == "" {
		return echo.NewHTTPError(400, "name is required")
	}
	// ✅ Update the name
	updatedName, err := usecase.EnsureUpdateName(c.Request().Context(), h.Pool, req.Name)
	if err != nil {
		log.Printf("UpdateNameHandler error for uid %s: %v", uid, err)
		return echo.NewHTTPError(500, "internal server error")
	}
	// ✅ Return the updated name
	log.Printf("Updated name for uid %s to %s", uid, updatedName)
	return c.JSON(200, map[string]string{"name": updatedName})
}
