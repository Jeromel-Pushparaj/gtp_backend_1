# Approval Service

A microservice for managing approval workflows through Slack integration with Kafka-based event-driven architecture.

## Overview

The Approval Service provides a comprehensive solution for handling approval requests within Slack workspaces. It enables users to create, manage, and process approval requests through interactive Slack messages, with full event tracking via Apache Kafka.

## Architecture

### Technology Stack

- **Language**: Go 1.21+
- **Web Framework**: Gin
- **Database**: SQLite (GORM ORM)
- **Message Broker**: Apache Kafka
- **Integration**: Slack API (Socket Mode)
- **Containerization**: Docker

### Core Components

1. **HTTP API Server**: RESTful API for approval request management
2. **Slack Integration**: Socket Mode handler for interactive Slack messages
3. **Kafka Producer**: Publishes approval events to Kafka topics
4. **Kafka Consumers**:
   - Approval Consumer: Processes approval requests and creates database records
   - Business Consumer: Handles approval completion events and triggers downstream actions
5. **Database Layer**: SQLite database for persistent storage of approval requests

## Features

### Approval Workflow

- Create approval requests via REST API
- Send interactive approval messages to Slack users via Direct Messages or channels
- Handle approve/reject actions through Slack Block Kit buttons
- Collect approver comments via modal dialogs
- Send automated notifications to requesters upon approval/rejection
- Track approval status and history in database

### Slack Integration

- User resolution by name or ID
- Channel resolution by name or ID
- Direct Message (DM) channel management
- Interactive message components (buttons, modals)
- Real-time event handling via Socket Mode
- Support for mentions and rich message formatting

### Event-Driven Architecture

- Asynchronous processing via Kafka
- Event sourcing for approval lifecycle
- Decoupled business logic execution
- Retry mechanisms for eventual consistency

## Installation

### Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose
- Slack workspace with bot token and app token
- Apache Kafka (provided via Docker Compose)

### Environment Variables

Create a `.env` file in the service root directory:

```env
# Service Configuration
SERVICE_NAME=approval-service
SERVICE_PORT=8083
SERVICE_HOST=0.0.0.0
ENVIRONMENT=development
LOG_LEVEL=debug

# Slack Configuration
SLACK_BOT_TOKEN=xoxb-your-bot-token
SLACK_APP_TOKEN=xapp-your-app-token

# Kafka Configuration
KAFKA_BROKERS=localhost:9092
KAFKA_GROUP_ID=approval-service-group
```

### Slack App Configuration

Required OAuth Scopes:

- `chat:write` - Send messages
- `users:read` - Read user information
- `channels:read` - Read channel information
- `im:write` - Open and write to DMs
- `im:read` - Read DM information

Required Event Subscriptions:

- Enable Socket Mode
- Subscribe to bot events: `message.im`, `app_mention`

### Setup Instructions

1. Clone the repository and navigate to the service directory:

```bash
cd services/approval-service
```

2. Install Go dependencies:

```bash
go mod download
```

3. Start infrastructure services (Kafka, Zookeeper, PostgreSQL, Redis):

```bash
cd ../..
docker-compose up -d
```

4. Wait for Kafka to be ready (use the helper script):

```bash
./services/approval-service/scripts/kafka-start.sh
```

5. Run the service:

```bash
cd services/approval-service
go run cmd/main.go
```

The service will start on `http://localhost:8083`

## Event-Driven Architecture Flow

### Complete Approval Workflow

```
Frontend/Client
   ↓
API Service (approval-service)
   ↓
Kafka Topic → approval.requested
   ↓
Approval Service (Kafka Consumer)
   ↓
Database (SQLite) + Slack Notifier
   ↓
Human Approves/Rejects (Slack)
   ↓
Slack sends response (Socket Mode)
   ↓
Approval Service (HandleApproval)
   ↓
Database Update + Requester Notification
   ↓
Kafka Topic → approval.completed
   ↓
Business Consumer Service
   ↓
Kafka Topic → action.executed / action.rejected
   ↓
Frontend/Downstream Services
```

### Endpoints Following This Flow

**Only 2 endpoints follow the complete Kafka event-driven flow:**

1. **POST** `/api/v1/approval/request` - Create generic approval request
2. **POST** `/api/v1/approval/domain-change` - Create domain change approval request

**All other 16 endpoints are direct synchronous operations** (Slack user/channel management, messaging, approval queries) and do NOT use Kafka.

### Detailed Flow Steps

1. **Request Creation**
   - Frontend/Client sends POST request to approval endpoint
   - Controller resolves user names to Slack user IDs
   - Controller opens DM channel with approver (if using DM)
   - Controller publishes message to `approval.requested` Kafka topic
   - Returns success response with request_id to client

