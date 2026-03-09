# GTP Backend Setup Instructions

## Current Issue

You're experiencing TLS certificate verification errors with Colima. This is typically caused by:
1. Corporate proxy/firewall intercepting HTTPS traffic
2. VPN interfering with certificate validation
3. Colima VM certificate store issues

## Solution Options

### Option 1: Use Docker Desktop (Recommended)

Docker Desktop is more reliable and handles certificates better on macOS.

```bash
# 1. Stop Colima
colima stop
colima delete -f

# 2. Install Docker Desktop (if not already installed)
# Download from: https://www.docker.com/products/docker-desktop

# 3. Start Docker Desktop from Applications

# 4. Switch Docker context
docker context use desktop-linux

# 5. Verify Docker is working
docker info

# 6. Start infrastructure
make infra-up
```

### Option 2: Fix Colima Certificate Issues

If you must use Colima:

```bash
# 1. Check if you're behind a proxy
echo $HTTP_PROXY
echo $HTTPS_PROXY

# 2. If proxy is set, you need to configure Colima to use it
# OR temporarily disable VPN/proxy

# 3. Restart Colima
colima stop
colima delete -f
colima start --cpu 4 --memory 8 --disk 60

# 4. Try pulling images
docker pull redis:7-alpine
```

### Option 3: Run Services Without Docker (Development Mode)

You can run services locally without Docker:

```bash
# Install dependencies locally
brew install postgresql@15
brew install redis
brew install kafka

# Start services
brew services start postgresql@15
brew services start redis

# Update service .env files to point to localhost
# Then run Go services directly
make jira-trigger
make chat-agent
# etc.
```

### Option 4: Use Pre-existing Infrastructure

If you have access to remote Kafka/PostgreSQL/Redis:

```bash
# Skip docker-compose
# Update .env files in each service to point to remote infrastructure

# Example in services/jira-trigger-service/.env:
DB_HOST=your-remote-db.com
DB_PORT=5432
KAFKA_BROKERS=your-kafka.com:9092
REDIS_HOST=your-redis.com
```

## Quick Start (After Fixing Docker)

Once Docker is working:

```bash
# 1. Run the automated setup
./setup.sh

# 2. Update environment files with your credentials
# Edit these files:
- sonar-shell-test/.env (add GITHUB_PAT and GITHUB_ORG)
- services/*/. env (add service-specific credentials)

# 3. Start a service
make gateway        # API Gateway on port 8080
make jira-trigger   # Jira Service on port 8081
```

## Verify Infrastructure is Running

```bash
# Check running containers
docker ps

# Check logs
make infra-logs

# Access services:
# - Kafka UI: http://localhost:8090
# - PostgreSQL: localhost:5432 (user: postgres, pass: postgres)
# - Redis: localhost:6379
```

## Troubleshooting

### Still getting certificate errors?

You're likely behind a corporate proxy. Ask your IT department for:
1. Proxy CA certificate
2. Proxy configuration details

Then configure Docker/Colima to use the proxy.

### Can't use Docker at all?

Run services in development mode without Docker (Option 3 above).

## Next Steps

1. Choose one of the options above
2. Get Docker working OR skip Docker entirely
3. Update .env files with credentials
4. Run services using `make` commands

## Need Help?

- Check Docker context: `docker context ls`
- Check Docker info: `docker info`
- Check Colima status: `colima status`
- Check for proxy: `env | grep -i proxy`

