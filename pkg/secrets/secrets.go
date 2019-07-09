package secrets

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
)

// SecretHandler handles encoding and decoding of secrets.
type SecretHandler struct {
	outputFormat string
}

// Secret is the representation of our secret.
// Our secret is very similar to the v1.Secret but with some small differences.
type Secret struct {
	Kind       string `yaml:"kind,omitempty" json:"kind,omitempty"`
	APIVersion string `yaml:"apiVersion,omitempty" json:"apiVersion,omitempty"`

	MetaData struct {
		Name              string            `yaml:"name" json:"name"`
		Namespace         string            `yaml:"namespace,omitempty" json:"namespace,omitempty"`
		CreationTimestamp string            `yaml:"creationTimestamp,omitempty" json:"creationTimestamp,omitempty"`
		Labels            map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`
		Annotations       map[string]string `yaml:"annotations,omitempty" json:"annotations,omitempty"`
	} `yaml:"metadata" json:"metadata"`

	Data       map[string]string `yaml:"data,omitempty" json:"data,omitempty"`
	StringData map[string]string `yaml:"stringData,omitempty" json:"stringData,omitempty"`
	Type       v1.SecretType     `yaml:"type,omitempty" json:"type,omitempty"`
}

// NewSecretHandler returns a new secret handler.
func NewSecretHandler(outputFormat string) *SecretHandler {
	return &SecretHandler{
		outputFormat: outputFormat,
	}
}

// Encode handles the base64 encoding of the 'data' and 'stringData' fields of a secret.
func (h *SecretHandler) Encode(data string) ([]byte, error) {
	secret := &Secret{}

	if h.outputFormat == "yaml" {
		err := yaml.Unmarshal([]byte(data), &secret)
		if err != nil {
			return nil, err
		}
	} else if h.outputFormat == "json" {
		err := json.Unmarshal([]byte(data), &secret)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("unsupported output format: %s", h.outputFormat)
	}

	for key, value := range secret.Data {
		secret.Data[key] = base64.StdEncoding.EncodeToString([]byte(value))
	}

	for key, value := range secret.StringData {
		secret.StringData[key] = base64.StdEncoding.EncodeToString([]byte(value))
	}

	if h.outputFormat == "yaml" {
		return yaml.Marshal(secret)
	} else if h.outputFormat == "json" {
		return json.Marshal(secret)
	}

	return nil, fmt.Errorf("unsupported output format: %s", h.outputFormat)
}

// Decode handles the base64 decoding of the 'data' and 'stringData' fields of a secret.
func (h *SecretHandler) Decode(data string) ([]byte, error) {
	secret := &Secret{}

	if h.outputFormat == "yaml" {
		err := yaml.Unmarshal([]byte(data), &secret)
		if err != nil {
			return nil, err
		}
	} else if h.outputFormat == "json" {
		err := json.Unmarshal([]byte(data), &secret)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("unsupported output format: %s", h.outputFormat)
	}

	for key, value := range secret.Data {
		decoded, err := base64.StdEncoding.DecodeString(value)
		if err != nil {
			return nil, err
		}

		secret.Data[key] = string(decoded)
	}

	for key, value := range secret.StringData {
		decoded, err := base64.StdEncoding.DecodeString(value)
		if err != nil {
			return nil, err
		}

		secret.StringData[key] = string(decoded)
	}

	if h.outputFormat == "yaml" {
		return yaml.Marshal(secret)
	} else if h.outputFormat == "json" {
		return json.Marshal(secret)
	}

	return nil, fmt.Errorf("unsupported output format: %s", h.outputFormat)
}
