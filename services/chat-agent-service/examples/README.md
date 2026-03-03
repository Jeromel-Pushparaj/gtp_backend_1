# Frontend Integration Examples

This directory contains examples of how to integrate the Chat Agent Service with your frontend application.

## Quick Start

### 1. Start the Backend API

```bash
cd sonar-shell-test
go run main.go -server -port=8080
```

### 2. Start the Chat Agent Service

```bash
cd services/chat-agent-service

# Set your Groq API key
export GROQ_API_KEY=your_groq_api_key_here

# Run the server
make run-http
```

### 3. Open the Example Frontend

Simply open `frontend-example.html` in your web browser:

```bash
open examples/frontend-example.html
```

Or use a local web server:

```bash
cd examples
python3 -m http.server 3000
# Then open http://localhost:3000/frontend-example.html
```

## Example Queries

Try these questions in the chat interface:

- "What is the health status of the backend?"
- "List all organizations"
- "Show me the organization members"
- "List all teams in the organization"
- "Check if the backend repository has a README"

## Integration Examples

### JavaScript/TypeScript

```javascript
async function chat(message) {
  const response = await fetch('http://localhost:8082/api/v1/chat', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ message }),
  });
  
  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }
  
  const data = await response.json();
  return data.response;
}

// Usage
const answer = await chat('What is the health status?');
console.log(answer);
```

### React Component

```jsx
import { useState } from 'react';

function ChatBot() {
  const [message, setMessage] = useState('');
  const [messages, setMessages] = useState([]);
  const [loading, setLoading] = useState(false);

  const sendMessage = async () => {
    if (!message.trim()) return;

    const userMessage = { role: 'user', content: message };
    setMessages([...messages, userMessage]);
    setMessage('');
    setLoading(true);

    try {
      const response = await fetch('http://localhost:8082/api/v1/chat', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ message }),
      });

      const data = await response.json();
      const botMessage = { role: 'bot', content: data.response };
      setMessages([...messages, userMessage, botMessage]);
    } catch (error) {
      console.error('Error:', error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      <div className="messages">
        {messages.map((msg, i) => (
          <div key={i} className={msg.role}>
            {msg.content}
          </div>
        ))}
      </div>
      <input
        value={message}
        onChange={(e) => setMessage(e.target.value)}
        onKeyPress={(e) => e.key === 'Enter' && sendMessage()}
      />
      <button onClick={sendMessage} disabled={loading}>
        Send
      </button>
    </div>
  );
}
```

### Vue.js Component

```vue
<template>
  <div class="chat-bot">
    <div class="messages">
      <div v-for="(msg, i) in messages" :key="i" :class="msg.role">
        {{ msg.content }}
      </div>
    </div>
    <input
      v-model="message"
      @keyup.enter="sendMessage"
      placeholder="Type your message..."
    />
    <button @click="sendMessage" :disabled="loading">Send</button>
  </div>
</template>

<script>
export default {
  data() {
    return {
      message: '',
      messages: [],
      loading: false,
    };
  },
  methods: {
    async sendMessage() {
      if (!this.message.trim()) return;

      this.messages.push({ role: 'user', content: this.message });
      const userMessage = this.message;
      this.message = '';
      this.loading = true;

      try {
        const response = await fetch('http://localhost:8082/api/v1/chat', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ message: userMessage }),
        });

        const data = await response.json();
        this.messages.push({ role: 'bot', content: data.response });
      } catch (error) {
        console.error('Error:', error);
      } finally {
        this.loading = false;
      }
    },
  },
};
</script>
```

## API Reference

### POST /api/v1/chat

Send a message to the chatbot.

**Request:**
```json
{
  "message": "Your question here"
}
```

**Response:**
```json
{
  "response": "The bot's answer"
}
```

**Error Response:**
```json
{
  "error": "Error message"
}
```

## CORS Configuration

The server is configured to allow CORS from all origins (`*`). For production, you should restrict this to your specific domain by modifying the `enableCORS` function in `server/http_server.go`.

## Production Considerations

1. **API Key Security**: Never expose your Groq API key in frontend code
2. **Rate Limiting**: Implement rate limiting to prevent abuse
3. **Authentication**: Add user authentication to your chat endpoint
4. **CORS**: Restrict CORS to your specific domain
5. **Error Handling**: Implement proper error handling and user feedback
6. **Logging**: Add logging for monitoring and debugging
7. **Caching**: Consider caching responses for common queries

