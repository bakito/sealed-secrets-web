package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/bakito/sealed-secrets-web/pkg/marshal"
	"github.com/bakito/sealed-secrets-web/pkg/version"
	"net/http"
)

type Handler struct {
	seal               func(secret string) ([]byte, error)
	indexHTML          string
	disableLoadSecrets bool
	marshaller         marshal.Marshaller
}

func New(indexHTML string, marshaller marshal.Marshaller, sealer func(secret string) ([]byte, error)) *Handler {
	return &Handler{
		marshaller: marshaller,
		seal:       sealer,
		indexHTML:  indexHTML,
	}
}

func (h *Handler) Index(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/html")
	c.String(http.StatusOK, h.indexHTML)
}

func (h *Handler) RedirectToIndex(ctx *gin.Context) {
	ctx.Redirect(http.StatusMovedPermanently, "/")
	ctx.Abort()
}

func (h *Handler) Version(c *gin.Context) {
	c.JSONP(http.StatusOK, map[string]string{"version": version.Version, "build": version.Build})
}
