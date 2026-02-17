package wellknown

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	proxy *httputil.ReverseProxy
}

func New() *Handler {
	target, err := url.Parse("https://matrix.robswebhub.net")
	if err != nil {
		log.Fatalf("Failed to parse matrix upstream URL: %v", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, err error) {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusBadGateway)
		_, _ = writer.Write([]byte(`{"error":"matrix well-known upstream unavailable"}`))
	}

	return &Handler{proxy: proxy}
}

func (h *Handler) RegisterRoute(rg *gin.RouterGroup) {
	rg.GET("/matrix/server", h.MatrixServer)
	rg.GET("/matrix/client", h.MatrixClient)
}

func (h *Handler) MatrixServer(c *gin.Context) {
	h.proxyWellKnown(c, "/.well-known/matrix/server")
}

func (h *Handler) MatrixClient(c *gin.Context) {
	h.proxyWellKnown(c, "/.well-known/matrix/client")
}

func (h *Handler) proxyWellKnown(c *gin.Context, path string) {
	c.Request.URL.Path = path
	c.Request.URL.RawPath = path
	c.Request.URL.RawQuery = ""
	h.proxy.ServeHTTP(c.Writer, c.Request)
}
