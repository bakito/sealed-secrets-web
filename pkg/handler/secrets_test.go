package handler

import (
	"context"
	"regexp"

	"github.com/bakito/sealed-secrets-web/pkg/config"
	ssv1alpha1 "github.com/bitnami-labs/sealed-secrets/pkg/apis/sealedsecrets/v1alpha1"
	ssfake "github.com/bitnami-labs/sealed-secrets/pkg/client/clientset/versioned/typed/sealedsecrets/v1alpha1/fake"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"
)

var _ = Describe("SecretsHandler", func() {
	Context("NamespacesMatch", func() {
		var handler *SecretsHandler

		BeforeEach(func() {
			handler = &SecretsHandler{}
		})

		Context("without regex", func() {
			BeforeEach(func() {
				handler.config = &config.Config{UseRegex: false}
			})

			DescribeTable("include namespaces",
				func(includeNamespaces []string, inputNamespaces []string, expectedMatches map[string]bool) {
					handler.config.IncludeNamespaces = includeNamespaces
					result := handler.NamespacesMatch(inputNamespaces)

					for ns, expected := range expectedMatches {
						Ω(result).Should(HaveKeyWithValue(ns, expected))
					}
				},
				Entry("includes specified namespaces",
					[]string{"ns1", "ns2"},
					[]string{"ns1", "ns2", "ns3"},
					map[string]bool{"ns1": true, "ns2": true}),
				Entry("includes only matching namespaces",
					[]string{"prod", "staging"},
					[]string{"prod", "dev", "test"},
					map[string]bool{"prod": true}),
			)

			DescribeTable("exclude namespaces",
				func(includeNamespaces, excludeNamespaces []string, inputNamespaces []string, expectedMatches map[string]bool) {
					handler.config.IncludeNamespaces = includeNamespaces
					handler.config.ExcludeNamespaces = excludeNamespaces
					result := handler.NamespacesMatch(inputNamespaces)

					for ns, expected := range expectedMatches {
						Ω(result).Should(HaveKeyWithValue(ns, expected))
					}
				},
				Entry("excludes from all namespaces when no include list",
					[]string{},
					[]string{"ns3"},
					[]string{"ns1", "ns2", "ns3"},
					map[string]bool{"ns1": true, "ns2": true, "ns3": false}),
				Entry("excludes from include list",
					[]string{"ns1", "ns2", "ns3"},
					[]string{"ns3"},
					[]string{"ns1", "ns2", "ns3"},
					map[string]bool{"ns1": true, "ns2": true, "ns3": false}),
			)

			It("should return empty map when no filters", func() {
				result := handler.NamespacesMatch([]string{"ns1", "ns2", "ns3"})
				Ω(result).Should(BeEmpty())
			})
		})

		Context("with regex", func() {
			BeforeEach(func() {
				handler.config = &config.Config{UseRegex: true}
			})

			DescribeTable("include namespaces with regex",
				func(patterns []string, inputNamespaces []string, expectedMatches map[string]bool) {
					handler.config.IncludeNamespaces = patterns
					var regexes []*regexp.Regexp
					for _, pattern := range patterns {
						regexes = append(regexes, regexp.MustCompile(pattern))
					}
					handler.config.IncludeNamespacesRegex = regexes

					result := handler.NamespacesMatch(inputNamespaces)

					for ns, expected := range expectedMatches {
						Ω(result).Should(HaveKeyWithValue(ns, expected))
					}
				},
				Entry("matches app namespaces",
					[]string{"app-.*"},
					[]string{"app-prod", "app-staging", "kube-system"},
					map[string]bool{"app-prod": true, "app-staging": true}),
				Entry("matches multiple patterns",
					[]string{"test-.*", "prod-.*"},
					[]string{"test-app", "prod-db", "dev-api"},
					map[string]bool{"test-app": true, "prod-db": true}),
			)

			DescribeTable("exclude namespaces with regex",
				func(includePatterns, excludePatterns []string, inputNamespaces []string, expectedMatches map[string]bool) {
					handler.config.IncludeNamespaces = includePatterns
					handler.config.ExcludeNamespaces = excludePatterns

					var includeRegexes []*regexp.Regexp
					for _, pattern := range includePatterns {
						includeRegexes = append(includeRegexes, regexp.MustCompile(pattern))
					}
					handler.config.IncludeNamespacesRegex = includeRegexes

					var excludeRegexes []*regexp.Regexp
					for _, pattern := range excludePatterns {
						excludeRegexes = append(excludeRegexes, regexp.MustCompile(pattern))
					}
					handler.config.ExcludeNamespacesRegex = excludeRegexes

					result := handler.NamespacesMatch(inputNamespaces)

					for ns, expected := range expectedMatches {
						Ω(result).Should(HaveKeyWithValue(ns, expected))
					}
				},
				Entry("excludes kube system namespaces",
					[]string{},
					[]string{"kube-.*"},
					[]string{"default", "kube-system", "kube-public", "app-ns"},
					map[string]bool{"default": true, "kube-system": false, "kube-public": false, "app-ns": true}),
				Entry("includes app but excludes test",
					[]string{"app-.*"},
					[]string{"app-test.*"},
					[]string{"app-prod", "app-staging", "app-test-1", "other-ns"},
					map[string]bool{"app-prod": true, "app-staging": true, "app-test-1": false}),
			)
		})
	})

	Context("list", func() {
		var (
			handler      *SecretsHandler
			fakeClient   *fake.Clientset
			fakeSSClient *ssfake.FakeBitnamiV1alpha1
			cfg          *config.Config
		)

		BeforeEach(func() {
			fakeClient = fake.NewClientset()
			fakeSSClient = &ssfake.FakeBitnamiV1alpha1{
				Fake: &ktesting.Fake{},
			}
			cfg = &config.Config{}
		})

		JustBeforeEach(func() {
			handler = NewHandler(fakeClient.CoreV1(), fakeSSClient, cfg)
		})

		Context("when load secrets is disabled", func() {
			BeforeEach(func() {
				cfg.DisableLoadSecrets = true
			})

			It("should return empty list", func() {
				result, err := handler.list(context.Background())
				Ω(err).ShouldNot(HaveOccurred())
				Ω(result).Should(BeEmpty())
			})
		})

		Context("when no filters are configured", func() {
			It("should return all sealed secrets", func() {
				setupSealedSecretsReactor(fakeSSClient, []ssv1alpha1.SealedSecret{
					{ObjectMeta: metav1.ObjectMeta{Name: "secret1", Namespace: "ns1"}},
					{ObjectMeta: metav1.ObjectMeta{Name: "secret2", Namespace: "ns2"}},
				})

				result, err := handler.list(context.Background())
				Ω(err).ShouldNot(HaveOccurred())
				Ω(result).Should(ConsistOf(
					Secret{Name: "secret1", Namespace: "ns1"},
					Secret{Name: "secret2", Namespace: "ns2"},
				))
			})
		})

		Context("with include namespaces filter", func() {
			BeforeEach(func() {
				cfg.IncludeNamespaces = []string{"ns1", "ns3"}
			})

			It("should return only secrets from included namespaces", func() {
				setupSealedSecretsReactor(fakeSSClient, []ssv1alpha1.SealedSecret{
					{ObjectMeta: metav1.ObjectMeta{Name: "secret1", Namespace: "ns1"}},
					{ObjectMeta: metav1.ObjectMeta{Name: "secret2", Namespace: "ns2"}},
					{ObjectMeta: metav1.ObjectMeta{Name: "secret3", Namespace: "ns3"}},
				})

				result, err := handler.list(context.Background())
				Ω(err).ShouldNot(HaveOccurred())
				Ω(result).Should(ConsistOf(
					Secret{Name: "secret1", Namespace: "ns1"},
					Secret{Name: "secret3", Namespace: "ns3"},
				))
			})
		})

		Context("with exclude namespaces filter", func() {
			BeforeEach(func() {
				cfg.ExcludeNamespaces = []string{"ns2"}
			})

			It("should return secrets from all namespaces except excluded ones", func() {
				// Setup namespaces in fake client
				fakeClient = fake.NewClientset(
					&v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "ns1"}},
					&v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "ns2"}},
					&v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "ns3"}},
				)

				setupSealedSecretsReactor(fakeSSClient, []ssv1alpha1.SealedSecret{
					{ObjectMeta: metav1.ObjectMeta{Name: "secret1", Namespace: "ns1"}},
					{ObjectMeta: metav1.ObjectMeta{Name: "secret2", Namespace: "ns2"}},
					{ObjectMeta: metav1.ObjectMeta{Name: "secret3", Namespace: "ns3"}},
				})

				handler = NewHandler(fakeClient.CoreV1(), fakeSSClient, cfg)
				result, err := handler.list(context.Background())
				Ω(err).ShouldNot(HaveOccurred())
				Ω(result).Should(ConsistOf(
					Secret{Name: "secret1", Namespace: "ns1"},
					Secret{Name: "secret3", Namespace: "ns3"},
				))
			})
		})

		Context("with regex filters", func() {
			BeforeEach(func() {
				cfg.UseRegex = true
				cfg.IncludeNamespaces = []string{"app-.*"}
				cfg.IncludeNamespacesRegex = []*regexp.Regexp{regexp.MustCompile("app-.*")}
			})

			It("should return secrets matching regex pattern", func() {
				// Setup namespaces in fake client
				fakeClient = fake.NewClientset(
					&v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "app-prod"}},
					&v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "kube-system"}},
					&v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "app-staging"}},
				)

				setupSealedSecretsReactor(fakeSSClient, []ssv1alpha1.SealedSecret{
					{ObjectMeta: metav1.ObjectMeta{Name: "secret1", Namespace: "app-prod"}},
					{ObjectMeta: metav1.ObjectMeta{Name: "secret2", Namespace: "kube-system"}},
					{ObjectMeta: metav1.ObjectMeta{Name: "secret3", Namespace: "app-staging"}},
				})

				handler = NewHandler(fakeClient.CoreV1(), fakeSSClient, cfg)
				result, err := handler.list(context.Background())
				Ω(err).ShouldNot(HaveOccurred())
				Ω(result).Should(ConsistOf(
					Secret{Name: "secret1", Namespace: "app-prod"},
					Secret{Name: "secret3", Namespace: "app-staging"},
				))
			})
		})

		Context("sorting", func() {
			It("should sort results by namespace then by name", func() {
				setupSealedSecretsReactor(fakeSSClient, []ssv1alpha1.SealedSecret{
					{ObjectMeta: metav1.ObjectMeta{Name: "z-secret", Namespace: "ns2"}},
					{ObjectMeta: metav1.ObjectMeta{Name: "b-secret", Namespace: "ns1"}},
					{ObjectMeta: metav1.ObjectMeta{Name: "a-secret", Namespace: "ns1"}},
				})

				result, err := handler.list(context.Background())
				Ω(err).ShouldNot(HaveOccurred())
				Ω(result).Should(Equal([]Secret{
					{Name: "a-secret", Namespace: "ns1"},
					{Name: "b-secret", Namespace: "ns1"},
					{Name: "z-secret", Namespace: "ns2"},
				}))
			})
		})

		Context("with synced status", func() {
			It("should include synced status in results", func() {
				setupSealedSecretsReactor(fakeSSClient, []ssv1alpha1.SealedSecret{
					createSealedSecretWithStatus("synced-secret", "ns1", true, ""),
					createSealedSecretWithStatus("failed-secret", "ns1", false, "decryption failed"),
					{ObjectMeta: metav1.ObjectMeta{Name: "no-status", Namespace: "ns1"}},
				})

				result, err := handler.list(context.Background())
				Ω(err).ShouldNot(HaveOccurred())
				Ω(len(result)).Should(Equal(3))

				// Check synced secret
				syncedSecret := findSecret(result, "synced-secret")
				Ω(syncedSecret).ShouldNot(BeNil())
				Ω(syncedSecret.Synced).ShouldNot(BeNil())
				Ω(*syncedSecret.Synced).Should(BeTrue())
				Ω(syncedSecret.Message).Should(Equal(""))

				// Check failed secret
				failedSecret := findSecret(result, "failed-secret")
				Ω(failedSecret).ShouldNot(BeNil())
				Ω(failedSecret.Synced).ShouldNot(BeNil())
				Ω(*failedSecret.Synced).Should(BeFalse())
				Ω(failedSecret.Message).Should(Equal("decryption failed"))

				// Check secret without status
				noStatusSecret := findSecret(result, "no-status")
				Ω(noStatusSecret).ShouldNot(BeNil())
				Ω(noStatusSecret.Synced).Should(BeNil())
			})
		})

		Context("with showOnlySyncedSecrets enabled", func() {
			BeforeEach(func() {
				cfg.ShowOnlySyncedSecrets = true
			})

			It("should only return synced secrets", func() {
				setupSealedSecretsReactor(fakeSSClient, []ssv1alpha1.SealedSecret{
					createSealedSecretWithStatus("synced-secret", "ns1", true, ""),
					createSealedSecretWithStatus("failed-secret", "ns2", false, "decryption failed"),
					{ObjectMeta: metav1.ObjectMeta{Name: "no-status", Namespace: "ns3"}},
				})

				result, err := handler.list(context.Background())
				Ω(err).ShouldNot(HaveOccurred())
				Ω(len(result)).Should(Equal(1))
				Ω(result[0].Name).Should(Equal("synced-secret"))
				Ω(result[0].Namespace).Should(Equal("ns1"))
				Ω(result[0].Synced).ShouldNot(BeNil())
				Ω(*result[0].Synced).Should(BeTrue())
			})
		})
	})
})

