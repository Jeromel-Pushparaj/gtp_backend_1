# MCP Client Setup Instructions

## Claude Desktop Setup

### Step 1: Locate Configuration File

On macOS:
```bash
~/Library/Application Support/Claude/claude_desktop_config.json
```

On Windows:
```
%APPDATA%\Claude\claude_desktop_config.json
```

On Linux:
```bash
~/.config/Claude/claude_desktop_config.json
```

### Step 2: Edit Configuration

Open the file and add the MCP server configuration:

```json
{
  "mcpServers": {
    "gtp-backend": {
      "command": "/Users/jeromelp/work/gtp_backend_1/services/chat-agent-service/mcp-server",
      "args": [],
      "env": {
        "API_BASE_URL": "http://localhost:8080",
        "API_KEY": ""
      }
    }
  }
}
```

Replace `/Users/jeromelp/work/gtp_backend_1` with your actual project path.

### Step 3: Restart Claude Desktop

Close and reopen Claude Desktop for the changes to take effect.

### Step 4: Verify

In Claude, you should see a small tool icon indicating MCP tools are available. Try asking:
```
Check the health of the GTP backend service
```

## Cursor IDE Setup

### Step 1: Open Settings

1. Open Cursor IDE
2. Click the settings icon (gear icon)
3. Look for "MCP" section
4. Click "Add new global MCP Server"

### Step 2: Add Configuration

Cursor will open `mcp.json`. Add:

```json
{
  "mcpServers": {
    "gtp-backend": {
      "command": "/Users/jeromelp/work/gtp_backend_1/services/chat-agent-service/mcp-server",
      "args": [],
      "env": {
        "API_BASE_URL": "http://localhost:8080",
        "API_KEY": ""
      }
    }
  }
}
```

### Step 3: Restart Cursor

Close and reopen Cursor IDE.

### Step 4: Verify

Open the chat panel and try asking:
```
List all repositories in the organization
```

## VS Code with MCP Extension

If using VS Code with MCP support:

### Step 1: Install MCP Extension

Search for "Model Context Protocol" in VS Code extensions and install.

### Step 2: Configure

Open VS Code settings and add to `settings.json`:

```json
{
  "mcp.servers": {
    "gtp-backend": {
      "command": "/Users/jeromelp/work/gtp_backend_1/services/chat-agent-service/mcp-server",
      "args": [],
      "env": {
        "API_BASE_URL": "http://localhost:8080",
        "API_KEY": ""
      }
    }
  }
}
```

## Troubleshooting

### Server Not Found

If the client can't find the server:

1. Verify the path is absolute (not relative)
2. Check file permissions: `chmod +x mcp-server`
3. Test the server manually: `./mcp-server`

### Connection Issues

If the server starts but doesn't respond:

1. Check that the GTP Backend API is running on port 8080
2. Verify API_BASE_URL is correct
3. Check if API_KEY is required and set correctly

### No Tools Available

If the client connects but shows no tools:

1. Check the client logs for errors
2. Verify the MCP server starts without errors
3. Try rebuilding: `cd services/chat-agent-service && make build`

### Testing the Server

You can test the server manually:

```bash
cd services/chat-agent-service
./mcp-server
```

The server should start and wait for input. Press Ctrl+C to stop.

## Environment Variables

You can override settings using environment variables:

```bash
API_BASE_URL=http://production-api.example.com:8080 \
API_KEY=your-secret-key \
./mcp-server
```

Or set them in the client configuration's `env` section.

## Multiple Environments

You can configure multiple MCP servers for different environments:

```json
{
  "mcpServers": {
    "gtp-backend-dev": {
      "command": "/path/to/mcp-server",
      "env": {
        "API_BASE_URL": "http://localhost:8080"
      }
    },
    "gtp-backend-prod": {
      "command": "/path/to/mcp-server",
      "env": {
        "API_BASE_URL": "https://api.production.com",
        "API_KEY": "prod-key"
      }
    }
  }
}
```

## Getting Help

If you encounter issues:

1. Check the server logs
2. Verify the GTP Backend API is accessible
3. Review the QUICKSTART.md guide
4. Check TOOLS.md for available tools
5. Consult the README.md for detailed documentation

