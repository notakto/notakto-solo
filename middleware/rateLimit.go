package middleware

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/rakshitg600/notakto-solo/ratelimit"
)

type KeyFunc func(echo.Context) (string, bool)

func IPKey(c echo.Context) (string, bool) {
	return c.RealIP(), true
}
func UIDKey(c echo.Context) (string, bool) {
	uid, ok := c.Get("uid").(string)
	return uid, ok
}

func RateLimit(
	limiter ratelimit.Limiter,
	keyFn KeyFunc,
) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			key, ok := keyFn(c)
			if !ok || key == "" {
				return next(c) // fail-open
			}

			res, err := limiter.Allow(c.Request().Context(), key)
			if err != nil {
				return next(c) // fail-open
			}

			if res != nil {
				c.Response().Header().Set(
					"X-RateLimit-Remaining",
					strconv.FormatInt(res.Remaining, 10),
				)
			}

			if !res.Allowed {
				return c.JSON(http.StatusTooManyRequests, map[string]string{
					"error": "rate limit exceeded",
				})
			}

			return next(c)
		}
	}
}
