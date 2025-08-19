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

// BuildClients builds the Kubernetes clients
// This function creates two clients: one for standard Kubernetes resources and one for Sealed Secrets
func BuildClients(
	clientConfig clientcmd.ClientConfig, // Configuration for the Kubernetes connection
	disableLoadSecrets bool, // Flag to disable loading secrets
) (corev1.CoreV1Interface, ssClient.BitnamiV1alpha1Interface, error) {
	// If loading secrets is disabled, return empty clients
	if disableLoadSecrets {
		return nil, nil, nil
	}

	// Create client configuration from the provided config
	conf, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, nil, err
	}

	// Create standard Kubernetes client for core resources (including Secrets)
	restClient, err := corev1.NewForConfig(conf)
	if err != nil {
		return nil, nil, err
	}

	// Create specialized client for Sealed Secrets
	ssCl, err := ssClient.NewForConfig(conf)
	if err != nil {
		return nil, nil, err
	}

	return restClient, ssCl, nil
}

// SecretsHandler manages all operations for secrets
type SecretsHandler struct {
	coreClient         corev1.CoreV1Interface            // Client for standard Kubernetes resources
	ssClient           ssClient.BitnamiV1alpha1Interface // Client for Sealed Secrets
	disableLoadSecrets bool                              // Flag whether secrets can be loaded
	includeNamespaces  map[string]bool                   // Map for quick checking if a namespace is included
	config             *config.Config                    // General configuration
}

// NewHandler creates a new secrets handler
func NewHandler(
	coreClient corev1.CoreV1Interface,
	ssCl ssClient.BitnamiV1alpha1Interface,
	cfg *config.Config,
) *SecretsHandler {
	// Create a map for quick lookups of included namespaces
	inMap := make(map[string]bool)
	for _, n := range cfg.IncludeNamespaces {
		inMap[n] = true
	}

	// Create and return the new handler
	return &SecretsHandler{
		ssClient:           ssCl,
		coreClient:         coreClient,
		disableLoadSecrets: cfg.DisableLoadSecrets,
		includeNamespaces:  inMap,
		config:             cfg,
	}
}

// NamespacesMatch checks if a namespace is allowed according to the filter rules
func (h *SecretsHandler) NamespacesMatch(namespaces []string) map[string]bool {
	matchedNamespaces := make(map[string]bool)
	// If regular expressions should be used for filtering
	if h.config.UseRegex {
		// Process inclusion rules with RegEx
		if len(h.config.IncludeNamespacesRegex) > 0 {
			for _, r := range h.config.IncludeNamespacesRegex {
				// Check all namespaces and include those matching the RegEx
				for _, ns := range namespaces {
					matched := r.FindString(ns) == ns
					if matched {
						matchedNamespaces[ns] = true
					}
				}
			}
		} else {
			// If no inclusion rules, use all namespaces
			for _, ns := range namespaces {
				matchedNamespaces[ns] = true
			}
		}

		// Process exclusion rules with RegEx
		for _, r := range h.config.ExcludeNamespacesRegex {
			// Remove namespaces that match the exclusion RegEx
			for ns := range matchedNamespaces {
				matched := r.FindString(ns) == ns
				if matched {
					// Remove the element from the slice (without preserving order)
					matchedNamespaces[ns] = false
				}
			}
		}
	} else {
		// Direct string comparisons for filtering (without RegEx)

		// Add all explicitly included namespaces
		for _, ns := range h.config.IncludeNamespaces {
			matchedNamespaces[ns] = true
		}

		// Apply exclusion logic
		if len(h.config.ExcludeNamespaces) > 0 {
			// If no inclusion rules are defined, use all available namespaces
			if len(h.config.IncludeNamespaces) < 1 {
				for _, ns := range namespaces {
					matchedNamespaces[ns] = true
				}
			}

			// Remove namespaces that are in the exclusion list
			for _, exc := range h.config.ExcludeNamespaces {
				for ns := range matchedNamespaces {
					if ns == exc {
						matchedNamespaces[ns] = false
					}
				}
			}
		}
	}
	return matchedNamespaces
}

// list returns a list of all secrets that match the filter criteria
func (h *SecretsHandler) list(ctx context.Context) ([]Secret, error) {
	var secrets []Secret

	// If loading secrets is disabled, return an empty list
	if h.disableLoadSecrets {
		return secrets, nil
	}

	// If namespace filters are defined (inclusion or exclusion)
	if len(h.config.ExcludeNamespaces) > 0 || len(h.config.IncludeNamespaces) > 0 {
		// Exclusion always takes precedence over inclusion
		var nsNameList []string

		// If no inclusion rules are defined, gather all available namespaces
		if len(h.config.IncludeNamespaces) <= 0 || h.config.UseRegex {
			nsList, err := h.coreClient.Namespaces().List(ctx, metav1.ListOptions{})
			if err != nil {
				return nil, err
			}
			for _, namespace := range nsList.Items {
				nsNameList = append(nsNameList, namespace.Name)
			}
		} else {
			nsNameList = h.config.IncludeNamespaces
		}

		matchedNamespaces := h.NamespacesMatch(nsNameList)

		// Get secrets for all matching namespaces
		for ns, v := range matchedNamespaces {
			if !v {
				continue
			}
			list, err := h.listForNamespace(ctx, ns)
			if err != nil {
				return nil, err
			}
			secrets = append(secrets, list...)
		}
	} else {
		// If no filters are specified, list all secrets in all namespaces
		// Empty string ("") means "all namespaces"
		list, err := h.listForNamespace(ctx, "")
		if err != nil {
			return nil, err
		}
		secrets = append(secrets, list...)
	}

	// Sort secrets: first by namespace, then by name
	sort.Slice(secrets, func(i, j int) bool {
		if secrets[i].Namespace == secrets[j].Namespace {
			return secrets[i].Name < secrets[j].Name
		}
		return secrets[i].Namespace < secrets[j].Namespace
	})

	return secrets, nil
}

