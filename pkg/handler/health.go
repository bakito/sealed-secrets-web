package handler

import (
	"github.com/gin-gonic/gin"
)

func (h *Handler) Health(c *gin.Context) {
	c.Writer.WriteHeader(200)
	_, _ = c.Writer.Write([]byte("OK"))
}
