# Include toolbox tasks
include ./.toolbox.mk

# Run go fmt against code
fmt: tb.golangci-lint
	$(TB_GOLANG_CI_LINT) run --fix

# Run go mod tidy
tidy:
	go mod tidy

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

docs: tb.helm-docs update-chart-version
	@$(TB_HELM_DOCSLOCALBIN)

update-chart-version: .semver
	@version=$$($(TB_SEMVER) -next); \
	versionNum=$$($(TB_SEMVER) -next -numeric); \
	sed -i "s/^version:.*$$/version: $${versionNum}/"    ./chart/Chart.yaml; \
	sed -i "s/^appVersion:.*$$/appVersion: $${version}/" ./chart/Chart.yaml

helm-lint: docs
	helm lint ./chart

helm-template:
	helm template ./chart -n sealed-secrets-web
	@echo "#######################"
	helm template ./chart -n sealed-secrets-web --set disableLoadSecrets=true --set sealedSecrets.serviceName=
