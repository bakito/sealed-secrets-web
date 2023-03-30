package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) Validate(c *gin.Context) {
	err := h.sealer.Validate(c.Request.Body)
	if err != nil {
		log.Printf("Error in %s: %v\n", Sanitize(c.Request.URL.Path), err)
		c.Data(http.StatusBadRequest, "text/plain", []byte(err.Error()))
	} else {
		c.Data(http.StatusOK, "text/plain", []byte("OK"))
	}
}
