# Quick Start Guide

## 🚨 Fix Docker First

### Option A: Use Docker Desktop (Easiest)
```bash
colima stop && colima delete -f
# Open Docker Desktop app
docker context use desktop-linux
docker info  # Verify it works
```

### Option B: Check Network Issues
```bash
# Are you behind a proxy/VPN?
env | grep -i proxy

# If yes, temporarily disable VPN and try again
```

## 🚀 Once Docker Works

### 1. Start Infrastructure
```bash
make infra-up
```

### 2. Verify Services
```bash
docker ps
# Should see: zookeeper, kafka, postgres, redis, kafka-ui
```

### 3. Access UIs
- Kafka UI: http://localhost:8090
- PostgreSQL: `localhost:5432` (user: `postgres`, pass: `postgres`)
- Redis: `localhost:6379`

## 📝 Configure Services

### Sonar Shell Test
```bash
# Edit sonar-shell-test/.env
GITHUB_PAT=ghp_your_token_here
GITHUB_ORG=your_org_name
```

### Run Sonar Server
```bash
cd sonar-shell-test
./bin/server -server
# Access at http://localhost:8080
```

## 🎯 Run Application Services

```bash
# API Gateway (Port 8080)
make gateway

# Jira Trigger Service (Port 8081)
make jira-trigger

# Chat Agent Service (Port 8082)
make chat-agent

# Approval Service (Port 8083)
make approval

# Service Catalog (Port 8084)
make service-catelog
```

## 🛠️ Useful Commands

```bash
# View all make targets
make help

# View infrastructure logs
make infra-logs

# Stop infrastructure
make infra-down

# Clean everything
make clean

# Run tests
make test

# Tidy dependencies
make tidy
```

## 📂 Important Files

- `CURRENT_STATUS.md` - What's done and what's next
- `SETUP_INSTRUCTIONS.md` - Detailed troubleshooting
- `.env` files - Configure your credentials here
- `Makefile` - All available commands

## ❓ Still Having Issues?

1. **Certificate errors**: See `SETUP_INSTRUCTIONS.md`
2. **Docker not working**: Try Docker Desktop
3. **Services won't start**: Check `.env` files
4. **Port conflicts**: Check what's running on ports 8080-8090

## 📞 Quick Diagnostics

```bash
# Check Docker
docker info
docker context ls

# Check Colima (if using)
colima status

# Check running containers
docker ps

# Check ports
lsof -i :8080
lsof -i :8090
```

## ✅ Success Checklist

- [ ] Docker is working (`docker info` succeeds)
- [ ] Infrastructure is running (`docker ps` shows 5 containers)
- [ ] Kafka UI accessible at http://localhost:8090
- [ ] `.env` files configured with credentials
- [ ] Services start without errors

---

**Current Status**: Environment files created, Docker needs fixing.  
**Next Step**: Fix Docker, then run `make infra-up`