2. **Request Processing**
   - Kafka consumer receives message from `approval.requested` topic
   - Consumer creates database record with status "pending"
   - Consumer sends interactive Slack message to approver with Approve/Reject buttons

3. **Human Interaction**
   - Approver sees message in Slack (DM or channel)
   - Approver clicks Approve or Reject button
   - Socket Mode handler receives interaction event
   - Handler opens modal dialog for approver to add comment
   - Approver submits modal with comment

4. **Approval Processing**
   - Service retrieves approval request from database (with retry logic for eventual consistency)
   - Service updates database record with decision and comment
   - Service updates original Slack message with approval status
   - Service sends notification DM to requester
   - Service publishes message to `approval.completed` Kafka topic

5. **Business Action Execution**
   - Business consumer receives message from `approval.completed` topic
   - If approved: publishes to `action.executed` topic
   - If rejected: publishes to `action.rejected` topic
   - Downstream services/Frontend can consume these topics for further processing

### Why Only 2 Endpoints Use Kafka?

**Event-Driven (Kafka) Endpoints:**

- Require asynchronous processing
- Involve human approval workflow
- Need to decouple request creation from execution
- Trigger downstream business actions

**Direct (Synchronous) Endpoints:**

- Simple CRUD operations
- Slack API interactions (user/channel lookup, messaging)
- Query operations (get all approvals, get pending, etc.)
- No need for asynchronous processing or event sourcing

## Kafka Topics

The service uses the following Kafka topics for event-driven communication:

### approval.requested

Published when a new approval request is created via API.

Message Schema:

```json
{
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "requester_id": "U9876543210",
  "requester_name": "John Doe",
  "approver_id": "U0123456789",
  "approver_name": "Jane Smith",
  "channel_id": "D0ABCDEF123",
  "request_type": "domain_change",
  "request_data": {
    "old_domain_name": "legacy-domain",
    "new_domain_name": "new-domain",
    "change_reason": "Rebranding initiative"
  },
  "message": "Domain change request details",
  "title": "Domain Name Change Request",
  "priority": "high",
  "category": "infrastructure"
}
```

### approval.completed

Published when an approver approves or rejects a request.

Message Schema:

```json
{
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "approved",
  "approved": true,
  "processed_by": "U0123456789",
  "processed_at": "2026-03-05T10:35:00Z",
  "reason": "Approved by manager",
  "approver_comment": "Looks good to proceed",
  "request_data": {
    "old_domain_name": "legacy-domain",
    "new_domain_name": "new-domain"
  }
}
```

### action.executed

Published by business consumer when an approved request triggers an action.

### action.rejected

Published by business consumer when a rejected request is logged.

## Database Schema

The service uses SQLite with the following schema:

### approval_requests Table

| Column           | Type     | Description                                  |
| ---------------- | -------- | -------------------------------------------- |
| id               | INTEGER  | Primary key (auto-increment)                 |
| request_id       | VARCHAR  | Unique request identifier (UUID)             |
| requester_id     | VARCHAR  | Slack user ID of requester                   |
| requester_name   | VARCHAR  | Display name of requester                    |
| approver_id      | VARCHAR  | Slack user ID of approver                    |
| approver_name    | VARCHAR  | Display name of approver                     |
| channel_id       | VARCHAR  | Slack channel/DM ID where message was sent   |
| message_ts       | VARCHAR  | Slack message timestamp                      |
| request_type     | VARCHAR  | Type of approval request                     |
| request_data     | TEXT     | JSON data containing request details         |
| title            | VARCHAR  | Request title                                |
| description      | TEXT     | Request description                          |
| priority         | VARCHAR  | Priority level (high, medium, low)           |
| category         | VARCHAR  | Request category                             |
| attachments      | TEXT     | Additional attachments                       |
| due_date         | DATETIME | Optional due date                            |
| status           | VARCHAR  | Current status (pending, approved, rejected) |
| approved         | BOOLEAN  | Approval decision                            |
| processed_by     | VARCHAR  | User ID who processed the request            |
| processed_at     | DATETIME | Timestamp of processing                      |
| reason           | VARCHAR  | Reason for decision                          |
| approver_comment | TEXT     | Comment from approver                        |
| created_at       | DATETIME | Record creation timestamp                    |
| updated_at       | DATETIME | Record update timestamp                      |

## Monitoring and Debugging

### Helper Scripts

The service includes several helper scripts in `scripts/` directory:

#### Start Kafka Infrastructure

```bash
./scripts/kafka-start.sh
```

Starts Zookeeper, Kafka, Kafka UI, PostgreSQL, and Redis with proper health checks and startup order.

