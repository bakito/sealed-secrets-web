release: semver
	@version=$$(semver); \
	git tag -s $$version -m"Release $$version"
	goreleaser --rm-dist
	cr upload --skip-existing

test-release:
	goreleaser --skip-publish --snapshot --rm-dist

semver:
ifeq (, $(shell which semver))
 $(shell go get -u github.com/bakito/semver)
endif
