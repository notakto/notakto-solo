package main

import (
	"database/sql"
	"log"
	"time"

	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"

	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/rakshitg600/notakto-solo/config"
	db "github.com/rakshitg600/notakto-solo/db/generated"
	"github.com/rakshitg600/notakto-solo/handlers"
	appMiddleware "github.com/rakshitg600/notakto-solo/middleware"
	"github.com/rakshitg600/notakto-solo/routes"
)

func main() {
	if err := config.InitEnv(); err != nil {
		log.Fatal("Failed to load environment variables:", err)
	}

	e := echo.New()
	e.Use(appMiddleware.CORSMiddleware)
	e.Use(echoMiddleware.ContextTimeout(5 * time.Second)) // only for http, not for websockets,etc
	// Set server timeouts
	e.Server.ReadTimeout = 5 * time.Second   //Max time to read the entire incoming request (headers + body)
	e.Server.WriteTimeout = 10 * time.Second //Max time to write the response back to client which includes handler execution + response write
	e.Server.IdleTimeout = 60 * time.Second  //Max time to wait for the next request when keep-alives are enabled
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

	port := config.MustGetEnv("PORT")

	log.Fatal(e.Start(":" + port))
}
