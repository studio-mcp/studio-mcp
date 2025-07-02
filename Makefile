.PHONY: help build test version sync-version release dev clean verify-release pre-release-checks

# Default target
help: ## Show this help message
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*##"; printf "\033[36m%-15s\033[0m %s\n", "Target", "Description"} /^[a-zA-Z_-]+:.*?##/ { printf "\033[36m%-15s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

# Get version from git tag or default
VERSION ?= $(shell git describe --tags --abbrev=0 2>/dev/null | sed 's/^v//' || echo "0.0.1")
COMMIT ?= $(shell git rev-parse --short HEAD)
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

build: ## Build the Go binary
	go build -ldflags "-X main.Version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)" -o bin/studio-mcp

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

verify-release: ## Verify the release setup locally
	@echo "🔍 Verifying release setup..."
	@echo ""
	@echo "1. Checking Go build..."
	$(MAKE) build
	@echo "✅ Go build successful"
	@echo ""
	@echo "2. Running tests..."
	$(MAKE) test
	@echo "✅ Tests passed"
	@echo ""
	@echo "3. Testing binary..."
	./bin/studio-mcp --version
	@echo "✅ Binary works"
	@echo ""
	@echo "4. Checking GoReleaser config..."
	goreleaser check || (echo "❌ GoReleaser config invalid. Install with: brew install goreleaser" && exit 1)
	@echo "✅ GoReleaser config valid"
	@echo ""
	@echo "5. Testing NPM postinstall script..."
	@if [ -f scripts/postinstall.js ]; then \
		echo "✅ Postinstall script exists"; \
	else \
		echo "❌ Postinstall script missing"; exit 1; \
	fi
	@echo ""
	@echo "6. Simulating GoReleaser build (dry-run)..."
	goreleaser build --snapshot --clean || (echo "❌ GoReleaser build failed" && exit 1)
	@echo "✅ GoReleaser build simulation successful"
	@echo ""
	@echo "🎉 All verification checks passed!"

pre-release-checks: ## Run pre-release checks
ifndef VERSION
	$(error VERSION is required. Usage: make pre-release-checks VERSION=1.2.3)
endif
	@echo "🔍 Pre-release checks for v$(VERSION)..."
	@echo ""
	@echo "1. Checking if version already exists..."
	@if git tag -l | grep -q "^v$(VERSION)$$"; then \
		echo "❌ Version v$(VERSION) already exists!"; \
		echo "Existing tags:"; git tag -l | tail -5; \
		exit 1; \
	else \
		echo "✅ Version v$(VERSION) is new"; \
	fi
	@echo ""
	@echo "2. Checking working directory is clean..."
	@if [ -n "$$(git status --porcelain)" ]; then \
		echo "❌ Working directory is not clean:"; \
		git status --short; \
		echo "Please commit or stash changes first"; \
		exit 1; \
	else \
		echo "✅ Working directory is clean"; \
	fi
	@echo ""
	@echo "3. Running verification..."
	$(MAKE) verify-release
	@echo ""
	@echo "🎉 All pre-release checks passed for v$(VERSION)!"

release: ## Create a release (usage: make release VERSION=1.2.3)
ifndef VERSION
	$(error VERSION is required. Usage: make release VERSION=1.2.3)
endif
	@echo "🚀 Starting release process for v$(VERSION)..."
	$(MAKE) pre-release-checks VERSION=$(VERSION)
	@echo ""
	@echo "Creating release v$(VERSION)..."
	$(MAKE) bump-version VERSION=$(VERSION)
	@echo ""
	@echo "Pushing tag v$(VERSION)..."
	git push origin v$(VERSION)
	@echo ""
	@echo "🎉 Release v$(VERSION) initiated!"
	@echo "📋 Next steps:"
	@echo "  1. Monitor GitHub Actions: https://github.com/studio-mcp/studio-mcp/actions"
	@echo "  2. Check GitHub Release: https://github.com/studio-mcp/studio-mcp/releases"
	@echo "  3. Verify NPM publish: https://www.npmjs.com/package/studio-mcp"
	@echo "  4. Test install: npm install studio-mcp@$(VERSION)"

dev: ## Run in development mode
	go run -ldflags "-X main.Version=dev -X main.commit=$(COMMIT) -X main.date=$(DATE)" . $(ARGS)

clean: ## Clean build artifacts
	rm -rf bin/studio-mcp dist/ studio-mcp
