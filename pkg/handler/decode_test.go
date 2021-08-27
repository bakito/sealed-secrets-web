package handler

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Decode", func() {
	Context("decodeData", func() {
	})
	It("should convert data to stringData and decode the values", func() {
		secretData := map[string]interface{}{
			"data": map[string]interface{}{
				"username": "Zm9v",
				"password": "YmFy",
			},
		}
		err := decodeData(secretData)
		Ω(err).ShouldNot(HaveOccurred())
		Ω(secretData).Should(HaveKey("stringData"))
		Ω(secretData).ShouldNot(HaveKey("data"))

		Ω(secretData["stringData"].(map[string]interface{})["username"]).Should(Equal("foo"))
		Ω(secretData["stringData"].(map[string]interface{})["password"]).Should(Equal("bar"))
	})
	It("should not change stringData", func() {
		secretData := map[string]interface{}{
			"stringData": map[string]interface{}{
				"username": "foo",
				"password": "bar",
			},
		}
		err := decodeData(secretData)
		Ω(err).ShouldNot(HaveOccurred())
		Ω(secretData).Should(HaveKey("stringData"))
		Ω(secretData).ShouldNot(HaveKey("data"))

		Ω(secretData["stringData"].(map[string]interface{})["username"]).Should(Equal("foo"))
		Ω(secretData["stringData"].(map[string]interface{})["password"]).Should(Equal("bar"))
	})
})
