SHELL := /bin/bash

MODULE := github.com/blairham/tuikit

.DEFAULT_GOAL := help

.PHONY: help
help: ## Display this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_\/-]+:.*?## / {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: build
build: ## Build all packages
	go build ./...

.PHONY: test
test: ## Run tests
	go test -race ./...

.PHONY: vet
vet: ## Run go vet
	go vet ./...

# No golangci-lint here, by design. Linting happens in the pre-commit hook at
# commit time and in CI as the merge gate — never from a Make target. A
# hand-started run tells you nothing the hook would not tell you moments later,
# and golangci-lint defaults its concurrency to NumCPU at gigabytes of RSS.
.PHONY: fmt
fmt: ## Format code (gofumpt, fieldalignment)
	@echo "Formatting code..."
	go tool gofumpt -l -w .
	go tool fieldalignment -fix ./...

.PHONY: check
check: build vet test ## Run all checks (build, vet, test) — this is what CI runs

.PHONY: sync
sync: ## Sync .tool-versions Go version from go.mod
	@GO_VERSION=$$(grep '^go ' go.mod | awk '{print $$2}'); \
	if [ -z "$$GO_VERSION" ]; then echo "ERROR: could not parse Go version from go.mod"; exit 1; fi; \
	sed -i '' "s/^golang .*/golang $$GO_VERSION/" .tool-versions; \
	echo "Synced .tool-versions: golang $$GO_VERSION"

.PHONY: tidy
tidy: ## go mod tidy
	go mod tidy

.PHONY: clean
clean: ## Remove build and test caches for this module
	go clean -cache -testcache
