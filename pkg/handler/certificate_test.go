package handler

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHandler_Certificate_IsSuccessfullyReturned(t *testing.T) {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	h := &Handler{
		sealer: successfulSealer{},
	}

	h.Certificate(c)

	assert.Equal(t, 200, recorder.Code)
	assert.Equal(t, validCertificate, recorder.Body.String())
	assert.Equal(t, "text/plain", recorder.Header().Get("Content-Type"))
}

func TestHandler_Certificate_Failed(t *testing.T) {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	h := &Handler{
		sealer: errorSealer{},
	}

	h.Certificate(c)

	assert.Equal(t, 500, recorder.Code)
	assert.Equal(t, "{\"error\":\"unexpected error\"}", recorder.Body.String())
	assert.Equal(t, "application/json; charset=utf-8", recorder.Header().Get("Content-Type"))
}
