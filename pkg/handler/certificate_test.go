package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/bakito/sealed-secrets-web/pkg/mocks/seal"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Handler ", func() {
	Context("Certificate", func() {
		var (
			recorder *httptest.ResponseRecorder
			c        *gin.Context
			mock     *gomock.Controller
			sealer   *seal.MockSealer
			h        *Handler
		)
		BeforeEach(func() {
			gin.SetMode(gin.ReleaseMode)
			recorder = httptest.NewRecorder()
			c, _ = gin.CreateTestContext(recorder)
			mock = gomock.NewController(GinkgoT())
			sealer = seal.NewMockSealer(mock)
			h = &Handler{
				sealer: sealer,
			}
		})
		It("should successfully return a certificate", func() {
			sealer.EXPECT().Certificate(gomock.Any()).Return([]byte(validCertificate), nil)
			h.Certificate(c)

			Ω(recorder.Code).Should(Equal(http.StatusOK))
			Ω(recorder.Body.String()).Should(Equal(validCertificate))
			Ω(recorder.Header().Get("Content-Type")).Should(Equal("text/plain; charset=utf-8"))
		})
		It("should successfully fail when requesting a certificate", func() {
			sealer.EXPECT().Certificate(gomock.Any()).Return(nil, fmt.Errorf("unexpected error"))
			h.Certificate(c)

			Ω(recorder.Code).Should(Equal(http.StatusInternalServerError))
			Ω(recorder.Body.String()).Should(Equal(`{"error":"unexpected error"}`))
			Ω(recorder.Header().Get("Content-Type")).Should(Equal("application/json; charset=utf-8"))
		})
	})
})