#### Check Kafka Status

```bash
./scripts/kafka-status.sh
```

Displays status of all Kafka infrastructure components, topics, and message counts.

#### View Kafka Messages (Prettified)

```bash
./scripts/kafka-pretty.sh approval.requested
./scripts/kafka-pretty.sh approval.completed
./scripts/kafka-pretty.sh action.executed
./scripts/kafka-pretty.sh action.rejected
```

#### Check Database

```bash
./scripts/db-check.sh all        # Show all requests
./scripts/db-check.sh pending    # Show pending requests
./scripts/db-check.sh approved   # Show approved requests
./scripts/db-check.sh rejected   # Show rejected requests
./scripts/db-check.sh stats      # Show statistics
./scripts/db-check.sh latest     # Show latest request
```

### Kafka UI

Access the Kafka UI web interface at `http://localhost:8090` to:

- View all topics and messages
- Monitor consumer groups
- Check message lag
- Inspect message payloads

### Database Inspection

Using SQLite CLI:

```bash
cd services/approval-service
sqlite3 approval_service.db

.headers on
.mode column
SELECT * FROM approval_requests ORDER BY created_at DESC LIMIT 10;
```

Using VS Code SQLite Viewer extension:

1. Right-click on `approval_service.db` in VS Code Explorer
2. Select "Open Database"
3. Browse tables and run queries in the visual interface

### Logs

The service provides detailed logging for debugging:

```bash
# Run service with logs
go run cmd/main.go

# View Kafka logs
docker logs kafka --tail 100 -f

# View Zookeeper logs
docker logs zookeeper --tail 100 -f
```

## Error Handling

### Retry Mechanism

The service implements retry logic for handling race conditions between Kafka message production and consumption:

- **HandleApproval**: Retries up to 10 times (1 second delay) when looking up approval requests
- This handles cases where Slack interaction arrives before Kafka consumer creates the database record

### Common Issues

#### "missing_scope" Error

- Cause: User has disabled DMs with the bot
- Solution: User must enable messages with the bot in Slack settings

#### "approval request not found" Error

- Cause: Race condition between Kafka consumer and approval handler
- Solution: Retry mechanism handles this automatically (up to 10 seconds)

#### "Connection refused" to Kafka

- Cause: Kafka is not running or not ready
- Solution: Use `./scripts/kafka-start.sh` to properly start Kafka infrastructure

#### "2/2 brokers are down"

- Cause: Kafka crashed or Zookeeper is not available
- Solution: Restart infrastructure with `docker-compose restart kafka zookeeper`

## Configuration

### Docker Compose

The service relies on infrastructure defined in the root `docker-compose.yml`:

- **Zookeeper**: Port 2181
- **Kafka**: Ports 9092 (external), 29092 (internal)
- **Kafka UI**: Port 8090
- **PostgreSQL**: Port 5432
- **Redis**: Port 6379

All services include:

- `restart: always` policy for automatic recovery
- Health checks for proper startup ordering
- Resource limits to prevent crashes

### Service Configuration

Default configuration (can be overridden via environment variables):

- **Service Port**: 8083
- **Service Host**: 0.0.0.0
- **Kafka Brokers**: localhost:9092
- **Kafka Group ID**: approval-service-group
- **Database Path**: approval_service.db
- **Log Level**: debug
- **Environment**: development

## Development

### Project Structure

```
services/approval-service/
├── cmd/
│   └── main.go                 # Application entry point
├── config/
│   └── config.go               # Configuration management
├── constants/
│   └── constants.go            # Application constants
├── controller/
│   └── v1/
│       ├── approval_controller.go
│       └── slack_controller.go
├── db/
│   ├── database.go             # Database initialization
│   └── repository.go           # Data access layer
├── middleware/
│   ├── logging.go              # Request logging
│   └── recovery.go             # Panic recovery
├── models/
│   └── approval.go             # Data models
├── resources/
│   └── *.go                    # Request/response DTOs
├── routes/
│   └── routes.go               # HTTP route definitions
├── scripts/
│   ├── kafka-start.sh          # Kafka startup script
│   ├── kafka-status.sh         # Kafka status checker
│   ├── kafka-pretty.sh         # Pretty Kafka consumer
│   └── db-check.sh             # Database query helper
├── service/
│   └── v1/
│       ├── approval_service.go # Core business logic
│       ├── consumer.go         # Kafka consumer
│       ├── business_consumer.go
│       ├── kafka_service.go    # Kafka producer
│       ├── slack_service.go    # Slack API client
│       └── socket_handler.go   # Slack Socket Mode
└── validator/
    └── *.go                    # Request validators
```
