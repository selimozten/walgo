.PHONY: help build install test clean fmt lint run dev doctor release

# Variables
BINARY_NAME=walgo
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(BUILD_DATE)"

help: ## Show this help message
	@echo "Walgo - Development Makefile"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	@go build $(LDFLAGS) -o $(BINARY_NAME) main.go
	@echo "✓ Built $(BINARY_NAME) ($(VERSION))"

install: build ## Build and install to /usr/local/bin
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	@sudo mv $(BINARY_NAME) /usr/local/bin/
	@echo "✓ Installed $(BINARY_NAME)"

install-user: build ## Build and install to ~/.local/bin
	@echo "Installing $(BINARY_NAME) to ~/.local/bin..."
	@mkdir -p ~/.local/bin
	@mv $(BINARY_NAME) ~/.local/bin/
	@echo "✓ Installed $(BINARY_NAME) to ~/.local/bin"
	@echo "  Make sure ~/.local/bin is in your PATH"

test: ## Run tests
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...
	@echo "✓ Tests passed"

test-coverage: test ## Run tests and show coverage
	@go tool cover -html=coverage.out

test-short: ## Run tests without race detector (faster)
	@go test -short ./...

clean: ## Remove build artifacts
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@rm -f coverage.out coverage.txt
	@rm -rf dist/
	@echo "✓ Cleaned"

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...
	@echo "✓ Formatted"

lint: ## Run linters
	@echo "Running linters..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
		echo "✓ Linting passed"; \
	else \
		echo "⚠ golangci-lint not installed, running basic checks..."; \
		go vet ./...; \
		echo "✓ Basic checks passed"; \
	fi

vet: ## Run go vet
	@go vet ./...

tidy: ## Tidy dependencies
	@echo "Tidying dependencies..."
	@go mod tidy
	@go mod verify
	@echo "✓ Dependencies tidied"

run: build ## Build and run walgo
	@./$(BINARY_NAME)

dev: ## Build and run doctor command
	@go run main.go doctor

doctor: build ## Build and run full diagnostics
	@./$(BINARY_NAME) doctor --verbose

# Release targets
release-dry: clean ## Dry run of release process
	@echo "Running release dry-run..."
	@goreleaser release --snapshot --clean --skip=publish
	@echo "✓ Release dry-run complete"

release-snapshot: clean ## Build snapshot release locally
	@goreleaser release --snapshot --clean
	@echo "✓ Snapshot release created in dist/"

# Cross-compilation targets
build-all: clean ## Build for all platforms
	@echo "Building for all platforms..."
	@mkdir -p dist
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 main.go
	@GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 main.go
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 main.go
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 main.go
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe main.go
	@echo "✓ Built binaries in dist/"
	@ls -lh dist/

# Docker targets
docker-build: ## Build Docker image
	@docker build -t $(BINARY_NAME):$(VERSION) .
	@docker tag $(BINARY_NAME):$(VERSION) $(BINARY_NAME):latest
	@echo "✓ Built Docker image $(BINARY_NAME):$(VERSION)"

docker-run: docker-build ## Build and run Docker image
	@docker run --rm $(BINARY_NAME):latest

# Development helpers
watch: ## Watch for changes and rebuild (requires entr)
	@if command -v entr >/dev/null 2>&1; then \
		find . -name '*.go' | entr -c make build; \
	else \
		echo "⚠ entr not installed. Install with: brew install entr (macOS)"; \
	fi

deps-install: ## Install development dependencies
	@echo "Installing development dependencies..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/goreleaser/goreleaser@latest
	@echo "✓ Installed development dependencies"

check: fmt vet test ## Run formatting, vetting, and tests

all: clean fmt vet test build ## Clean, format, test, and build

# Quick development workflow
quick: ## Quick build and test
	@make -s build && make -s test-short

# Version info
version: ## Show version information
	@echo "Version: $(VERSION)"
	@echo "Commit:  $(COMMIT)"
	@echo "Date:    $(BUILD_DATE)"
