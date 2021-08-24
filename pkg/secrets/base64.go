package secrets

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Encode encodes the data field in a secret.
func (h *Handler) Encode(data string) ([]byte, error) {
	secretData := make(map[string]interface{})

	if h.outputFormat == "yaml" {
		err := yaml.Unmarshal([]byte(data), &secretData)
		if err != nil {
			return nil, err
		}
		encodeData(secretData)
		return yaml.Marshal(secretData)
	} else if h.outputFormat == "json" {
		err := json.Unmarshal([]byte(data), &secretData)
		if err != nil {
			return nil, err
		}
		encodeData(secretData)
		return json.MarshalIndent(secretData, "", "  ")
	}

	return nil, fmt.Errorf("unsupported output format: %s", h.outputFormat)
}

func encodeData(secretData map[string]interface{}) {

	if _, ok := secretData["data"]; ok {
		// already encoded
		return
	}
	// set empty data
	secretData["data"] = make(map[string]interface{})
	if m, ok, _ := unstructured.NestedMap(secretData, "stringData"); ok {
		for key, value := range m {
			b := []byte(value.(string))
			_ = unstructured.SetNestedField(secretData, base64.StdEncoding.EncodeToString(b), "data", key)
		}
	}
	delete(secretData, "stringData")
}

// Decode decodes the data field in a secret.
func (h *Handler) Decode(data string) ([]byte, error) {
	secretData := make(map[string]interface{})

	if h.outputFormat == "yaml" {
		err := yaml.Unmarshal([]byte(data), &secretData)
		if err != nil {
			return nil, err
		}

		if err = decodeData(secretData); err != nil {
			return nil, err
		}

		return yaml.Marshal(secretData)
	} else if h.outputFormat == "json" {
		err := json.Unmarshal([]byte(data), &secretData)
		if err != nil {
			return nil, err
		}

		if err = decodeData(secretData); err != nil {
			return nil, err
		}

		return json.MarshalIndent(secretData, "", "  ")
	}

	return nil, fmt.Errorf("unsupported output format: %s", h.outputFormat)
}

func decodeData(secretData map[string]interface{}) error {
	if _, ok := secretData["stringData"]; ok {
		// already decoded
		return nil
	}
	secretData["stringData"] = make(map[string]interface{})
	if m, ok, _ := unstructured.NestedMap(secretData, "data"); ok {
		for key, value := range m {
			decoded, err := base64.StdEncoding.DecodeString(value.(string))
			if err != nil {
				return err
			}
			_ = unstructured.SetNestedField(secretData, string(decoded), "stringData", key)
		}
	}
	delete(secretData, "data")
	return nil
}
