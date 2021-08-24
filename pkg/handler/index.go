package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/ricoberger/sealed-secrets-web/pkg/marshal"
	"net/http"
)

type Handler struct {
	seal               func(secret string) ([]byte, error)
	indexHTML          string
	disableLoadSecrets bool
	marshaller         marshal.Marshaller
}

func New(indexHTML string, marshaller marshal.Marshaller, sealer func(secret string) ([]byte, error), disableLoadSecrets bool) *Handler {

	return &Handler{
		marshaller:         marshaller,
		seal:               sealer,
		indexHTML:          indexHTML,
		disableLoadSecrets: disableLoadSecrets,
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
