package middleware

import (
	"net/http"
	"sync/atomic"
	"time"

	"github.com/labstack/echo/v4"
)

// CooldownMiddleware rejects requests if less than the given duration has
// passed since the last successful request. Pure in-memory, zero dependencies.
func CooldownMiddleware(cooldown time.Duration) echo.MiddlewareFunc {
	var lastUnixNano atomic.Int64

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			now := time.Now().UnixNano()
			prev := lastUnixNano.Load()

			if prev != 0 && now-prev < cooldown.Nanoseconds() {
				return c.JSON(http.StatusTooManyRequests, map[string]string{
					"message": "Too many requests, try again later",
				})
			}

			lastUnixNano.Store(now)
			return next(c)
		}
	}
}
