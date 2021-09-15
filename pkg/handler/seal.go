package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
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

	h.filter.Apply(sec)

	if ss, err = h.marshaller.Marshal(sec); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	data.Secret = string(ss)
	c.JSON(http.StatusOK, data)
}
