package test

import (
	"github.com/onsi/gomega"
)

func SubMap(data map[string]interface{}, fields ...string) map[string]interface{} {
	gomega.Ω(fields).ShouldNot(gomega.BeEmpty())
	value := data[fields[0]]
	gomega.Ω(value).Should(gomega.BeAssignableToTypeOf(make(map[string]interface{})))
	sm := value.(map[string]interface{})
	if len(fields) == 1 {
		return sm
	}
	return SubMap(sm, fields[1:]...)
}
