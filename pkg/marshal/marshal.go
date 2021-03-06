package marshal

import (
	"bytes"
	"encoding/json"
	"strings"

	"gopkg.in/yaml.v3"
)

type Marshaller interface {
	Marshal(interface{}) ([]byte, error)
	Unmarshal([]byte, interface{}) error
}

func For(format string) Marshaller {
	if strings.EqualFold(format, "yaml") {
		return &yamlMarshaller{}
	}
	return &jsonMarshaller{}
}

type yamlMarshaller struct{}

func (m yamlMarshaller) Marshal(in interface{}) ([]byte, error) {
	var b bytes.Buffer
	yamlEncoder := yaml.NewEncoder(&b)
	yamlEncoder.SetIndent(2)
	err := yamlEncoder.Encode(in)
	return b.Bytes(), err
}

func (m yamlMarshaller) Unmarshal(in []byte, out interface{}) error {
	return yaml.Unmarshal(in, out)
}

type jsonMarshaller struct{}

func (m jsonMarshaller) Marshal(in interface{}) ([]byte, error) {
	return json.MarshalIndent(in, "", "  ")
}

func (m jsonMarshaller) Unmarshal(in []byte, out interface{}) error {
	return json.Unmarshal(in, out)
}
