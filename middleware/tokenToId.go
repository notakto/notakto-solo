package middleware

import (
	"context"
	"net/http"
	"strings"

	"firebase.google.com/go/v4/auth"
	"github.com/labstack/echo/v4"
	"github.com/rakshitg600/notakto-solo/contextkey"
	"github.com/rakshitg600/notakto-solo/usecase"
)

// FirebaseAuthMiddleware validates the Bearer token and injects UID into the request context.
func FirebaseAuthMiddleware(authClient *auth.Client) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "Missing Authorization header")
			}
			if !strings.HasPrefix(authHeader, "Bearer ") {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid Authorization header format")
			}

			ctx := c.Request().Context()
			idToken := authHeader[len("Bearer "):]
			uid, err := usecase.VerifyFirebaseToken(ctx, authClient, idToken)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token")
			}
			req := c.Request().WithContext(context.WithValue(ctx, contextkey.UID, uid))
			c.SetRequest(req)
			return next(c)
		}
	}
}
