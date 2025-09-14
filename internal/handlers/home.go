package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gitznik/robswebhub/internal/database"
	"github.com/gitznik/robswebhub/internal/templates/pages"
)

type Handler struct {
	queries *database.Queries
}

func New(queries *database.Queries) *Handler {
	return &Handler{
		queries: queries,
	}
}

func (h *Handler) Home(c *gin.Context) {
	component := pages.Home()
	if err := component.Render(c.Request.Context(), c.Writer); err != nil {
		c.String(http.StatusInternalServerError, "Failed to render page")
		return
	}
}

func (h *Handler) HomeHead(c *gin.Context) {
	c.Status(http.StatusOK)
}
