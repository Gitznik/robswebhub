package middleware

import (
	"log"
	"net/http"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/gitznik/robswebhub/internal/auth"
)

// IsAuthenticated is a middleware that checks if
// the user has already been authenticated previously.
func IsAuthenticated(ctx *gin.Context) {
	if sessions.Default(ctx).Get("profile") == nil {
		ctx.Redirect(http.StatusSeeOther, "/login")
	} else {
		ctx.Next()
	}
}

var LoginKey = "IsLoggedIn"

func LoginStatus(ctx *gin.Context) {
	// FIXME: solve privilege persistence
	log.Printf("Is logged in: %v", sessions.Default(ctx).Get("profile") != nil)
	profile := sessions.Default(ctx).Get("profile").(*auth.UserProfile)
	if profile != nil {
		if hub := sentrygin.GetHubFromContext(ctx); hub != nil {
			hub.ConfigureScope(func(scope *sentry.Scope) {
				scope.SetUser(sentry.User{ID: profile.Name})
			})
		}
	}

	ctx.Set(LoginKey, profile != nil)
	ctx.Next()
}
