package handler

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"

	"github.com/bakito/sealed-secrets-web/pkg/handler/binding"
	"github.com/bakito/sealed-secrets-web/pkg/handler/render"
	"github.com/gin-gonic/gin"
)

func (h *Handler) Dencode(c *gin.Context) {
	secretData := make(map[string]interface{})
	var err error
	switch c.ContentType() {
	case gin.MIMEJSON:
		err = c.ShouldBindJSON(&secretData)
	case gin.MIMEYAML:
		// create custom binding because gin still uses yaml v2 => remove when gin starts using "gopkg.in/yaml.v3"
		err = binding.YAML.Bind(c.Request, &secretData)
	default: // case MIMEPOSTForm:
		err := fmt.Errorf("unsupported media type: %s", c.ContentType())
		log.Printf("Error in %s: %s\n", Sanitize(c.Request.URL.Path), Sanitize(err.Error()))
		c.JSON(http.StatusUnsupportedMediaType, gin.H{"error": err.Error()})
		return
	}

	if err != nil {
		log.Printf("Error in %s: %s\n", Sanitize(c.Request.URL.Path), Sanitize(err.Error()))
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	dencoded, err := h.dencode(secretData)
	if err != nil {
		log.Printf("Error in %s: %s\n", Sanitize(c.Request.URL.Path), Sanitize(err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	switch c.NegotiateFormat(gin.MIMEJSON, gin.MIMEYAML) {
	case gin.MIMEJSON:
		c.JSON(http.StatusOK, dencoded)
	case gin.MIMEYAML:
		c.Render(http.StatusOK, render.YAML{Data: dencoded})
	default:
		c.AbortWithStatus(http.StatusNotAcceptable)
	}
}

func (h *Handler) dencode(secretData map[string]interface{}) (map[string]interface{}, error) {
	h.filter.Apply(secretData)

	// https://kubernetes.io/docs/concepts/configuration/secret/#restriction-names-data
	// If a key appears in both the data and the stringData field, the value specified in the stringData field takes precedence.
	if _, ok := secretData["stringData"]; ok {
		err := encodeDataA(secretData)
		if err != nil {
			return nil, err
		}

		return secretData, nil
	}
	err := decodeDataA(secretData)
	if err != nil {
		return nil, err
	}

	return secretData, nil
}

func encodeDataA(secretData map[string]interface{}) error {
	dm, err := asMap(secretData, "data")
	if err != nil {
		return err
	}
	sdm, err := asMap(secretData, "stringData")
	if err != nil {
		return err
	}
	for key, value := range sdm {
		str, ok := value.(string)
		if ok {
			dm[key] = base64.StdEncoding.EncodeToString([]byte(str))
		}
	}
	delete(secretData, "stringData")
	return nil
}

func decodeDataA(secretData map[string]interface{}) error {
	dm, err := asMap(secretData, "data")
	if err != nil {
		return err
	}
	sdm, err := asMap(secretData, "stringData")
	if err != nil {
		return err
	}
	for key, value := range dm {
		str, ok := value.(string)
		if ok {
			decoded, err := base64.StdEncoding.DecodeString(str)
			if err != nil {
				return fmt.Errorf("base64 decoding for data.%s failed with %w", key, err)
			}
			sdm[key] = string(decoded)
		}
	}
	delete(secretData, "data")
	return nil
}

func asMap(m map[string]interface{}, field string) (map[string]interface{}, error) {
	if _, ok := m[field]; !ok {
		m[field] = make(map[string]interface{})
	}

	val, ok := m[field].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("%v accessor error: %v is of the type %T, expected map[string]interface{}", field, m[field], m[field])
	}
	return val, nil
}
