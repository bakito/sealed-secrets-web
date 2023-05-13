package handler

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) Validate(c *gin.Context) {
	if h.cfg.SealedSecrets.CertURL != "" {
		configError := fmt.Errorf("validate can't be used with CertURL (%s)", h.cfg.SealedSecrets.CertURL)
		c.Data(http.StatusConflict, "text/plain", []byte(configError.Error()))
		return
	}
	err := h.sealer.Validate(c, c.Request.Body)

	if err != nil {
		log.Printf("Error in %s: %v\n", Sanitize(c.Request.URL.Path), err)
		c.Data(http.StatusBadRequest, "text/plain", []byte(err.Error()))
	} else {
		c.Data(http.StatusOK, "text/plain", []byte("OK"))
	}
}
