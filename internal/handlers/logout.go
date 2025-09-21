package handlers

import (
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

func (h *Handler) Logout(c *gin.Context) {
	c.SetCookie(
		"auth-session",       // cookie name
		"",                   // empty value
		-1,                   // max age -1 means delete now
		"/",                  // path
		"",                   // domain (empty = current domain)
		c.Request.TLS != nil, // secure flag (true if HTTPS)
		false,                // httpOnly
	)

	logoutUrl, err := url.Parse("https://" + h.cfg.Auth.Auth0Domain + "/v2/logout")
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	scheme := "http"
	if h.cfg.Application.Environment != "dev" {
		scheme = "https"
	}

	returnTo, err := url.Parse(scheme + "://" + c.Request.Host)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	parameters := url.Values{}
	parameters.Add("returnTo", returnTo.String())
	parameters.Add("client_id", h.cfg.Auth.Auth0ClientId)
	logoutUrl.RawQuery = parameters.Encode()

	c.Redirect(http.StatusTemporaryRedirect, logoutUrl.String())
}
