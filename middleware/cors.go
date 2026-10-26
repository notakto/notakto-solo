package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

func CORSMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		origin := c.Request().Header.Get("Origin")

		if origin == "http://localhost:3000" ||
			origin == "https://notakto.xyz" ||
			origin == "https://www.notakto.xyz" ||
			origin == "https://notakto.vercel.app" ||
			origin == "https://notakto-website.vercel.app" ||
			origin == "https://notakto-staging.netlify.app" ||
			(strings.HasSuffix(origin, "--notakto-staging.netlify.app") &&
				strings.HasPrefix(origin, "https://deploy-preview-")) {

			h := c.Response().Header()
			h.Set("Access-Control-Allow-Origin", origin)
			h.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			h.Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
			h.Set("Access-Control-Allow-Credentials", "true")
		} else if origin != "" {
			log.Printf("CORS origin not allowed: origin=%q method=%s path=%s", origin, c.Request().Method, c.Request().URL.Path)
		}

		if c.Request().Method == http.MethodOptions {
			return c.NoContent(http.StatusNoContent)
		}

		return next(c)
	}
}
