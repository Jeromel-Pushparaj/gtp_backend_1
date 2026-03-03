# Chat Agent Service - Architecture

## System Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                         FRONTEND APPLICATION                         │
│                    (React, Vue, Angular, etc.)                       │
│                                                                       │
│  User Interface:                                                     │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │  Chat Input: "What is the health status of the backend?"     │  │
│  │  [Send Button]                                                │  │
│  └──────────────────────────────────────────────────────────────┘  │
│                                                                       │
│  JavaScript:                                                         │
│  fetch('http://localhost:8082/api/v1/chat', {                       │
│    method: 'POST',                                                   │
│    body: JSON.stringify({ message: userInput })                     │
│  })                                                                  │
└───────────────────────────┬─────────────────────────────────────────┘
                            │
                            │ HTTP POST
                            │ Content-Type: application/json
                            │ {"message": "..."}
                            ▼
┌─────────────────────────────────────────────────────────────────────┐
│              CHAT AGENT SERVICE (Port 8082)                          │
│                                                                       │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                    HTTP Server Layer                         │   │
│  │  - Handles incoming HTTP requests                            │   │
│  │  - CORS middleware                                            │   │
│  │  - Request validation                                         │   │
│  │  - Response formatting                                        │   │
│  └────────────────────────┬────────────────────────────────────┘   │
│                            │                                          │
│                            ▼                                          │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                    Chat Agent Layer                          │   │
│  │  - Orchestrates conversation flow                            │   │
│  │  - Manages message history                                    │   │
│  │  - Coordinates between Groq and Tools                        │   │
│  │  - Handles multi-turn interactions                           │   │
│  └────────────────────────┬────────────────────────────────────┘   │
│                            │                                          │
│              ┌─────────────┴─────────────┐                          │
│              ▼                           ▼                           │
│  ┌──────────────────────┐    ┌──────────────────────┐              │
│  │   Groq Client        │    │   Tool Executor      │              │
│  │  - API calls         │    │  - Execute tools     │              │
│  │  - Tool calling      │    │  - Backend API calls │              │
│  │  - Response parsing  │    │  - Error handling    │              │
│  └──────────┬───────────┘    └──────────┬───────────┘              │
│             │                            │                           │
└─────────────┼────────────────────────────┼───────────────────────────┘
              │                            │
              │                            │
              ▼                            ▼
┌──────────────────────────┐    ┌──────────────────────────┐
│      Groq API            │    │   GTP Backend API        │
│  (External Service)      │    │   (Port 8080)            │
│                          │    │                          │
│  - LLM Processing        │    │  - GitHub metrics        │
│  - Tool call decisions   │    │  - Jira metrics          │
│  - Response generation   │    │  - SonarCloud metrics    │
└──────────────────────────┘    └──────────────────────────┘
```

## Request Flow

### 1. User Sends Message

```
Frontend → POST /api/v1/chat
{
  "message": "What is the health status of the backend?"
}
```

### 2. HTTP Server Receives Request

```go
// server/http_server.go
func (s *HTTPServer) chatHandler(w http.ResponseWriter, r *http.Request) {
    var req agent.ChatRequest
    json.NewDecoder(r.Body).Decode(&req)
    
    resp, err := s.chatAgent.ProcessMessage(req.Message)
    
    json.NewEncoder(w).Encode(resp)
}
```

### 3. Chat Agent Processes Message

```go
// agent/chat_agent.go
func (a *ChatAgent) ProcessMessage(userMessage string) (*ChatResponse, error) {
    messages := []client.ChatMessage{
        {Role: "system", Content: "You are a helpful assistant..."},
        {Role: "user", Content: userMessage},
    }
    
    // Call Groq with tools
    resp, err := a.groqClient.Chat(messages, a.tools)
    
    // If Groq wants to call tools...
    if resp.FinishReason == "tool_calls" {
        // Execute tools
        for _, toolCall := range toolCalls {
            result := a.toolExecutor.ExecuteTool(toolCall.Name, toolCall.Args)
            messages = append(messages, result)
        }
        
        // Call Groq again with tool results
        resp = a.groqClient.Chat(messages, nil)
    }
    
    return &ChatResponse{Response: resp.Message.Content}
}
```

### 4. Groq Client Calls API

```go
// client/groq_client.go
func (c *GroqClient) Chat(messages []ChatMessage, tools []Tool) (*ChatResponse, error) {
    req := ChatRequest{
        Model: "llama-3.3-70b-versatile",
        Messages: messages,
        Tools: tools,
    }
    
    // POST to https://api.groq.com/openai/v1/chat/completions
    resp := c.httpClient.Do(req)
    
    return resp
}
```

### 5. Groq Decides to Call Tools

```json
{
  "finish_reason": "tool_calls",
  "message": {
    "tool_calls": [
      {
        "id": "call_123",
        "type": "function",
        "function": {
          "name": "health_check",
          "arguments": "{}"
        }
      }
    ]
  }
}
```

### 6. Tool Executor Calls Backend

```go
// client/tool_executor.go
func (t *ToolExecutor) ExecuteTool(toolName string, args map[string]interface{}) (string, error) {
    endpoint, method := t.mapToolToEndpoint(toolName, args)
    
    // GET http://localhost:8080/health
    result := t.makeRequest(method, endpoint, nil, nil)
    
    return result
}
```

### 7. Backend Responds

```json
{
  "status": "healthy",
  "service": "sonar-automation",
  "organization": "teknex-poc"
}
```

### 8. Chat Agent Sends Results to Groq

```go
messages = append(messages, ChatMessage{
    Role: "user",
    Content: "Tool result: {\"status\":\"healthy\",...}"
})

