# Include toolbox tasks
include ./.toolbox.mk

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
	$(GINKGO) --cover ./...

release: goreleaser semver
	@version=$$($(SEMVER)); \
	git tag -s $$version -m"Release $$version"
	$(GORELEASER) --clean

test-release: goreleaser
	$(GORELEASER) --skip=publish --snapshot --clean

mocks: mockgen
	$(MOCKGEN) -destination pkg/mocks/core/mock.go     --package core     k8s.io/client-go/kubernetes/typed/core/v1 CoreV1Interface,SecretInterface
	$(MOCKGEN) -destination pkg/mocks/ssclient/mock.go --package ssclient github.com/bitnami-labs/sealed-secrets/pkg/client/clientset/versioned/typed/sealedsecrets/v1alpha1 BitnamiV1alpha1Interface,SealedSecretInterface
	$(MOCKGEN) -destination pkg/mocks/seal/mock.go --package seal github.com/bakito/sealed-secrets-web/pkg/seal Sealer

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