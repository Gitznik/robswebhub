package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/getsentry/sentry-go"

	"github.com/gitznik/robswebhub/internal/auth"
	"github.com/gitznik/robswebhub/internal/config"
	"github.com/gitznik/robswebhub/internal/database"
	"github.com/gitznik/robswebhub/internal/router"
)

func main() {
	log.Print("Starting up!")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := sentry.Init(sentry.ClientOptions{
		Dsn:                cfg.Telemetry.SentryDSN,
		EnableTracing:      true,
		TracesSampleRate:   1.0,
		EnableLogs:         true,
		Environment:        cfg.Application.Environment,
		IgnoreTransactions: []string{"HEAD /"},
		SendDefaultPII:     true,
	}); err != nil {
		log.Fatalf("Sentry initialization failed: %v\n", err)
	}
	log.Print("Started Sentry")
	defer sentry.Flush(2 * time.Second)

	// Connect to database
	db, err := database.Connect(cfg.Database.ConnectionString)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Print("Established database connection")

	// Run migrations
	if err := database.RunMigrations(cfg.Database.ConnectionString); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Print("Ran migrations")

	// Create queries instance
	queries := database.New(db)

	// Setup Gin router
	auth, err := auth.New(&cfg.Auth)
	if err != nil {
		log.Fatalf("Could not setup authenticator: %v", err)
	}
	router := router.SetupRouter(cfg, queries, auth)

	// Create HTTP server
	srv := &http.Server{
		Addr:    cfg.Address(),
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Starting server on %s", cfg.Address())
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Print("Server exiting")
}
