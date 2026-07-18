package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"

	"github.com/bakito/sealed-secrets-web/pkg/config"
	"github.com/bakito/sealed-secrets-web/pkg/version"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const helloWorld = `<html>Hello World</html>`

var _ = Describe("Handler ", func() {
	Context("common", func() {
		var (
			recorder *httptest.ResponseRecorder
			c        *gin.Context
			h        *Handler
		)
		BeforeEach(func() {
			gin.SetMode(gin.ReleaseMode)
			recorder = httptest.NewRecorder()
			c, _ = gin.CreateTestContext(recorder)
			h = New(helloWorld, nil, &config.Config{})
		})
		Context("Health", func() {
			It("should return OK", func() {
				c.Request, _ = http.NewRequest(http.MethodGet, "/_health", http.NoBody)
				h.Health(c)

				Ω(recorder.Code).Should(Equal(http.StatusOK))
				Ω(recorder.Body.String()).Should(Equal("OK"))
				Ω(recorder.Header().Get("Content-Type")).Should(Equal("text/plain; charset=utf-8"))
			})
		})
		Context("Index", func() {
			It("should return the index html", func() {
				c.Request, _ = http.NewRequest(http.MethodGet, "/", http.NoBody)
				h.Index(c)

				Ω(recorder.Code).Should(Equal(http.StatusOK))
				Ω(recorder.Body.String()).Should(Equal(helloWorld))
				Ω(recorder.Header().Get("Content-Type")).Should(Equal("text/html; charset=utf-8"))
			})
		})
		Context("RedirectToIndex", func() {
			It("should redirect to index", func() {
				c.Request, _ = http.NewRequest(http.MethodGet, "/foo", http.NoBody)
				h.RedirectToIndex("/ssw/")(c)

				Ω(recorder.Code).Should(Equal(http.StatusMovedPermanently))
				Ω(recorder.Body.String()).Should(Equal(`<a href="/ssw/">Moved Permanently</a>.

`))
				Ω(recorder.Header().Get("Content-Type")).Should(Equal("text/html; charset=utf-8"))
			})
		})
		Context("Version", func() {
			It("should return the current version", func() {
				c.Request, _ = http.NewRequest(http.MethodGet, "/version", http.NoBody)
				h.Version(c)

				Ω(recorder.Code).Should(Equal(http.StatusOK))
				Ω(
					recorder.Body.String(),
				).Should(Equal(fmt.Sprintf(`{"build":"%v","version":"%v"}`, version.Build, version.Version)))
				Ω(recorder.Header().Get("Content-Type")).Should(Equal("application/json; charset=utf-8"))
			})
		})
	})
})
