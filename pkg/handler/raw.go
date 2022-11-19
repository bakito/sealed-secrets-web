package handler

import (
	"log"
	"net/http"

	"github.com/bakito/sealed-secrets-web/pkg/seal"
	"github.com/gin-gonic/gin"
)

type secret struct {
	Secret string `json:"secret"`
}

func (h *Handler) Raw(c *gin.Context) {
	data := &seal.Raw{}
	if err := c.ShouldBindJSON(&data); err != nil {
		log.Printf("Error in %s: %v\n", Sanitize(c.Request.URL.Path), err)
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	r, err := h.sealer.Raw(*data)
	if err != nil {
		log.Printf("Error in %s: %v\n", Sanitize(c.Request.URL.Path), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	sec := secret{}
	sec.Secret = string(r)
	c.JSON(http.StatusOK, sec)
}
