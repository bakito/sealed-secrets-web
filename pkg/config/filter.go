package config

import "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

type FieldFilter struct {
	Skip      [][]string `yaml:"skip"`
	SkipIfNil [][]string `yaml:"skipIfNil"`
}

func (ff *FieldFilter) Apply(sec map[string]interface{}) {
	for _, fieldPath := range ff.Skip {
		unstructured.RemoveNestedField(sec, fieldPath...)
	}

	for _, fieldPath := range ff.SkipIfNil {
		removeFieldIfNull(sec, fieldPath...)
	}
}

func removeFieldIfNull(sec map[string]interface{}, fields ...string) {
	path := fields[:len(fields)-1]
	name := fields[len(fields)-1]
	if m, ok, _ := unstructured.NestedMap(sec, path...); ok {
		f := m[name]
		if f == nil {
			delete(m, name)
			_ = unstructured.SetNestedMap(sec, m, path...)
		}
	}
}
