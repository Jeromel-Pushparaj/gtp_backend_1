# Troubleshooting Guide

## Issue: AI is Hallucinating Instead of Calling Tools

### Symptoms
- The chatbot responds with made-up information
- No requests are made to the backend API
- Server logs don't show tool execution

### Solution

The latest code has been updated with:
1. **Proper tool call handling** - The `ChatMessage` struct now includes `ToolCalls` field
2. **Correct message format** - Tool results are sent with `role: "tool"` and `tool_call_id`
3. **Detailed logging** - You can now see exactly what's happening

### Steps to Fix

1. **Rebuild the server**:
   ```bash
   cd services/chat-agent-service
   go build -o chat-agent-server cmd/server/main.go
   ```

2. **Run with your Groq API key**:
   ```bash
   GROQ_API_KEY=your_actual_key ./chat-agent-server
   ```

3. **Test and watch the logs**:
   ```bash
   # In another terminal
   ./test_with_logs.sh
   ```

4. **Check the server logs** - You should see:
   ```
   Sending request to Groq API with 10 tools
   Tool choice: auto
   Received response with 1 choices
   Groq response - Finish reason: tool_calls
   Tool calls in message: 1
   Processing 1 tool calls
   Executing tool: health_check with args: {}
   Calling backend: GET http://localhost:8080/health
   Backend call successful, result length: 78
   Making second call to Groq with tool results
   ```

### What Changed

**Before (Broken):**
```go
type ChatMessage struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}
```

**After (Fixed):**
```go
type ChatMessage struct {
    Role       string     `json:"role"`
    Content    string     `json:"content,omitempty"`
    ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
    ToolCallID string     `json:"tool_call_id,omitempty"`
}
```

**Before (Broken):**
```go
// Incorrect message format
messages = append(messages, ChatMessage{
    Role:    "user",
    Content: fmt.Sprintf("Tool result: %s", result),
})
```

**After (Fixed):**
```go
// Correct OpenAI-compatible format
messages = append(messages, ChatMessage{
    Role:       "tool",
    Content:    result,
    ToolCallID: toolCall.ID,
})
```

## Issue: Backend API Not Being Called

### Check 1: Is the backend running?

```bash
curl http://localhost:8080/health
```

Expected response:
```json
{"status":"healthy","service":"sonar-automation","organization":"teknex-poc"}
```

If not running:
```bash
cd sonar-shell-test
go run main.go -server -port=8080
```

### Check 2: Is the backend URL correct?

Check your `.env` file:
```env
BACKEND_URL=http://localhost:8080
```

### Check 3: Are tools being called?

Look for this in server logs:
```
Executing tool: health_check with args: {}
Calling backend: GET http://localhost:8080/health
```

If you don't see this, the issue is with Groq not calling tools.

## Issue: Groq Not Calling Tools

### Possible Causes

1. **Wrong model** - Make sure you're using a model that supports tool calling
   - ✅ `llama-3.3-70b-versatile` (supports tools)
   - ✅ `llama-3.1-70b-versatile` (supports tools)
   - ✅ `llama-3.1-8b-instant` (supports tools)

2. **Invalid API key** - Check your Groq API key
   ```bash
   echo $GROQ_API_KEY
   ```

3. **Tool definitions incorrect** - Check server logs for:
   ```
   Sending request to Groq API with 10 tools
   Tool choice: auto
   ```

### Debug Steps

1. **Enable verbose logging** - Already enabled in the latest code

2. **Check Groq API response**:
   Look for in logs:
   ```
   Groq response - Finish reason: tool_calls
   Tool calls in message: 1
   ```

3. **If finish_reason is "stop" instead of "tool_calls"**:
   - The model decided not to use tools
   - Try a more explicit question: "Use the health_check tool to check the backend status"

## Issue: Server Won't Start

### Error: "GROQ_API_KEY is required"

**Solution**: Set your API key
```bash
export GROQ_API_KEY=gsk_your_actual_key_here
./chat-agent-server
```

Or add to `.env`:
```env
GROQ_API_KEY=gsk_your_actual_key_here
```

### Error: "bind: address already in use"

**Solution**: Port 8082 is already in use
```bash
# Find what's using the port
lsof -i :8082

# Kill it or use a different port
PORT=8083 ./chat-agent-server
```

## Issue: CORS Errors

### Symptoms
- Browser console shows CORS errors
- Requests fail from frontend

### Solution

The server already has CORS enabled for all origins. If you still see errors:

1. **Check the server is running**: `curl http://localhost:8082/health`
2. **Check the URL in frontend**: Make sure it's `http://localhost:8082`
3. **Check browser console**: Look for the specific CORS error

## Testing Checklist

- [ ] Backend API is running on port 8080
- [ ] Chat Agent server is running on port 8082
- [ ] GROQ_API_KEY is set correctly
- [ ] Server logs show "Sending request to Groq API with 10 tools"
- [ ] Server logs show "Tool calls in message: X" when asking about backend
- [ ] Server logs show "Calling backend: GET http://localhost:8080/..."
- [ ] Response includes actual data from backend, not hallucinated info

## Getting Help

If you're still having issues:

1. **Capture the logs**:
   ```bash
   ./chat-agent-server 2>&1 | tee server.log
   ```

2. **Test with curl**:
   ```bash
   curl -X POST http://localhost:8082/api/v1/chat \
     -H "Content-Type: application/json" \
     -d '{"message": "What is the health status of the backend?"}' | jq '.'
   ```

3. **Check the logs** for:
   - Groq API calls
   - Tool call detection
   - Backend API calls
   - Any error messages

4. **Verify the response** contains real data from your backend, not made-up information

