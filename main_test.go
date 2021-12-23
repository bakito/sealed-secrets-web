package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/bakito/sealed-secrets-web/pkg/config"
	"github.com/bakito/sealed-secrets-web/pkg/marshal"
	"github.com/bakito/sealed-secrets-web/pkg/mocks/core"
	"github.com/bakito/sealed-secrets-web/pkg/mocks/ssclient"
	"github.com/bitnami-labs/sealed-secrets/pkg/apis/sealed-secrets/v1alpha1"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Main", func() {
	Context("the router is initialized successfully", func() {
		var (
			name         string
			namespace    string
			w            *httptest.ResponseRecorder
			router       *gin.Engine
			mock         *gomock.Controller
			alpha1Client *ssclient.MockBitnamiV1alpha1Interface
			ssClient     *ssclient.MockSealedSecretInterface
			coreClient   *core.MockCoreV1Interface
			secrets      *core.MockSecretInterface
			cfg          *config.Config
		)

		BeforeEach(func() {
			cfg = &config.Config{
				OutputFormat: "yaml",
				FieldFilter:  &config.FieldFilter{},
				Marshaller:   marshal.For("yaml"),
			}
			name = uuid.NewString()
			namespace = uuid.NewString()
			w = httptest.NewRecorder()
			mock = gomock.NewController(GinkgoT())
			alpha1Client = ssclient.NewMockBitnamiV1alpha1Interface(mock)
			ssClient = ssclient.NewMockSealedSecretInterface(mock)
			coreClient = core.NewMockCoreV1Interface(mock)
			secrets = core.NewMockSecretInterface(mock)
			router = setupRouter(coreClient, alpha1Client, cfg)
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

		It("list sealed secrets", func() {
			alpha1Client.EXPECT().SealedSecrets("").Return(ssClient)
			ssClient.EXPECT().List(gomock.Any()).Return(&v1alpha1.SealedSecretList{
				Items: []v1alpha1.SealedSecret{
					{
						ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name},
						Spec:       v1alpha1.SealedSecretSpec{Template: v1alpha1.SecretTemplateSpec{}},
					},
				},
			}, nil)
			req, _ := http.NewRequest("GET", "/api/secrets", nil)
			router.ServeHTTP(w, req)
			Ω(w.Code).Should(Equal(200))
			Ω(w.Body.String()).Should(Equal(fmt.Sprintf(`[{"namespace":"%s","name":"%s"}]`, namespace, name)))
		})

		It("list secret of namespace", func() {
			coreClient.EXPECT().Secrets(namespace).Return(secrets)
			secrets.EXPECT().Get(name, gomock.Any()).Return(&corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name},
			}, nil)
			req, _ := http.NewRequest("GET", fmt.Sprintf("/api/secret/%s/%s", namespace, name), nil)
			router.ServeHTTP(w, req)
			Ω(w.Code).Should(Equal(200))
			Ω(w.Body.String()).Should(Equal(fmt.Sprintf(`{"secret":"apiVersion: v1\nkind: Secret\nmetadata:\n    creationTimestamp: null\n    name: %s\n    namespace: %s\n"}`, name, namespace)))
		})

		It("secrets endpoints are disabled", func() {
			cfg.DisableLoadSecrets = true
			router = setupRouter(coreClient, alpha1Client, cfg)
			req, _ := http.NewRequest("GET", "/api/secrets", nil)
			router.ServeHTTP(w, req)
			Ω(w.Code).Should(Equal(403))
			req, _ = http.NewRequest("GET", "/api/secret/namespace/name", nil)
			router.ServeHTTP(w, req)
			Ω(w.Code).Should(Equal(403))
		})
	})
})

const (
	encodeBody = `{"secret":"apiVersion: v1\nkind: Secret\nmetadata:\n  name: mysecretname\n  namespace: mysecretnamespace\nstringData:\n  password: admin\n  username: admin\ntype: Opaque\n"}`
	decodeBody = `{"secret":"apiVersion: v1\ndata:\n  password: YWRtaW4=\n  username: YWRtaW4=\nkind: Secret\nmetadata:\n  name: mysecretname\n  namespace: mysecretnamespace\ntype: Opaque\n"}`
)
