package handler

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Sanitize", func() {
	DescribeTable("Sanitize the path correctly",
		func(value string, expected string) {
			Ω(Sanitize(value)).Should(Equal(expected))
		},
		Entry("When value only", "foo", "foo"),
		Entry("When value has \n", "foo\nbar", "foobar"),
		Entry("When value has rn", "foo\r\nbar", "foobar"),
		Entry("When value has //", "foo//bar", "foo/bar"),
		Entry("When value has ..", "foo/../bar", "bar"),
	)
})
