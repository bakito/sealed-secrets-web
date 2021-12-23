package secrets

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/bakito/sealed-secrets-web/pkg/config"
	ssClient "github.com/bitnami-labs/sealed-secrets/pkg/client/clientset/versioned/typed/sealed-secrets/v1alpha1"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"
)

// BuildClients build the  k82 clients
func BuildClients(clientConfig clientcmd.ClientConfig, disableLoadSecrets bool) (corev1.CoreV1Interface, ssClient.BitnamiV1alpha1Interface, error) {
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

// Handler handles our secrets operations.
type Handler struct {
	outputFormat       string
	coreClient         corev1.CoreV1Interface
	ssClient           ssClient.BitnamiV1alpha1Interface
	disableLoadSecrets bool
	includeNamespaces  map[string]bool
}

// NewHandler creates a new secrets handler.
func NewHandler(coreClient corev1.CoreV1Interface, ssCl ssClient.BitnamiV1alpha1Interface, cfg *config.Config) *Handler {
	inMap := make(map[string]bool)
	for _, n := range cfg.IncludeNamespaces {
		inMap[n] = true
	}
	return &Handler{
		outputFormat:       cfg.OutputFormat,
		ssClient:           ssCl,
		coreClient:         coreClient,
		disableLoadSecrets: cfg.DisableLoadSecrets,
		includeNamespaces:  inMap,
	}
}

// List returns a list of all secrets.
func (h *Handler) list() ([]Secret, error) {
	var secrets []Secret
	if h.disableLoadSecrets {
		return secrets, nil
	}
	ssList, err := h.ssClient.SealedSecrets("").List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, item := range ssList.Items {
		if len(h.includeNamespaces) == 0 || h.includeNamespaces[item.Namespace] {
			secrets = append(secrets, Secret{Namespace: item.Namespace, Name: item.Name})
		}
	}

	sort.Slice(secrets, func(i, j int) bool {
		if secrets[i].Namespace == secrets[j].Namespace {
			return secrets[i].Name < secrets[j].Name
		}
		return secrets[i].Namespace < secrets[j].Namespace
	})

	return secrets, nil
}

// GetSecret returns a secret by name in the given namespace.
func (h *Handler) GetSecret(namespace, name string) ([]byte, error) {
	if h.disableLoadSecrets {
		return nil, nil
	}
	if len(h.includeNamespaces) > 0 && !h.includeNamespaces[namespace] {
		return nil, fmt.Errorf("namespace '%s' is not allowed", namespace)
	}
	secret, err := h.coreClient.Secrets(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	secret.TypeMeta = metav1.TypeMeta{
		APIVersion: "v1",
		Kind:       "Secret",
	}
	secret.ObjectMeta.ManagedFields = nil

	jsonData, err := json.MarshalIndent(secret, "", "  ")
	if err != nil {
		return nil, err
	}

	if strings.EqualFold(h.outputFormat, "yaml") {
		secretMap := make(map[string]interface{})

		err = json.Unmarshal(jsonData, &secretMap)
		if err != nil {
			return nil, err
		}

		return yaml.Marshal(secretMap)
	} else if h.outputFormat == "json" {
		return jsonData, nil
	}

	return nil, fmt.Errorf("unsupported output format: %s", h.outputFormat)
}

func (h *Handler) AllSecrets(c *gin.Context) {
	if h.disableLoadSecrets {
		c.JSON(http.StatusForbidden, gin.H{"error": "Loading secrets is disabled"})
		return
	}

	sec, err := h.list()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, sec)
}

func (h *Handler) Secret(c *gin.Context) {
	if h.disableLoadSecrets {
		c.JSON(http.StatusForbidden, gin.H{"error": "Loading secrets is disabled"})
		return
	}

	// Load existing secret.
	namespace := c.Param("namespace")
	name := c.Param("name")
	secret, err := h.GetSecret(namespace, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	data := struct {
		Secret string `json:"secret"`
	}{
		string(secret),
	}

	c.JSON(http.StatusOK, &data)
}

type Secret struct {
	Namespace string `json:"namespace" yaml:"namespace"`
	Name      string `json:"name" yaml:"name"`
}
