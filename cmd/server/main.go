package main

import (
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
	db "github.com/rakshitg600/notakto-solo/db/generated"
	"github.com/rakshitg600/notakto-solo/handlers"
	"github.com/rakshitg600/notakto-solo/routes"
	"github.com/rakshitg600/notakto-solo/valkey"
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
	conn.SetMaxOpenConns(20)
	conn.SetMaxIdleConns(10)
	conn.SetConnMaxLifetime(30 * time.Minute)
	conn.SetConnMaxIdleTime(5 * time.Minute)

	if err := conn.Ping(); err != nil {
		log.Fatal("failed to connect to database:", err)
	}
	defer conn.Close()

	if err := conn.Ping(); err != nil {
		log.Fatal("failed to connect to database:", err)
	}

	queries := db.New(conn)
	handler := handlers.NewHandler(queries)

	valkeyURL := os.Getenv("VALKEY_URL")
	if valkeyURL == "" {
		panic("VALKEY_URL is not set")
	}
	valkeyClient := valkey.NewClient(
		valkeyURL,
		"", // password
		0,  // default DB
	)
	e := echo.New()
	routes.RegisterRoutes(e, handler, valkeyClient)
	e.Logger.Fatal(e.Start(":1323"))
}
