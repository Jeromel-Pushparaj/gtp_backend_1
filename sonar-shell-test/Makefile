# Makefile for Sonar Shell Test

# Variables
APP_NAME=sonar-automation
BINARY_NAME=sonar-shell-test
BIN_DIR=./bin
SERVER_BINARY=$(BIN_DIR)/server
SEED_BINARY=$(BIN_DIR)/seed
MAIN_FILE=main.go
PORT?=8080
GO=go
GOFLAGS=-v

# Colors for output
GREEN=\033[0;32m
YELLOW=\033[1;33m
RED=\033[0;31m
NC=\033[0m # No Color

.PHONY: help
help: ## Show this help message
	@echo "$(GREEN)Available commands:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(YELLOW)%-20s$(NC) %s\n", $$1, $$2}'

.PHONY: all
all: clean deps build ## Clean, install dependencies, and build

.PHONY: deps
deps: ## Install Go dependencies
	@echo "$(GREEN)Installing dependencies...$(NC)"
	$(GO) mod download
	$(GO) mod tidy

.PHONY: build
build: ## Build the application
	@echo "$(GREEN)Building $(BINARY_NAME)...$(NC)"
	@mkdir -p $(BIN_DIR)
	$(GO) build $(GOFLAGS) -o $(SERVER_BINARY) $(MAIN_FILE)
	@echo "$(GREEN)✅ Build complete: $(SERVER_BINARY)$(NC)"

.PHONY: build-seed
build-seed: ## Build the seed command
	@echo "$(GREEN)Building seed binary...$(NC)"
	@mkdir -p $(BIN_DIR)
	$(GO) build $(GOFLAGS) -o $(SEED_BINARY) ./cmd/seed.go
	@echo "$(GREEN)✅ Seed build complete: $(SEED_BINARY)$(NC)"

.PHONY: run
run: build ## Build and run the CLI application
	@echo "$(GREEN)Running $(BINARY_NAME) in CLI mode...$(NC)"
	$(SERVER_BINARY)

.PHONY: server
server: build ## Build and run the API server
	@echo "$(GREEN)Starting API server on port $(PORT)...$(NC)"
	$(SERVER_BINARY) -server -port=$(PORT)

.PHONY: up
up: server ## Alias for 'server' - start the API server

.PHONY: dev
dev: ## Run the server in development mode with auto-reload (requires air)
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "$(RED)Error: 'air' not found. Install it with: go install github.com/cosmtrek/air@latest$(NC)"; \
		echo "$(YELLOW)Falling back to regular server mode...$(NC)"; \
		$(MAKE) server; \
	fi

.PHONY: list-secrets
list-secrets: build ## List all secrets in repositories
	@echo "$(GREEN)Listing secrets...$(NC)"
	$(SERVER_BINARY) -list-secrets

.PHONY: add-env-secrets
add-env-secrets: build ## Add environment secrets to repositories
	@echo "$(GREEN)Adding environment secrets...$(NC)"
	$(SERVER_BINARY) -add-env-secrets

.PHONY: update-workflows
update-workflows: build ## Update workflows to use environment
	@echo "$(GREEN)Updating workflows...$(NC)"
	$(SERVER_BINARY) -update-workflows

.PHONY: full-setup
full-setup: build ## Run full setup (create projects, add workflows, fetch results)
	@echo "$(GREEN)Running full setup...$(NC)"
	$(SERVER_BINARY) -full-setup

.PHONY: fetch-results
fetch-results: build ## Fetch analysis results from SonarCloud
	@echo "$(GREEN)Fetching results...$(NC)"
	$(SERVER_BINARY) -fetch-results

.PHONY: test
test: ## Run tests
	@echo "$(GREEN)Running tests...$(NC)"
	$(GO) test -v ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	@echo "$(GREEN)Running tests with coverage...$(NC)"
	$(GO) test -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report generated: coverage.html$(NC)"

.PHONY: clean
clean: ## Clean build artifacts
	@echo "$(YELLOW)Cleaning build artifacts...$(NC)"
	rm -rf $(BIN_DIR)
	rm -f coverage.out coverage.html
	@echo "$(GREEN)✅ Clean complete$(NC)"

.PHONY: fmt
fmt: ## Format Go code
	@echo "$(GREEN)Formatting code...$(NC)"
	$(GO) fmt ./...

.PHONY: lint
lint: ## Run linter (requires golangci-lint)
	@if command -v golangci-lint > /dev/null; then \
		echo "$(GREEN)Running linter...$(NC)"; \
		golangci-lint run; \
	else \
		echo "$(RED)Error: 'golangci-lint' not found. Install it from https://golangci-lint.run/$(NC)"; \
	fi

.PHONY: vet
vet: ## Run go vet
	@echo "$(GREEN)Running go vet...$(NC)"
	$(GO) vet ./...

.PHONY: check
check: fmt vet test ## Run format, vet, and tests

.PHONY: install
install: build ## Install the binary to $GOPATH/bin
	@echo "$(GREEN)Installing $(BINARY_NAME)...$(NC)"
	$(GO) install

.PHONY: docker-build
docker-build: ## Build Docker image (if Dockerfile exists)
	@if [ -f Dockerfile ]; then \
		echo "$(GREEN)Building Docker image...$(NC)"; \
		docker build -t $(BINARY_NAME):latest .; \
	else \
		echo "$(RED)Error: Dockerfile not found$(NC)"; \
	fi

.PHONY: docker-run
docker-run: ## Run Docker container
	@echo "$(GREEN)Running Docker container...$(NC)"
	docker run -p $(PORT):$(PORT) --env-file .env $(BINARY_NAME):latest

.PHONY: logs
logs: ## Show application logs (if running in background)
	@echo "$(GREEN)Showing logs...$(NC)"
	@tail -f logs/*.log 2>/dev/null || echo "$(YELLOW)No log files found$(NC)"

.PHONY: env-check
env-check: ## Check if required environment variables are set
	@echo "$(GREEN)Checking environment variables...$(NC)"
	@if [ ! -f .env ]; then \
		echo "$(RED)Error: .env file not found$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)✅ .env file exists$(NC)"

