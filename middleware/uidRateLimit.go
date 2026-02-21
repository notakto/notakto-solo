package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rakshitg600/notakto-solo/contextkey"
	"github.com/redis/go-redis/v9"
)

// UIDRateLimitMiddleware enforces the given requests/min limit per authenticated user.
// Must run after FirebaseAuthMiddleware.
func UIDRateLimitMiddleware(rdb *redis.Client, limit int) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			uid, ok := contextkey.UIDFromContext(c.Request().Context())
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "Missing UID")
			}

			allowed, retryAfter, err := checkRateLimit(c, rdb, "rl:uid", uid, limit)
			if err != nil {
				c.Logger().Errorf("uid-rate-limit: %v", err)
				return next(c)
			}
			if !allowed {
				return rejectRateLimit(c, retryAfter)
			}
			return next(c)
		}
	}
}
