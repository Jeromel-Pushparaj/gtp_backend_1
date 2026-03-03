.PHONY: help build build-http build-mcp run run-http run-mcp clean deps test

help:
	@echo "Available commands:"
	@echo "  make build        - Build both HTTP server and MCP server"
	@echo "  make build-http   - Build the HTTP server (for frontend integration)"
	@echo "  make build-mcp    - Build the MCP server (for AI tools)"
	@echo "  make run-http     - Run the HTTP server on port 8082"
	@echo "  make run-mcp      - Run the MCP server (stdio mode)"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make deps         - Download dependencies"
	@echo "  make test         - Run tests"

build: build-http build-mcp

build-http:
	@echo "Building HTTP server..."
	go build -o chat-agent-server cmd/server/main.go
	@echo "Build complete: chat-agent-server"

build-mcp:
	@echo "Building MCP server..."
	go build -o mcp-server cmd/main.go
	@echo "Build complete: mcp-server"

run-http:
	@echo "Running HTTP server on port 8082..."
	go run cmd/server/main.go

run-mcp:
	@echo "Running MCP server (stdio mode)..."
	go run cmd/main.go

clean:
	@echo "Cleaning build artifacts..."
	rm -f mcp-server chat-agent-server
	@echo "Clean complete"

deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy
	@echo "Dependencies updated"

test:
	@echo "Running tests..."
	go test ./...

