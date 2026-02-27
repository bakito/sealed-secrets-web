package handler

import (
	"bytes"
	"encoding/base64"
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"
)

const errInvalidBase64 = "Data must be uniformly base64-encoded or in plain text. Use .data for encoded or .stringData for plaintext"

func (h *Handler) KubeSeal(c *gin.Context) {
	outputContentType, outputFormat, done := NegotiateFormat(c)
	if done {
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("Error reading body in %s: %v\n", Sanitize(c.Request.URL.Path), err)
		contextNegotiate(c, http.StatusInternalServerError, gin.Negotiate{
			Offered: []string{outputContentType},
			Data:    gin.H{"error": err.Error()},
		})
		return
	}

	if err := validateBase64Data(body); err != nil {
		contextNegotiate(c, http.StatusUnprocessableEntity, gin.Negotiate{
			Offered: []string{outputContentType},
			Data:    gin.H{"error": err.Error()},
		})
		return
	}

	ss, err := h.sealer.Seal(outputFormat, bytes.NewReader(body))
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

// validateBase64Data parses the body as a raw map to get the original string
// values in .data before k8s decodes them, and validates each is valid base64.
// Returns an error if any value fails decoding, or nil if the body has no .data
// or all values are valid.
func validateBase64Data(body []byte) error {
	var raw map[string]interface{}
	if err := yaml.Unmarshal(body, &raw); err != nil {
		return nil // not parseable; let the sealer produce its own error
	}
	dataField, ok := raw["data"]
	if !ok {
		return nil
	}
	dataMap, ok := dataField.(map[string]interface{})
	if !ok {
		return nil
	}
	for _, v := range dataMap {
		strVal, ok := v.(string)
		if !ok {
			return errors.New(errInvalidBase64)
		}
		if _, err := base64.StdEncoding.DecodeString(strVal); err != nil {
			return errors.New(errInvalidBase64)
		}
	}
	return nil
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
	switch contentType {
	case "":
		c.AbortWithStatus(http.StatusNotAcceptable)
		return "", "", true
	case gin.MIMEJSON:
		outputFormat = "json"
	default:
		outputFormat = "yaml"
	}
	return contentType, outputFormat, false
}
