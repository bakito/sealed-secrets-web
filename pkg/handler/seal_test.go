package handler

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Seal", func() {
	Context("removeNullFields", func() {

		It("should remove nil fields", func() {
			secretData := map[string]interface{}{
				"metadata": map[string]interface{}{
					"creationTimestamp": nil,
				},
				"spec": map[string]interface{}{
					"template": map[string]interface{}{
						"data": nil,
						"metadata": map[string]interface{}{
							"creationTimestamp": nil,
						},
					},
				},
			}

			removeNullFields(secretData)

			Ω(secretData["metadata"]).ShouldNot(HaveKey("creationTimestamp"))
			Ω(subMap(secretData, "spec", "template")).ShouldNot(HaveKey("data"))
			Ω(subMap(secretData, "spec", "template", "metadata")).ShouldNot(HaveKey("creationTimestamp"))
		})
		It("should keep non nil fields", func() {
			secretData := map[string]interface{}{
				"metadata": map[string]interface{}{
					"creationTimestamp": "00:00",
				},
				"spec": map[string]interface{}{
					"template": map[string]interface{}{
						"data": map[string]interface{}{},
						"metadata": map[string]interface{}{
							"creationTimestamp": "00:00",
						},
					},
				},
			}

			removeNullFields(secretData)

			Ω(secretData["metadata"]).Should(HaveKey("creationTimestamp"))
			Ω(subMap(secretData, "spec", "template")).Should(HaveKey("data"))
			Ω(subMap(secretData, "spec", "template", "metadata")).Should(HaveKey("creationTimestamp"))
		})
	})
})

func subMap(data map[string]interface{}, fields ...string) map[string]interface{} {
	Ω(fields).ShouldNot(BeEmpty())
	value := data[fields[0]]
	Ω(value).Should(BeAssignableToTypeOf(make(map[string]interface{})))
	sm := value.(map[string]interface{})
	if len(fields) == 1 {
		return sm
	}
	return subMap(sm, fields[1:]...)
}
