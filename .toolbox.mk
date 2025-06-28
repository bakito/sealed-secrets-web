## toolbox - start
## Generated with https://github.com/bakito/toolbox

## Current working directory
TB_LOCALDIR ?= $(shell which cygpath > /dev/null 2>&1 && cygpath -m $$(pwd) || pwd)
## Location to install dependencies to
TB_LOCALBIN ?= $(TB_LOCALDIR)/bin
$(TB_LOCALBIN):
	if [ ! -e $(TB_LOCALBIN) ]; then mkdir -p $(TB_LOCALBIN); fi

## Tool Binaries
TB_GINKGO ?= $(TB_LOCALBIN)/ginkgo
TB_GOFUMPT ?= $(TB_LOCALBIN)/gofumpt
TB_GOLANGCI_LINT ?= $(TB_LOCALBIN)/golangci-lint
TB_GOLINES ?= $(TB_LOCALBIN)/golines
TB_GORELEASER ?= $(TB_LOCALBIN)/goreleaser
TB_HELM_DOCS ?= $(TB_LOCALBIN)/helm-docs
TB_MOCKGEN ?= $(TB_LOCALBIN)/mockgen
TB_SEMVER ?= $(TB_LOCALBIN)/semver

## Tool Versions
TB_GINKGO_VERSION ?= v2.23.4
TB_GOFUMPT_VERSION ?= v0.8.0
TB_GOLANGCI_LINT_VERSION ?= v2.1.6
TB_GOLINES_VERSION ?= v0.12.2
TB_GORELEASER_VERSION ?= v2.10.2
TB_HELM_DOCS_VERSION ?= v1.14.2
TB_MOCKGEN_VERSION ?= v0.5.2
TB_SEMVER_VERSION ?= v1.1.3

## Tool Installer
.PHONY: tb.ginkgo
tb.ginkgo: $(TB_GINKGO) ## Download ginkgo locally if necessary.
$(TB_GINKGO): $(TB_LOCALBIN)
	test -s $(TB_LOCALBIN)/ginkgo || GOBIN=$(TB_LOCALBIN) go install github.com/onsi/ginkgo/v2/ginkgo@$(TB_GINKGO_VERSION)
.PHONY: tb.gofumpt
tb.gofumpt: $(TB_GOFUMPT) ## Download gofumpt locally if necessary.
$(TB_GOFUMPT): $(TB_LOCALBIN)
	test -s $(TB_LOCALBIN)/gofumpt || GOBIN=$(TB_LOCALBIN) go install mvdan.cc/gofumpt@$(TB_GOFUMPT_VERSION)
.PHONY: tb.golangci-lint
tb.golangci-lint: $(TB_GOLANGCI_LINT) ## Download golangci-lint locally if necessary.
$(TB_GOLANGCI_LINT): $(TB_LOCALBIN)
	test -s $(TB_LOCALBIN)/golangci-lint || GOBIN=$(TB_LOCALBIN) go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(TB_GOLANGCI_LINT_VERSION)
.PHONY: tb.golines
tb.golines: $(TB_GOLINES) ## Download golines locally if necessary.
$(TB_GOLINES): $(TB_LOCALBIN)
	test -s $(TB_LOCALBIN)/golines || GOBIN=$(TB_LOCALBIN) go install github.com/segmentio/golines@$(TB_GOLINES_VERSION)
.PHONY: tb.goreleaser
tb.goreleaser: $(TB_GORELEASER) ## Download goreleaser locally if necessary.
$(TB_GORELEASER): $(TB_LOCALBIN)
	test -s $(TB_LOCALBIN)/goreleaser || GOBIN=$(TB_LOCALBIN) go install github.com/goreleaser/goreleaser/v2@$(TB_GORELEASER_VERSION)
.PHONY: tb.helm-docs
tb.helm-docs: $(TB_HELM_DOCS) ## Download helm-docs locally if necessary.
$(TB_HELM_DOCS): $(TB_LOCALBIN)
	test -s $(TB_LOCALBIN)/helm-docs || GOBIN=$(TB_LOCALBIN) go install github.com/norwoodj/helm-docs/cmd/helm-docs@$(TB_HELM_DOCS_VERSION)
.PHONY: tb.mockgen
tb.mockgen: $(TB_MOCKGEN) ## Download mockgen locally if necessary.
$(TB_MOCKGEN): $(TB_LOCALBIN)
	test -s $(TB_LOCALBIN)/mockgen || GOBIN=$(TB_LOCALBIN) go install go.uber.org/mock/mockgen@$(TB_MOCKGEN_VERSION)
.PHONY: tb.semver
tb.semver: $(TB_SEMVER) ## Download semver locally if necessary.
$(TB_SEMVER): $(TB_LOCALBIN)
	test -s $(TB_LOCALBIN)/semver || GOBIN=$(TB_LOCALBIN) go install github.com/bakito/semver@$(TB_SEMVER_VERSION)

## Reset Tools
.PHONY: tb.reset
tb.reset:
	@rm -f \
		$(TB_LOCALBIN)/ginkgo \
		$(TB_LOCALBIN)/gofumpt \
		$(TB_LOCALBIN)/golangci-lint \
		$(TB_LOCALBIN)/golines \
		$(TB_LOCALBIN)/goreleaser \
		$(TB_LOCALBIN)/helm-docs \
		$(TB_LOCALBIN)/mockgen \
		$(TB_LOCALBIN)/semver

## Update Tools
.PHONY: tb.update
tb.update: tb.reset
	toolbox makefile -f $(TB_LOCALDIR)/Makefile \
		github.com/onsi/ginkgo/v2/ginkgo \
		mvdan.cc/gofumpt@github.com/mvdan/gofumpt \
		github.com/golangci/golangci-lint/v2/cmd/golangci-lint \
		github.com/segmentio/golines \
		github.com/goreleaser/goreleaser/v2 \
		github.com/norwoodj/helm-docs/cmd/helm-docs \
		go.uber.org/mock/mockgen@github.com/uber-go/mock \
		github.com/bakito/semver
## toolbox - end