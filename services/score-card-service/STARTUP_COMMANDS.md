# 🚀 Service Startup Commands

## Why Docker is Needed

The score-card service has these dependencies:

1. **PostgreSQL** (Port 5432) - For storing scorecard history (v1 API only)
2. **Kafka** (Port 9092) - For event-driven communication between services
3. **Redis** (Port 6379) - For caching
4. **Metrics API** (Port 8080) - External API that provides GitHub/Jira/SonarCloud metrics

**Important:** The v2 API endpoints (used in the curl commands) **do NOT require PostgreSQL, Kafka, or Redis**. They only need:

- The score-card service itself (Port 8085)
- The metrics API (Port 8080) to fetch GitHub/Jira/SonarCloud data

However, the full infrastructure is started for completeness and future features.

## Prerequisites

Ensure Docker is running before starting services.

---

## 🎯 Option A: Minimal Startup (No Docker Required)

If you only want to test the v2 API curl commands and don't need database/Kafka/Redis:

### Step 1: Start Metrics API (sonar-shell-test)

```bash
cd sonar-shell-test
./bin/server -server -port 8080
```

### Step 2: Start Score Card Service

```bash
# In a new terminal
cd /Users/srilekhar/Desktop/get_to_prod/backend1_v1/gtp_backend_1
make score-card
```

### Step 3: Test with Curl Commands

Now you can run all 5 curl commands from the [Test with Curl Commands](#step-6-test-with-curl-commands) section below.

**Note:** The service will log warnings about database/Kafka/Redis not being available, but the v2 API endpoints will work fine.

---

## 🏗️ Option B: Full Startup (With Docker Infrastructure)

Use this option if you need the complete infrastructure (database, Kafka, Redis).

## Step 1: Start Docker

```bash
# Option A: Start Docker Desktop (Recommended)
open -a Docker

# Option B: Start Colima (if using Colima)
colima start

# Verify Docker is running
docker ps
```

## Step 2: Start Infrastructure Services

```bash
# Start Kafka, PostgreSQL, Redis, Zookeeper
make infra-up
```

**Wait 30 seconds for services to initialize**

## Step 3: Verify Infrastructure

```bash
# Check all containers are running
docker ps

# Expected containers:
# - zookeeper
# - kafka
# - kafka-ui
# - postgres
# - redis
```

## Step 4: Start Application Services

### Terminal 1 - Score Card Service (Port 8085)

```bash
cd /Users/srilekhar/Desktop/get_to_prod/backend1_v1/gtp_backend_1
make score-card
```

### Terminal 2 - API Gateway (Port 8080) [Optional]

```bash
cd /Users/srilekhar/Desktop/get_to_prod/backend1_v1/gtp_backend_1
make gateway
```

### Terminal 3 - Jira Trigger Service (Port 8081) [Optional]

```bash
cd /Users/srilekhar/Desktop/get_to_prod/backend1_v1/gtp_backend_1
make jira-trigger
```

### Terminal 4 - Chat Agent Service (Port 8082) [Optional]

```bash
cd /Users/srilekhar/Desktop/get_to_prod/backend1_v1/gtp_backend_1
make chat-agent
```

### Terminal 5 - Approval Service (Port 8083) [Optional]

```bash
cd /Users/srilekhar/Desktop/get_to_prod/backend1_v1/gtp_backend_1
make approval
```

### Terminal 6 - Service Catalog (Port 8084) [Optional]

```bash
cd /Users/srilekhar/Desktop/get_to_prod/backend1_v1/gtp_backend_1
make service-catelog
```

## Step 5: Verify Services Are Running

### Health Check - Score Card Service

```bash
curl http://localhost:8085/health
```

### Health Check - API Gateway

```bash
curl http://localhost:8080/health
```

## Step 6: Test with Curl Commands

### 1️⃣ Code Quality Scorecard

```bash
curl "http://localhost:8085/api/v2/scorecards/code-quality?service_name=delivery-management-frontend" | jq '.'
```

**Show only achieved level metrics:**

```bash
curl "http://localhost:8085/api/v2/scorecards/code-quality?service_name=delivery-management-frontend&only_achieved_level=true" | jq '.'
```

### 2️⃣ Service Health Scorecard

```bash
curl "http://localhost:8085/api/v2/scorecards/service-health?service_name=delivery-management-frontend" | jq '.'
```

### 3️⃣ Security Maturity Scorecard

```bash
curl "http://localhost:8085/api/v2/scorecards/security-maturity?service_name=delivery-management-frontend" | jq '.'
```

### 4️⃣ Production Readiness Scorecard

```bash
curl "http://localhost:8085/api/v2/scorecards/production-readiness?service_name=delivery-management-frontend" | jq '.'
```

### 5️⃣ PR Metrics Scorecard

```bash
curl "http://localhost:8085/api/v2/scorecards/pr-metrics?service_name=delivery-management-frontend" | jq '.'
```

### All 5 Scorecards Summary

```bash
echo "=== Scorecard Levels for delivery-management-frontend ==="
echo ""
echo "1. Code Quality:         $(curl -s "http://localhost:8085/api/v2/scorecards/code-quality?service_name=delivery-management-frontend" | jq -r '.evaluation.achieved_level_name')"
echo "2. Service Health:       $(curl -s "http://localhost:8085/api/v2/scorecards/service-health?service_name=delivery-management-frontend" | jq -r '.evaluation.achieved_level_name')"
echo "3. Security Maturity:    $(curl -s "http://localhost:8085/api/v2/scorecards/security-maturity?service_name=delivery-management-frontend" | jq -r '.evaluation.achieved_level_name')"
echo "4. Production Readiness: $(curl -s "http://localhost:8085/api/v2/scorecards/production-readiness?service_name=delivery-management-frontend" | jq -r '.evaluation.achieved_level_name')"
echo "5. PR Metrics:           $(curl -s "http://localhost:8085/api/v2/scorecards/pr-metrics?service_name=delivery-management-frontend" | jq -r '.evaluation.achieved_level_name')"
echo ""
```

## Infrastructure Access

- **Kafka UI**: http://localhost:8090
- **PostgreSQL**: localhost:5432 (user: postgres, pass: postgres)
- **Redis**: localhost:6379
- **Kafka**: localhost:9092

## Shutdown Commands

### Stop Application Services

```bash
# Press Ctrl+C in each terminal running a service
```

### Stop Infrastructure

```bash
make infra-down
```

### Clean Everything (Remove volumes)

```bash
make clean
```

## Troubleshooting

### Docker not running

```bash
# Check Docker status
docker info

# If error, start Docker Desktop or Colima
open -a Docker
```

### Port already in use

```bash
# Find process using port (e.g., 8085)
lsof -i :8085

# Kill process
kill -9 <PID>
```

### Infrastructure not starting

```bash
# View logs
make infra-logs

# Restart infrastructure
make infra-restart
```
