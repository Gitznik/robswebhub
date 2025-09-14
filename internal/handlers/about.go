package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gitznik/robswebhub/internal/templates/pages"
)

func (h *Handler) About(c *gin.Context) {
	component := pages.About()
	if err := component.Render(c.Request.Context(), c.Writer); err != nil {
		c.String(http.StatusInternalServerError, "Failed to render page")
		return
	}
}
