package secrets

import (
	"encoding/json"
	"fmt"

	ssClient "github.com/bitnami-labs/sealed-secrets/pkg/client/clientset/versioned/typed/sealed-secrets/v1alpha1"
	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"
)

// Handler handles our secrets operations.
type Handler struct {
	clientConfig clientcmd.ClientConfig
	outputFormat string
	restClient   *corev1.CoreV1Client
	ssClient     *ssClient.BitnamiV1alpha1Client
}

// NewHandler creates a new secrets handler.
func NewHandler(clientConfig clientcmd.ClientConfig, outputFormat string) (*Handler, error) {
	conf, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	ssCl, err := ssClient.NewForConfig(conf)
	if err != nil {
		return nil, err
	}

	restClient, err := corev1.NewForConfig(conf)
	if err != nil {
		return nil, err
	}

	return &Handler{
		clientConfig: clientConfig,
		outputFormat: outputFormat,
		ssClient:     ssCl,
		restClient:   restClient,
	}, nil
}

// List returns a list of all secrets.
func (h *Handler) List() (map[string]interface{}, error) {
	ssList, err := h.ssClient.SealedSecrets("").List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var secrets map[string]interface{}
	secrets = make(map[string]interface{})

	for _, item := range ssList.Items {
		secrets[item.Name] = item.Namespace
	}

	return secrets, nil
}

// GetSecret returns a secret by name in the given namespace.
func (h *Handler) GetSecret(namespace, name string) ([]byte, error) {
	secret, err := h.restClient.Secrets(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(secret)
	if err != nil {
		return nil, err
	}

	if h.outputFormat == "yaml" {
		var secretMap map[string]interface{}
		secretMap = make(map[string]interface{})

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