// listForNamespace retrieves all Sealed Secrets in a specific namespace
func (h *SecretsHandler) listForNamespace(ctx context.Context, ns string) ([]Secret, error) {
	var secrets []Secret

	// API call to retrieve all SealedSecrets in the specified namespace
	// Empty string ("") means "all namespaces"
	ssList, err := h.ssClient.SealedSecrets(ns).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	// Convert SealedSecrets to the simpler Secret structure
	for _, item := range ssList.Items {
		secrets = append(secrets, Secret{Namespace: item.Namespace, Name: item.Name})
	}
	return secrets, nil
}

// GetSecret returns a single secret by namespace and name
func (h *SecretsHandler) GetSecret(ctx context.Context, namespace, name string) (*v1.Secret, error) {
	// If loading secrets is disabled, return null
	if h.disableLoadSecrets {
		return nil, nil
	}

	// Check if the namespace is allowed according to the filter rules
	if len(h.config.ExcludeNamespaces) > 0 || len(h.config.IncludeNamespaces) > 0 {
		namespaces := h.NamespacesMatch([]string{namespace})
		if !namespaces[namespace] {
			return nil, fmt.Errorf("namespace '%s' is not allowed", namespace)
		}
	}

	// Retrieve the secret from the Kubernetes cluster
	secret, err := h.coreClient.Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// Clean up secret metadata (remove unnecessary fields)
	secret.TypeMeta = metav1.TypeMeta{
		APIVersion: "v1",
		Kind:       "Secret",
	}
	secret.ManagedFields = nil
	secret.OwnerReferences = nil
	secret.CreationTimestamp = metav1.Time{}
	secret.ResourceVersion = ""
	secret.UID = ""

	return secret, nil
}

// AllSecrets is an HTTP handler that returns a list of all available secrets
func (h *SecretsHandler) AllSecrets(c *gin.Context) {
	// If loading secrets is disabled, return an error
	if h.disableLoadSecrets {
		c.JSON(http.StatusForbidden, gin.H{"error": "Loading secrets is disabled"})
		return
	}

	// Retrieve secrets
	sec, err := h.list(c)
	if err != nil {
		// Log error and return it to the client
		log.Printf("Error in %s: %v\n", Sanitize(c.Request.URL.Path), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Successful response with the list of secrets
	c.JSON(http.StatusOK, gin.H{"secrets": sec})
}

// Secret is an HTTP handler that returns a single secret
func (h *SecretsHandler) Secret(c *gin.Context) {
	// Determine the response format (JSON or YAML)
	contentType, outputFormat, done := NegotiateFormat(c)
	if done {
		return // If the format is not supported, an error response was already sent
	}

	// If loading secrets is disabled, return an error
	if h.disableLoadSecrets {
		c.JSON(http.StatusForbidden, gin.H{"error": "Loading secrets is disabled"})
		return
	}

	// Extract and sanitize parameters from the request
	namespace := Sanitize(c.Param("namespace"))
	name := Sanitize(c.Param("name"))

	// Retrieve the secret
	secret, err := h.GetSecret(c, namespace, name)
	if err != nil {
		// Log error and return it to the client
		log.Printf("Error in %s: %v\n", Sanitize(c.Request.URL.Path), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Encode the secret in the desired format
	encode, err := encodeSecret(secret, outputFormat)
	if err != nil {
		// Log encoding error and return it
		log.Printf("Error in %s: %v\n", Sanitize(c.Request.URL.Path), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Successful response with the encoded secret
	c.Data(http.StatusOK, contentType, encode)
}

// encodeSecret encodes a Secret object into the specified format (JSON or YAML)
func encodeSecret(secret *v1.Secret, outputFormat string) ([]byte, error) {
	var contentType string

	// Determine content type based on the desired output format
	switch strings.ToLower(outputFormat) {
	case "json", "":
		contentType = runtime.ContentTypeJSON
	case "yaml":
		contentType = runtime.ContentTypeYAML
	default:
		return nil, fmt.Errorf("unsupported output format: %s", outputFormat)
	}

	// Get serializer for the desired format
	info, ok := runtime.SerializerInfoForMediaType(scheme.Codecs.SupportedMediaTypes(), contentType)
	if !ok {
		return nil, fmt.Errorf("unsupported output format: %s", outputFormat)
	}

	// Use "pretty" serializer if available, otherwise use standard serializer
	prettyEncoder := info.PrettySerializer
	if prettyEncoder == nil {
		prettyEncoder = info.Serializer
	}

	// Create encoder for the API version
	encoder := scheme.Codecs.EncoderForVersion(prettyEncoder, schema.GroupVersion{Group: "", Version: "v1"})

	// Encode and return the secret
	return runtime.Encode(encoder, secret)
}

// Secret represents the basic data of a Kubernetes Secret
type Secret struct {
	Namespace string `json:"namespace" yaml:"namespace"` // Namespace where the secret is located
	Name      string `json:"name"      yaml:"name"`      // Name of the secret
}
