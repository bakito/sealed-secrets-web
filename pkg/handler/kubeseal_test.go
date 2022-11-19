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
	Context("KubeSeal", func() {
		var (
			recorder *httptest.ResponseRecorder
			c        *gin.Context
		)
		BeforeEach(func() {
			gin.SetMode(gin.ReleaseMode)
			recorder = httptest.NewRecorder()
			c, _ = gin.CreateTestContext(recorder)
		})

		It("should kubeseal input as json and output as json", func() {
			c.Request, _ = http.NewRequest("POST", "/v1/kubeseal", bytes.NewReader([]byte(stringDataAsJSON)))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Request.Header.Set("Accept", "application/json")
			h := &Handler{
				sealer: successfulSealer{},
			}

			h.KubeSeal(c)

			Ω(recorder.Code).Should(Equal(http.StatusOK))
			Ω(recorder.Body.String()).Should(Equal(sealAsJSON))
			Ω(recorder.Header().Get("Content-Type")).Should(Equal("application/json"))
		})

		It("should kubeseal input as yaml and output as yaml", func() {
			c.Request, _ = http.NewRequest("POST", "/v1/kubeseal", bytes.NewReader([]byte(stringDataAsYAML)))
			c.Request.Header.Set("Content-Type", "application/x-yaml")
			c.Request.Header.Set("Accept", "application/x-yaml")
			h := &Handler{
				sealer: successfulSealer{},
			}

			h.KubeSeal(c)

			Ω(recorder.Code).Should(Equal(http.StatusOK))
			Ω(recorder.Body.String()).Should(Equal(sealedAsYAML))
			Ω(recorder.Header().Get("Content-Type")).Should(Equal("application/x-yaml"))
		})
	})
})
