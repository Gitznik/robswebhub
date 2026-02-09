package router

import (
	"log"

	sentrygin "github.com/getsentry/sentry-go/gin"

	"github.com/gin-gonic/gin"
	"github.com/gitznik/robswebhub/internal/auth"
	"github.com/gitznik/robswebhub/internal/config"
	"github.com/gitznik/robswebhub/internal/database"
	"github.com/gitznik/robswebhub/internal/handlers"
	authhandler "github.com/gitznik/robswebhub/internal/handlers/auth"
	gameshandler "github.com/gitznik/robswebhub/internal/handlers/games"
	scorehandler "github.com/gitznik/robswebhub/internal/handlers/scorekeeper"
	wellknownhandler "github.com/gitznik/robswebhub/internal/handlers/wellknown"
	"github.com/gitznik/robswebhub/internal/middleware"
	"github.com/gitznik/robswebhub/internal/sessions"
)

func SetupRouter(cfg *config.Config, queries *database.Queries, authenticator *auth.Authenticator) *gin.Engine {
	if cfg.Application.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()
	router.Use(sentrygin.New(sentrygin.Options{Repanic: true}))
	router.Use(middleware.ErrorHandler)

	sessionMiddleware, err := sessions.SetupSessionMiddleware(cfg.Auth.CookieAuthKey, cfg.Auth.CookieEncryptionKey)
	if err != nil {
		log.Fatalf("Could not attach session information: %v", err)
	}
	router.Use(sessionMiddleware)
	router.Use(middleware.LoginStatus)

	router.Static("/static", "./static")
	router.Static("/images", "./static/images")
	router.StaticFile("/favicon.ico", "./static/images/favicon.ico")

	rg := router.Group("")
	h := handlers.New(queries, cfg)
	h.RegisterRoute(rg)

	skh := scorehandler.New(queries)
	scores := router.Group("/scores")
	skh.RegisterRoute(scores)

	ah := authhandler.New(cfg, authenticator)
	ah.RegisterRoute(rg)

	gkh := gameshandler.New(queries)
	gamekeeper := router.Group("/gamekeeper")
	gkh.RegisterRoute(gamekeeper)

	wkh := wellknownhandler.New()
	wellknown := router.Group("/.well-known")
	wkh.RegisterRoute(wellknown)

	return router
}
