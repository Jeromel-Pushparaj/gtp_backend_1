# 🌟 GTP Backend API Gateway - Makefile
# Convenient commands for building, running, and managing the gateway

.PHONY: help run build clean test install deps docker-build docker-run

# 🎯 Default target - show help
help:
	@echo "🌟 GTP Backend API Gateway - Available Commands"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "  make run          - 🚀 Run the gateway (development mode)"
	@echo "  make build        - 🔨 Build the gateway binary"
	@echo "  make install      - 📦 Install dependencies"
	@echo "  make deps         - 🔄 Download and tidy dependencies"
	@echo "  make clean        - 🧹 Clean build artifacts"
	@echo "  make test         - 🧪 Run tests"
	@echo "  make docker-build - 🐳 Build Docker image"
	@echo "  make docker-run   - 🐳 Run Docker container"
	@echo "  make fmt          - 🎨 Format code"
	@echo "  make lint         - 🔍 Run linter"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# 🚀 Run the gateway in development mode
run:
	@echo "🚀 Starting API Gateway..."
	go run main.go

# 🔨 Build the gateway binary
build:
	@echo "🔨 Building API Gateway..."
	go build -o bin/api-gateway main.go
	@echo "✅ Build complete! Binary: bin/api-gateway"

# 📦 Install dependencies
install: deps

# 🔄 Download and tidy dependencies
deps:
	@echo "📦 Installing dependencies..."
	go mod download
	go mod tidy
	@echo "✅ Dependencies installed!"

# 🧹 Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -rf bin/
	rm -f api-gateway
	@echo "✅ Clean complete!"

# 🧪 Run tests
test:
	@echo "🧪 Running tests..."
	go test -v ./...

# 🎨 Format code
fmt:
	@echo "🎨 Formatting code..."
	go fmt ./...
	@echo "✅ Code formatted!"

# 🔍 Run linter (requires golangci-lint)
lint:
	@echo "🔍 Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "⚠️  golangci-lint not installed. Install it from https://golangci-lint.run/"; \
	fi

# 🐳 Build Docker image
docker-build:
	@echo "🐳 Building Docker image..."
	docker build -t gtp-api-gateway:latest .
	@echo "✅ Docker image built: gtp-api-gateway:latest"

# 🐳 Run Docker container
docker-run:
	@echo "🐳 Running Docker container..."
	docker run -p 8000:8000 --env-file .env gtp-api-gateway:latest

# 🔧 Development mode with auto-reload (requires air)
dev:
	@echo "🔧 Starting development mode with auto-reload..."
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "⚠️  air not installed. Install it with: go install github.com/cosmtrek/air@latest"; \
		echo "📝 Falling back to regular run..."; \
		make run; \
	fi

# 📊 Show project statistics
stats:
	@echo "📊 Project Statistics"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "📁 Go files:"
	@find . -name "*.go" | wc -l
	@echo "📝 Lines of code:"
	@find . -name "*.go" -exec wc -l {} + | tail -1
	@echo "📦 Dependencies:"
	@go list -m all | wc -l

# 🏥 Health check
health:
	@echo "🏥 Checking gateway health..."
	@curl -s http://localhost:8000/health | jq . || echo "❌ Gateway not responding"

# 🔐 Setup environment
setup:
	@echo "🔐 Setting up environment..."
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo "✅ Created .env file from .env.example"; \
		echo "📝 Please edit .env with your configuration"; \
	else \
		echo "⚠️  .env file already exists"; \
	fi

