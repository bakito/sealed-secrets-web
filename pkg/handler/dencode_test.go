package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHandler_encode_InputAsJson_OutputAsJson(t *testing.T) {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request, _ = http.NewRequest("POST", "/v1/dencode", bytes.NewReader([]byte(stringDataAsJSON)))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Request.Header.Set("Accept", "application/json")
	h := &Handler{
		sealer: successfulSealer{},
	}
	h.Dencode(c)

	assert.Equal(t, 200, recorder.Code)
	assert.Equal(t, dataAsJSON, recorder.Body.String())
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
}

func TestHandler_Decode_InputAsJson_OutputAsJson(t *testing.T) {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request, _ = http.NewRequest("POST", "/v1/dencode", bytes.NewReader([]byte(dataAsJSON)))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Request.Header.Set("Accept", "application/json")
	h := &Handler{
		sealer: successfulSealer{},
	}
	h.Dencode(c)

	assert.Equal(t, 200, recorder.Code)
	assert.Equal(t, stringDataAsJSON, recorder.Body.String())
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
}

func TestHandler_encode_InputAsYaml_OutputAsYaml(t *testing.T) {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request, _ = http.NewRequest("POST", "/v1/dencode", bytes.NewReader([]byte(stringDataAsYAML)))
	c.Request.Header.Set("Content-Type", "application/x-yaml")
	c.Request.Header.Set("Accept", "application/x-yaml")
	h := &Handler{
		sealer: successfulSealer{},
	}
	h.Dencode(c)

	assert.Equal(t, 200, recorder.Code)
	assert.Equal(t, dataAsYAML, recorder.Body.String())
	assert.Equal(t, "application/x-yaml", recorder.Header().Get("Content-Type"))
}

func TestHandler_Decode_InputAsYaml_OutputAsYaml(t *testing.T) {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request, _ = http.NewRequest("POST", "/v1/dencode", bytes.NewReader([]byte(dataAsYAML)))
	c.Request.Header.Set("Content-Type", "application/x-yaml")
	c.Request.Header.Set("Accept", "application/x-yaml")
	h := &Handler{
		sealer: successfulSealer{},
	}
	h.Dencode(c)

	assert.Equal(t, 200, recorder.Code)
	assert.Equal(t, stringDataAsYAML, recorder.Body.String())
	assert.Equal(t, "application/x-yaml", recorder.Header().Get("Content-Type"))
}

func TestHandler_encode_InputAsJson_OutputAsText_NotAcceptable(t *testing.T) {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request, _ = http.NewRequest("POST", "/v1/dencode", bytes.NewReader([]byte(dataAsJSON)))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Request.Header.Set("Accept", "text/plain")
	h := &Handler{
		sealer: successfulSealer{},
	}
	h.Dencode(c)

	assert.Equal(t, 406, recorder.Code)
	assert.Equal(t, "", recorder.Body.String())
	assert.Equal(t, "", recorder.Header().Get("Content-Type"))
}

func TestHandler_encode_InputAsJson_OutputAsText_UnprocessableEntity(t *testing.T) {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request, _ = http.NewRequest("POST", "/v1/dencode", bytes.NewReader([]byte("invalidInputSecret")))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Request.Header.Set("Accept", "application/json")
	h := &Handler{
		sealer: successfulSealer{},
	}
	h.Dencode(c)

	assert.Equal(t, 422, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "{\"error\":")
	assert.Equal(t, "application/json; charset=utf-8", recorder.Header().Get("Content-Type"))
}