func setupSealedSecretsReactor(fakeSSClient *ssfake.FakeBitnamiV1alpha1, sealedSecrets []ssv1alpha1.SealedSecret) {
	fakeSSClient.AddReactor("list", "sealedsecrets", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
		listAction := action.(ktesting.ListAction)
		requestedNamespace := listAction.GetNamespace()

		ssList := &ssv1alpha1.SealedSecretList{}
		for _, ss := range sealedSecrets {
			// If namespace is empty string, return all secrets
			// Otherwise, only return secrets from the requested namespace
			if requestedNamespace == "" || ss.Namespace == requestedNamespace {
				ssList.Items = append(ssList.Items, ss)
			}
		}

		return true, ssList, nil
	})
}

// Helper function to create a SealedSecret with synced status
func createSealedSecretWithStatus(name, namespace string, synced bool, message string) ssv1alpha1.SealedSecret {
	status := v1.ConditionTrue
	if !synced {
		status = v1.ConditionFalse
	}

	return ssv1alpha1.SealedSecret{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Status: &ssv1alpha1.SealedSecretStatus{
			Conditions: []ssv1alpha1.SealedSecretCondition{
				{
					Type:    "Synced",
					Status:  status,
					Message: message,
				},
			},
		},
	}
}

// Helper function to find a secret by name in a list
func findSecret(secrets []Secret, name string) *Secret {
	for _, s := range secrets {
		if s.Name == name {
			return &s
		}
	}
	return nil
}
