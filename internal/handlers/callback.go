package handlers

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/gitznik/robswebhub/internal/auth"
)

func (h *Handler) MakeCallback(auth *auth.Authenticator) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		if c.Query("state") != session.Get("state") {
			c.String(http.StatusBadRequest, "Invalid state parameter.")
			return
		}

		// Exchange an authorization code for a token.
		token, err := auth.Exchange(c.Request.Context(), c.Query("code"))
		if err != nil {
			c.String(http.StatusUnauthorized, "Failed to convert an authorization code into a token.")
			return
		}

		idToken, err := auth.VerifyIDToken(c.Request.Context(), token)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to verify ID Token.")
			return
		}

		var profile map[string]interface{}
		if err := idToken.Claims(&profile); err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		session.Set("access_token", token.AccessToken)
		session.Set("profile", profile)
		if err := session.Save(); err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		// Redirect to logged in page.
		c.Redirect(http.StatusTemporaryRedirect, "/")
	}
}
