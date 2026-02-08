package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	firebase "firebase.google.com/go/v4"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"google.golang.org/api/option"

	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/rakshitg600/notakto-solo/config"
	appMiddleware "github.com/rakshitg600/notakto-solo/middleware"
	"github.com/rakshitg600/notakto-solo/routes"
)

// main initializes environment and dependencies, configures the Echo HTTP server and middleware, connects to the database, registers routes, and starts listening on the configured port.
func main() {
	if err := config.InitEnv(); err != nil {
		log.Fatal("Failed to load environment variables:", err)
	}

	// Initialize Firebase Admin SDK (ServiceAccount type avoids deprecated WithCredentialsJSON)
	credJSON := config.MustGetEnv("FIREBASE_CREDENTIALS_JSON")
	firebaseApp, err := firebase.NewApp(context.Background(), nil, option.WithAuthCredentialsJSON(option.ServiceAccount, []byte(credJSON)))
	if err != nil {
		log.Fatal("failed to initialize Firebase app:", err)
	}
	authClient, err := firebaseApp.Auth(context.Background())
	if err != nil {
		log.Fatal("failed to get Firebase Auth client:", err)
	}

	e := echo.New()
	e.Use(appMiddleware.CORSMiddleware)
	e.Use(echoMiddleware.ContextTimeout(5 * time.Second)) // only for http, not for websockets,etc
	// Set server timeouts
	e.Server.ReadTimeout = 5 * time.Second   //Max time to read the entire incoming request (headers + body)
	e.Server.WriteTimeout = 10 * time.Second //Max time to write the response back to client which includes handler execution + response write
	e.Server.IdleTimeout = 60 * time.Second  //Max time to wait for the next request when keep-alives are enabled
	dbURL := config.MustGetEnv("DATABASE_URL")

	poolConfig, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		log.Fatal("failed to parse DATABASE_URL:", err)
	}
	// Pool tuning (adjust as needed)
	poolConfig.MaxConns = 10
	poolConfig.MinConns = 2
	poolConfig.MaxConnLifetime = 30 * time.Minute
	poolConfig.MaxConnIdleTime = 5 * time.Minute
	poolConfig.HealthCheckPeriod = 30 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		log.Fatal("failed to create pgx pool:", err)
	}

	if err := pool.Ping(ctx); err != nil {
		log.Fatal("failed to connect to database:", err)
	}

	routes.SetupRoutes(e, pool, authClient)
	port := config.MustGetEnv("PORT")
	serverErr := make(chan error, 1)
	go func() {
		if err := e.Start(":" + port); err != nil && err != http.ErrServerClosed {
			log.Println("server error:", err)
			serverErr <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-quit:
		log.Println("shutdown signal received")
	case err := <-serverErr:
		log.Println("server failed, initiating shutdown:", err)
	}
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := e.Shutdown(shutdownCtx); err != nil {
		log.Println("server shutdown failed:", err)
	}
	log.Println("closing database pool...")
	pool.Close()
	log.Println("server exited gracefully")
}
