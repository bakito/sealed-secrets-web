package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) Certificate(c *gin.Context) {
	keyLabel := c.GetHeader("X-Sealing-Key")
	certificate, err := h.sealer.CertificateWithKey(c, keyLabel)
	if err != nil {
		log.Printf("Error in reading Certificate %s\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Data(http.StatusOK, "text/plain; charset=utf-8", certificate)
}
