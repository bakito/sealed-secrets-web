package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"k8s.io/apimachinery/pkg/runtime"
)

func (h *Handler) KubeSeal(c *gin.Context) {
	result := []byte{}
	outputContentType, outputFormat, done := NegotiateFormat(c)
	if done {
		return
	}

	data, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		return err
	}

	yamlFiles := strings.Split(string(data), "
---
")
	for _, yamlFile := range yamlFiles {
		var obj map[string]interface{}
		if err := yaml.Unmarshal([]byte(yamlFile), &obj); err != nil {
			return err
		}

		ss, err := h.sealer.Seal(outputFormat, obj)
		if err != nil {
			log.Printf("Error in %s: %v\n", Sanitize(c.Request.URL.Path), err)
			c.Negotiate(http.StatusInternalServerError, gin.Negotiate{
				Offered: []string{gin.MIMEJSON, gin.MIMEYAML},
				Data:    gin.H{"error": err.Error()},
			})
			return
		}
		result = append(result, ss)
  }

	c.Data(http.StatusOK, outputContentType, result)
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
