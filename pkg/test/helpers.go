package test

import (
	"github.com/onsi/gomega"
)

func SubMap(data map[string]any, fields ...string) map[string]any {
	gomega.Ω(fields).ShouldNot(gomega.BeEmpty())
	value, ok := data[fields[0]]
	gomega.Ω(ok).Should(gomega.BeTrue())
	gomega.Ω(value).Should(gomega.BeAssignableToTypeOf(make(map[string]any)))
	sm, ok := value.(map[string]any)
	gomega.Ω(ok).Should(gomega.BeTrue())
	if len(fields) == 1 {
		return sm
	}
	return SubMap(sm, fields[1:]...)
}
