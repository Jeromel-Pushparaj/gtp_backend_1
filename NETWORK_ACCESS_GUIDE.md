# Network Access Guide 🌐

## Your Local IP Address
**10.140.8.28**

## 🚀 Recommended Approach: Use the API Gateway

The **API Gateway** is the single entry point for all services and is the recommended way to expose your application to friends on the local network.

### Step 1: Start the API Gateway
```bash
cd gateway/api-gateway
go run main.go
```

The gateway will start on port **8089** and automatically route requests to all backend services.

### Step 2: Share the Gateway URL with Your Friends

Your friends should use your IP address **10.140.8.28** with the gateway port:

#### 🌟 API Gateway (Port 8089) - Single Entry Point for All Services
```
http://10.140.8.28:8089
```

### Step 3: Available Routes Through the Gateway

All services are accessible through the gateway with these paths:

- **🎫 Jira Service**: `http://10.140.8.28:8089/jira/*`
- **🤖 Chat Agent**: `http://10.140.8.28:8089/chat/*`
- **✅ Approval Service**: `http://10.140.8.28:8089/approval/*`
- **📦 Service Catalog**: `http://10.140.8.28:8089/onboarding/*`
- **📊 Score Card**: `http://10.140.8.28:8089/scorecard/*`
- **🔍 Sonar Shell**: `http://10.140.8.28:8089/sonar/*`

---

## Alternative: Direct Service Access (Not Recommended)

If you need to access individual services directly (without the gateway), you can start them individually:

```bash
# Start individual services:
cd services/chat-agent-service/cmd/server
go run main.go

cd services/approval-service/cmd
go run main.go

# etc...
```

#### Direct Service URLs:
- **Sonar Shell Service** (Port 8080): `http://10.140.8.28:8080`
- **Chat Agent Service** (Port 8082): `http://10.140.8.28:8082`
- **Approval Service** (Port 8083): `http://10.140.8.28:8083`
- **Service Catalog** (Port 8084): `http://10.140.8.28:8084`
- **Score Card Service** (Port 8085): `http://10.140.8.28:8085`
- **Jira Trigger Service** (Port 8086): `http://10.140.8.28:8086`

---

## 🧪 Testing the Connection

### Step 4: Test the Gateway

Your friends can test if they can reach the gateway by opening this in their browser:

```
http://10.140.8.28:8089/health
```

Or using curl:
```bash
curl http://10.140.8.28:8089/health
```

Expected response:
```json
{
  "status": "healthy",
  "service": "api-gateway",
  "version": "1.0.0",
  "message": "Gateway is running smoothly"
}
```

### Example API Calls Through the Gateway

#### 🎫 Create Jira Issue (via Gateway)
```bash
curl -X POST http://10.140.8.28:8089/jira/api/create-issue \
  -H "Content-Type: application/json" \
  -d '{
    "project": "PROJ",
    "summary": "Test issue from network",
    "description": "Testing network access"
  }'
```

#### 🤖 Chat with AI Agent (via Gateway)
```bash
curl -X POST http://10.140.8.28:8089/chat/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Hello from the network!"}'
```

#### 📦 List Services (via Gateway)
```bash
curl http://10.140.8.28:8089/onboarding/api/services
```

#### 🔍 Check Sonar Metrics (via Gateway)
```bash
curl http://10.140.8.28:8089/sonar/health
```

## Firewall Configuration (macOS)

If your friends can't connect, you may need to allow incoming connections:

### Option 1: Allow specific ports
```bash
# Allow port 8089 (API Gateway)
sudo /usr/libexec/ApplicationFirewall/socketfilterfw --add /usr/local/go/bin/go
sudo /usr/libexec/ApplicationFirewall/socketfilterfw --unblockapp /usr/local/go/bin/go
```

### Option 2: Temporarily disable firewall (for testing only)
```bash
sudo /usr/libexec/ApplicationFirewall/socketfilterfw --setglobalstate off
```

To re-enable:
```bash
sudo /usr/libexec/ApplicationFirewall/socketfilterfw --setglobalstate on
```

---

## ⚙️ Gateway Configuration

The gateway is already configured to accept network connections:

- **Host**: `0.0.0.0` (listens on all network interfaces)
- **Port**: `8089` (default, configurable via `GATEWAY_PORT` env variable)
- **CORS**: Enabled for all origins (configurable)
- **Rate Limiting**: 100 requests per minute (configurable)

You can customize these settings by creating a `.env` file in `gateway/api-gateway/`:

```bash
cd gateway/api-gateway
cp .env.example .env
# Edit .env to customize settings
```

---

## 🔒 Important Security Notes

⚠️ **Security Considerations:**
1. This only works on the **same WiFi network** (local network)
2. Your friends must be connected to the same network as you
3. The gateway provides built-in security features:
   - ✅ Rate limiting (prevents abuse)
   - ✅ CORS protection
   - ✅ Request logging
   - ✅ Error handling
4. For production use, add authentication (JWT tokens)

✅ **Why Use the Gateway:**
- **Single entry point**: One URL for all services
- **Security**: Built-in rate limiting and middleware
- **Monitoring**: Centralized logging
- **Flexibility**: Easy to add authentication later

## Troubleshooting

### Friends can't connect?
1. **Check if you're on the same WiFi**: Both devices must be on the same network
2. **Check firewall**: macOS firewall might be blocking connections
3. **Verify service is running**: Make sure the service started successfully
4. **Check the port**: Ensure you're using the correct port number

### How to check if a service is listening on the network:
```bash
netstat -an | grep LISTEN | grep 8089
```

You should see something like:
```
tcp4       0      0  *.8089                 *.*                    LISTEN
```

The `*` means it's listening on all interfaces (good!).
If you see `127.0.0.1`, it's only listening on localhost (bad).

---

## 📱 Quick Start for Friends

**Just share this with your friends:**

> "Hey! I'm running an API Gateway on my local network. You can access all services through:
>
> **Main Gateway**: `http://10.140.8.28:8089`
>
> **Health Check**: `http://10.140.8.28:8089/health`
>
> **Available Services**:
> - Jira: `http://10.140.8.28:8089/jira/*`
> - Chat: `http://10.140.8.28:8089/chat/*`
> - Approval: `http://10.140.8.28:8089/approval/*`
> - Onboarding: `http://10.140.8.28:8089/onboarding/*`
> - ScoreCard: `http://10.140.8.28:8089/scorecard/*`
> - Sonar: `http://10.140.8.28:8089/sonar/*`
>
> Make sure you're connected to the same WiFi as me!"

---

## 🚀 Quick Start Commands

### Start the Gateway
```bash
cd gateway/api-gateway
go run main.go
```

### Check if Gateway is Running
```bash
# From your machine
curl http://localhost:8089/health

# From your friend's machine (on same network)
curl http://10.140.8.28:8089/health
```

### View Gateway Logs
The gateway will show all incoming requests and proxy them to the appropriate backend services.

