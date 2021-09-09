package main

import (
	"net/http"
	"net/http/httptest"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/gin-gonic/gin"
)

var _ = Describe("Main", func() {

	Context("the router is initialized successfully", func() {
		var (
			w      *httptest.ResponseRecorder
			router *gin.Engine
		)

		BeforeEach(func() {
			w = httptest.NewRecorder()
			format := "yaml"
			outputFormat = &format
			disabled := true
			disableLoadSecrets = &disabled
			router = setupRouter()
		})
		It("return OK on health", func() {
			req, _ := http.NewRequest("GET", "/_health", nil)
			router.ServeHTTP(w, req)
			Ω(w.Code).Should(Equal(200))
			Ω(w.Body.String()).Should(Equal("OK"))
		})
		It("return version info on version", func() {
			req, _ := http.NewRequest("GET", "/api/version", nil)
			router.ServeHTTP(w, req)
			Ω(w.Code).Should(Equal(200))
			Ω(w.Body.String()).Should(Equal(`{"build":"","version":"dev"}`))
		})

		It("return the index page", func() {
			req, _ := http.NewRequest("GET", "/", nil)
			router.ServeHTTP(w, req)
			Ω(w.Code).Should(Equal(200))
		})

		It("redirect on any other url", func() {
			req, _ := http.NewRequest("GET", "/foo/bar", nil)
			router.ServeHTTP(w, req)
			Ω(w.Code).Should(Equal(301))
			Ω(w.Body.String()).Should(ContainSubstring("Moved Permanently"))
		})

		It("encode a secret", func() {
			req, _ := http.NewRequest("POST", "/api/encode", strings.NewReader(encodeBody))
			router.ServeHTTP(w, req)
			Ω(w.Code).Should(Equal(200))
			Ω(w.Body.String()).Should(Equal(decodeBody))
		})

		It("decode a secret", func() {
			req, _ := http.NewRequest("POST", "/api/decode", strings.NewReader(decodeBody))
			router.ServeHTTP(w, req)
			Ω(w.Code).Should(Equal(200))
			Ω(w.Body.String()).Should(Equal(encodeBody))
		})

		It("decode secrets endpoints are disabled", func() {
			req, _ := http.NewRequest("GET", "/api/secrets", nil)
			router.ServeHTTP(w, req)
			Ω(w.Code).Should(Equal(403))
			req, _ = http.NewRequest("GET", "/api/secrets/namespace/name", nil)
			router.ServeHTTP(w, req)
			Ω(w.Code).Should(Equal(403))
		})
	})
})

const (
	encodeBody = `{"secret":"apiVersion: v1\nkind: Secret\nmetadata:\n  name: mysecretname\n  namespace: mysecretnamespace\nstringData:\n  password: admin\n  username: admin\ntype: Opaque\n"}`
	decodeBody = `{"secret":"apiVersion: v1\ndata:\n  password: YWRtaW4=\n  username: YWRtaW4=\nkind: Secret\nmetadata:\n  name: mysecretname\n  namespace: mysecretnamespace\ntype: Opaque\n"}`
)
