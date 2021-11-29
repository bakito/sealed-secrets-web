package handler

import (
	"encoding/base64"
	"net/http"

	"github.com/gin-gonic/gin"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func (h *Handler) Decode(c *gin.Context) {
	data := &secret{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	encoded, err := h.decode(data.Secret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	data.Secret = string(encoded)
	c.JSON(http.StatusOK, data)
}

func (h *Handler) decode(data string) ([]byte, error) {
	sec := make(map[string]interface{})

	err := h.marshaller.Unmarshal([]byte(data), &sec)
	if err != nil {
		return nil, err
	}

	h.filter.Apply(sec)

	if err = decodeData(sec); err != nil {
		return nil, err
	}

	return h.marshaller.Marshal(sec)
}

func decodeData(secretData map[string]interface{}) error {
	if _, ok := secretData["stringData"]; ok {
		// already decoded
		return nil
	}
	secretData["stringData"] = make(map[string]interface{})
	if m, ok, _ := unstructured.NestedMap(secretData, "data"); ok {
		for key, value := range m {
			decoded, err := base64.StdEncoding.DecodeString(value.(string))
			if err != nil {
				return err
			}
			_ = unstructured.SetNestedField(secretData, string(decoded), "stringData", key)
		}
	}
	delete(secretData, "data")
	return nil
}
