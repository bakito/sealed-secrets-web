package handler

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Encode", func() {
	Context("encodeData", func() {
		It("should convert stringData to data and decode the values", func() {
			secretData := map[string]interface{}{
				"stringData": map[string]interface{}{
					"username": "foo",
					"password": "bar",
				},
			}
			encodeData(secretData)
			Ω(secretData).Should(HaveKey("data"))
			Ω(secretData).ShouldNot(HaveKey("stringData"))

			Ω(subMap(secretData, "data")["username"]).Should(Equal("Zm9v"))
			Ω(subMap(secretData, "data")["password"]).Should(Equal("YmFy"))
		})
		It("should not change data", func() {
			secretData := map[string]interface{}{
				"stringData": map[string]interface{}{
					"username": "foo",
					"password": "bar",
				},
			}
			encodeData(secretData)
			Ω(secretData).Should(HaveKey("data"))
			Ω(secretData).ShouldNot(HaveKey("stringData"))

			Ω(subMap(secretData, "data")["username"]).Should(Equal("Zm9v"))
			Ω(subMap(secretData, "data")["password"]).Should(Equal("YmFy"))
		})
	})
})
