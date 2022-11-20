package handler

import (
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Handler ", func() {
	Context("Certificate", func() {
		var (
			recorder *httptest.ResponseRecorder
			c        *gin.Context
		)
		BeforeEach(func() {
			gin.SetMode(gin.ReleaseMode)
			recorder = httptest.NewRecorder()
			c, _ = gin.CreateTestContext(recorder)
		})
		It("should successfully return a certificate", func() {
			h := &Handler{
				sealer: successfulSealer{},
			}
			h.Certificate(c)

			Ω(recorder.Code).Should(Equal(http.StatusOK))
			Ω(recorder.Body.String()).Should(Equal(validCertificate))
			Ω(recorder.Header().Get("Content-Type")).Should(Equal("text/plain; charset=utf-8"))
		})
		It("should successfully fail when requesting a certificate", func() {
			h := &Handler{
				sealer: errorSealer{},
			}
			h.Certificate(c)

			Ω(recorder.Code).Should(Equal(http.StatusInternalServerError))
			Ω(recorder.Body.String()).Should(Equal(`{"error":"unexpected error"}`))
			Ω(recorder.Header().Get("Content-Type")).Should(Equal("application/json; charset=utf-8"))
		})
	})
})
