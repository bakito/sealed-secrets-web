package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/bakito/sealed-secrets-web/pkg/config"
	"github.com/bakito/sealed-secrets-web/pkg/mocks/core"
	"github.com/bakito/sealed-secrets-web/pkg/mocks/ssclient"
	"github.com/bitnami-labs/sealed-secrets/pkg/apis/sealedsecrets/v1alpha1"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
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
				FieldFilter: &config.FieldFilter{},
			}
			name = uuid.NewString()
			namespace = uuid.NewString()
			w = httptest.NewRecorder()
			mock = gomock.NewController(GinkgoT())
			alpha1Client = ssclient.NewMockBitnamiV1alpha1Interface(mock)
			ssClient = ssclient.NewMockSealedSecretInterface(mock)
			coreClient = core.NewMockCoreV1Interface(mock)
			secrets = core.NewMockSecretInterface(mock)
			router = setupRouter(coreClient, alpha1Client, cfg, nil)
		})
		It("return OK on health", func() {
			req, _ := http.NewRequest("GET", "/_health", nil)
			router.ServeHTTP(w, req)
			Ω(w.Code).Should(Equal(http.StatusOK))
			Ω(w.Body.String()).Should(Equal("OK"))
		})
		It("return version info on version", func() {
			req, _ := http.NewRequest("GET", "/api/version", nil)
			router.ServeHTTP(w, req)
			Ω(w.Code).Should(Equal(http.StatusOK))
			Ω(w.Body.String()).Should(Equal(`{"build":"","version":"dev"}`))
		})

		It("return the index page", func() {
			req, _ := http.NewRequest("GET", "/", nil)
			router.ServeHTTP(w, req)
			Ω(w.Code).Should(Equal(http.StatusOK))
		})

		It("redirect on any other url", func() {
			req, _ := http.NewRequest("GET", "/foo/bar", nil)
			router.ServeHTTP(w, req)
			Ω(w.Code).Should(Equal(301))
			Ω(w.Body.String()).Should(ContainSubstring("Moved Permanently"))
		})

		It("list sealed secrets for all namespaces", func() {
			cfg.IncludeNamespaces = nil
			alpha1Client.EXPECT().SealedSecrets("").Return(ssClient)
			ssClient.EXPECT().List(gomock.Any(), gomock.Any()).Return(&v1alpha1.SealedSecretList{
				Items: []v1alpha1.SealedSecret{
					{
						ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name},
						Spec:       v1alpha1.SealedSecretSpec{Template: v1alpha1.SecretTemplateSpec{}},
					},
				},
			}, nil)
			req, _ := http.NewRequest("GET", "/api/secrets", nil)
			router.ServeHTTP(w, req)
			Ω(w.Code).Should(Equal(http.StatusOK))
			Ω(
				w.Body.String(),
			).Should(Equal(fmt.Sprintf(`{"secrets":[{"namespace":"%s","name":"%s"}]}`, namespace, name)))
		})

		It("list sealed secrets only for given namespaces", func() {
			cfg.IncludeNamespaces = []string{"a", "b"}
			router = setupRouter(coreClient, alpha1Client, cfg, nil)
			alpha1Client.EXPECT().SealedSecrets("a").Return(ssClient)
			ssClient.EXPECT().List(gomock.Any(), gomock.Any()).Return(&v1alpha1.SealedSecretList{
				Items: []v1alpha1.SealedSecret{
					{
						ObjectMeta: metav1.ObjectMeta{Namespace: "a", Name: name},
						Spec:       v1alpha1.SealedSecretSpec{Template: v1alpha1.SecretTemplateSpec{}},
					},
				},
			}, nil)
			alpha1Client.EXPECT().SealedSecrets("b").Return(ssClient)
			ssClient.EXPECT().List(gomock.Any(), gomock.Any()).Return(&v1alpha1.SealedSecretList{
				Items: []v1alpha1.SealedSecret{
					{
						ObjectMeta: metav1.ObjectMeta{Namespace: "b", Name: name},
						Spec:       v1alpha1.SealedSecretSpec{Template: v1alpha1.SecretTemplateSpec{}},
					},
				},
			}, nil)
			req, _ := http.NewRequest("GET", "/api/secrets", nil)
			router.ServeHTTP(w, req)
			Ω(w.Code).Should(Equal(http.StatusOK))
			Ω(
				w.Body.String(),
			).Should(Equal(fmt.Sprintf(`{"secrets":[{"namespace":"%s","name":"%s"},{"namespace":"%s","name":"%s"}]}`, "a", name, "b", name)))
		})

		It("get secret from namespace by name", func() {
			coreClient.EXPECT().Secrets(namespace).Return(secrets)
			secrets.EXPECT().Get(gomock.Any(), name, gomock.Any()).Return(&corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name},
			}, nil)
			req, _ := http.NewRequest("GET", fmt.Sprintf("/api/secret/%s/%s", namespace, name), nil)
			router.ServeHTTP(w, req)
			Ω(w.Code).Should(Equal(http.StatusOK))
			Ω(w.Body.String()).Should(Equal(fmt.Sprintf(`{
  "kind": "Secret",
  "apiVersion": "v1",
  "metadata": {
    "name": "%s",
    "namespace": "%s"
  }
}`, name, namespace)))
		})

		It("secrets endpoints are disabled", func() {
			cfg.DisableLoadSecrets = true
			router = setupRouter(coreClient, alpha1Client, cfg, nil)
			req, _ := http.NewRequest("GET", "/api/secrets", nil)
			router.ServeHTTP(w, req)
			Ω(w.Code).Should(Equal(403))
			req, _ = http.NewRequest("GET", "/api/secret/namespace/name", nil)
			router.ServeHTTP(w, req)
			Ω(w.Code).Should(Equal(403))
		})
	})
})
