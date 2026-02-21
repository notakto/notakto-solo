package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rakshitg600/notakto-solo/lua"
	"github.com/redis/go-redis/v9"
)

// IPRateLimitMiddleware enforces the given requests/min limit per IP address.
func IPRateLimitMiddleware(rdb *redis.Client, limit int) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip := c.RealIP()
			allowed, retryAfter, err := checkRateLimit(c, rdb, "rl:ip", ip, limit)
			if err != nil {
				c.Logger().Errorf("ip-rate-limit: %v", err)
				// fail open: let the request through if Valkey is down
				return next(c)
			}
			if !allowed {
				return rejectRateLimit(c, retryAfter)
			}
			return next(c)
		}
	}
}

// checkRateLimit atomically checks and increments a sliding window counter.
// Returns whether the request is allowed, and retryAfter seconds if rejected.
func checkRateLimit(c echo.Context, rdb *redis.Client, keyPrefix, id string, limit int) (allowed bool, retryAfter int, err error) {
	now := time.Now().Unix()
	windowID := now / 60
	elapsed := now % 60
	weight := 1.0 - float64(elapsed)/60.0
	ttlSecs := 60 - elapsed

	prevKey := fmt.Sprintf("%s:%s:%d", keyPrefix, id, windowID-1)
	currKey := fmt.Sprintf("%s:%s:%d", keyPrefix, id, windowID)

	ctx := c.Request().Context()
	result, err := lua.RateLimit.Run(ctx, rdb, []string{prevKey, currKey},
		fmt.Sprintf("%.6f", weight), limit, ttlSecs,
	).Int64Slice()
	if err != nil {
		return false, 0, err
	}

	if result[0] == -1 {
		return false, int(result[1]), nil
	}
	return true, 0, nil
}

func rejectRateLimit(c echo.Context, retryAfter int) error {
	c.Response().Header().Set("Retry-After", fmt.Sprintf("%d", retryAfter))
	return c.JSON(http.StatusTooManyRequests, map[string]string{
		"message": "Rate limit exceeded, try again later",
	})
}
