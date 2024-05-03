package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/runtime"
)

func mapToReader(inputMap map[string]interface{}) (io.Reader, error) {
	jsonBytes, err := json.Marshal(inputMap)
	if err != nil {
		return nil, err
	}

	reader := bytes.NewReader(jsonBytes)
	return reader, nil
}

func (h *Handler) KubeSeal(c *gin.Context) {
	result := []byte{}
	outputContentType, outputFormat, done := NegotiateFormat(c)
	if done {
		return
	}

	data, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return
	}

	yamlFiles := strings.Split(string(data), "---\n")
	for _, yamlFile := range yamlFiles {
		var obj map[string]interface{}
		if err := yaml.Unmarshal([]byte(yamlFile), &obj); err != nil {
			return
		}
		reader, err := mapToReader(obj)
		if err != nil {
			return
		}
		ss, err := h.sealer.Seal(outputFormat, reader)
		if err != nil {
			log.Printf("Error in %s: %v\n", Sanitize(c.Request.URL.Path), err)
			c.Negotiate(http.StatusInternalServerError, gin.Negotiate{
				Offered: []string{gin.MIMEJSON, gin.MIMEYAML},
				Data:    gin.H{"error": err.Error()},
			})
			return
		}
		result = append(result, ss...)
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
