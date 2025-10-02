package scorekeeper

import (
	"github.com/gin-gonic/gin"
	"github.com/gitznik/robswebhub/internal/database"
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
	{
		rg.GET("", h.ScoresIndex)
		rg.POST("/single", h.ScoresSingle)
		rg.POST("/batch", h.ScoresBatch)
		rg.GET("/single-form", h.SingleScoreForm)
		rg.GET("/batch-form", h.BatchScoreForm)
		rg.GET("/chart/:id", h.ScoresChart)
	}
}
