# Include toolbox tasks
include ./.toolbox.mk

lint: tb.golangci-lint
	$(TB_GOLANGCI_LINT) run --fix

# Run go mod tidy
tidy:
	go mod tidy

fmt: tb.golines tb.gofumpt
	$(TB_GOLINES) --base-formatter="$(TB_GOFUMPT)" --max-len=120 --write-output .

# Run tests
test: mocks tidy fmt helm-lint test-cover
# Run coverage tests
test-cover: tb.ginkgo
	$(TB_GINKGO) --cover ./...

release: tb.goreleaser tb.semver
	@version=$$($(TB_SEMVER)); \
	git tag -s $$version -m"Release $$version"
	$(TB_GORELEASER) --clean

test-release: tb.goreleaser
	$(TB_GORELEASER) --skip=publish --snapshot --clean

mocks: tb.mockgen
	$(TB_MOCKGEN) -destination pkg/mocks/core/mock.go     --package core     k8s.io/client-go/kubernetes/typed/core/v1 CoreV1Interface,SecretInterface
	$(TB_MOCKGEN) -destination pkg/mocks/ssclient/mock.go --package ssclient github.com/bitnami-labs/sealed-secrets/pkg/client/clientset/versioned/typed/sealedsecrets/v1alpha1 BitnamiV1alpha1Interface,SealedSecretInterface
	$(TB_MOCKGEN) -destination pkg/mocks/seal/mock.go --package seal github.com/bakito/sealed-secrets-web/pkg/seal Sealer

build:
	podman build --build-arg VERSION=dev --build-arg BUILD=dev --build-arg TARGETPLATFORM=linux/amd64 -t sealed-secrets-web .

build-arm:
	podman build --build-arg VERSION=dev --build-arg BUILD=dev --build-arg TARGETPLATFORM=linux/arm64 -t sealed-secrets-web .

helm-docs: tb.helm-docs
	@$(TB_HELM_DOCS)

# Detect OS
OS := $(shell uname)

# Define the sed command based on OS
SED := $(if $(filter Darwin, $(OS)), sed -i "", sed -i)

update-chart-version: tb.semver
	@version=$$($(TB_SEMVER) -next); \
	versionNum=$$($(TB_SEMVER) -next -numeric); \
	$(SED) "s/^version:.*$$/version: $${versionNum}/"    ./chart/Chart.yaml; \
	$(SED) "s/^appVersion:.*$$/appVersion: $${version}/" ./chart/Chart.yaml

helm-lint: helm-docs
	helm lint ./chart

helm-template:
	helm template ./chart -n sealed-secrets-web
	@echo "#######################"
	helm template ./chart -n sealed-secrets-web --set disableLoadSecrets=true --set sealedSecrets.serviceName=

helm-chainsaw:
	./testdata/e2e/chainsaw/run.sh
