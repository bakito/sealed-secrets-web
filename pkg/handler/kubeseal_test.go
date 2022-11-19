package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHandler_KubeSeal_InputAsJson_OutputAsJson(t *testing.T) {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request, _ = http.NewRequest("POST", "/v1/kubeseal", bytes.NewReader([]byte(stringDataAsYAML)))
	c.Request.Header.Set("Content-Type", "application/x-yaml")
	c.Request.Header.Set("Accept", "application/x-yaml")
	h := &Handler{
		sealer: successfulSealer{},
	}

	h.KubeSeal(c)

	assert.Equal(t, 200, recorder.Code)
	assert.Equal(t, sealedAsYAML, recorder.Body.String())
	assert.Equal(t, "application/x-yaml", recorder.Header().Get("Content-Type"))
}

func TestHandler_KubeSeal_InputAsYaml_OutputAsYaml(t *testing.T) {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request, _ = http.NewRequest("POST", "/v1/kubeseal", bytes.NewReader([]byte(stringDataAsYAML)))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Request.Header.Set("Accept", "application/json")
	h := &Handler{
		sealer: successfulSealer{},
	}

	h.KubeSeal(c)

	assert.Equal(t, 200, recorder.Code)
	assert.Equal(t, sealAsJSON, recorder.Body.String())
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
}
