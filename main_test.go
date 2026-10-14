package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/bitnami/sealed-secrets/pkg/apis/sealedsecrets/v1alpha1"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
	corev1 "k8s.io/api/core/v1"

	"github.com/bakito/sealed-secrets-web/pkg/config"
	"github.com/bakito/sealed-secrets-web/pkg/matcher"
	"github.com/bakito/sealed-secrets-web/pkg/mocks/core"
	"github.com/bakito/sealed-secrets-web/pkg/mocks/seal"
	"github.com/bakito/sealed-secrets-web/pkg/mocks/ssclient"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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
			req, _ := http.NewRequest(http.MethodGet, "/_health", http.NoBody)
			router.ServeHTTP(w, req)
			Ω(w.Code).Should(Equal(http.StatusOK))
			Ω(w.Body.String()).Should(Equal("OK"))
		})
		It("return version info on version", func() {
			req, _ := http.NewRequest(http.MethodGet, "/api/version", http.NoBody)
			router.ServeHTTP(w, req)
			Ω(w.Code).Should(Equal(http.StatusOK))
			Ω(w.Body.String()).Should(Equal(`{"build":"","version":"dev"}`))
		})

		It("return the index page", func() {
			req, _ := http.NewRequest(http.MethodGet, "/", http.NoBody)
			router.ServeHTTP(w, req)
			Ω(w.Code).Should(Equal(http.StatusOK))
		})

		It("redirect on any other url", func() {
			req, _ := http.NewRequest(http.MethodGet, "/foo/bar", http.NoBody)
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
						Namespace: namespace, Name: name,
						Spec: v1alpha1.SealedSecretSpec{Template: v1alpha1.SecretTemplateSpec{}},
					},
				},
			}, nil)
			req, _ := http.NewRequest(http.MethodGet, "/api/secrets", http.NoBody)
			router.ServeHTTP(w, req)
			Ω(w.Code).Should(Equal(http.StatusOK))
			Ω(
				w.Body.String(),
			).Should(Equal(fmt.Sprintf(`{"secrets":[{"namespace":%q,"name":%q}]}`, namespace, name)))
		})

		It("list sealed secrets only for given namespaces", func() {
			cfg.IncludeNamespaces = []string{"a", "b"}
			router = setupRouter(coreClient, alpha1Client, cfg, nil)
			alpha1Client.EXPECT().SealedSecrets("a").Return(ssClient)
			ssClient.EXPECT().List(gomock.Any(), gomock.Any()).Return(&v1alpha1.SealedSecretList{
				Items: []v1alpha1.SealedSecret{
					{
						Namespace: "a", Name: name,
						Spec: v1alpha1.SealedSecretSpec{Template: v1alpha1.SecretTemplateSpec{}},
					},
				},
			}, nil)
			alpha1Client.EXPECT().SealedSecrets("b").Return(ssClient)
			ssClient.EXPECT().List(gomock.Any(), gomock.Any()).Return(&v1alpha1.SealedSecretList{
				Items: []v1alpha1.SealedSecret{
					{
						Namespace: "b", Name: name,
						Spec: v1alpha1.SealedSecretSpec{Template: v1alpha1.SecretTemplateSpec{}},
					},
				},
			}, nil)
			req, _ := http.NewRequest(http.MethodGet, "/api/secrets", http.NoBody)
			router.ServeHTTP(w, req)
			Ω(w.Code).Should(Equal(http.StatusOK))
			Ω(
				w.Body.String(),
			).Should(Equal(fmt.Sprintf(`{"secrets":[{"namespace":%q,"name":%q},{"namespace":%q,"name":%q}]}`, "a", name, "b", name)))
		})

		It("get secret from namespace by name", func() {
			coreClient.EXPECT().Secrets(namespace).Return(secrets)
			secrets.EXPECT().Get(gomock.Any(), name, gomock.Any()).Return(&corev1.Secret{
				Namespace: namespace, Name: name,
			}, nil)
			req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/secret/%s/%s", namespace, name), http.NoBody)
			router.ServeHTTP(w, req)
			Ω(w.Code).Should(Equal(http.StatusOK))
			Ω(w.Body.String()).Should(matcher.EqualDiff(fmt.Sprintf(`{
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
			req, _ := http.NewRequest(http.MethodGet, "/api/secrets", http.NoBody)
			router.ServeHTTP(w, req)
			Ω(w.Code).Should(Equal(403))
			req, _ = http.NewRequest(http.MethodGet, "/api/secret/namespace/name", http.NoBody)
			router.ServeHTTP(w, req)
			Ω(w.Code).Should(Equal(403))
		})
	})

	Context("multi-key sealing", func() {
		var (
			w            *httptest.ResponseRecorder
			router       *gin.Engine
			mock         *gomock.Controller
			mockSealer   *seal.MockSealerWithKeys
			alpha1Client *ssclient.MockBitnamiV1alpha1Interface
			coreClient   *core.MockCoreV1Interface
			cfg          *config.Config
		)

		BeforeEach(func() {
			cfg = &config.Config{
				FieldFilter: &config.FieldFilter{},
				SealedSecretsKeys: []config.SealedSecretsKey{
					{Label: "primary", Service: "sealed-secrets", Namespace: "sealed-secrets"},
					{Label: "secondary", Service: "sealed-secrets-secondary", Namespace: "sealed-secrets-secondary"},
				},
			}
			w = httptest.NewRecorder()
			mock = gomock.NewController(GinkgoT())
			mockSealer = seal.NewMockSealerWithKeys(mock)
			alpha1Client = ssclient.NewMockBitnamiV1alpha1Interface(mock)
			coreClient = core.NewMockCoreV1Interface(mock)
			router = setupRouter(coreClient, alpha1Client, cfg, mockSealer)
		})

		It("should return the list of keys", func() {
			mockSealer.EXPECT().Keys().Return([]string{"primary", "secondary"})

			req, _ := http.NewRequest(http.MethodGet, "/api/keys", http.NoBody)
			router.ServeHTTP(w, req)
			Ω(w.Code).Should(Equal(http.StatusOK))
			Ω(w.Body.String()).Should(Equal(`{"keys":["primary","secondary"]}`))
		})

		It("should kubeseal with the primary key", func() {
			secretYAML := `apiVersion: v1
kind: Secret
metadata:
  name: my-secret
  namespace: default
stringData:
  password: secret123`
			sealed := `apiVersion: bitnami.com/v1alpha1
kind: SealedSecret
metadata:
  name: my-secret
  namespace: default`

			mockSealer.EXPECT().
				SealWithKey(gomock.Any(), "primary", "yaml", gomock.Any()).
				Return([]byte(sealed), nil)

			req, _ := http.NewRequest(http.MethodPost, "/api/kubeseal", bytes.NewReader([]byte(secretYAML)))
			req.Header.Set("Content-Type", "application/yaml")
			req.Header.Set("Accept", "application/yaml")
			req.Header.Set("X-Sealing-Key", "primary")
			router.ServeHTTP(w, req)
			Ω(w.Code).Should(Equal(http.StatusOK))
			Ω(w.Body.String()).Should(Equal(sealed))
		})

		It("should kubeseal with the secondary key", func() {
			secretYAML := `apiVersion: v1
kind: Secret
metadata:
  name: my-secret
  namespace: default
stringData:
  password: secret123`
			sealed := `apiVersion: bitnami.com/v1alpha1
kind: SealedSecret
metadata:
  name: my-secret
  namespace: default
spec:
  encryptedSecret: secondary-key-data`

			mockSealer.EXPECT().
				SealWithKey(gomock.Any(), "secondary", "yaml", gomock.Any()).
				Return([]byte(sealed), nil)

			req, _ := http.NewRequest(http.MethodPost, "/api/kubeseal", bytes.NewReader([]byte(secretYAML)))
			req.Header.Set("Content-Type", "application/yaml")
			req.Header.Set("Accept", "application/yaml")
			req.Header.Set("X-Sealing-Key", "secondary")
			router.ServeHTTP(w, req)
			Ω(w.Code).Should(Equal(http.StatusOK))
			Ω(w.Body.String()).Should(ContainSubstring("secondary-key-data"))
		})

		It("should kubeseal without key header using default key", func() {
			secretYAML := `apiVersion: v1
kind: Secret
metadata:
  name: my-secret
  namespace: default
stringData:
  password: secret123`
			sealed := `apiVersion: bitnami.com/v1alpha1
kind: SealedSecret
metadata:
  name: my-secret
  namespace: default`

			mockSealer.EXPECT().
				SealWithKey(gomock.Any(), "", "yaml", gomock.Any()).
				Return([]byte(sealed), nil)

			req, _ := http.NewRequest(http.MethodPost, "/api/kubeseal", bytes.NewReader([]byte(secretYAML)))
			req.Header.Set("Content-Type", "application/yaml")
			req.Header.Set("Accept", "application/yaml")
			router.ServeHTTP(w, req)
			Ω(w.Code).Should(Equal(http.StatusOK))
		})

		It("should return certificate for the primary key", func() {
			cert := `-----BEGIN CERTIFICATE-----
primary-key-cert-data
-----END CERTIFICATE-----`

			mockSealer.EXPECT().
				CertificateWithKey(gomock.Any(), "primary").
				Return([]byte(cert), nil)

			req, _ := http.NewRequest(http.MethodGet, "/api/certificate", http.NoBody)
			req.Header.Set("X-Sealing-Key", "primary")
			router.ServeHTTP(w, req)
			Ω(w.Code).Should(Equal(http.StatusOK))
			Ω(w.Body.String()).Should(ContainSubstring("primary-key-cert-data"))
		})

		It("should return certificate for the secondary key", func() {
			cert := `-----BEGIN CERTIFICATE-----
secondary-key-cert-data
-----END CERTIFICATE-----`

			mockSealer.EXPECT().
				CertificateWithKey(gomock.Any(), "secondary").
				Return([]byte(cert), nil)

			req, _ := http.NewRequest(http.MethodGet, "/api/certificate", http.NoBody)
			req.Header.Set("X-Sealing-Key", "secondary")
			router.ServeHTTP(w, req)
			Ω(w.Code).Should(Equal(http.StatusOK))
			Ω(w.Body.String()).Should(ContainSubstring("secondary-key-cert-data"))
		})

		It("should return certificate without key header using default key", func() {
			cert := `-----BEGIN CERTIFICATE-----
default-cert-data
-----END CERTIFICATE-----`

			mockSealer.EXPECT().
				CertificateWithKey(gomock.Any(), "").
				Return([]byte(cert), nil)

			req, _ := http.NewRequest(http.MethodGet, "/api/certificate", http.NoBody)
			router.ServeHTTP(w, req)
			Ω(w.Code).Should(Equal(http.StatusOK))
			Ω(w.Body.String()).Should(ContainSubstring("default-cert-data"))
		})

		It("should raw seal with the primary key", func() {
			rawBody := `{"name":"my-secret","namespace":"default","data":{"password":"c2VjcmV0"}}`

			mockSealer.EXPECT().
				RawWithKey(gomock.Any(), "primary", gomock.Any()).
				Return([]byte("sealed-primary-data"), nil)

			req, _ := http.NewRequest(http.MethodPost, "/api/raw", bytes.NewReader([]byte(rawBody)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Sealing-Key", "primary")
			router.ServeHTTP(w, req)
			Ω(w.Code).Should(Equal(http.StatusOK))
			Ω(w.Body.String()).Should(ContainSubstring("sealed-primary-data"))
		})

		It("should raw seal with the secondary key", func() {
			rawBody := `{"name":"my-secret","namespace":"default","data":{"password":"c2VjcmV0"}}`

			mockSealer.EXPECT().
				RawWithKey(gomock.Any(), "secondary", gomock.Any()).
				Return([]byte("sealed-secondary-data"), nil)

			req, _ := http.NewRequest(http.MethodPost, "/api/raw", bytes.NewReader([]byte(rawBody)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Sealing-Key", "secondary")
			router.ServeHTTP(w, req)
			Ω(w.Code).Should(Equal(http.StatusOK))
			Ω(w.Body.String()).Should(ContainSubstring("sealed-secondary-data"))
		})

		It("should render index page with multi-key support", func() {
			req, _ := http.NewRequest(http.MethodGet, "/", http.NoBody)
			router.ServeHTTP(w, req)
			Ω(w.Code).Should(Equal(http.StatusOK))
			Ω(w.Body.String()).Should(ContainSubstring(`:items="sealingKeys"`))
		})
	})
})
