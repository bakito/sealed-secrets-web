package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"k8s.io/apimachinery/pkg/runtime"
)

func (h *Handler) KubeSeal(c *gin.Context) {
	outputContentType, outputFormat, done := NegotiateFormat(c)
	if done {
		return
	}

	ss, err := h.sealer.Seal(outputFormat, c.Request.Body)
	if err != nil {
		log.Printf("Error in %s: %v\n", Sanitize(c.Request.URL.Path), err)
		c.Negotiate(http.StatusInternalServerError, gin.Negotiate{
			Offered: []string{gin.MIMEJSON, gin.MIMEYAML2},
			Data:    gin.H{"error": err.Error()},
		})
		return
	}

	c.Data(http.StatusOK, outputContentType, ss)
}

func NegotiateFormat(c *gin.Context) (string, string, bool) {
	contentType := c.NegotiateFormat(gin.MIMEJSON, gin.MIMEYAML2, runtime.ContentTypeYAML)
	var outputFormat string
	if contentType == "" {
		c.AbortWithStatus(http.StatusNotAcceptable)
		return "", "", true
	} else if contentType == gin.MIMEJSON {
		outputFormat = "json"
	} else {
		outputFormat = "yaml"
	}
	return contentType, outputFormat, false
}
