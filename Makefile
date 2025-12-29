# Makefile for Go project

# Variables
BINARY_NAME=app
GO=go
GOFLAGS=-v
GOTEST=$(GO) test
GOVET=$(GO) vet
GOFMT=$(GO) fmt
GOLINT=golangci-lint
DOCKER_IMAGE=app
DOCKER_TAG=latest
COVERAGE_FILE=coverage.out
COVERAGE_HTML=coverage.html

# Directories
CMD_DIR=./cmd
INTERNAL_DIR=./internal
PKG_DIR=./pkg
BUILD_DIR=./build

# ANSI color codes
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[0;33m
BLUE=\033[0;34m
MAGENTA=\033[0;35m
CYAN=\033[0;36m
WHITE=\033[0;37m
BOLD=\033[1m
RESET=\033[0m

.PHONY: help
help: ## Display this help message
	@echo ""
	@echo "$(BOLD)$(CYAN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo "$(BOLD)$(CYAN)  Available Make Targets$(RESET)"
	@echo "$(BOLD)$(CYAN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo ""
	@echo "$(BOLD)$(GREEN)Build Targets:$(RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; /^build/ || /^install/ || /^docker-build/ {printf "  $(YELLOW)%-20s$(RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(BOLD)$(BLUE)Test Targets:$(RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; /^test/ || /^coverage/ {printf "  $(YELLOW)%-20s$(RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(BOLD)$(MAGENTA)Quality Targets:$(RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; /^lint/ || /^fmt/ || /^vet/ {printf "  $(YELLOW)%-20s$(RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(BOLD)$(RED)Development Targets:$(RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; /^run/ || /^clean/ {printf "  $(YELLOW)%-20s$(RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(BOLD)$(CYAN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo ""

.PHONY: build
build: ## Build the application binary
	@echo "$(BOLD)$(GREEN)Building $(BINARY_NAME)...$(RESET)"
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)/...
	@echo "$(BOLD)$(GREEN)✓ Build complete: $(BUILD_DIR)/$(BINARY_NAME)$(RESET)"

.PHONY: test
test: generate-fixtures ## Run all tests
	@echo "$(BOLD)$(BLUE)Running all tests...$(RESET)"
	$(GOTEST) $(GOFLAGS) -race -timeout 5m ./...
	@echo "$(BOLD)$(BLUE)✓ All tests passed$(RESET)"

.PHONY: test-unit
test-unit: ## Run unit tests only
	@echo "$(BOLD)$(BLUE)Running unit tests...$(RESET)"
	$(GOTEST) $(GOFLAGS) -short -race -timeout 30s ./...
	@echo "$(BOLD)$(BLUE)✓ Unit tests passed$(RESET)"

.PHONY: test-integration
test-integration: generate-fixtures ## Run integration tests only
	@echo "$(BOLD)$(BLUE)Running integration tests...$(RESET)"
	$(GOTEST) $(GOFLAGS) -run Integration -race -timeout 5m ./tests/integration/...
	@echo "$(BOLD)$(BLUE)✓ Integration tests passed$(RESET)"

.PHONY: test-bench
test-bench: generate-fixtures ## Run benchmark tests
	@echo "$(BOLD)$(BLUE)Running benchmark tests...$(RESET)"
	$(GOTEST) $(GOFLAGS) -bench=. -benchmem -run=^$$ ./tests/integration/...
	@echo "$(BOLD)$(BLUE)✓ Benchmarks complete$(RESET)"

.PHONY: generate-fixtures
generate-fixtures: ## Generate test fixtures
	@echo "$(BOLD)$(CYAN)Generating test fixtures...$(RESET)"
	@if [ ! -f "testdata/epub/valid/minimal.epub" ]; then \
		echo "  Generating EPUB fixtures..."; \
		cd testdata/epub && $(GO) run generate_fixtures.go .; \
	else \
		echo "  EPUB fixtures already exist"; \
	fi
	@if [ ! -f "testdata/pdf/valid/minimal.pdf" ]; then \
		echo "  Generating PDF fixtures..."; \
		cd testdata/pdf && $(GO) run generate_fixtures.go .; \
	else \
		echo "  PDF fixtures already exist"; \
	fi
	@echo "$(BOLD)$(CYAN)✓ Test fixtures ready$(RESET)"

.PHONY: clean-fixtures
clean-fixtures: ## Clean generated test fixtures
	@echo "$(BOLD)$(RED)Cleaning test fixtures...$(RESET)"
	@rm -rf testdata/epub/valid/*.epub testdata/epub/invalid/*.epub
	@rm -rf testdata/pdf/valid/*.pdf testdata/pdf/invalid/*.pdf
	@echo "$(BOLD)$(RED)✓ Test fixtures cleaned$(RESET)"

.PHONY: coverage
coverage: generate-fixtures ## Generate test coverage report
	@echo "$(BOLD)$(BLUE)Generating coverage report...$(RESET)"
	$(GOTEST) -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	$(GO) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "$(BOLD)$(BLUE)✓ Coverage report generated: $(COVERAGE_HTML)$(RESET)"
	@$(GO) tool cover -func=$(COVERAGE_FILE) | tail -n 1

.PHONY: lint
lint: ## Run linter on the codebase
	@echo "$(BOLD)$(MAGENTA)Running linter...$(RESET)"
	@if command -v $(GOLINT) > /dev/null 2>&1; then \
		$(GOLINT) run ./...; \
		echo "$(BOLD)$(MAGENTA)✓ Linting complete$(RESET)"; \
	else \
		echo "$(BOLD)$(RED)✗ golangci-lint not installed. Install with: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin$(RESET)"; \
		exit 1; \
	fi

.PHONY: fmt
fmt: ## Format code with go fmt
	@echo "$(BOLD)$(MAGENTA)Formatting code...$(RESET)"
	$(GOFMT) ./...
	@echo "$(BOLD)$(MAGENTA)✓ Code formatted$(RESET)"

.PHONY: vet
vet: ## Run go vet on the codebase
	@echo "$(BOLD)$(MAGENTA)Running go vet...$(RESET)"
	$(GOVET) ./...
	@echo "$(BOLD)$(MAGENTA)✓ Vet complete$(RESET)"

.PHONY: clean
clean: ## Clean build artifacts and cache
	@echo "$(BOLD)$(RED)Cleaning build artifacts...$(RESET)"
	@rm -rf $(BUILD_DIR)
	@rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)
	$(GO) clean -cache -testcache -modcache
	@echo "$(BOLD)$(RED)✓ Clean complete$(RESET)"

.PHONY: install
install: ## Install dependencies
	@echo "$(BOLD)$(GREEN)Installing dependencies...$(RESET)"
	$(GO) mod download
	$(GO) mod tidy
	@echo "$(BOLD)$(GREEN)✓ Dependencies installed$(RESET)"

.PHONY: run
run: ## Run the application
	@echo "$(BOLD)$(RED)Running application...$(RESET)"
	$(GO) run $(CMD_DIR)/...

.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "$(BOLD)$(GREEN)Building Docker image...$(RESET)"
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "$(BOLD)$(GREEN)✓ Docker image built: $(DOCKER_IMAGE):$(DOCKER_TAG)$(RESET)"

.DEFAULT_GOAL := help
