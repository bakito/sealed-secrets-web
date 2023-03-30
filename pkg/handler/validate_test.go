package handler

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"

	"github.com/bakito/sealed-secrets-web/pkg/mocks/seal"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Handler ", func() {
	Context("Validate", func() {
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

		It("should return success if validation succeeds", func() {
			c.Request, _ = http.NewRequest("POST", "/v1/validate", bytes.NewReader([]byte(stringDataAsYAML)))
			c.Request.Header.Set("Content-Type", "application/x-yaml")

			sealer.EXPECT().Validate(gomock.Any()).Return(nil)

			h.Validate(c)

			Ω(recorder.Code).Should(Equal(http.StatusOK))
			Ω(recorder.Body.String()).Should(Equal("OK"))
			Ω(recorder.Header().Get("Content-Type")).Should(Equal("text/plain"))
		})

		It("should return an error if validation fails", func() {
			c.Request, _ = http.NewRequest("POST", "/v1/validate", bytes.NewReader([]byte(stringDataAsYAML)))
			c.Request.Header.Set("Content-Type", "application/x-yaml")

			sealer.EXPECT().Validate(gomock.Any()).Return(errors.New("Validation failed"))

			h.Validate(c)

			Ω(recorder.Code).Should(Equal(http.StatusBadRequest))
			Ω(recorder.Body.String()).Should(Equal("Validation failed"))
			Ω(recorder.Header().Get("Content-Type")).Should(Equal("text/plain"))
		})
	})
})
