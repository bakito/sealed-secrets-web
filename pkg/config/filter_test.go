package config

import (
	"flag"
	"io"
	"os"

	. "github.com/bakito/sealed-secrets-web/pkg/test"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var testConfigFile = "../../testdata/config.yaml"

var _ = Describe("Filter", func() {
	Context("removeNullFields", func() {
		var ff *FieldFilter
		BeforeEach(func() {
			resetFlagsForTesting()
			cfg, err := Parse()
			Ω(err).ShouldNot(HaveOccurred())
			ff = cfg.FieldFilter
		})
		It("should remove nil fields", func() {
			secretData := map[string]interface{}{
				"spec": map[string]interface{}{
					"template": map[string]interface{}{
						"data": nil,
						"metadata": map[string]interface{}{
							"creationTimestamp": nil,
						},
					},
				},
			}

			ff.Apply(secretData)

			Ω(SubMap(secretData, "spec", "template")).ShouldNot(HaveKey("data"))
			Ω(SubMap(secretData, "spec", "template", "metadata")).ShouldNot(HaveKey("creationTimestamp"))
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

			ff.Apply(secretData)

			Ω(secretData["metadata"]).Should(HaveKey("creationTimestamp"))
			Ω(SubMap(secretData, "spec", "template")).Should(HaveKey("data"))
		})
	})
	Context("removeRuntimeFields", func() {
		var ff *FieldFilter
		BeforeEach(func() {
			resetFlagsForTesting()
			f := newFlags()
			f.config = &testConfigFile
			cfg, err := parse(f)
			Ω(err).ShouldNot(HaveOccurred())
			ff = cfg.FieldFilter
		})
		It("should remove the fields", func() {
			secretData := map[string]interface{}{
				"metadata": map[string]interface{}{
					"creationTimestamp": "foo",
					"managedFields":     "foo",
					"resourceVersion":   "foo",
					"selfLink":          "foo",
					"uid":               "foo",
					"annotations": map[string]interface{}{
						"kubectl.kubernetes.io/last-applied-configuration": "foo",
						"foo": "bar",
					},
				},
			}

			ff.Apply(secretData)
			Ω(secretData["metadata"]).ShouldNot(HaveKey("creationTimestamp"))
			Ω(secretData["metadata"]).ShouldNot(HaveKey("managedFields"))
			Ω(secretData["metadata"]).ShouldNot(HaveKey("resourceVersion"))
			Ω(secretData["metadata"]).ShouldNot(HaveKey("selfLink"))
			Ω(secretData["metadata"]).ShouldNot(HaveKey("uid"))
			Ω(
				SubMap(secretData, "metadata", "annotations"),
			).ShouldNot(HaveKey("kubectl.kubernetes.io/last-applied-configuration"))
			Ω(SubMap(secretData, "metadata", "annotations")).Should(HaveKey("foo"))
		})
	})
})

func resetFlagsForTesting() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	flag.CommandLine.SetOutput(io.Discard)
}
