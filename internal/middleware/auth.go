package middleware

import (
	"log"
	"net/http"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"github.com/gitznik/robswebhub/internal/sessions"
)

// IsAuthenticated is a middleware that checks if
// the user has already been authenticated previously.
func IsAuthenticated(ctx *gin.Context) {
	profile, err := sessions.GetProfile(ctx)
	if profile == nil || err != nil || profile.IsExpired() {
		ctx.Redirect(http.StatusSeeOther, "/login")
	} else {
		ctx.Next()
	}
}

var LoginKey = "IsLoggedIn"

func LoginStatus(ctx *gin.Context) {
	profile, err := sessions.GetProfile(ctx)
	if profile != nil && err != nil {
		if hub := sentrygin.GetHubFromContext(ctx); hub != nil {
			hub.ConfigureScope(func(scope *sentry.Scope) {
				scope.SetUser(sentry.User{ID: profile.Name})
			})
		}
	}

	ctx.Set(LoginKey, profile != nil)
	ctx.Next()
}

func ErrorHandler(ctx *gin.Context) {
	ctx.Next()
	if len(ctx.Errors) > 0 {
		err := ctx.Errors.Last().Err
		log.Printf("Encountered unhandled error: %v", err)

		ctx.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"message": "Internal server error",
		})
	}
}
