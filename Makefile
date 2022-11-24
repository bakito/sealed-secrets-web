# Run go fmt against code
fmt: golangci-lint
	$(LOCALBIN)/golangci-lint run --fix

# Run go mod tidy
tidy:
	go mod tidy

# Run tests
test: mocks tidy fmt helm-lint test-cover
# Run coverage tests
test-cover: ginkgo
	$(LOCALBIN)/ginkgo --cover ./...

release: semver
	@version=$$($(LOCALBIN)/semver); \
	git tag -s $$version -m"Release $$version"
	goreleaser --rm-dist

test-release:
	goreleaser --skip-publish --snapshot --rm-dist

mocks: mockgen
	$(LOCALBIN)/mockgen -destination pkg/mocks/core/mock.go     --package core     k8s.io/client-go/kubernetes/typed/core/v1 CoreV1Interface,SecretInterface
	$(LOCALBIN)/mockgen -destination pkg/mocks/ssclient/mock.go --package ssclient github.com/bitnami-labs/sealed-secrets/pkg/client/clientset/versioned/typed/sealedsecrets/v1alpha1 BitnamiV1alpha1Interface,SealedSecretInterface
	$(LOCALBIN)/mockgen -destination pkg/mocks/seal/mock.go --package seal github.com/bakito/sealed-secrets-web/pkg/seal Sealer

## toolbox - start
## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
SEMVER ?= $(LOCALBIN)/semver
HELM_DOCS ?= $(LOCALBIN)/helm-docs
MOCKGEN ?= $(LOCALBIN)/mockgen
GINKGO ?= $(LOCALBIN)/ginkgo
GOLANGCI_LINT ?= $(LOCALBIN)/golangci-lint

## Tool Versions
SEMVER_VERSION ?= v1.1.3
HELM_DOCS_VERSION ?= v1.11.0
MOCKGEN_VERSION ?= v1.6.0
GINKGO_VERSION ?= v2.5.1
GOLANGCI_LINT_VERSION ?= v1.50.1

## Tool Installer
.PHONY: semver
semver: $(SEMVER) ## Download semver locally if necessary.
$(SEMVER): $(LOCALBIN)
	test -s $(LOCALBIN)/semver || GOBIN=$(LOCALBIN) go install github.com/bakito/semver@$(SEMVER_VERSION)
.PHONY: helm-docs
helm-docs: $(HELM_DOCS) ## Download helm-docs locally if necessary.
$(HELM_DOCS): $(LOCALBIN)
	test -s $(LOCALBIN)/helm-docs || GOBIN=$(LOCALBIN) go install github.com/norwoodj/helm-docs/cmd/helm-docs@$(HELM_DOCS_VERSION)
.PHONY: mockgen
mockgen: $(MOCKGEN) ## Download mockgen locally if necessary.
$(MOCKGEN): $(LOCALBIN)
	test -s $(LOCALBIN)/mockgen || GOBIN=$(LOCALBIN) go install github.com/golang/mock/mockgen@$(MOCKGEN_VERSION)
.PHONY: ginkgo
ginkgo: $(GINKGO) ## Download ginkgo locally if necessary.
$(GINKGO): $(LOCALBIN)
	test -s $(LOCALBIN)/ginkgo || GOBIN=$(LOCALBIN) go install github.com/onsi/ginkgo/v2/ginkgo@$(GINKGO_VERSION)
.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT) ## Download golangci-lint locally if necessary.
$(GOLANGCI_LINT): $(LOCALBIN)
	test -s $(LOCALBIN)/golangci-lint || GOBIN=$(LOCALBIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

## Update Tools
.PHONY: update-toolbox-tools
update-toolbox-tools:
	toolbox makefile -f $$(pwd)/Makefile \
		github.com/bakito/semver \
		github.com/norwoodj/helm-docs/cmd/helm-docs \
		github.com/golang/mock/mockgen \
		github.com/onsi/ginkgo/v2/ginkgo \
		github.com/golangci/golangci-lint/cmd/golangci-lint
## toolbox - end

build:
	podman build --build-arg VERSION=dev --build-arg BUILD=dev --build-arg TARGETPLATFORM=linux/amd64 -t sealed-secrets-web .

build-arm:
	podman build --build-arg VERSION=dev --build-arg BUILD=dev --build-arg TARGETPLATFORM=linux/arm64 -t sealed-secrets-web .

docs: helm-docs update-chart-version
	@$(LOCALBIN)/helm-docs

update-chart-version: semver
	@version=$$($(LOCALBIN)/semver -next); \
	versionNum=$$($(LOCALBIN)/semver -next -numeric); \
	sed -i "s/^version:.*$$/version: $${versionNum}/"    ./chart/Chart.yaml; \
	sed -i "s/^appVersion:.*$$/appVersion: $${version}/" ./chart/Chart.yaml

helm-lint: docs
	helm lint ./chart

helm-template:
	helm template ./chart -n sealed-secrets-web
	@echo "#######################"
	helm template ./chart -n sealed-secrets-web --set disableLoadSecrets=true --set sealedSecrets.serviceName=
