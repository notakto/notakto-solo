package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rakshitg600/notakto-solo/types"
	"github.com/redis/go-redis/v9"
)

var rateLimitLua = redis.NewScript(`
local key = KEYS[1]
local limit = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local now = tonumber(ARGV[3])

local current = redis.call("INCR", key)

if current == 1 then
	redis.call("EXPIRE", key, window)
end

local ttl = redis.call("TTL", key)

if current > limit then
	return {0, ttl}
end

return {limit - current, ttl}
`)

func RateLimit(redisClient *redis.Client, cfg types.RateLimiterConfig) echo.MiddlewareFunc {
	if cfg.Prefix == "" {
		cfg.Prefix = "ratelimit:"
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip := c.RealIP()
			key := cfg.Prefix + ip

			ctx := c.Request().Context()

			res, err := rateLimitLua.Run(
				ctx,
				redisClient,
				[]string{key},
				cfg.Limit,
				int(cfg.Window.Seconds()),
				time.Now().Unix(),
			).Result()

			if err != nil {
				// Fail-open (important for prod)
				return next(c)
			}

			values := res.([]interface{})
			remaining := values[0].(int64)
			reset := values[1].(int64)

			c.Response().Header().Set("X-RateLimit-Limit", strconv.Itoa(cfg.Limit))
			c.Response().Header().Set("X-RateLimit-Remaining", strconv.FormatInt(remaining, 10))
			c.Response().Header().Set("X-RateLimit-Reset", strconv.FormatInt(reset, 10))

			if remaining <= 0 {
				return c.JSON(http.StatusTooManyRequests, map[string]string{
					"error": "rate limit exceeded",
				})
			}

			return next(c)
		}
	}
}
