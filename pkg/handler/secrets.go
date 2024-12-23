package handler

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"

	"github.com/bakito/sealed-secrets-web/pkg/config"
	ssClient "github.com/bitnami-labs/sealed-secrets/pkg/client/clientset/versioned/typed/sealedsecrets/v1alpha1"
	"github.com/gin-gonic/gin"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"
)

// BuildClients build the  k82 clients
func BuildClients(
	clientConfig clientcmd.ClientConfig,
	disableLoadSecrets bool,
) (corev1.CoreV1Interface, ssClient.BitnamiV1alpha1Interface, error) {
	if disableLoadSecrets {
		return nil, nil, nil
	}
	conf, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, nil, err
	}

	restClient, err := corev1.NewForConfig(conf)
	if err != nil {
		return nil, nil, err
	}

	ssCl, err := ssClient.NewForConfig(conf)
	if err != nil {
		return nil, nil, err
	}

	return restClient, ssCl, nil
}

// SecretsHandler handles our secrets operations.
type SecretsHandler struct {
	coreClient         corev1.CoreV1Interface
	ssClient           ssClient.BitnamiV1alpha1Interface
	disableLoadSecrets bool
	includeNamespaces  map[string]bool
}

// NewHandler creates a new secret handler.
func NewHandler(
	coreClient corev1.CoreV1Interface,
	ssCl ssClient.BitnamiV1alpha1Interface,
	cfg *config.Config,
) *SecretsHandler {
	inMap := make(map[string]bool)
	for _, n := range cfg.IncludeNamespaces {
		inMap[n] = true
	}
	return &SecretsHandler{
		ssClient:           ssCl,
		coreClient:         coreClient,
		disableLoadSecrets: cfg.DisableLoadSecrets,
		includeNamespaces:  inMap,
	}
}

// List returns a list of all secrets.
func (h *SecretsHandler) list(ctx context.Context) ([]Secret, error) {
	var secrets []Secret
	fmt.Printf("secrets: %#v\n", secrets)
	if h.disableLoadSecrets {
		return secrets, nil
	}

	if len(h.includeNamespaces) > 0 {
		for ns := range h.includeNamespaces {
			list, err := h.listForNamespace(ctx, ns)
			if err != nil {
				return nil, err
			}
			secrets = append(secrets, list...)
		}
	} else {
		list, err := h.listForNamespace(ctx, "")
		if err != nil {
			return nil, err
		}
		secrets = append(secrets, list...)
	}

	sort.Slice(secrets, func(i, j int) bool {
		if secrets[i].Namespace == secrets[j].Namespace {
			return secrets[i].Name < secrets[j].Name
		}
		return secrets[i].Namespace < secrets[j].Namespace
	})

	return secrets, nil
}

func (h *SecretsHandler) listForNamespace(ctx context.Context, ns string) ([]Secret, error) {
	var secrets []Secret
	ssList, err := h.ssClient.SealedSecrets(ns).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, item := range ssList.Items {
		secrets = append(secrets, Secret{Namespace: item.Namespace, Name: item.Name})
	}
	return secrets, nil
}

// GetSecret returns a secret by name in the given namespace.
func (h *SecretsHandler) GetSecret(ctx context.Context, namespace, name string) (*v1.Secret, error) {
	if h.disableLoadSecrets {
		return nil, nil
	}
	if len(h.includeNamespaces) > 0 && !h.includeNamespaces[namespace] {
		return nil, fmt.Errorf("namespace '%s' is not allowed", namespace)
	}
	secret, err := h.coreClient.Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	secret.TypeMeta = metav1.TypeMeta{
		APIVersion: "v1",
		Kind:       "Secret",
	}
	secret.ObjectMeta.ManagedFields = nil
	secret.ObjectMeta.OwnerReferences = nil
	secret.ObjectMeta.CreationTimestamp = metav1.Time{}
	secret.ObjectMeta.ResourceVersion = ""
	secret.ObjectMeta.UID = ""

	return secret, nil
}

func (h *SecretsHandler) AllSecrets(c *gin.Context) {
	if h.disableLoadSecrets {
		c.JSON(http.StatusForbidden, gin.H{"error": "Loading secrets is disabled"})
		return
	}

	sec, err := h.list(c)
	if err != nil {
		log.Printf("Error in %s: %v\n", Sanitize(c.Request.URL.Path), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"secrets": sec})
}

func (h *SecretsHandler) Secret(c *gin.Context) {
	contentType, outputFormat, done := NegotiateFormat(c)
	if done {
		return
	}

	if h.disableLoadSecrets {
		c.JSON(http.StatusForbidden, gin.H{"error": "Loading secrets is disabled"})
		return
	}

	// Load existing secret.
	namespace := Sanitize(c.Param("namespace"))
	name := Sanitize(c.Param("name"))
	secret, err := h.GetSecret(c, namespace, name)
	if err != nil {
		log.Printf("Error in %s: %v\n", Sanitize(c.Request.URL.Path), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	encode, err := encodeSecret(secret, outputFormat)
	if err != nil {
		log.Printf("Error in %s: %v\n", Sanitize(c.Request.URL.Path), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Data(http.StatusOK, contentType, encode)
}

func encodeSecret(secret *v1.Secret, outputFormat string) ([]byte, error) {
	var contentType string
	switch strings.ToLower(outputFormat) {
	case "json", "":
		contentType = runtime.ContentTypeJSON
	case "yaml":
		contentType = runtime.ContentTypeYAML
	default:
		return nil, fmt.Errorf("unsupported output format: %s", outputFormat)
	}

	info, ok := runtime.SerializerInfoForMediaType(scheme.Codecs.SupportedMediaTypes(), contentType)
	if !ok {
		return nil, fmt.Errorf("unsupported output format: %s", outputFormat)
	}

	prettyEncoder := info.PrettySerializer
	if prettyEncoder == nil {
		prettyEncoder = info.Serializer
	}
	encoder := scheme.Codecs.EncoderForVersion(prettyEncoder, schema.GroupVersion{Group: "", Version: "v1"})

	return runtime.Encode(encoder, secret)
}

type Secret struct {
	Namespace string `json:"namespace" yaml:"namespace"`
	Name      string `json:"name"      yaml:"name"`
}
