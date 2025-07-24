# MCP Todo Server Makefile
# Comprehensive build automation for Go MCP Todo Server

# Variables
BINARY_NAME := mcp-todo-server
PACKAGE := github.com/user/mcp-todo-server
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DATE := $(shell date -u '+%Y-%m-%d %H:%M:%S')
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod
GOFMT := gofmt
GOVET := $(GOCMD) vet

# Build flags
LDFLAGS := -ldflags "-X 'main.Version=$(VERSION)' -X 'main.BuildDate=$(BUILD_DATE)' -X 'main.Commit=$(COMMIT)'"
BUILD_FLAGS := -v

# Directories
BUILD_DIR := ./build
COVERAGE_DIR := ./coverage
SRC_DIRS := ./... 

# Test flags
TEST_FLAGS := -v -race
BENCH_FLAGS := -bench=. -benchmem
COVERAGE_FLAGS := -coverprofile=$(COVERAGE_DIR)/coverage.out -covermode=atomic

# Colors for output
GREEN := \033[0;32m
YELLOW := \033[0;33m
RED := \033[0;31m
NC := \033[0m # No Color

# Default target
.PHONY: all
all: clean test build

# Build targets
.PHONY: build
build: ## Build the server binary
	@echo "$(GREEN)Building $(BINARY_NAME)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "$(GREEN)Build complete: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

.PHONY: build-all
build-all: ## Build for multiple platforms
	@echo "$(GREEN)Building for multiple platforms...$(NC)"
	@mkdir -p $(BUILD_DIR)
	# Darwin AMD64
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 .
	# Darwin ARM64 (M1/M2)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 .
	# Linux AMD64
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .
	# Linux ARM64
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 .
	# Windows AMD64
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .
	@echo "$(GREEN)Multi-platform build complete$(NC)"

.PHONY: install
install: build ## Build and install to $GOPATH/bin
	@echo "$(GREEN)Installing $(BINARY_NAME)...$(NC)"
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/$(BINARY_NAME)
	@echo "$(GREEN)Installed to $(GOPATH)/bin/$(BINARY_NAME)$(NC)"

.PHONY: clean
clean: ## Remove build artifacts and coverage files
	@echo "$(YELLOW)Cleaning up...$(NC)"
	@rm -rf $(BUILD_DIR)
	@rm -rf $(COVERAGE_DIR)
	@rm -f $(BINARY_NAME)
	@rm -f *.out *.html
	@rm -f server.log
	@echo "$(GREEN)Clean complete$(NC)"

# Test targets
.PHONY: test
test: ## Run all unit tests
	@echo "$(GREEN)Running tests...$(NC)"
	$(GOTEST) $(TEST_FLAGS) -timeout 30s $(SRC_DIRS)

.PHONY: test-verbose
test-verbose: ## Run tests with verbose output
	@echo "$(GREEN)Running tests (verbose)...$(NC)"
	$(GOTEST) $(TEST_FLAGS) -v -timeout 30s $(SRC_DIRS)

.PHONY: test-race
test-race: ## Run tests with race detector
	@echo "$(GREEN)Running tests with race detector...$(NC)"
	$(GOTEST) -race -timeout 60s $(SRC_DIRS)

.PHONY: test-integration
test-integration: ## Run integration tests only
	@echo "$(GREEN)Running integration tests...$(NC)"
	$(GOTEST) $(TEST_FLAGS) -tags=integration -timeout 60s $(SRC_DIRS)

.PHONY: test-coverage
test-coverage: ## Generate coverage report and open HTML
	@echo "$(GREEN)Running tests with coverage...$(NC)"
	@mkdir -p $(COVERAGE_DIR)
	$(GOTEST) $(TEST_FLAGS) $(COVERAGE_FLAGS) -timeout 30s $(SRC_DIRS)
	@$(GOCMD) tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@echo "$(GREEN)Coverage report generated: $(COVERAGE_DIR)/coverage.html$(NC)"
	@if command -v open > /dev/null; then \
		open $(COVERAGE_DIR)/coverage.html; \
	elif command -v xdg-open > /dev/null; then \
		xdg-open $(COVERAGE_DIR)/coverage.html; \
	fi

.PHONY: test-bench
test-bench: ## Run benchmark tests
	@echo "$(GREEN)Running benchmarks...$(NC)"
	$(GOTEST) $(BENCH_FLAGS) $(SRC_DIRS)

.PHONY: test-short
test-short: ## Run short tests (exclude integration)
	@echo "$(GREEN)Running short tests...$(NC)"
	$(GOTEST) -short $(TEST_FLAGS) -timeout 20s $(SRC_DIRS)

.PHONY: test-quick
test-quick: ## Quick test run for CI/Claude Code (no race detector, 20s timeout)
	@echo "$(GREEN)Running quick tests...$(NC)"
	$(GOTEST) -v -timeout 20s $(SRC_DIRS)

.PHONY: test-claude
test-claude: ## Test run optimized for Claude Code (exits cleanly, clear output)
	@echo "$(GREEN)Running tests for Claude Code...$(NC)"
	@if $(GOTEST) -timeout 20s $(SRC_DIRS) > /tmp/test.log 2>&1; then \
		echo "$(GREEN)✓ All tests passed!$(NC)"; \
		exit 0; \
	else \
		echo "$(RED)✗ Tests failed! Output:$(NC)"; \
		cat /tmp/test.log | grep -E "FAIL|Error:|panic:" | head -20; \
		exit 1; \
	fi

# Development targets
.PHONY: run
run: build ## Build and run server in HTTP mode
	@echo "$(GREEN)Starting server in HTTP mode...$(NC)"
	$(BUILD_DIR)/$(BINARY_NAME) -transport http -port 8080

.PHONY: run-stdio
run-stdio: build ## Build and run server in STDIO mode
	@echo "$(GREEN)Starting server in STDIO mode...$(NC)"
	$(BUILD_DIR)/$(BINARY_NAME) -transport stdio

.PHONY: lint
lint: ## Run golangci-lint
	@echo "$(GREEN)Running linter...$(NC)"
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run $(SRC_DIRS); \
	else \
		echo "$(YELLOW)golangci-lint not installed. Run 'make tools' to install.$(NC)"; \
	fi

.PHONY: fmt
fmt: ## Format code with go fmt
	@echo "$(GREEN)Formatting code...$(NC)"
	@$(GOFMT) -w .
	@echo "$(GREEN)Code formatting complete$(NC)"

.PHONY: fmt-check
fmt-check: ## Check if code is formatted
	@echo "$(GREEN)Checking code formatting...$(NC)"
	@if [ -n "$$($(GOFMT) -l .)" ]; then \
		echo "$(RED)The following files need formatting:$(NC)"; \
		$(GOFMT) -l .; \
		exit 1; \
	else \
		echo "$(GREEN)All files are properly formatted$(NC)"; \
	fi

.PHONY: vet
vet: ## Run go vet for static analysis
	@echo "$(GREEN)Running go vet...$(NC)"
	$(GOVET) $(SRC_DIRS)

.PHONY: mod-tidy
mod-tidy: ## Clean up go.mod dependencies
	@echo "$(GREEN)Tidying go.mod...$(NC)"
	$(GOMOD) tidy

.PHONY: mod-download
mod-download: ## Download dependencies
	@echo "$(GREEN)Downloading dependencies...$(NC)"
	$(GOMOD) download

.PHONY: mod-verify
mod-verify: ## Verify dependencies
	@echo "$(GREEN)Verifying dependencies...$(NC)"
	$(GOMOD) verify

# Server management
.PHONY: server-http
server-http: build ## Start HTTP server (port 8080)
	@echo "$(GREEN)Starting HTTP server on port 8080...$(NC)"
	$(BUILD_DIR)/$(BINARY_NAME) -transport http -port 8080

.PHONY: server-stdio
server-stdio: build ## Start STDIO server
	@echo "$(GREEN)Starting STDIO server...$(NC)"
	$(BUILD_DIR)/$(BINARY_NAME) -transport stdio

.PHONY: server-dev
server-dev: ## Start with development settings
	@echo "$(GREEN)Starting server in development mode...$(NC)"
	@LOG_LEVEL=debug $(BUILD_DIR)/$(BINARY_NAME) -transport http -port 8080

.PHONY: test-e2e
test-e2e: build ## Run comprehensive end-to-end tests
	@echo "$(GREEN)Running end-to-end tests...$(NC)"
	@chmod +x ./scripts/test/e2e/test_comprehensive.sh
	@./scripts/test/e2e/test_comprehensive.sh

# Documentation & Release
.PHONY: docs
docs: ## Generate godoc documentation
	@echo "$(GREEN)Generating documentation...$(NC)"
	@echo "Documentation available at http://localhost:6060/pkg/$(PACKAGE)/"
	@godoc -http=:6060

.PHONY: check
check: fmt-check vet lint test ## Run all checks (fmt, vet, lint, test)
	@echo "$(GREEN)All checks passed!$(NC)"

.PHONY: release
release: clean test build-all ## Build release binaries for all platforms
	@echo "$(GREEN)Creating release artifacts...$(NC)"
	@mkdir -p $(BUILD_DIR)/release
	@cp $(BUILD_DIR)/$(BINARY_NAME)-* $(BUILD_DIR)/release/
	@cd $(BUILD_DIR)/release && \
		for file in *; do \
			if [ -f "$$file" ]; then \
				tar czf "$$file.tar.gz" "$$file"; \
				rm "$$file"; \
			fi; \
		done
	@echo "$(GREEN)Release artifacts created in $(BUILD_DIR)/release/$(NC)"

# Utility targets
.PHONY: setup
setup: ## Initial project setup (deps, tools)
	@echo "$(GREEN)Running initial setup...$(NC)"
	@chmod +x ./scripts/setup/setup.sh
	@./scripts/setup/setup.sh

.PHONY: tools
tools: ## Install development tools
	@echo "$(GREEN)Installing development tools...$(NC)"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install golang.org/x/tools/cmd/godoc@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@echo "$(GREEN)Tools installed$(NC)"

.PHONY: todo-stats
todo-stats: build ## Generate project todo statistics
	@echo "$(GREEN)Generating todo statistics...$(NC)"
	@$(BUILD_DIR)/$(BINARY_NAME) -transport stdio < <(echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"todo_stats","arguments":{}}}')

.PHONY: clean-todos
clean-todos: ## Clean test todo files
	@echo "$(YELLOW)Cleaning test todo files...$(NC)"
	@rm -rf /tmp/test-todos
	@rm -rf ~/.claude/todos/test-*
	@echo "$(GREEN)Test todos cleaned$(NC)"

# Helper targets
.PHONY: help
help: ## Display this help message
	@echo "$(GREEN)MCP Todo Server - Available targets:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  $(YELLOW)%-20s$(NC) %s\n", $$1, $$2}'

.PHONY: version
version: ## Display version information
	@echo "$(GREEN)MCP Todo Server$(NC)"
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT)"
	@echo "Build Date: $(BUILD_DATE)"

# Quick test scripts
.PHONY: test-http-quick
test-http-quick: build ## Quick HTTP transport test
	@echo "$(GREEN)Running quick HTTP test...$(NC)"
	@chmod +x ./scripts/test/http/test_http.sh
	@./scripts/test/http/test_http.sh

.PHONY: test-stdio-quick
test-stdio-quick: build ## Quick STDIO transport test
	@echo "$(GREEN)Running quick STDIO test...$(NC)"
	@chmod +x ./scripts/test/stdio/test_server.sh
	@./scripts/test/stdio/test_server.sh

# Docker targets
.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "$(GREEN)Building Docker image...$(NC)"
	docker build -t mcp-todo-server:$(VERSION) -t mcp-todo-server:latest .

.PHONY: docker-run
docker-run: docker-build ## Run Docker container
	@echo "$(GREEN)Running Docker container...$(NC)"
	docker run -it --rm \
		-p 8080:8080 \
		-v ~/.claude/todos:/home/mcp/.claude/todos \
		--name mcp-todo-server \
		mcp-todo-server:latest

.PHONY: docker-compose-up
docker-compose-up: ## Start services with docker-compose
	@echo "$(GREEN)Starting services with docker-compose...$(NC)"
	docker-compose up -d

.PHONY: docker-compose-down
docker-compose-down: ## Stop services with docker-compose
	@echo "$(GREEN)Stopping services with docker-compose...$(NC)"
	docker-compose down

.PHONY: docker-compose-logs
docker-compose-logs: ## View docker-compose logs
	docker-compose logs -f

.PHONY: docker-push
docker-push: docker-build ## Push Docker image to registry
	@echo "$(GREEN)Pushing Docker image...$(NC)"
	@echo "$(YELLOW)Note: You need to tag with your registry URL first$(NC)"
	@echo "Example: docker tag mcp-todo-server:latest your-registry/mcp-todo-server:latest"

# Workflow testing
.PHONY: workflow-test
workflow-test: ## Test GitHub workflows locally with act
	@if command -v act > /dev/null; then \
		echo "$(GREEN)Testing CI workflow...$(NC)"; \
		echo "$(YELLOW)Note: Make sure Docker is running and you've selected a container image in act$(NC)"; \
		act push -j lint --container-architecture linux/amd64 || true; \
		act push -j test --container-architecture linux/amd64 -P ubuntu-latest=catthehacker/ubuntu:act-latest || true; \
		act push -j build --container-architecture linux/amd64 || true; \
	else \
		echo "$(YELLOW)act not installed. Install with: brew install act$(NC)"; \
		echo "Then run: act -l to list workflows"; \
	fi

.PHONY: workflow-lint
workflow-lint: ## Lint GitHub workflow files
	@if command -v actionlint > /dev/null; then \
		echo "$(GREEN)Linting workflow files...$(NC)"; \
		actionlint .github/workflows/*.yml; \
	else \
		echo "$(YELLOW)actionlint not installed. Install with: brew install actionlint$(NC)"; \
	fi

# Shortcuts
.PHONY: b
b: build ## Shortcut for build

.PHONY: t
t: test ## Shortcut for test

.PHONY: tc
tc: test-claude ## Shortcut for test-claude (AI-friendly)

.PHONY: tq
tq: test-quick ## Shortcut for test-quick

.PHONY: r
r: run ## Shortcut for run

.PHONY: c
c: clean ## Shortcut for clean

.PHONY: d
d: docker-run ## Shortcut for docker-run

# Default help
.DEFAULT_GOAL := help