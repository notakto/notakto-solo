package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rakshitg600/notakto-solo/contextkey"
	"github.com/rakshitg600/notakto-solo/lua"
	"github.com/redis/go-redis/v9"
)

const (
	lockTTL       = 10 * time.Second
	lockRetryWait = 50 * time.Millisecond
)

// UIDLockMiddleware returns middleware that serializes requests per UID using
// a distributed lock in Valkey/Redis. Must run after FirebaseAuthMiddleware.
func UIDLockMiddleware(rdb *redis.Client) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()

			uid, ok := contextkey.UIDFromContext(ctx)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "Missing UID")
			}

			lockKey := "lock:uid:" + uid // The key in key val pairs

			// value in key value pair, this value is unique to each request ==> uid,requestIdentifier map
			nonce := make([]byte, 16)
			if _, err := rand.Read(nonce); err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate lock nonce")
			}
			lockVal := hex.EncodeToString(nonce)
			ticker := time.NewTicker(lockRetryWait)
			defer ticker.Stop()

			// Retry acquiring the lock until the request context expires.
			for {
				ok, err := rdb.SetNX(ctx, lockKey, lockVal, lockTTL).Result()
				if err != nil {
					return echo.NewHTTPError(http.StatusInternalServerError, "Lock service unavailable")
				}
				if ok {
					break
				}

				select {
				case <-ctx.Done():
					return echo.NewHTTPError(http.StatusTooManyRequests, "Could not acquire lock, try again later")
				case <-ticker.C:
				}
			}

			// Ensure unlock runs after the handler, even on panic.
			// Use context.Background() because the request ctx may be canceled.
			defer func() {
				unlockCtx, unlockCancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer unlockCancel()
				if err := lua.Unlock.Run(unlockCtx, rdb, []string{lockKey}, lockVal).Err(); err != nil {
					c.Logger().Errorf("uid-lock: failed to unlock %s: %v", lockKey, err)
				}
			}()

			return next(c)
		}
	}
}
