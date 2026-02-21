package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (h *Handler) HealthHeadHandler(c echo.Context) error {
	return c.NoContent(http.StatusOK)
}

func (h *Handler) HealthGetHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status": "ok",
	})
}
