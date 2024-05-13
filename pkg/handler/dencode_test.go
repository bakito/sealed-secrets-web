package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Handler ", func() {
	Context("dencode", func() {
		var (
			recorder *httptest.ResponseRecorder
			c        *gin.Context
			h        *Handler
		)
		BeforeEach(func() {
			gin.SetMode(gin.ReleaseMode)
			recorder = httptest.NewRecorder()
			c, _ = gin.CreateTestContext(recorder)
			h = &Handler{}
		})
		It("should encode input as json and output as json", func() {
			c.Request, _ = http.NewRequest("POST", "/v1/dencode", bytes.NewReader([]byte(stringDataAsJSON)))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Request.Header.Set("Accept", "application/json")
			h.Dencode(c)

			Ω(recorder.Code).Should(Equal(http.StatusOK))
			Ω(recorder.Body.String()).Should(Equal(dataAsJSON))
			Ω(recorder.Header().Get("Content-Type")).Should(Equal("application/json"))
		})
		It("should decode input as json and output as json", func() {
			c.Request, _ = http.NewRequest("POST", "/v1/dencode", bytes.NewReader([]byte(dataAsJSON)))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Request.Header.Set("Accept", "application/json")
			h.Dencode(c)

			Ω(recorder.Code).Should(Equal(http.StatusOK))
			Ω(recorder.Body.String()).Should(Equal(stringDataAsJSON))
			Ω(recorder.Header().Get("Content-Type")).Should(Equal("application/json"))
		})
		It("should encode input as yaml and output as yaml", func() {
			c.Request, _ = http.NewRequest("POST", "/v1/dencode", bytes.NewReader([]byte(stringDataAsYAML)))
			c.Request.Header.Set("Content-Type", "application/yaml")
			c.Request.Header.Set("Accept", "application/yaml")
			h.Dencode(c)

			Ω(recorder.Code).Should(Equal(http.StatusOK))
			Ω(recorder.Body.String()).Should(Equal(dataAsYAML))
			Ω(recorder.Header().Get("Content-Type")).Should(Equal("application/yaml"))
		})
		It("should decode input as yaml and output as yaml", func() {
			c.Request, _ = http.NewRequest("POST", "/v1/dencode", bytes.NewReader([]byte(dataAsYAML)))
			c.Request.Header.Set("Content-Type", "application/yaml")
			c.Request.Header.Set("Accept", "application/yaml")
			h.Dencode(c)

			Ω(recorder.Code).Should(Equal(http.StatusOK))
			Ω(recorder.Body.String()).Should(Equal(stringDataAsYAML))
			Ω(recorder.Header().Get("Content-Type")).Should(Equal("application/yaml"))
		})
		It("should encode input as json and output as text not acceptable", func() {
			c.Request, _ = http.NewRequest("POST", "/v1/dencode", bytes.NewReader([]byte(dataAsJSON)))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Request.Header.Set("Accept", "text/plain")
			h.Dencode(c)

			Ω(recorder.Code).Should(Equal(http.StatusNotAcceptable))
			Ω(recorder.Body.String()).Should(BeEmpty())
			Ω(recorder.Header().Get("Content-Type")).Should(BeEmpty())
		})
		It("should encode input as json and output as text unprocessable entity", func() {
			c.Request, _ = http.NewRequest("POST", "/v1/dencode", bytes.NewReader([]byte("invalidInputSecret")))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Request.Header.Set("Accept", "application/json")
			h.Dencode(c)

			Ω(recorder.Code).Should(Equal(http.StatusUnprocessableEntity))
			Ω(recorder.Body.String()).Should(ContainSubstring(`{"error":`))
			Ω(recorder.Header().Get("Content-Type")).Should(Equal("application/json; charset=utf-8"))
		})
	})
})
