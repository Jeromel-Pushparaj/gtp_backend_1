# Auto-Regression Platform Makefile

.PHONY: help build test clean run-server run-worker run-scheduler docker-build

# Variables
BINARY_DIR=bin
SERVER_BINARY=$(BINARY_DIR)/server
WORKER_BINARY=$(BINARY_DIR)/worker
SCHEDULER_BINARY=$(BINARY_DIR)/scheduler
AGENT_RUNNER_BINARY=$(BINARY_DIR)/agent-runner
CLI_BINARY=$(BINARY_DIR)/cli

GO=go
GOFLAGS=-v
LDFLAGS=-ldflags "-s -w"

# Default target
help:
	@echo "Auto-Regression Platform - Available targets:"
	@echo "  make build              - Build all binaries"
	@echo "  make build-server       - Build API server"
	@echo "  make build-worker       - Build worker"
	@echo "  make build-executor     - Build executor"
	@echo "  make test               - Run unit tests"
	@echo "  make test-integration   - Run integration tests"
	@echo "  make test-coverage      - Run tests with coverage"
	@echo "  make lint               - Run linters"
	@echo "  make fmt                - Format code"
	@echo "  make clean              - Clean build artifacts"
	@echo "  make run-server         - Run API server"
	@echo "  make run-worker         - Run worker"
	@echo "  make run-executor       - Run executor"
	@echo "  make docker-build       - Build all Docker images"
	@echo "  make docker-up          - Start all services"
	@echo "  make docker-down        - Stop all services"
	@echo "  make docker-logs        - View service logs"
	@echo "  make health             - Check service health"

# Build targets
build: build-server build-worker build-scheduler build-agent-runner build-cli

build-server:
	@echo "Building API server..."
	@mkdir -p $(BINARY_DIR)
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(SERVER_BINARY) ./cmd/server

build-worker:
	@echo "Building worker..."
	@mkdir -p $(BINARY_DIR)
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(WORKER_BINARY) ./cmd/worker

build-scheduler:
	@echo "Building scheduler..."
	@mkdir -p $(BINARY_DIR)
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(SCHEDULER_BINARY) ./cmd/scheduler

build-agent-runner:
	@echo "Building agent runner..."
	@mkdir -p $(BINARY_DIR)
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(AGENT_RUNNER_BINARY) ./cmd/agent-runner

build-cli:
	@echo "Building CLI..."
	@mkdir -p $(BINARY_DIR)
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(CLI_BINARY) ./cmd/cli

# Test targets
test:
	@echo "Running unit tests..."
	$(GO) test -v -race -short ./...

test-integration:
	@echo "Running integration tests..."
	$(GO) test -v -race -tags=integration ./tests/integration/...

test-e2e:
	@echo "Running end-to-end tests..."
	$(GO) test -v -tags=e2e ./tests/integration/...

test-coverage:
	@echo "Running tests with coverage..."
	$(GO) test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Code quality targets
lint:
	@echo "Running linters..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed" && exit 1)
	golangci-lint run ./...

fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...
	@which goimports > /dev/null && goimports -w . || echo "goimports not installed, skipping"

# Clean target
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BINARY_DIR)
	rm -f coverage.out coverage.html
	$(GO) clean -cache -testcache

# Run targets
run-server: build-server
	@echo "Starting API server..."
	$(SERVER_BINARY)

run-worker: build-worker
	@echo "Starting worker..."
	$(WORKER_BINARY)

run-scheduler: build-scheduler
	@echo "Starting scheduler..."
	$(SCHEDULER_BINARY)

# Docker targets
docker-build:
	@echo "Building Docker images..."
	docker build -f deployments/docker/Dockerfile.server -t auto-regression/server:latest .
	docker build -f deployments/docker/Dockerfile.worker -t auto-regression/worker:latest .
	docker build -f deployments/docker/Dockerfile.scheduler -t auto-regression/scheduler:latest .

# Database migration targets
migrate-up:
	@echo "Running database migrations up..."
	@which migrate > /dev/null || (echo "golang-migrate not installed" && exit 1)
	migrate -path migrations/postgres -database "$(DATABASE_URL)" up

migrate-down:
	@echo "Running database migrations down..."
	@which migrate > /dev/null || (echo "golang-migrate not installed" && exit 1)
	migrate -path migrations/postgres -database "$(DATABASE_URL)" down

# Development helpers
dev-setup:
	@echo "Setting up development environment..."
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	@which goimports > /dev/null || (echo "Installing goimports..." && go install golang.org/x/tools/cmd/goimports@latest)
	@which migrate > /dev/null || (echo "Installing golang-migrate..." && go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest)
	@echo "Development environment ready!"

# Docker targets
docker-up:
	@echo "Starting all services..."
	cd deployments/docker && docker-compose up -d
	@echo "Services started! Access:"
	@echo "  - Server:     http://localhost:8080"
	@echo "  - Grafana:    http://localhost:3000 (admin/admin)"
	@echo "  - Prometheus: http://localhost:9090"
	@echo "  - Jaeger:     http://localhost:16686"
	@echo "  - MinIO:      http://localhost:9001 (minioadmin/minioadmin)"

docker-down:
	@echo "Stopping all services..."
	cd deployments/docker && docker-compose down

docker-down-volumes:
	@echo "Stopping all services and removing volumes..."
	cd deployments/docker && docker-compose down -v

docker-logs:
	cd deployments/docker && docker-compose logs -f

docker-logs-server:
	cd deployments/docker && docker-compose logs -f server

docker-logs-worker:
	cd deployments/docker && docker-compose logs -f worker-1 worker-2 worker-3

docker-ps:
	cd deployments/docker && docker-compose ps

docker-restart:
	cd deployments/docker && docker-compose restart

# Executor targets
build-executor:
	@echo "Building executor..."
	@mkdir -p $(BINARY_DIR)
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BINARY_DIR)/executor ./cmd/executor

run-executor:
	@echo "Running executor with example manifest..."
	$(GO) run ./cmd/executor -manifest examples/manifests/petstore-test-manifest.json

# Health check
health:
	@echo "Checking service health..."
	@curl -s http://localhost:8080/health | jq . || echo "Server not responding"
	@docker exec auto-regression-redis redis-cli ping || echo "Redis not responding"
	@docker exec auto-regression-postgres pg_isready -U postgres || echo "PostgreSQL not responding"

.DEFAULT_GOAL := help

