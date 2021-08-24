package marshal

import (
	"gopkg.in/yaml.v3"
	"strings"
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
	return yaml.Marshal(in)
}

func (m yamlMarshaller) Unmarshal(in []byte, out interface{}) error {
	return yaml.Unmarshal(in, out)
}

type jsonMarshaller struct{}

func (m jsonMarshaller) Marshal(in interface{}) ([]byte, error) {
	return yaml.Marshal(in)
}

func (m jsonMarshaller) Unmarshal(in []byte, out interface{}) error {
	return yaml.Unmarshal(in, out)
}
