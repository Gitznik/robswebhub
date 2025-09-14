package main

import (
	"context"
	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gitznik/robswebhub/internal/config"
	"github.com/gitznik/robswebhub/internal/database"
	"github.com/gitznik/robswebhub/internal/handlers"
)

func main() {

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := sentry.Init(sentry.ClientOptions{
		Dsn:			cfg.Telemetry.SentryDSN,
		EnableTracing:    true,
		TracesSampleRate: 1.0,
		EnableLogs:       true,
		Environment:      cfg.Application.Environment,
	}); err != nil {
		log.Fatalf("Sentry initialization failed: %v\n", err)
	}
	defer sentry.Flush(2 * time.Second)

	ctx := context.Background()
	logger := sentry.NewLogger(ctx)
	// Connect to database
	db, err := database.Connect(cfg.Database.ConnectionString)
	if err != nil {
		logger.Fatal().Emitf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := database.RunMigrations(cfg.Database.ConnectionString); err != nil {
		logger.Fatal().Emitf("Failed to run migrations: %v", err)
	}

	// Create queries instance
	queries := database.New(db)

	// Setup Gin router
	router := setupRouter(cfg, queries)

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
		logger.Fatal().Emitf("Server forced to shutdown: %v", err)
	}

	logger.Info().Emit("Server exiting")
}

func setupRouter(cfg *config.Config, queries *database.Queries) *gin.Engine {
	// Set Gin mode based on environment
	if os.Getenv("APP_ENVIRONMENT") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()
	router.Use(sentrygin.New(sentrygin.Options{Repanic: true}))

	// Serve static files
	router.Static("/static", "./static")
	router.Static("/images", "./static/images")
	router.StaticFile("/favicon.ico", "./static/images/favicon.ico")

	// Create handlers
	h := handlers.New(queries)

	// Routes
	router.GET("/", h.Home)
	router.HEAD("/", h.HomeHead)
	router.GET("/about", h.About)

	// Scores routes
	scores := router.Group("/scores")
	{
		scores.GET("", h.ScoresIndex)
		scores.POST("/single", h.ScoresSingle)
		scores.POST("/batch", h.ScoresBatch)
		scores.GET("/single-form", h.SingleScoreForm)
		scores.GET("/batch-form", h.BatchScoreForm)
		scores.GET("/chart/:id", h.ScoresChart)
	}

	return router
}
