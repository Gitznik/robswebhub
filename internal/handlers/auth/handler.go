package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/gitznik/robswebhub/internal/auth"
	"github.com/gitznik/robswebhub/internal/config"
)

type Handler struct {
	cfg           *config.Config
	authenticator *auth.Authenticator
}

func New(cfg *config.Config, authenticator *auth.Authenticator) *Handler {
	return &Handler{
		cfg:           cfg,
		authenticator: authenticator,
	}
}

func (h *Handler) RegisterRoute(rg *gin.RouterGroup) {
	rg.GET("/login", h.Login)
	rg.GET("/logout", h.Logout)
	rg.GET("/callback", h.Callback)
}
