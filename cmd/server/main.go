package main

import (
	"database/sql"
	"log"

	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"

	"github.com/rakshitg600/notakto-solo/config"
	db "github.com/rakshitg600/notakto-solo/db/generated"
	"github.com/rakshitg600/notakto-solo/handlers"
	"github.com/rakshitg600/notakto-solo/middleware"
	"github.com/rakshitg600/notakto-solo/routes"
)

func main() {
	// Init config once
	if err := config.InitEnv(); err != nil {
		log.Fatal("Failed to load environment variables:", err)
	}

	e := echo.New()
	e.Use(middleware.CORSMiddleware)

	dbURL := config.MustGetEnv("DATABASE_URL")

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

	routes.RegisterRoutes(e, handler)

	port, ok := config.GetEnv("PORT")
	if !ok {
		log.Fatal("PORT is not set")
	}

	log.Println("Server starting on port", port)
	log.Fatal(e.Start(":" + port))
}
