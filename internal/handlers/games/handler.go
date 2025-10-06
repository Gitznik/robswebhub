package games

import (
	"github.com/gin-gonic/gin"
	"github.com/gitznik/robswebhub/internal/database"
	"github.com/gitznik/robswebhub/internal/middleware"
)

type Handler struct {
	queries *database.Queries
}

func New(queries *database.Queries) *Handler {
	return &Handler{
		queries: queries,
	}
}

func (h *Handler) RegisterRoute(rg *gin.RouterGroup) {
	rg.Use(middleware.IsAuthenticated)
	{
		rg.GET("", h.GamesIndex)
		rg.GET("/signup", h.SignUp)
		rg.POST("/signup", h.DoSignUp)
	}
}
