# Why the MCP Server Behaves This Way

## The Issue You Saw

When you ran `./mcp-server`, you saw:
```
MCP Server starting with base URL: http://localhost:8080
Server ready and waiting for stdio input from MCP client...
```

Then it appeared to "hang" or "stop running" - but this is **CORRECT BEHAVIOR**.

## Why This Happens

### MCP Servers Use stdio Communication

MCP (Model Context Protocol) servers communicate via **stdio** (standard input/output), not HTTP:

```
┌─────────────────┐
│  Claude Desktop │
│   (MCP Client)  │
└────────┬────────┘
         │
         │ stdio (pipes)
         │ stdin → sends commands
         │ stdout ← receives responses
         │
┌────────▼────────┐
│   MCP Server    │
│  (Your binary)  │
└─────────────────┘
```

### What "Waiting for stdio input" Means

The server is:
1. Running (process is alive)
2. Listening on stdin (standard input)
3. Waiting for commands from an MCP client
4. Ready to respond on stdout (standard output)

**It's NOT:**
- Crashed
- Frozen
- Broken
- Stopped

## How to Verify It's Working

### Test 1: Check if Process is Running

```bash
cd services/chat-agent-service
./mcp-server &
ps aux | grep mcp-server
```

Output should show:
```
jeromelp  49996  0.0  0.1  411836656  10976  ??  SN  7:47AM  0:00.02 ./mcp-server
```

The process is RUNNING.

### Test 2: Check Backend Connection

```bash
cd services/chat-agent-service
./test_connection.sh
```

All 3 tests should pass.

### Test 3: Let Claude Manage It

Don't run it manually. Configure Claude Desktop and let Claude start/stop it automatically.

## The Fix Applied

I updated the code to make it clearer:

**Before:**
```go
if err := server.Serve(); err != nil {
    log.Fatalf("Server error: %v", err)
}
// Program exits here
```

**After:**
```go
log.Println("Server ready and waiting for stdio input from MCP client...")

if err := server.Serve(); err != nil {
    log.Fatalf("Server error: %v", err)
}

select {} // Block forever, waiting for stdio input
```

Now the server:
1. Prints a clear message
2. Starts the stdio listener
3. Blocks forever (using `select {}`)
4. Waits for Claude to send commands

## How MCP Communication Works

### Step-by-Step Flow

1. **You ask Claude a question:**
   ```
   "Check the health of the GTP backend"
   ```

2. **Claude starts the MCP server:**
   ```bash
   /path/to/mcp-server
   ```

3. **Server starts and waits:**
   ```
   MCP Server starting...
   Server ready and waiting for stdio input...
   (waiting on stdin)
   ```

4. **Claude sends command via stdin:**
   ```json
   {"jsonrpc":"2.0","method":"tools/call","params":{"name":"health_check"}}
   ```

5. **Server processes and responds via stdout:**
   ```json
   {"jsonrpc":"2.0","result":{"content":[{"type":"text","text":"..."}]}}
   ```

6. **Claude receives response and shows you:**
   ```
   The backend is healthy. Status: healthy, Service: sonar-automation
   ```

7. **When done, Claude stops the server**

## Why You Should NOT Run It Manually

### Running Manually:
```bash
./mcp-server
# Output:
# MCP Server starting...
# Server ready and waiting for stdio input...
# (appears to hang - this is normal!)
```

**What's happening:**
- Server is waiting for stdin input
- No input comes (you're not sending MCP protocol messages)
- Server keeps waiting (correct behavior)
- Looks like it's "stuck" (but it's not)

### Running via Claude:
```
Claude Desktop → starts mcp-server → sends commands → gets responses → stops server
```

**What's happening:**
- Claude manages the entire lifecycle
- Sends proper MCP protocol messages
- Gets responses
- Everything works automatically

## Current Status

### Default Base URL Issue

I noticed the default was `http://localhost:8093` but your backend runs on `8080`.

**Fixed in the code:**
```go
baseURL := flag.String("base-url", getEnv("API_BASE_URL", "http://localhost:8080"), ...)
```

Now defaults to port 8080.

### How to Override

**Option 1: Environment variable**
```bash
export API_BASE_URL=http://localhost:8080
./mcp-server
```

**Option 2: Command line flag**
```bash
./mcp-server -base-url=http://localhost:8080
```

**Option 3: .env file**
```bash
cd services/chat-agent-service
echo "API_BASE_URL=http://localhost:8080" > .env
./mcp-server
```

**Option 4: Claude config (recommended)**
```json
{
  "mcpServers": {
    "gtp-backend": {
      "command": "/path/to/mcp-server",
      "env": {
        "API_BASE_URL": "http://localhost:8080"
      }
    }
  }
}
```

## Summary

### The MCP Server IS Working

- Process runs and stays alive ✓
- Waits for stdio input ✓
- Ready to receive commands ✓
- Can connect to backend API ✓

### What Changed

1. Fixed default port from 8093 to 8080
2. Added `select {}` to keep server running
3. Added clearer log message

### Next Steps

1. Don't run the server manually
2. Configure Claude Desktop with the correct path
3. Set `API_BASE_URL=http://localhost:8080` in Claude config
4. Restart Claude Desktop
5. Ask Claude to check the backend health

The server will start automatically when Claude needs it!

