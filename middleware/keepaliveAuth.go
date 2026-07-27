package middleware

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"

	"github.com/labstack/echo/v4"
	"github.com/rakshitg600/notakto-solo/contextkey"
)

// KeepaliveAuthMiddleware validates the protected keepalive token and records
// the result in the request context for the handler to log and enforce.
func KeepaliveAuthMiddleware(expectedToken string) echo.MiddlewareFunc {
	expectedTokenHash := sha256.Sum256([]byte(expectedToken))

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			providedToken := c.Request().Header.Get("X-Keepalive-Token")
			providedTokenHash := sha256.Sum256([]byte(providedToken))
			tokenMatches := subtle.ConstantTimeCompare(providedTokenHash[:], expectedTokenHash[:]) == 1
			authorized := providedToken != "" && tokenMatches

			ctx := context.WithValue(c.Request().Context(), contextkey.KeepaliveAuthorized, authorized)
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	}
}
