package middleware

import (
	"net/http"
	"strings"

	"github.com/rakshitg600/notakto-solo/functions"

	"github.com/labstack/echo/v4"
)

func FirebaseAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
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
		uid, _, _, _, err := functions.VerifyFirebaseToken(ctx, idToken) //underscore here means ignore photo,name,email for middleware
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token")
		}
		c.Set("uid", uid)
		c.Set("idToken", idToken)
		return next(c)
	}
}
