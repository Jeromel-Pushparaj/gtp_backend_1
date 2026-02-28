# Quick Start Guide

## Prerequisites

1. Ensure the GTP Backend API is running (default: http://localhost:8080)
2. Go 1.21 or higher installed

## Setup

1. Navigate to the service directory:
```bash
cd services/chat-agent-service
```

2. Copy the environment file:
```bash
cp .env.example .env
```

3. Edit `.env` and configure:
```env
API_BASE_URL=http://localhost:8080
API_KEY=your_api_key_if_needed
```

4. Download dependencies:
```bash
go mod download
```

5. Build the server:
```bash
make build
```

## Running the MCP Server

Run the server:
```bash
./mcp-server
```

Or run directly without building:
```bash
go run cmd/main.go
```

## Configuring with Claude Desktop

1. Open Claude Desktop settings
2. Click "Developer" tab
3. Click "Edit Config"
4. Add the following to `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "gtp-backend": {
      "command": "/full/path/to/services/chat-agent-service/mcp-server",
      "args": [],
      "env": {
        "API_BASE_URL": "http://localhost:8080",
        "API_KEY": ""
      }
    }
  }
}
```

5. Restart Claude Desktop

## Configuring with Cursor IDE

1. Open Cursor settings
2. Find "MCP" section
3. Click "Add new global MCP Server"
4. Add the same configuration as above to `mcp.json`
5. Restart Cursor

## Testing

Once configured, you can ask Claude or Cursor questions like:

- "Check the health of the GTP backend service"
- "List all pull requests for repository X"
- "Get GitHub metrics for repository Y"
- "Show me open bugs in Jira project Z"
- "Fetch SonarCloud metrics for repository A"

The AI assistant will use the MCP tools to interact with your GTP Backend API.

## Troubleshooting

If the server doesn't work:

1. Check that the API_BASE_URL is correct
2. Verify the GTP Backend API is running
3. Check if API_KEY is required and set correctly
4. Look at the logs when running the server
5. Ensure the binary path in the config is absolute

## Available Commands

```bash
make build    # Build the MCP server
make run      # Run the MCP server
make clean    # Clean build artifacts
make deps     # Download dependencies
make test     # Run tests
```

