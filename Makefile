.DEFAULT_GOAL := dev

export PATH := $(pwd)tools/bin:$(PATH)
OS := $(shell uname -s)
ARCH = $(shell uname -m)

.PHONY: dev
dev: ## dev build
dev: clean install generate vet fmt test cli-test mod-tidy

.PHONY: ci
ci: ## CI build
ci: dev diff

.PHONY: clean
clean: ## remove files created during build pipeline
	$(call print-target)
	rm -rf dist
	rm -f coverage.*

.PHONY: install
install: tools/bin/golangci-lint tools/bin/goreleaser tools/bin/crane
install: ## go install tools

tools/bin/golangci-lint:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b tools/bin v1.48.0

tools/bin/goreleaser:
	curl -sfL https://github.com/goreleaser/goreleaser/releases/download/v1.10.3/goreleaser_${OS}_${ARCH}.tar.gz | tar xz -C tools/bin goreleaser

tools/bin/crane:
	curl -sfL https://github.com/google/go-containerregistry/releases/download/v0.11.0/go-containerregistry_${OS}_${ARCH}.tar.gz | tar xz -C tools/bin crane

.PHONY: generate
generate: ## go generate
	$(call print-target)
	go generate ./...

.PHONY: vet
vet: ## go vet
	$(call print-target)
	go vet ./...

.PHONY: fmt
fmt: ## format source code
	$(call print-target)
	go fmt ./...
	golangci-lint run --fix || true

.PHONY: lint
lint: ## golangci-lint
	$(call print-target)
	golangci-lint run

.PHONY: test
test: ## go test with race detector and code covarage
	$(call print-target)
	go test -race -covermode=atomic -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

.PHONY: cli-test
cli-test: build
cli-test: ## go test with race detector and code covarage
	$(call print-target)
	./tests/structure_test_tests.sh

.PHONY: mod-tidy
mod-tidy: ## go mod tidy
	$(call print-target)
	go mod tidy

.PHONY: diff
diff: ## git diff
	$(call print-target)
	git diff --exit-code
	RES=$$(git status --porcelain) ; if [ -n "$$RES" ]; then echo $$RES && exit 1 ; fi

.PHONY: build
build: ## goreleaser --snapshot --skip-publish --rm-dist
build: install
	$(call print-target)
	goreleaser --snapshot --skip-publish --rm-dist

.PHONY: release
release: ## goreleaser --rm-dist
release: install
	$(call print-target)
	goreleaser --rm-dist

.PHONY: run
run: ## go run
	@go run -race .

.PHONY: go-clean
go-clean: ## go clean build, test and modules caches
	$(call print-target)
	go clean -r -i -cache -testcache -modcache

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

define print-target
    @printf "Executing target: \033[36m$@\033[0m\n"
endef
