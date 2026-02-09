package wellknown

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct{}

func New() *Handler {
	return &Handler{}
}

func (h *Handler) RegisterRoute(rg *gin.RouterGroup) {
	rg.GET("/matrix/server", h.MatrixServer)
	rg.GET("/matrix/client", h.MatrixClient)
}

func (h *Handler) MatrixServer(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.JSON(http.StatusOK, gin.H{
		"m.server": "matrix.robswebhub.net:443",
	})
}

func (h *Handler) MatrixClient(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.JSON(http.StatusOK, gin.H{
		"m.homeserver": gin.H{
			"base_url": "https://matrix.robswebhub.net",
		},
	})
}
