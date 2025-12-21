# Shadowforge Makefile
# Quantum-resistant steganography tool with DDD/CQRS architecture

# ============================================================================
# Configuration
# ============================================================================

# Project metadata
PROJECT_NAME := shadowforge
MODULE_NAME := github.com/greysquirr3l/shadowforge
VERSION_FILE := VERSION
VERSION := $(shell cat $(VERSION_FILE) 2>/dev/null || echo "0.0.0-dev")

# Build targets
CLI_BINARY := bin/shadowforge
API_BINARY := bin/shadowforge-api
CLI_CMD := ./cmd/cli
API_CMD := ./cmd/api

# Build configuration
GO := go
GOFMT := gofmt
GOVET := $(GO) vet
GOLINT := golangci-lint
GOTEST := $(GO) test
GOBUILD := $(GO) build

# Version package for ldflags injection
VERSION_PKG := $(MODULE_NAME)/pkg/version

# Git information
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GIT_TAG := $(shell git describe --tags --exact-match 2>/dev/null || echo "untagged")
GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
GIT_DIRTY := $(shell git diff --quiet 2>/dev/null || echo "true")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S_UTC')
BUILD_USER := $(shell whoami)
BUILD_HOST := $(shell hostname)
GO_VERSION := $(shell $(GO) version | awk '{print $$3}')
CGO_ENABLED := $(shell go env CGO_ENABLED)

# Build flags
LDFLAGS := -s -w \
	-X '$(VERSION_PKG).GitCommit=$(GIT_COMMIT)' \
	-X '$(VERSION_PKG).GitTag=$(GIT_TAG)' \
	-X '$(VERSION_PKG).GitBranch=$(GIT_BRANCH)' \
	-X '$(VERSION_PKG).GitDirty=$(GIT_DIRTY)' \
	-X '$(VERSION_PKG).BuildTime=$(BUILD_TIME)' \
	-X '$(VERSION_PKG).BuildUser=$(BUILD_USER)' \
	-X '$(VERSION_PKG).BuildHost=$(BUILD_HOST)' \
	-X '$(VERSION_PKG).CGOEnabled=$(CGO_ENABLED)'

# Build tags
BUILD_TAGS :=

# Go build flags
GOFLAGS := -trimpath
RACE_FLAGS := -race

# Test configuration
TEST_FLAGS := -v -race -cover -timeout 10m
TEST_COVERAGE_FILE := coverage.out
TEST_COVERAGE_HTML := coverage.html

# Docker configuration
DOCKER_IMAGE := $(PROJECT_NAME)
DOCKER_TAG := $(VERSION)

