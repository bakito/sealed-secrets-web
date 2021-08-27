package handler

import (
	"encoding/base64"
	"net/http"

	"github.com/gin-gonic/gin"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func (h *Handler) Encode(c *gin.Context) {
	data := &secret{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	encoded, err := h.encode(data.Secret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	data.Secret = string(encoded)
	c.JSON(http.StatusOK, data)
}

func (h *Handler) encode(data string) ([]byte, error) {
	secretData := make(map[string]interface{})
	if err := h.marshaller.Unmarshal([]byte(data), &secretData); err != nil {
		return nil, err
	}
	encodeData(secretData)

	return h.marshaller.Marshal(secretData)
}

func encodeData(secretData map[string]interface{}) {

	if _, ok := secretData["data"]; ok {
		// already encoded
		return
	}
	// set empty data
	secretData["data"] = make(map[string]interface{})
	if m, ok, _ := unstructured.NestedMap(secretData, "stringData"); ok {
		for key, value := range m {
			b := []byte(value.(string))
			_ = unstructured.SetNestedField(secretData, base64.StdEncoding.EncodeToString(b), "data", key)
		}
	}
	delete(secretData, "stringData")
}
