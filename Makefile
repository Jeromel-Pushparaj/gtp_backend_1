.PHONY: help setup infra-up infra-down infra-logs clean

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

setup: ## Initial setup - copy .env.example files
	@echo "Setting up environment files..."
	@cp -n .env.example .env || true
	@cp -n services/jira-trigger-service/.env.example services/jira-trigger-service/.env || true
	@cp -n services/chat-agent-service/.env.example services/chat-agent-service/.env || true
	@cp -n services/approval-service/.env.example services/approval-service/.env || true
	@cp -n services/onboarding-service/.env.example services/onboarding-service/.env || true
	@cp -n gateway/api-gateway/.env.example gateway/api-gateway/.env || true
	@echo "Environment files created. Please update them with your configuration."

infra-up: ## Start infrastructure (Kafka, PostgreSQL, Redis)
	@echo "Starting infrastructure services..."
	docker-compose up -d
	@echo "Infrastructure is running!"
	@echo "Kafka UI: http://localhost:8090"
	@echo "PostgreSQL: localhost:5432"
	@echo "Redis: localhost:6379"

infra-down: ## Stop infrastructure
	@echo "Stopping infrastructure services..."
	docker-compose down

infra-logs: ## View infrastructure logs
	docker-compose logs -f

infra-restart: ## Restart infrastructure
	@echo "Restarting infrastructure services..."
	docker-compose restart

clean: ## Clean up infrastructure volumes and data
	@echo "Cleaning up infrastructure..."
	docker-compose down -v
	@echo "Infrastructure cleaned!"

# Service-specific targets
jira-trigger: ## Run jira-trigger-service
	cd services/jira-trigger-service && go run cmd/main.go

chat-agent: ## Run chat-agent-service
	cd services/chat-agent-service && go run cmd/main.go

approval: ## Run approval-service
	cd services/approval-service && go run cmd/main.go

service-catelog: ## Run service-service
	cd services/service-catelog && go run cmd/main.go

score-card: ## Run score-card-service
	cd services/score-card-service && go run cmd/main.go

gateway: ## Run api-gateway
	cd gateway/api-gateway && go run cmd/main.go

# Testing targets
test: ## Run all tests
	@echo "Running tests for all services..."
	cd services/jira-trigger-service && go test ./... || true
	cd services/chat-agent-service && go test ./... || true
	cd services/approval-service && go test ./... || true
	cd services/onboarding-service && go test ./... || true
	cd services/score-card-service && go test ./... || true
	cd gateway/api-gateway && go test ./... || true

# Dependency management
tidy: ## Run go mod tidy for all services
	@echo "Tidying dependencies..."
	cd services/jira-trigger-service && go mod tidy
	cd services/chat-agent-service && go mod tidy
	cd services/approval-service && go mod tidy
	cd services/onboarding-service && go mod tidy
	cd services/score-card-service && go mod tidy
	cd gateway/api-gateway && go mod tidy
	@echo "Dependencies tidied!"