# Colors for output
COLOR_RESET := \033[0m
COLOR_BOLD := \033[1m
COLOR_GREEN := \033[32m
COLOR_YELLOW := \033[33m
COLOR_BLUE := \033[34m
COLOR_CYAN := \033[36m

# ============================================================================
# Default Target
# ============================================================================

.DEFAULT_GOAL := help

.PHONY: help
help: ## Display this help message
	@echo "$(COLOR_BOLD)Shadowforge - Quantum-Resistant Steganography$(COLOR_RESET)"
	@echo "$(COLOR_CYAN)Version: $(VERSION)$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BOLD)Usage:$(COLOR_RESET)"
	@echo "  make $(COLOR_GREEN)<target>$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BOLD)Build Targets:$(COLOR_RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  $(COLOR_GREEN)%-20s$(COLOR_RESET) %s\n", $$1, $$2}' | \
		grep -E "(build|install|clean|docker)"
	@echo ""
	@echo "$(COLOR_BOLD)Development Targets:$(COLOR_RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  $(COLOR_GREEN)%-20s$(COLOR_RESET) %s\n", $$1, $$2}' | \
		grep -E "(fmt|vet|lint|test|coverage)"
	@echo ""
	@echo "$(COLOR_BOLD)Utility Targets:$(COLOR_RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  $(COLOR_GREEN)%-20s$(COLOR_RESET) %s\n", $$1, $$2}' | \
		grep -E "(version|deps|tidy|generate)"

# ============================================================================
# Build Targets
# ============================================================================

.PHONY: all
all: clean deps test build ## Run all checks and build all binaries

.PHONY: build
build: build-cli ## Build all binaries (alias for build-cli)

.PHONY: build-cli
build-cli: ## Build CLI binary (shadowforge)
	@echo "$(COLOR_BLUE)Building CLI binary...$(COLOR_RESET)"
	@mkdir -p bin
	@$(GOBUILD) $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(CLI_BINARY) $(CLI_CMD)
	@echo "$(COLOR_GREEN)✓ CLI binary built: $(CLI_BINARY)$(COLOR_RESET)"
	@$(CLI_BINARY) version

.PHONY: build-api
build-api: ## Build API server binary (shadowforge-api)
	@echo "$(COLOR_BLUE)Building API server binary...$(COLOR_RESET)"
	@mkdir -p bin
	@$(GOBUILD) $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(API_BINARY) $(API_CMD)
	@echo "$(COLOR_GREEN)✓ API binary built: $(API_BINARY)$(COLOR_RESET)"

.PHONY: build-all
build-all: build-cli build-api ## Build all binaries (CLI + API)

.PHONY: build-debug
build-debug: ## Build CLI with debug symbols (no -s -w flags)
	@echo "$(COLOR_BLUE)Building CLI binary with debug symbols...$(COLOR_RESET)"
	@mkdir -p bin
	@$(GOBUILD) $(GOFLAGS) -gcflags="all=-N -l" -o $(CLI_BINARY) $(CLI_CMD)
	@echo "$(COLOR_GREEN)✓ Debug CLI binary built: $(CLI_BINARY)$(COLOR_RESET)"

.PHONY: build-race
build-race: ## Build CLI with race detector enabled
	@echo "$(COLOR_BLUE)Building CLI binary with race detector...$(COLOR_RESET)"
	@mkdir -p bin
	@$(GOBUILD) $(GOFLAGS) $(RACE_FLAGS) -ldflags "$(LDFLAGS)" -o $(CLI_BINARY) $(CLI_CMD)
	@echo "$(COLOR_GREEN)✓ Race-enabled CLI binary built: $(CLI_BINARY)$(COLOR_RESET)"

.PHONY: install
install: build-cli ## Install CLI binary to $GOPATH/bin
	@echo "$(COLOR_BLUE)Installing CLI binary...$(COLOR_RESET)"
	@$(GO) install -ldflags "$(LDFLAGS)" $(CLI_CMD)
	@echo "$(COLOR_GREEN)✓ CLI binary installed to $(shell go env GOPATH)/bin/$(PROJECT_NAME)$(COLOR_RESET)"

# ============================================================================
# Development Targets
# ============================================================================

.PHONY: fmt
fmt: ## Format Go source code
	@echo "$(COLOR_BLUE)Formatting code...$(COLOR_RESET)"
	@$(GOFMT) -w -s .
	@echo "$(COLOR_GREEN)✓ Code formatted$(COLOR_RESET)"

.PHONY: vet
vet: ## Run go vet on all packages
	@echo "$(COLOR_BLUE)Running go vet...$(COLOR_RESET)"
	@$(GOVET) ./...
	@echo "$(COLOR_GREEN)✓ go vet passed$(COLOR_RESET)"

.PHONY: lint
lint: ## Run golangci-lint
	@echo "$(COLOR_BLUE)Running golangci-lint...$(COLOR_RESET)"
	@if command -v $(GOLINT) >/dev/null 2>&1; then \
		$(GOLINT) run; \
		echo "$(COLOR_GREEN)✓ golangci-lint passed$(COLOR_RESET)"; \
	else \
		echo "$(COLOR_YELLOW)⚠ golangci-lint not installed, skipping...$(COLOR_RESET)"; \
		echo "$(COLOR_YELLOW)  Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest$(COLOR_RESET)"; \
	fi

.PHONY: test
test: ## Run all tests with race detector
	@echo "$(COLOR_BLUE)Running tests...$(COLOR_RESET)"
	@$(GOTEST) $(TEST_FLAGS) ./...
	@echo "$(COLOR_GREEN)✓ All tests passed$(COLOR_RESET)"

.PHONY: test-short
test-short: ## Run short tests only
	@echo "$(COLOR_BLUE)Running short tests...$(COLOR_RESET)"
	@$(GOTEST) -v -short -race ./...
	@echo "$(COLOR_GREEN)✓ Short tests passed$(COLOR_RESET)"

.PHONY: test-integration
test-integration: ## Run integration tests only
	@echo "$(COLOR_BLUE)Running integration tests...$(COLOR_RESET)"
	@$(GOTEST) -v -race -run Integration ./tests/integration/...
	@echo "$(COLOR_GREEN)✓ Integration tests passed$(COLOR_RESET)"

.PHONY: test-unit
test-unit: ## Run unit tests only (exclude integration)
	@echo "$(COLOR_BLUE)Running unit tests...$(COLOR_RESET)"
	@$(GOTEST) -v -short -race ./internal/...
	@echo "$(COLOR_GREEN)✓ Unit tests passed$(COLOR_RESET)"

.PHONY: coverage
coverage: ## Generate test coverage report
	@echo "$(COLOR_BLUE)Generating coverage report...$(COLOR_RESET)"
	@$(GOTEST) -race -coverprofile=$(TEST_COVERAGE_FILE) -covermode=atomic ./...
	@$(GO) tool cover -html=$(TEST_COVERAGE_FILE) -o $(TEST_COVERAGE_HTML)
	@echo "$(COLOR_GREEN)✓ Coverage report generated$(COLOR_RESET)"
	@echo "  - Text report: $(TEST_COVERAGE_FILE)"
	@echo "  - HTML report: $(TEST_COVERAGE_HTML)"
	@$(GO) tool cover -func=$(TEST_COVERAGE_FILE) | tail -1

.PHONY: bench
bench: ## Run benchmarks
	@echo "$(COLOR_BLUE)Running benchmarks...$(COLOR_RESET)"
	@$(GO) test -bench=. -benchmem ./...

# ============================================================================
# Code Generation Targets
# ============================================================================

.PHONY: generate
generate: ## Run go generate
	@echo "$(COLOR_BLUE)Running go generate...$(COLOR_RESET)"
	@$(GO) generate ./...
	@echo "$(COLOR_GREEN)✓ Code generation complete$(COLOR_RESET)"

.PHONY: mocks
mocks: ## Generate mocks using mockgen
	@echo "$(COLOR_BLUE)Generating mocks...$(COLOR_RESET)"
	@if command -v mockgen >/dev/null 2>&1; then \
		$(GO) generate ./...; \
		echo "$(COLOR_GREEN)✓ Mocks generated$(COLOR_RESET)"; \
	else \
		echo "$(COLOR_YELLOW)⚠ mockgen not installed$(COLOR_RESET)"; \
		echo "$(COLOR_YELLOW)  Install with: go install go.uber.org/mock/mockgen@latest$(COLOR_RESET)"; \
	fi

# ============================================================================
# Dependency Management
# ============================================================================

.PHONY: deps
deps: ## Download dependencies
	@echo "$(COLOR_BLUE)Downloading dependencies...$(COLOR_RESET)"
	@$(GO) mod download
	@echo "$(COLOR_GREEN)✓ Dependencies downloaded$(COLOR_RESET)"

.PHONY: tidy
tidy: ## Tidy go.mod and go.sum
	@echo "$(COLOR_BLUE)Tidying go.mod...$(COLOR_RESET)"
	@$(GO) mod tidy
	@echo "$(COLOR_GREEN)✓ go.mod tidied$(COLOR_RESET)"

.PHONY: verify
verify: ## Verify dependencies
	@echo "$(COLOR_BLUE)Verifying dependencies...$(COLOR_RESET)"
	@$(GO) mod verify
	@echo "$(COLOR_GREEN)✓ Dependencies verified$(COLOR_RESET)"

.PHONY: deps-update
deps-update: ## Update all dependencies
	@echo "$(COLOR_BLUE)Updating dependencies...$(COLOR_RESET)"
	@$(GO) get -u ./...
	@$(GO) mod tidy
	@echo "$(COLOR_GREEN)✓ Dependencies updated$(COLOR_RESET)"

# ============================================================================
# Clean Targets
# ============================================================================

.PHONY: clean
clean: ## Remove build artifacts and caches
	@echo "$(COLOR_BLUE)Cleaning build artifacts...$(COLOR_RESET)"
	@rm -rf bin/
	@rm -f $(TEST_COVERAGE_FILE) $(TEST_COVERAGE_HTML)
	@$(GO) clean -cache -testcache -modcache
	@echo "$(COLOR_GREEN)✓ Build artifacts cleaned$(COLOR_RESET)"

.PHONY: clean-bin
clean-bin: ## Remove binaries only
	@echo "$(COLOR_BLUE)Cleaning binaries...$(COLOR_RESET)"
	@rm -rf bin/
	@echo "$(COLOR_GREEN)✓ Binaries cleaned$(COLOR_RESET)"

# ============================================================================
# Docker Targets
# ============================================================================

.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "$(COLOR_BLUE)Building Docker image...$(COLOR_RESET)"
	@docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@docker tag $(DOCKER_IMAGE):$(DOCKER_TAG) $(DOCKER_IMAGE):latest
	@echo "$(COLOR_GREEN)✓ Docker image built: $(DOCKER_IMAGE):$(DOCKER_TAG)$(COLOR_RESET)"

.PHONY: docker-run
docker-run: ## Run Docker container
	@echo "$(COLOR_BLUE)Running Docker container...$(COLOR_RESET)"
	@docker run --rm -it $(DOCKER_IMAGE):$(DOCKER_TAG)

# ============================================================================
# Version and Info Targets
# ============================================================================

.PHONY: version
version: ## Display version information
	@echo "$(COLOR_BOLD)Shadowforge Version Information$(COLOR_RESET)"
	@echo "  Version:      $(COLOR_CYAN)$(VERSION)$(COLOR_RESET)"
	@echo "  Git Commit:   $(COLOR_CYAN)$(GIT_COMMIT)$(COLOR_RESET)"
	@echo "  Git Tag:      $(COLOR_CYAN)$(GIT_TAG)$(COLOR_RESET)"
	@echo "  Git Branch:   $(COLOR_CYAN)$(GIT_BRANCH)$(COLOR_RESET)"
	@echo "  Git Dirty:    $(COLOR_CYAN)$(GIT_DIRTY)$(COLOR_RESET)"
	@echo "  Build Time:   $(COLOR_CYAN)$(BUILD_TIME)$(COLOR_RESET)"
	@echo "  Build User:   $(COLOR_CYAN)$(BUILD_USER)$(COLOR_RESET)"
	@echo "  Build Host:   $(COLOR_CYAN)$(BUILD_HOST)$(COLOR_RESET)"
	@echo "  Go Version:   $(COLOR_CYAN)$(GO_VERSION)$(COLOR_RESET)"
	@echo "  CGO Enabled:  $(COLOR_CYAN)$(CGO_ENABLED)$(COLOR_RESET)"

.PHONY: info
info: version ## Alias for version

# ============================================================================
# CI/CD Targets
# ============================================================================

.PHONY: ci
ci: deps fmt vet lint test build ## Run CI pipeline (format, vet, lint, test, build)

.PHONY: ci-full
ci-full: ci test-integration coverage ## Run full CI pipeline with integration tests and coverage

# ============================================================================
# Release Targets
# ============================================================================

.PHONY: release-dry-run
release-dry-run: ## Dry run of release process
	@echo "$(COLOR_BLUE)Release dry run for version $(VERSION)...$(COLOR_RESET)"
	@echo "  - Would build CLI binary"
	@echo "  - Would build API binary"
	@echo "  - Would run tests"
	@echo "  - Would create git tag: v$(VERSION)"
	@echo "$(COLOR_GREEN)✓ Release dry run complete$(COLOR_RESET)"

# ============================================================================
# Development Utilities
# ============================================================================

.PHONY: run-cli
run-cli: build-cli ## Build and run CLI
	@$(CLI_BINARY) $(ARGS)

.PHONY: run-api
run-api: build-api ## Build and run API server
	@$(API_BINARY) $(ARGS)

.PHONY: watch
watch: ## Watch for changes and rebuild (requires entr)
	@if command -v entr >/dev/null 2>&1; then \
		echo "$(COLOR_BLUE)Watching for changes...$(COLOR_RESET)"; \
		find . -name '*.go' | entr -r make build-cli; \
	else \
		echo "$(COLOR_YELLOW)⚠ entr not installed$(COLOR_RESET)"; \
		echo "$(COLOR_YELLOW)  Install with: brew install entr (macOS) or apt-get install entr (Linux)$(COLOR_RESET)"; \
	fi

.PHONY: pre-commit
pre-commit: fmt vet lint test-short ## Run pre-commit checks

# ============================================================================
# Special Targets
# ============================================================================

# Prevent make from deleting intermediate files
.SECONDARY:

# Disable built-in rules
.SUFFIXES:
