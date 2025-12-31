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

	// Set database connection pool settings
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

	// Set server timeouts
	e.Server.ReadTimeout = 5 * time.Second   //Max time to read the entire incoming request (headers + body)
	e.Server.WriteTimeout = 10 * time.Second //Max time to write the response back to client which includes handler execution + response write
	e.Server.IdleTimeout = 60 * time.Second  //Max time to wait for the next request when keep-alives are enabled

	routes.RegisterRoutes(e, handler, valkeyClient)
	e.Logger.Fatal(e.Start(":1323"))
}
