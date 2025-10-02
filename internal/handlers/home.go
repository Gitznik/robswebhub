package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gitznik/robswebhub/internal/config"
	"github.com/gitznik/robswebhub/internal/database"
	"github.com/gitznik/robswebhub/internal/middleware"
	"github.com/gitznik/robswebhub/internal/templates/pages"
)

type Handler struct {
	queries *database.Queries
	cfg     *config.Config
}

func New(queries *database.Queries, cfg *config.Config) *Handler {
	return &Handler{
		queries: queries,
		cfg:     cfg,
	}
}

func (h *Handler) Home(c *gin.Context) {
	redirectError := c.Query("error")
	component := pages.Home(redirectError, c.GetBool(middleware.LoginKey))
	if err := component.Render(c.Request.Context(), c.Writer); err != nil {
		_ = c.Error(errors.New("Failed to render page"))
		return
	}
}

func (h *Handler) HomeHead(c *gin.Context) {
	c.Status(http.StatusOK)
}
