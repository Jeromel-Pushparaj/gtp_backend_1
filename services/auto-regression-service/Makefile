# Auto-Regression Service Makefile
# Orchestrates both auto_regression_agent and auto_regression_api

.PHONY: help build up down logs clean test deps

# Colors for output
BLUE := \033[0;34m
GREEN := \033[0;32m
YELLOW := \033[0;33m
NC := \033[0m # No Color

help: ## Show this help message
	@echo "$(BLUE)Auto-Regression Service - Available Commands:$(NC)"
	@echo ""
	@echo "$(GREEN)Main Commands:$(NC)"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(YELLOW)%-20s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# ============================================================================
# Build Commands
# ============================================================================

build: build-agent build-api ## Build both agent and API components
	@echo "$(GREEN)✓ All components built successfully$(NC)"

build-agent: ## Build the auto-regression agent
	@echo "$(BLUE)Building auto-regression agent...$(NC)"
	cd auto_regression_agent && $(MAKE) build

build-api: ## Build the auto-regression API
	@echo "$(BLUE)Building auto-regression API...$(NC)"
	cd auto_regression_api && go build -o middleman main.go

# ============================================================================
# Run Commands
# ============================================================================

up: ## Start all services (agent + API)
	@echo "$(BLUE)Starting auto-regression services...$(NC)"
	cd auto_regression_agent && docker-compose up -d
	@echo "$(GREEN)✓ Services started$(NC)"
	@echo "$(YELLOW)Agent API: http://localhost:8080$(NC)"
	@echo "$(YELLOW)Auto-Regression API: http://localhost:8092$(NC)"

down: ## Stop all services
	@echo "$(BLUE)Stopping auto-regression services...$(NC)"
	cd auto_regression_agent && docker-compose down
	@echo "$(GREEN)✓ Services stopped$(NC)"

restart: down up ## Restart all services

start: up run-api ## Start all services and run API on port 8092

run-all: ## Run both agent and API without Docker (development mode)
	@echo "$(BLUE)Starting services without Docker...$(NC)"
	@echo "$(YELLOW)This will start:$(NC)"
	@echo "  • Agent API on port 8080"
	@echo "  • Auto-Regression API on port 8092"
	@echo ""
	@echo "$(YELLOW)Press Ctrl+C to stop all services$(NC)"
	@echo ""
	@$(MAKE) run-agent & $(MAKE) run-api

# ============================================================================
# Development Commands
# ============================================================================

run-agent: ## Run the agent server locally (without Docker)
	@echo "$(BLUE)Running agent server on port 8080 (without Docker)...$(NC)"
	@echo "$(YELLOW)⚠️  Running in development mode (in-memory storage)$(NC)"
	cd auto_regression_agent && go run cmd/server/main.go

run-api: ## Run the API middleman locally on port 8092
	@echo "$(BLUE)Running Auto-Regression API on port 8092...$(NC)"
	cd auto_regression_api && PORT=8092 go run main.go

run-worker: ## Run the agent worker
	@echo "$(BLUE)Running agent worker...$(NC)"
	cd auto_regression_agent && $(MAKE) run-worker

# ============================================================================
# Docker Commands
# ============================================================================

docker-build: ## Build Docker images for all components
	@echo "$(BLUE)Building Docker images...$(NC)"
	cd auto_regression_agent && docker-compose build
	cd auto_regression_api && docker build -t auto-regression-api:latest .
	@echo "$(GREEN)✓ Docker images built$(NC)"

docker-up: ## Start services using Docker Compose
	@echo "$(BLUE)Starting services with Docker...$(NC)"
	cd auto_regression_agent && docker-compose up -d
	@echo "$(GREEN)✓ Services running in Docker$(NC)"

docker-down: ## Stop Docker services
	@echo "$(BLUE)Stopping Docker services...$(NC)"
	cd auto_regression_agent && docker-compose down
	@echo "$(GREEN)✓ Docker services stopped$(NC)"

docker-logs: ## View Docker logs
	cd auto_regression_agent && docker-compose logs -f

docker-ps: ## Show running Docker containers
	cd auto_regression_agent && docker-compose ps

# ============================================================================
# Testing Commands
# ============================================================================

test: test-agent test-api ## Run all tests

test-agent: ## Run agent tests
	@echo "$(BLUE)Running agent tests...$(NC)"
	cd auto_regression_agent && $(MAKE) test

test-api: ## Run API tests
	@echo "$(BLUE)Running API tests...$(NC)"
	cd auto_regression_api && go test -v ./...

test-integration: ## Run integration tests
	@echo "$(BLUE)Running integration tests...$(NC)"
	cd auto_regression_agent && $(MAKE) test-integration

# ============================================================================
# Dependency Management
# ============================================================================

deps: deps-agent deps-api ## Install all dependencies

deps-agent: ## Install agent dependencies
	@echo "$(BLUE)Installing agent dependencies...$(NC)"
	cd auto_regression_agent && go mod download && go mod tidy

deps-api: ## Install API dependencies
	@echo "$(BLUE)Installing API dependencies...$(NC)"
	cd auto_regression_api && go mod download && go mod tidy

# ============================================================================
# Cleanup Commands
# ============================================================================

clean: clean-agent clean-api ## Clean all build artifacts
	@echo "$(GREEN)✓ Cleanup complete$(NC)"

clean-agent: ## Clean agent build artifacts
	@echo "$(BLUE)Cleaning agent artifacts...$(NC)"
	cd auto_regression_agent && $(MAKE) clean

clean-api: ## Clean API build artifacts
	@echo "$(BLUE)Cleaning API artifacts...$(NC)"
	cd auto_regression_api && rm -f middleman
	cd auto_regression_api && go clean -cache -testcache

clean-all: clean docker-down ## Clean everything including Docker volumes
	@echo "$(BLUE)Removing Docker volumes...$(NC)"
	cd auto_regression_agent && docker-compose down -v
	@echo "$(GREEN)✓ Complete cleanup done$(NC)"

# ============================================================================
# Utility Commands
# ============================================================================

logs: ## View logs from all services
	cd auto_regression_agent && docker-compose logs -f

logs-agent: ## View agent logs
	cd auto_regression_agent && docker-compose logs -f server

logs-worker: ## View worker logs
	cd auto_regression_agent && docker-compose logs -f worker

health: ## Check health of all services
	@echo "$(BLUE)Checking service health...$(NC)"
	@echo "$(YELLOW)Agent API (port 8080):$(NC)"
	@curl -s http://localhost:8080/health | jq '.' || echo "$(YELLOW)  ⚠️  Agent not responding$(NC)"
	@echo ""
	@echo "$(YELLOW)Auto-Regression API (port 8092):$(NC)"
	@curl -s http://localhost:8092/health | jq '.' || echo "$(YELLOW)  ⚠️  API not responding$(NC)"

fmt: ## Format code in both components
	@echo "$(BLUE)Formatting code...$(NC)"
	cd auto_regression_agent && go fmt ./...
	cd auto_regression_api && go fmt ./...

.DEFAULT_GOAL := help

