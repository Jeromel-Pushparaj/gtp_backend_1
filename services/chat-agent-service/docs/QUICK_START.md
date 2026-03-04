# Quick Start - Chat Agent Service for Frontend

## What You Have Now

✅ **HTTP Server** running on port 8082  
✅ **Groq AI Integration** for intelligent responses  
✅ **Tool Calling** to access your GTP Backend API  
✅ **CORS Enabled** for frontend integration  
✅ **Example Frontend** ready to use  

## 3-Step Setup

### Step 1: Get a Groq API Key (2 minutes)

1. Go to https://console.groq.com
2. Sign up (it's free!)
3. Click "API Keys" in the sidebar
4. Click "Create API Key"
5. Copy the key

### Step 2: Configure and Run (1 minute)

```bash
cd services/chat-agent-service

# Create .env file
cp .env.example .env

# Edit .env and paste your Groq API key
# Change this line:
# GROQ_API_KEY=your_groq_api_key_here
# To:
# GROQ_API_KEY=gsk_your_actual_key_here

# Build and run
make build-http
./chat-agent-server
```

You should see:

```
🤖 Chat Agent Service
=====================
Port: 8082
Backend URL: http://localhost:8080
Groq API Key: gsk_...

🚀 Chat Agent Service starting on port 8082
📡 Endpoints:
   GET  /health           - Health check
   POST /api/v1/chat      - Chat with AI agent
```

### Step 3: Test It (30 seconds)

**Option A: Use the example frontend**

```bash
# Open the example HTML file in your browser
open examples/frontend-example.html
```

**Option B: Use curl**

```bash
curl -X POST http://localhost:8082/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "What is the health status of the backend?"}'
```

**Option C: Use the test script**

```bash
./test_http_server.sh
```

## Integrate with Your Frontend

### JavaScript/Fetch

```javascript
async function chat(message) {
  const response = await fetch('http://localhost:8082/api/v1/chat', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ message }),
  });
  
  const data = await response.json();
  return data.response;
}

// Usage
const answer = await chat('List all repositories');
console.log(answer);
```

### React

```jsx
const [response, setResponse] = useState('');

const handleChat = async (message) => {
  const res = await fetch('http://localhost:8082/api/v1/chat', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ message }),
  });
  
  const data = await res.json();
  setResponse(data.response);
};
```

### Vue

```vue
<script setup>
import { ref } from 'vue';

const response = ref('');

const chat = async (message) => {
  const res = await fetch('http://localhost:8082/api/v1/chat', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ message }),
  });
  
  const data = await res.json();
  response.value = data.response;
};
</script>
```

## Example Questions to Try

- "What is the health status of the backend?"
- "List all organizations"
- "Show me the organization members"
- "List all teams"
- "Check if the backend repository has a README"

## What the Bot Can Do

The chatbot has access to these tools:

✅ Health checks  
✅ List pull requests  
✅ List commits  
✅ List issues  
✅ Check README files  
✅ List branches  
✅ List organization members  
✅ List organization teams  
✅ Fetch organizations  
✅ Fetch repositories  

## Troubleshooting

### "Server is not running"

Make sure you started the server:

```bash
cd services/chat-agent-service
./chat-agent-server
```

### "GROQ_API_KEY is required"

Add your API key to `.env`:

```env
GROQ_API_KEY=gsk_your_actual_key_here
```

### "Connection refused"

Make sure the backend API is running:

```bash
curl http://localhost:8080/health
```

If not, start it:

```bash
cd sonar-shell-test
go run main.go -server -port=8080
```

### CORS errors in browser

The server is already configured for CORS. If you still see errors, check:
- Server is running on port 8082
- You're using the correct URL in your frontend
- Browser console for specific error messages

## File Structure

```
services/chat-agent-service/
├── cmd/
│   ├── main.go              # MCP server (stdio mode)
│   └── server/
│       └── main.go          # HTTP server (for frontend)
├── client/
│   ├── groq_client.go       # Groq API integration
│   ├── tool_executor.go     # Tool execution logic
│   └── tools.go             # Tool definitions
├── agent/
│   └── chat_agent.go        # Chat orchestration
├── server/
│   └── http_server.go       # HTTP server
├── examples/
│   ├── frontend-example.html # Working demo
│   └── README.md            # Integration examples
├── docs/
│   └── HTTP_SERVER_GUIDE.md # Detailed documentation
├── .env.example             # Configuration template
├── Makefile                 # Build commands
└── QUICK_START.md          # This file
```

## Next Steps

1. ✅ You've set up the service
2. ✅ You've tested it works
3. 📝 Now integrate it with your frontend
4. 🚀 Deploy to production (see DEPLOYMENT_GUIDE.md)

## Need Help?

- **Detailed docs**: See `docs/HTTP_SERVER_GUIDE.md`
- **Examples**: Check `examples/README.md`
- **Deployment**: Read `DEPLOYMENT_GUIDE.md`
- **Test script**: Run `./test_http_server.sh`

## Summary

You now have a fully functional chatbot backend that:
- Runs on port 8082
- Uses Groq AI for intelligent responses
- Can access your GTP Backend API
- Is ready for frontend integration
- Has CORS enabled
- Includes working examples

**Just add your Groq API key and you're ready to go!** 🚀

