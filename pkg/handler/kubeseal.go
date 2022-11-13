package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) KubeSeal(c *gin.Context) {
	contentType := c.NegotiateFormat(gin.MIMEJSON, gin.MIMEYAML)
	var outputFormat string
	if contentType == "" {
		c.AbortWithStatus(http.StatusNotAcceptable)
		return
	} else if contentType == gin.MIMEJSON {
		outputFormat = "json"
	} else {
		outputFormat = "yaml"
	}

	ss, err := h.sealer.Seal(outputFormat, c.Request.Body)
	if err != nil {
		log.Printf("Error in %s: %v\n", Sanitize(c.Request.URL.Path), err)
		c.Negotiate(http.StatusInternalServerError, gin.Negotiate{
			Offered: []string{gin.MIMEJSON, gin.MIMEYAML},
			Data:    gin.H{"error": err.Error()},
		})
		return
	}

	c.Data(http.StatusOK, contentType, ss)
}
