package handlers

import (
	"log"

	"github.com/labstack/echo/v4"
	"github.com/rakshitg600/notakto-solo/functions"
)

type UpdatePlayerNameRequest struct {
	Name string `json:"name"`
}

func (h *Handler) UpdateNameHandler(c echo.Context) error {
	// ✅ Get UID
	uid, ok := c.Get("uid").(string)
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
	updatedName, err := functions.EnsureUpdateName(c.Request().Context(), h.Queries, req.Name, uid)
	if err != nil {
		return echo.NewHTTPError(500, err.Error())
	}
	// ✅ Return the updated name
	log.Printf("Updated name for uid %s to %s", uid, updatedName)
	return c.JSON(200, map[string]string{"name": updatedName})
}
