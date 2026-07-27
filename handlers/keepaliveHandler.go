package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rakshitg600/notakto-solo/contextkey"
)

const (
	keepaliveTimeout = 5 * time.Second
	keepaliveKey     = "notakto:system:keepalive"
	keepaliveValue   = "active"
	keepaliveTTL     = 60 * time.Second
)

type keepaliveResponse struct {
	Status   string `json:"status"`
	Postgres string `json:"postgres"`
	Redis    string `json:"redis"`
}

type keepaliveDependencyResult struct {
	name string
	ok   bool
}

func (h *Handler) KeepaliveHandler(c echo.Context) error {
	start := time.Now()
	authorized, _ := contextkey.KeepaliveAuthorizedFromContext(c.Request().Context())
	postgresStatus := "skipped"
	redisStatus := "skipped"
	httpStatus := http.StatusUnauthorized

	defer func() {
		log.Printf(
			"keepalive request: method=%s path=%s authorized=%t postgres=%s redis=%s http_status=%d duration=%s",
			c.Request().Method,
			c.Request().URL.Path,
			authorized,
			postgresStatus,
			redisStatus,
			httpStatus,
			time.Since(start),
		)
	}()

	if !authorized {
		return c.NoContent(httpStatus)
	}

	postgresStatus = "down"
	redisStatus = "down"

	ctx, cancel := context.WithTimeout(c.Request().Context(), keepaliveTimeout)
	defer cancel()

	results := make(chan keepaliveDependencyResult, 2)
	deadline, _ := ctx.Deadline()
	go func() {
		var result int
		err := h.Pool.QueryRow(ctx, "SELECT 1").Scan(&result)
		completedWithinDeadline := !time.Now().After(deadline)
		results <- keepaliveDependencyResult{
			name: "postgres",
			ok:   err == nil && result == 1 && completedWithinDeadline,
		}
	}()
	go func() {
		err := h.ValkeyClient.Set(ctx, keepaliveKey, keepaliveValue, keepaliveTTL).Err()
		completedWithinDeadline := !time.Now().After(deadline)
		results <- keepaliveDependencyResult{
			name: "redis",
			ok:   err == nil && completedWithinDeadline,
		}
	}()

	recordResult := func(result keepaliveDependencyResult) {
		status := "down"
		if result.ok {
			status = "ok"
		}

		switch result.name {
		case "postgres":
			postgresStatus = status
		case "redis":
			redisStatus = status
		}
	}

	remaining := cap(results)
	for remaining > 0 {
		select {
		case result := <-results:
			recordResult(result)
			remaining--
		case <-ctx.Done():
			for remaining > 0 {
				select {
				case result := <-results:
					recordResult(result)
					remaining--
				default:
					remaining = 0
				}
			}
		}
	}

	responseStatus := "degraded"
	httpStatus = http.StatusServiceUnavailable
	if postgresStatus == "ok" && redisStatus == "ok" {
		responseStatus = "ok"
		httpStatus = http.StatusOK
	}

	return c.JSON(httpStatus, keepaliveResponse{
		Status:   responseStatus,
		Postgres: postgresStatus,
		Redis:    redisStatus,
	})
}
