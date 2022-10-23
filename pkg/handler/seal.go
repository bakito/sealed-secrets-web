package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type secret struct {
	Secret string `json:"secret"`
}

func (h *Handler) Seal(c *gin.Context) {
	data := &secret{}
	if err := c.ShouldBindJSON(&data); err != nil {
		log.Printf("Error in %s: %v\n", Sanitize(c.Request.URL.Path), err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ss, err := h.sealer.Secret(data.Secret)
	if err != nil {
		log.Printf("Error in %s: %v\n", Sanitize(c.Request.URL.Path), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// unmarshal result to json
	sec := make(map[string]interface{})
	if err := json.Unmarshal(ss, &sec); err != nil {
		log.Printf("Error in %s: %v\n", Sanitize(c.Request.URL.Path), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.filter.Apply(sec)

	if ss, err = h.marshaller.Marshal(sec); err != nil {
		log.Printf("Error in %s: %v\n", Sanitize(c.Request.URL.Path), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	data.Secret = string(ss)
	c.JSON(http.StatusOK, data)
}
