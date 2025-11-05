package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
	db "github.com/rakshitg600/notakto-solo/db/generated"
	"github.com/rakshitg600/notakto-solo/handlers"
	"github.com/rakshitg600/notakto-solo/routes"
)

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	conn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	if err := conn.Ping(); err != nil {
		log.Fatal("failed to connect to database:", err)
	}

	queries := db.New(conn)
	handler := handlers.NewHandler(queries)

	e := echo.New()
	// ✅ Enable CORS for frontend (Next.js at localhost:3000)
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			origin := c.Request().Header.Get("Origin")

			// Allow local dev and any Netlify preview or staging URLs
			if origin == "http://localhost:3000" ||
				origin == "https://staging-notakto.netlify.app" ||
				(strings.HasSuffix(origin, "--staging-notakto.netlify.app") &&
					strings.HasPrefix(origin, "https://deploy-preview-")) {

				c.Response().Header().Set("Access-Control-Allow-Origin", origin)
				c.Response().Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				c.Response().Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
				c.Response().Header().Set("Access-Control-Allow-Credentials", "true")
			}

			// Handle preflight requests explicitly
			if c.Request().Method == http.MethodOptions {
				return c.NoContent(http.StatusNoContent)
			}

			return next(c)
		}
	})

	routes.RegisterRoutes(e, handler)
	e.Logger.Fatal(e.Start(":1323"))
}
