API Endpoints

Base URL

`http://localhost:8083`

## Endpoints List

### Health

1. GET `/health` - Health check

### Slack Channels

2. POST `/api/v1/slack/channel/create` - Create Slack channel
3. GET `/api/v1/slack/channels/all` - Get all Slack channels
4. POST `/api/v1/slack/channel/by-name` - Get channel by name
5. POST `/api/v1/slack/channel/by-id` - Get channel by ID

### Slack Users

6. GET `/api/v1/slack/users/all` - Get all Slack users
7. POST `/api/v1/slack/user/by-name` - Get user by name
8. POST `/api/v1/slack/user/by-id` - Get user by ID
9. GET `/api/v1/slack/apps/all` - Get all Slack apps

### Slack Messaging

10. POST `/api/v1/slack/member/add` - Add member to channel
11. POST `/api/v1/slack/message/send` - Send message to channel
12. POST `/api/v1/slack/dm-channel/get` - Get DM channel ID
13. POST `/api/v1/slack/approval-form-button/send` - Send approval form button to channel

### Approval Requests (Query Operations)

14. GET `/api/v1/approval/all` - Get all approval requests
15. GET `/api/v1/approval/pending` - Get pending approval requests
16. POST `/api/v1/approval/by-id` - Get approval request by ID

### Approval Requests (Event-Driven - Kafka Flow)

17. POST `/api/v1/approval/request` - Create approval request
18. POST `/api/v1/approval/domain-change` - Create domain change approval request
19. POST `/api/v1/approval/generic` - Create generic approval request (NEW)

---

## Event-Driven Endpoints

The following 3 endpoints follow the complete Kafka event-driven flow:

- `/api/v1/approval/request`
- `/api/v1/approval/domain-change`
- `/api/v1/approval/generic`

**Event Flow:**

1. API receives request and publishes to `approval.requested` Kafka topic
2. Kafka consumer processes message and sends Slack message with interactive buttons
3. Human approves/rejects via Slack modal (with optional/required comments)
4. Socket Mode handler updates database and publishes to `approval.completed` topic
5. Business consumer processes completion and publishes to `action.executed` or `action.rejected` topic

All other endpoints are synchronous operations.

---

## New Endpoint: Generic Approval Request

### POST `/api/v1/approval/generic`

**Description:** A flexible approval endpoint that supports any type of approval request.

**Features:**

- ✅ User mentions for approver and requester
- ✅ Slack interactive buttons (Approve/Reject)
- ✅ Comment required for reject, optional for approve
- ✅ Sends approval request to approver's DM
- ✅ Flexible `request_data` field for any structured data
- ✅ Kafka event-driven architecture

**Request Body:**

```json
{
  "bot_id": "U0AGPDSLH0V",
  "approver_name": "Sarumathi S",
  "requester_name": "Jeromel Pushparaj",
  "request_type": "deployment",
  "message": "Please approve production deployment",
  "request_data": {
    "service": "api-gateway",
    "version": "v2.1.0",
    "environment": "production"
  },
  "use_app_dm": true,
  "app_bot_user_id": "U0AGPDSLH0V"
}
```

**Fields:**

- `bot_id` (optional): Bot ID for reference
- `approver_id` (optional): Slack user ID of approver (use this OR approver_name)
- `approver_name` (optional): Display name of approver (use this OR approver_id)
- `requester_id` (optional): Slack user ID of requester (use this OR requester_name)
- `requester_name` (optional): Display name of requester (use this OR requester_id)
- `request_type` (required): Type of approval request (e.g., "deployment", "access_request", "code_review")
- `message` (required): Detailed message about the approval request
- `request_data` (optional): Additional structured data about the request (flexible JSON object)
- `use_app_dm` (required): Must be true - sends to approver's DM
- `app_bot_user_id` (required when use_app_dm=true): Bot user ID for DM

**Response:**

```json
{
  "success": true,
  "message": "Generic approval request created and published to Kafka",
  "request_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Example Use Cases:**

1. **Deployment Approval:**

```bash
curl -X POST http://localhost:8083/api/v1/approval/generic \
  -H "Content-Type: application/json" \
  -d '{
    "approver_name": "Sarumathi S",
    "requester_name": "Jeromel Pushparaj",
    "request_type": "deployment",
    "message": "Please approve production deployment",
    "request_data": {
      "service": "api-gateway",
      "version": "v2.1.0",
      "environment": "production"
    },
    "use_app_dm": true,
    "app_bot_user_id": "U0AGPDSLH0V"
  }'
```

2. **Code Review Approval:**

```bash
curl -X POST http://localhost:8083/api/v1/approval/generic \
  -H "Content-Type: application/json" \
  -d '{
    "approver_name": "Jeromel Pushparaj",
    "requester_name": "Sarumathi S",
    "request_type": "code_review",
    "message": "Please review PR #123 for the new feature",
    "request_data": {
      "pr_number": "123",
      "repository": "backend-api",
      "branch": "feature/new-approval-system",
      "lines_changed": 450
    },
    "use_app_dm": true,
    "app_bot_user_id": "U0AGPDSLH0V"
  }'
```

3. **Access Request Approval:**

```bash
curl -X POST http://localhost:8083/api/v1/approval/generic \
  -H "Content-Type: application/json" \
  -d '{
    "approver_name": "Sarumathi S",
    "requester_name": "Jeromel Pushparaj",
    "request_type": "access_request",
    "message": "Requesting access to production database",
    "request_data": {
      "resource": "production-db",
      "access_level": "read-only",
      "duration": "24 hours"
    },
    "use_app_dm": true,
    "app_bot_user_id": "U0AGPDSLH0V"
  }'
```

**Slack Message Format:**

The approver receives a DM with:

```
*deployment*

*Requester:* @Jeromel Pushparaj
*Approver:* @Sarumathi S

Please approve production deployment

*Request Details:*
• *service:* api-gateway
• *version:* v2.1.0
• *environment:* production

[Approve] [Reject]
```

**Interactive Buttons:**

- **Approve Button (Green):** Opens modal with optional comment field
- **Reject Button (Red):** Opens modal with required comment field

**Kafka Topics:**

- Publishes to: `approval.requested`
- Completion published to: `approval.completed`
- Business action published to: `action.executed` or `action.rejected`
