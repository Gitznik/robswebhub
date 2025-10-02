package handlers

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/gitznik/robswebhub/internal/middleware"
	"github.com/gitznik/robswebhub/internal/templates/pages"
)

func (h *Handler) About(c *gin.Context) {
	component := pages.About(c.GetBool(middleware.LoginKey))
	if err := component.Render(c.Request.Context(), c.Writer); err != nil {
		_ = c.Error(errors.New("Failed to render page"))
		return
	}
}
