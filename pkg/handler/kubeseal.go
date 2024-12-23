package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
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
		contextNegotiate(c, http.StatusInternalServerError, gin.Negotiate{
			Offered: []string{outputContentType},
			Data:    gin.H{"error": err.Error()},
		})
		c.Data(http.StatusInternalServerError, outputContentType, ss)
		return
	}

	c.Data(http.StatusOK, outputContentType, ss)
}

// fox for gin 1.10 incomplete yaml handling https://github.com/gin-gonic/gin/issues/3965
func contextNegotiate(c *gin.Context, code int, config gin.Negotiate) {
	switch c.NegotiateFormat(config.Offered...) {
	case binding.MIMEJSON:
		data := config.Data
		c.JSON(code, data)

	case binding.MIMEHTML:
		data := config.Data
		c.HTML(code, config.HTMLName, data)

	case binding.MIMEXML:
		data := config.Data
		c.XML(code, data)

	case binding.MIMEYAML:
	case binding.MIMEYAML2:
		data := config.Data
		c.YAML(code, data)

	case binding.MIMETOML:
		data := config.Data
		c.TOML(code, data)

	default:
		_ = c.AbortWithError(
			http.StatusNotAcceptable,
			errors.New("the accepted formats are not offered by the server"),
		)
	}
}

func NegotiateFormat(c *gin.Context) (string, string, bool) {
	contentType := c.NegotiateFormat(gin.MIMEJSON, gin.MIMEYAML, runtime.ContentTypeYAML)
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
