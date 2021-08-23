package secrets

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
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
	secretData["data"] = make(map[string]interface{})
	for key, value := range secretData["stringData"].(map[string]interface{}) {
		secretData["data"].(map[string]interface{})[key] = base64.StdEncoding.EncodeToString([]byte(value.(string)))
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
	secretData["stringData"] = make(map[string]interface{})
	for key, value := range secretData["data"].(map[string]interface{}) {
		decoded, err := base64.StdEncoding.DecodeString(value.(string))
		if err != nil {
			return err
		}
		secretData["stringData"].(map[string]interface{})[key] = string(decoded)
	}
	delete(secretData, "data")
	return nil
}
