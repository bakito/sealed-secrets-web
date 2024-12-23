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
	Context("Raw", func() {
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

		It("should return ras data for the given content", func() {
			c.Request, _ = http.NewRequest("POST", "/v1/raw", bytes.NewReader([]byte(rawData)))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Request.Header.Set("Accept", "application/json")

			sealer.EXPECT().Raw(gomock.Any()).Return([]byte("foo"), nil)

			h.Raw(c)

			Ω(recorder.Code).Should(Equal(http.StatusOK))
			Ω(recorder.Body.String()).Should(Equal(`{"secret":"foo"}`))
			Ω(recorder.Header().Get("Content-Type")).Should(Equal("application/json; charset=utf-8"))
		})

		It("should return an error if body can not be parsed as json", func() {
			c.Request, _ = http.NewRequest("POST", "/v1/raw", bytes.NewReader([]byte("foo")))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Request.Header.Set("Accept", "application/json")

			h.Raw(c)

			Ω(recorder.Code).Should(Equal(http.StatusUnprocessableEntity))
			Ω(
				recorder.Body.String(),
			).Should(Equal(`{"error":"invalid character 'o' in literal false (expecting 'a')"}`))
			Ω(recorder.Header().Get("Content-Type")).Should(Equal("application/json; charset=utf-8"))
		})

		It("should return an error if body can not be parsed", func() {
			c.Request, _ = http.NewRequest("POST", "/v1/raw", bytes.NewReader([]byte(rawData)))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Request.Header.Set("Accept", "application/json")

			sealer.EXPECT().Raw(gomock.Any()).Return(nil, errors.New("error processing raw"))

			h.Raw(c)

			Ω(recorder.Code).Should(Equal(http.StatusInternalServerError))
			Ω(recorder.Body.String()).Should(Equal(`{"error":"error processing raw"}`))
			Ω(recorder.Header().Get("Content-Type")).Should(Equal("application/json; charset=utf-8"))
		})
	})
})

const rawData = `{
  "name": "a-name",
  "namespace": "a-namespace",
  "value": "some value"
}
`
