package secrets

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v2"
)

// Encode encodes the data field in a secret.
func (h *Handler) Encode(data string) ([]byte, error) {
	var secretData map[string]interface{}
	secretData = make(map[string]interface{})

	if h.outputFormat == "yaml" {
		err := yaml.Unmarshal([]byte(data), &secretData)
		if err != nil {
			return nil, err
		}

		for key, value := range secretData["data"].(map[interface{}]interface{}) {
			secretData["data"].(map[interface{}]interface{})[key] = base64.StdEncoding.EncodeToString([]byte(value.(string)))
		}

		return yaml.Marshal(secretData)
	} else if h.outputFormat == "json" {
		err := json.Unmarshal([]byte(data), &secretData)
		if err != nil {
			return nil, err
		}

		for key, value := range secretData["data"].(map[string]interface{}) {
			secretData["data"].(map[string]interface{})[key] = base64.StdEncoding.EncodeToString([]byte(value.(string)))
		}

		return json.Marshal(secretData)
	}

	return nil, fmt.Errorf("unsupported output format: %s", h.outputFormat)
}

// Decode decodes the data field in a secret.
func (h *Handler) Decode(data string) ([]byte, error) {
	var secretData map[string]interface{}
	secretData = make(map[string]interface{})

	if h.outputFormat == "yaml" {
		err := yaml.Unmarshal([]byte(data), &secretData)
		if err != nil {
			return nil, err
		}

		for key, value := range secretData["data"].(map[interface{}]interface{}) {
			decoded, err := base64.StdEncoding.DecodeString(value.(string))
			if err != nil {
				return nil, err
			}

			secretData["data"].(map[interface{}]interface{})[key] = string(decoded)
		}

		return yaml.Marshal(secretData)
	} else if h.outputFormat == "json" {
		err := json.Unmarshal([]byte(data), &secretData)
		if err != nil {
			return nil, err
		}

		for key, value := range secretData["data"].(map[string]interface{}) {
			decoded, err := base64.StdEncoding.DecodeString(value.(string))
			if err != nil {
				return nil, err
			}

			secretData["data"].(map[string]interface{})[key] = string(decoded)
		}

		return json.Marshal(secretData)
	}

	return nil, fmt.Errorf("unsupported output format: %s", h.outputFormat)
}
