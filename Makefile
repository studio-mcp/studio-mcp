.PHONY: help build test version sync-version release dev clean

# Default target
help: ## Show this help message
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*##"; printf "\033[36m%-15s\033[0m %s\n", "Target", "Description"} /^[a-zA-Z_-]+:.*?##/ { printf "\033[36m%-15s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

# Get version from git tag or default
VERSION ?= $(shell git describe --tags --abbrev=0 2>/dev/null | sed 's/^v//' || echo "0.0.1")
COMMIT ?= $(shell git rev-parse --short HEAD)
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

build: ## Build the Go binary
	go build -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)" -o bin/studio-mcp

test: ## Run tests
	go test ./...

version: ## Show current version information
	@echo "Version: $(VERSION)"
	@echo "Commit:  $(COMMIT)"
	@echo "Date:    $(DATE)"
	@echo "Package.json version: $(shell grep '"version"' package.json | cut -d'"' -f4)"
	@echo ""
	@echo "Binary version:"
	@if [ -f bin/studio-mcp ]; then ./bin/studio-mcp --version; else echo "Binary not built. Run 'make build' first."; fi

sync-version: ## Sync package.json version with git tags
	./scripts/sync-version.sh

bump-version: ## Bump version (usage: make bump-version VERSION=1.2.3)
ifndef VERSION
	$(error VERSION is required. Usage: make bump-version VERSION=1.2.3)
endif
	./scripts/sync-version.sh $(VERSION)

release: ## Create a release (usage: make release VERSION=1.2.3)
ifndef VERSION
	$(error VERSION is required. Usage: make release VERSION=1.2.3)
endif
	$(MAKE) bump-version VERSION=$(VERSION)
	git push origin v$(VERSION)
	@echo "Release v$(VERSION) pushed. GitHub Actions will handle the rest."

dev: ## Run in development mode
	go run -ldflags "-X main.version=dev -X main.commit=$(COMMIT) -X main.date=$(DATE)" . $(ARGS)

clean: ## Clean build artifacts
	rm -rf bin/ dist/
