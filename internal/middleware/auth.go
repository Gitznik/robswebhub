package middleware

import (
	"log"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// IsAuthenticated is a middleware that checks if
// the user has already been authenticated previously.
func IsAuthenticated(ctx *gin.Context) {
	if sessions.Default(ctx).Get("profile") == nil {
		ctx.Redirect(http.StatusSeeOther, "/")
	} else {
		ctx.Next()
	}
}

var LoginKey = "IsLoggedIn"

func LoginStatus(ctx *gin.Context) {
	// FIXME: solve privilege persistence
	log.Printf("Is logged in: %v", sessions.Default(ctx).Get("profile") != nil)
	ctx.Set(LoginKey, sessions.Default(ctx).Get("profile") != nil)
	ctx.Next()
}
