package handlers

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

func (h *Handler) HealthHeadHandler(c echo.Context) error {
	log.Default().Println("Health check HEAD request received")
	return c.NoContent(http.StatusOK)
}

func (h *Handler) HealthGetHandler(c echo.Context) error {
	log.Default().Println("Health check GET request received")
	return c.JSON(http.StatusOK, map[string]string{
		"status":  "ok",
		"uptime":  "running",
		"service": "my-echo-server",
	})
}
