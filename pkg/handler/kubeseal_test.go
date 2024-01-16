package handler

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"

	"github.com/bakito/sealed-secrets-web/pkg/mocks/seal"
	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

var _ = Describe("Handler ", func() {
	Context("KubeSeal", func() {
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

		It("should kubeseal input as json and output as json", func() {
			c.Request, _ = http.NewRequest("POST", "/v1/kubeseal", bytes.NewReader([]byte(stringDataAsJSON)))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Request.Header.Set("Accept", "application/json")

			sealer.EXPECT().Seal("json", gomock.Any()).Return([]byte(sealAsJSON), nil)

			h.KubeSeal(c)

			Ω(recorder.Code).Should(Equal(http.StatusOK))
			Ω(recorder.Body.String()).Should(Equal(sealAsJSON))
			Ω(recorder.Header().Get("Content-Type")).Should(Equal("application/json"))
		})

		It("should kubeseal input as yaml and output as yaml", func() {
			c.Request, _ = http.NewRequest("POST", "/v1/kubeseal", bytes.NewReader([]byte(stringDataAsYAML)))
			c.Request.Header.Set("Content-Type", "application/x-yaml")
			c.Request.Header.Set("Accept", "application/x-yaml")

			sealer.EXPECT().Seal("yaml", gomock.Any()).Return([]byte(sealedAsYAML), nil)

			h.KubeSeal(c)

			Ω(recorder.Code).Should(Equal(http.StatusOK))
			Ω(recorder.Body.String()).Should(Equal(sealedAsYAML))
			Ω(recorder.Header().Get("Content-Type")).Should(Equal("application/x-yaml"))
		})

		It("should return an error if seal is not successful", func() {
			c.Request, _ = http.NewRequest("POST", "/v1/kubeseal", bytes.NewReader([]byte(stringDataAsYAML)))
			c.Request.Header.Set("Content-Type", "application/x-yaml")
			c.Request.Header.Set("Accept", "application/x-yaml")

			sealer.EXPECT().Seal(gomock.Any(), gomock.Any()).Return(nil, errors.New("error sealing"))

			h.KubeSeal(c)

			Ω(recorder.Code).Should(Equal(http.StatusInternalServerError))
			Ω(recorder.Body.String()).Should(Equal("error: error sealing\n"))
			Ω(recorder.Header().Get("Content-Type")).Should(Equal("application/x-yaml; charset=utf-8"))
		})
	})
})
