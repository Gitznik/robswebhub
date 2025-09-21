package main

import (
	"context"
	"encoding/gob"
	"encoding/hex"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"

	"github.com/gin-gonic/gin"
	"github.com/gitznik/robswebhub/internal/auth"
	"github.com/gitznik/robswebhub/internal/config"
	"github.com/gitznik/robswebhub/internal/database"
	"github.com/gitznik/robswebhub/internal/handlers"
	"github.com/gitznik/robswebhub/internal/middleware"
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
	router := setupRouter(cfg, queries, auth)

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

func setupRouter(cfg *config.Config, queries *database.Queries, auth *auth.Authenticator) *gin.Engine {
	// Set Gin mode based on environment
	if os.Getenv("APP_ENVIRONMENT") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()
	router.Use(sentrygin.New(sentrygin.Options{Repanic: true}))

	// Setup Auth
	gob.Register(map[string]interface{}{})

	session_auth_key, err := hex.DecodeString(cfg.Auth.CookieAuthKey)
	if err != nil {
		log.Fatalf("Could not decode auth key")
	}
	session_encryption_key, err := hex.DecodeString(cfg.Auth.CookieEncryptionKey)
	if err != nil {
		log.Fatalf("Could not decode encryption key")
	}
	store := cookie.NewStore(session_auth_key, session_encryption_key)
	router.Use(sessions.Sessions("auth-session", store))
	router.Use(middleware.LoginStatus)

	// Serve static files
	router.Static("/static", "./static")
	router.Static("/images", "./static/images")
	router.StaticFile("/favicon.ico", "./static/images/favicon.ico")

	// Create handlers
	h := handlers.New(queries, cfg)

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

	router.GET("/login", h.MakeLogin(auth))
	router.GET("/logout", h.Logout)
	router.GET("/callback", h.MakeCallback(auth))
	scoresV2 := router.Group("/scoresV2")
	scoresV2.Use(middleware.IsAuthenticated)
	{
		scoresV2.GET("", h.ScoresIndex)
	}

	return router
}
