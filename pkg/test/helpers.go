package test

import (
	. "github.com/onsi/gomega"
)

func SubMap(data map[string]interface{}, fields ...string) map[string]interface{} {
	Ω(fields).ShouldNot(BeEmpty())
	value := data[fields[0]]
	Ω(value).Should(BeAssignableToTypeOf(make(map[string]interface{})))
	sm := value.(map[string]interface{})
	if len(fields) == 1 {
		return sm
	}
	return SubMap(sm, fields[1:]...)
}
