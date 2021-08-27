package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type secret struct {
	Secret string `json:"secret"`
}

func (h *Handler) Seal(c *gin.Context) {
	data := &secret{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ss, err := h.sealer.Secret(data.Secret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// unmarshal result to json
	sec := make(map[string]interface{})
	if err := json.Unmarshal(ss, &sec); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	removeFieldIfNull(sec, "metadata", "creationTimestamp")
	removeFieldIfNull(sec, "spec", "template", "data")
	removeFieldIfNull(sec, "spec", "template", "metadata", "creationTimestamp")

	if ss, err = h.marshaller.Marshal(sec); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	data.Secret = string(ss)
	c.JSON(http.StatusOK, data)
}

func removeFieldIfNull(sec map[string]interface{}, fields ...string) {
	path := fields[:len(fields)-1]
	name := fields[len(fields)-1]
	if m, ok, _ := unstructured.NestedMap(sec, path...); ok {
		f := m[name]
		if f == nil {
			delete(m, name)
			_ = unstructured.SetNestedMap(sec, m, path...)
		}
	}
}
