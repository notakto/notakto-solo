package middleware

import (
	"net/http"
	"strings"

	"firebase.google.com/go/v4/auth"
	"github.com/labstack/echo/v4"
	"github.com/rakshitg600/notakto-solo/usecase"
)

// FirebaseAuthMiddleware returns an Echo middleware that validates a Firebase ID
// token from the Authorization header using the provided auth client.
// On success it sets "uid" and "idToken" in the Echo context.
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
			c.Set("uid", uid)
			c.Set("idToken", idToken)
			return next(c)
		}
	}
}
