# Run go fmt against code
fmt:
	go fmt ./...
	gofmt -s -w .

# Run go vet against code
vet:
	go vet ./...

# Run golangci-lint
lint:
	golangci-lint run

# Run go mod tidy
tidy:
	go mod tidy

# Run tests
test: tidy fmt vet
	go test ./...  -coverprofile=coverage.out
	go tool cover -func=coverage.out

release: semver
	@version=$$(semver); \
	git tag -s $$version -m"Release $$version"
	goreleaser --rm-dist
	cr upload --skip-existing
	cr index

test-release:
	goreleaser --skip-publish --snapshot --rm-dist

semver:
ifeq (, $(shell which semver))
 $(shell go get -u github.com/bakito/semver)
endif