resp = groqClient.Chat(messages, nil)
```

### 9. Groq Generates Final Response

```json
{
  "choices": [
    {
      "message": {
        "content": "The backend is healthy. Service: sonar-automation, Organization: teknex-poc"
      }
    }
  ]
}
```

### 10. Frontend Receives Response

```json
{
  "response": "The backend is healthy. Service: sonar-automation, Organization: teknex-poc"
}
```

## Component Responsibilities

### HTTP Server (`server/http_server.go`)
- Accept HTTP requests
- Validate input
- Handle CORS
- Return responses
- Error handling

### Chat Agent (`agent/chat_agent.go`)
- Orchestrate conversation
- Manage message flow
- Coordinate Groq and tools
- Handle multi-turn interactions

### Groq Client (`client/groq_client.go`)
- Call Groq API
- Handle tool calling
- Parse responses
- Manage API errors

### Tool Executor (`client/tool_executor.go`)
- Map tool names to endpoints
- Execute HTTP requests
- Call backend API
- Return results

### Tools (`client/tools.go`)
- Define available tools
- Specify parameters
- Provide descriptions

## Data Flow

```
User Input
    ↓
HTTP Request
    ↓
Chat Agent
    ↓
Groq API (with tools)
    ↓
Tool Calls Decision
    ↓
Tool Executor
    ↓
Backend API
    ↓
Tool Results
    ↓
Groq API (with results)
    ↓
Final Response
    ↓
HTTP Response
    ↓
User sees answer
```

## Technology Stack

- **Language**: Go 1.21
- **AI Provider**: Groq (https://groq.com)
- **Model**: llama-3.3-70b-versatile
- **Protocol**: HTTP/REST
- **Transport**: JSON
- **CORS**: Enabled for all origins

## Configuration

```env
PORT=8082                              # HTTP server port
GROQ_API_KEY=gsk_...                   # Groq API key
BACKEND_URL=http://localhost:8080      # Backend API URL
BACKEND_API_KEY=                       # Optional backend auth
```

## Scalability Considerations

1. **Stateless Design**: Each request is independent
2. **Horizontal Scaling**: Can run multiple instances
3. **Load Balancing**: Use nginx/traefik
4. **Caching**: Add Redis for common queries
5. **Rate Limiting**: Implement per-user limits

## Security Considerations

1. **API Key Protection**: Never expose in frontend
2. **CORS**: Restrict to specific domains
3. **Authentication**: Add user auth
4. **Rate Limiting**: Prevent abuse
5. **Input Validation**: Sanitize user input

## Monitoring Points

1. **HTTP Server**: Request count, latency, errors
2. **Groq API**: API calls, tokens used, errors
3. **Backend API**: Tool execution success/failure
4. **Overall**: Response time, error rate

## Error Handling

```
User Input → Validation Error → 400 Bad Request
          → Groq API Error → 500 Internal Server Error
          → Backend Error → Tool execution fails → Groq handles gracefully
          → Network Error → Retry logic → Eventually fails
```

## Future Enhancements

1. **Conversation History**: Store past messages
2. **User Authentication**: Add auth layer
3. **More Tools**: Expand tool library
4. **Streaming**: Real-time response streaming
5. **Analytics**: Track usage patterns
6. **Caching**: Cache common responses

