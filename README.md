# gtp_backend_1

Great — since you're building a **multi-service API platform** (Jira trigger, Chat agent, Approval via Slack, Onboarding service, etc.), you should structure this like a **platform**, not like a normal single backend.

Think in terms of:

👉 Scalability
👉 Ownership (each person owns a service)
👉 Versioning
👉 Independent deployment

---

# 🚀 Quick Start

## Start the API Gateway (Recommended)

The easiest way to run the entire platform is through the API Gateway:

```bash
# Option 1: Use the startup script
./start-gateway.sh

# Option 2: Manual start
cd gateway/api-gateway
go run main.go
```

The gateway will be available at:
- **Local**: `http://localhost:8089`
- **Network**: `http://<your-ip>:8089` (for friends on same WiFi)

### Access All Services Through Gateway

- 🎫 **Jira**: `http://localhost:8089/jira/*`
- 🤖 **Chat**: `http://localhost:8089/chat/*`
- ✅ **Approval**: `http://localhost:8089/approval/*`
- 📦 **Onboarding**: `http://localhost:8089/onboarding/*`
- 📊 **ScoreCard**: `http://localhost:8089/scorecard/*`
- 🔍 **Sonar**: `http://localhost:8089/sonar/*`

### Network Access for Friends

See [NETWORK_ACCESS_GUIDE.md](./NETWORK_ACCESS_GUIDE.md) for detailed instructions on exposing the gateway to friends on your local network.

---

# 🧱 1. High-Level Architecture Approach

Use a **Modular Monorepo (Service-Oriented Structure)**

Instead of:

❌ One big backend
OR
❌ Fully separate repos (chaos in early stage)

Use:

✅ One repo
➡️ Multiple services inside
➡️ Shared contracts
➡️ Versioned APIs

---

# 📁 2. Recommended Folder Structure

```
platform-root/
│
├── services/
│   ├── jira-trigger-service/
│   │   ├── cmd/
│   │   ├── internal/
│   │   ├── api/
│   │   │   └── v1/
│   │   ├── kafka/
│   │   ├── domain/
│   │   └── Dockerfile
│   │
│   ├── chat-agent-service/
│   │   ├── api/v1/
│   │   ├── agent/
│   │   ├── orchestrator/
│   │   └── Dockerfile
│   │
│   ├── approval-service/
│   │   ├── api/v1/
│   │   ├── slack/
│   │   ├── workflow/
│   │   └── Dockerfile
│   │
│   └── onboarding-service/
│       ├── api/v1/
│       ├── business/
│       └── Dockerfile
│
├── shared/
│   ├── contracts/        # Event schemas
│   ├── middleware/
│   ├── auth/
│   └── utils/
│
├── gateway/
│   └── api-gateway/
│
├── infra/
│   ├── kafka/
│   ├── docker/
│   └── terraform/
│
├── docs/
│   └── openapi/
│
└── Makefile
```

---

# 🧠 3. Why This Works

| Need                      | Solved By                   |
| ------------------------- | --------------------------- |
| Independent development   | Each service self-contained |
| Kafka async flow          | kafka/ inside service       |
| API versioning            | api/v1, api/v2              |
| Reuse logic               | shared/                     |
| Future microservice split | Easy                        |
| CI/CD                     | Per service deploy          |

---

# 🔖 4. API Versioning Strategy

Inside every service:

```
api/
 ├── v1/
 │   ├── handler.go
 │   ├── routes.go
 │
 └── v2/ (future)
```

Routing example:

```
/jira/v1/create-trigger
/chat/v1/respond
/approval/v1/request
/onboard/v1/register
```

---

# 🔄 5. Version Control Strategy (Git)

## Branching Model

Use **Service-Based Feature Branching**

```
main
develop

feature/jira-trigger
feature/chat-agent
feature/approval-flow
feature/onboarding-logic
```

---

## Ownership Mapping

| Service      | Owner     |
| ------------ | --------- |
| Jira Trigger | Keerthana |
| Chat Agent   | Jeromel   |
| Approval     | Sarumathi |
| Onboarding   | Rashmi    |

Each works ONLY inside:

```
services/<their-service>/
```

---

# 🧩 6. Shared Contracts (Very Important)

Since Kafka is involved:

Put all event schemas here:

```
shared/contracts/

approval_requested.json
approval_completed.json
jira_trigger_created.json
```

So all services speak same language.

---

# 🚀 7. Deployment Friendly

Later this becomes:

```
jira-trigger → deployable
chat-agent → deployable
approval → deployable
onboarding → deployable
```

Without rewriting.

---

# 🔐 8. Change Safety

Use:

### CODEOWNERS

```
/services/jira-trigger-service/   @keerthana
/services/chat-agent-service/     @jeromel
/services/approval-service/       @sarumathi
/services/onboarding-service/     @rashmi
```

No one accidentally breaks others.

---

# 🧭 9. Event Driven Ready

Kafka topics mapped like:

```
jira.trigger.created
approval.requested
approval.completed
service.onboarded
```

Each service has:

```
kafka/
 ├── producer.go
 └── consumer.go
```

---

# 🏁 10. Working Rule for Team

| Rule                       | Why            |
| -------------------------- | -------------- |
| No cross-service logic     | Loose coupling |
| Shared only via contracts  | Prevent chaos  |
| API version must not break | Stability      |
| Kafka events immutable     | Reliability    |

---

# 🔥 Outcome

You get:

✔ Independent work
✔ Clean versioning
✔ Kafka-ready
✔ Scalable to microservices
✔ Easy ownership

---
