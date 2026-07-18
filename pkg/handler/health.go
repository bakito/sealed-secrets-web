package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (*Handler) Health(c *gin.Context) {
	c.String(http.StatusOK, "OK")
}
