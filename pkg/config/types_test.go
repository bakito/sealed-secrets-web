package config

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Types", func() {
	Context("sanitizeWebContext", func() {
		var cfg *Config
		BeforeEach(func() {
			cfg = &Config{Web: Web{}}
		})
		DescribeTable("web context is formatted correctly",
			func(context string, expected string) {
				cfg.Web.Context = context
				Ω(sanitizeWebContext(cfg)).Should(Equal(expected))
			},
			Entry("trailing / is added", "/ssw", "/ssw/"),
			Entry("leading / is added", "ssw/", "/ssw/"),
			Entry("leading  and trailing / are added", "ssw", "/ssw/"),
			Entry("correct path is not changed", "/ssw/", "/ssw/"),

			Entry("trailing / is added with http", "http://ssw", "http://ssw/"),
			Entry("http with trailing / should not be changed", "http://ssw/", "http://ssw/"),

			Entry("trailing / is added with https", "https://ssw", "https://ssw/"),
			Entry("https with trailing / should not be changed", "https://ssw/", "https://ssw/"),
		)
	})
	Context("SealedSecrets", func() {
		var ss *SealedSecrets
		BeforeEach(func() {
			ss = &SealedSecrets{}
		})
		Context("String", func() {
			It("should print the cert URL", func() {
				ss.CertURL = "https://cert.url"
				Ω(ss.String()).Should(Equal("Cert URL: https://cert.url"))
			})
			It("should print the service name and namespace", func() {
				ss.Namespace = "sealed-secrets"
				ss.Service = "sealed-secrets-svc"
				Ω(ss.String()).Should(Equal("Namespace: sealed-secrets / ServiceName: sealed-secrets-svc"))
			})
		})
	})
	Context("Parse", func() {
		var (
			cfg *Config
			err error
			f   *flags
		)
		BeforeEach(func() {
			resetFlagsForTesting()
			f = newFlags()
			f.config = &testConfigFile
		})
		It("should set the sealedSecretsCertURL", func() {
			f.sealedSecretsCertURL = ptr("cert.url")
			cfg, err = parse(f)
			Ω(cfg.SealedSecrets.CertURL).Should(Equal("cert.url"))
			Ω(cfg.SealedSecrets.Namespace).Should(Equal("sealed-secrets"))
			Ω(cfg.SealedSecrets.Service).Should(Equal("sealed-secrets"))
		})
		It("should set the service namespace and name", func() {
			f.sealedSecretsServiceName = ptr("name")
			f.sealedSecretsServiceNamespace = ptr("namespace")
			cfg, err = parse(f)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(cfg.SealedSecrets.CertURL).Should(BeEmpty())
			Ω(cfg.SealedSecrets.Namespace).Should(Equal("namespace"))
			Ω(cfg.SealedSecrets.Service).Should(Equal("name"))
		})
		It("should set included namespaces correctly", func() {
			f.includeNamespaces = ptr("foo bar")
			cfg, err = parse(f)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(cfg.IncludeNamespaces).Should(ContainElements("foo", "bar"))
		})
		It("should read the initial secrets file", func() {
			f.initialSecretFile = &testConfigFile
			cfg, err = parse(f)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(cfg.InitialSecret).ShouldNot(BeEmpty())
		})
	})
})

func ptr(v string) *string {
	return &v
}
