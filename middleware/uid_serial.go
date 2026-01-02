package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rakshitg600/notakto-solo/concurrency"
)

func UIDSerial(guard concurrency.Guard) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			uid, ok := c.Get("uid").(string)
			if !ok || uid == "" {
				return next(c)
			}

			res, err := guard.Acquire(c.Request().Context(), uid)
			if err != nil {
				return next(c) // fail-open
			}

			if !res.Acquired {
				return c.JSON(http.StatusTooManyRequests, map[string]string{
					"error": "another request in progress",
				})
			}

			defer guard.Release(c.Request().Context(), uid)

			return next(c)
		}
	}
}
