# Example Approval Requests

This document contains example approval requests for different scenarios.

## 1. Deployment Approval

### Production Deployment
```json
{
  "bot_id": "U0AGPDSLH0V",
  "approver_name": "Sarumathi S",
  "requester_name": "Jeromel Pushparaj",
  "request_type": "deployment",
  "message": "Please approve production deployment for api-gateway v2.1.0",
  "request_data": {
    "service": "api-gateway",
    "version": "v2.1.0",
    "environment": "production",
    "deployment_time": "2026-03-09T14:00:00Z",
    "rollback_plan": "Automatic rollback on health check failure"
  },
  "use_app_dm": true,
  "app_bot_user_id": "U0AGPDSLH0V"
}
```

### Staging Deployment
```json
{
  "approver_name": "Tech Lead",
  "requester_name": "Developer",
  "request_type": "deployment",
  "message": "Deploying new feature to staging environment",
  "request_data": {
    "service": "payment-service",
    "version": "v1.5.0",
    "environment": "staging",
    "features": ["payment-gateway-integration", "refund-processing"]
  },
  "use_app_dm": true,
  "app_bot_user_id": "U0AGPDSLH0V"
}
```

## 2. Access Request

### Database Access
```json
{
  "approver_name": "Database Admin",
  "requester_name": "Backend Developer",
  "request_type": "access_request",
  "message": "Requesting read-only access to production database for debugging",
  "request_data": {
    "resource": "production-db",
    "access_level": "read-only",
    "duration": "24 hours",
    "reason": "Investigate customer reported issue #12345",
    "databases": ["users", "transactions"]
  },
  "use_app_dm": true,
  "app_bot_user_id": "U0AGPDSLH0V"
}
```

### AWS Console Access
```json
{
  "approver_name": "DevOps Lead",
  "requester_name": "Junior Developer",
  "request_type": "access_request",
  "message": "Need AWS console access to debug CloudWatch logs",
  "request_data": {
    "resource": "aws-console",
    "access_level": "read-only",
    "duration": "8 hours",
    "services": ["CloudWatch", "Lambda", "S3"],
    "account": "production-account"
  },
  "use_app_dm": true,
  "app_bot_user_id": "U0AGPDSLH0V"
}
```

### API Key Request
```json
{
  "approver_name": "Security Team",
  "requester_name": "Integration Developer",
  "request_type": "access_request",
  "message": "Requesting API key for third-party integration",
  "request_data": {
    "resource": "external-api-key",
    "access_level": "full",
    "duration": "permanent",
    "purpose": "Stripe payment integration",
    "scopes": ["payments.read", "payments.write", "customers.read"]
  },
  "use_app_dm": true,
  "app_bot_user_id": "U0AGPDSLH0V"
}
```

## 3. Code Review

### Feature Pull Request
```json
{
  "approver_name": "Senior Developer",
  "requester_name": "Mid-level Developer",
  "request_type": "code_review",
  "message": "Please review PR #123 for the new user authentication feature",
  "request_data": {
    "pr_number": "123",
    "repository": "backend-api",
    "branch": "feature/oauth2-authentication",
    "lines_changed": 450,
    "files_changed": 12,
    "tests_added": true,
    "breaking_changes": false
  },
  "use_app_dm": true,
  "app_bot_user_id": "U0AGPDSLH0V"
}
```

### Hotfix Review
```json
{
  "approver_name": "Tech Lead",
  "requester_name": "On-call Engineer",
  "request_type": "code_review",
  "message": "URGENT: Hotfix for production bug - memory leak in payment processor",
  "request_data": {
    "pr_number": "456",
    "repository": "payment-service",
    "branch": "hotfix/memory-leak-fix",
    "lines_changed": 25,
    "files_changed": 2,
    "severity": "critical",
    "incident_id": "INC-2024-001"
  },
  "use_app_dm": true,
  "app_bot_user_id": "U0AGPDSLH0V"
}
```

## 4. Other Request Types

### Infrastructure Change
```json
{
  "approver_name": "Infrastructure Lead",
  "requester_name": "DevOps Engineer",
  "request_type": "other",
  "message": "Requesting approval to scale up production database instances",
  "request_data": {
    "change_type": "infrastructure",
    "resource": "RDS PostgreSQL",
    "current_instance": "db.t3.medium",
    "new_instance": "db.r5.large",
    "estimated_cost_increase": "$200/month",
    "reason": "Performance degradation during peak hours"
  },
  "use_app_dm": true,
  "app_bot_user_id": "U0AGPDSLH0V"
}
```

### Budget Approval
```json
{
  "approver_name": "Finance Manager",
  "requester_name": "Engineering Manager",
  "request_type": "other",
  "message": "Requesting budget approval for new monitoring tools",
  "request_data": {
    "category": "budget",
    "amount": "$5000",
    "purpose": "Datadog Enterprise subscription",
    "duration": "annual",
    "justification": "Improved observability and reduced MTTR"
  },
  "use_app_dm": true,
  "app_bot_user_id": "U0AGPDSLH0V"
}
```

### Security Exception
```json
{
  "approver_name": "Security Officer",
  "requester_name": "Product Manager",
  "request_type": "other",
  "message": "Requesting temporary security policy exception for demo",
  "request_data": {
    "exception_type": "security_policy",
    "policy": "MFA requirement",
    "duration": "2 hours",
    "reason": "Client demo with external stakeholders",
    "mitigation": "Demo environment isolated from production"
  },
  "use_app_dm": true,
  "app_bot_user_id": "U0AGPDSLH0V"
}
```

## Using User IDs Instead of Names

You can also use Slack user IDs instead of names:

```json
{
  "approver_id": "U0AH3EX5VV2",
  "requester_id": "U0AGQHFQ0MC",
  "request_type": "deployment",
  "message": "Please approve staging deployment",
  "request_data": {
    "service": "payment-service",
    "version": "v1.5.0",
    "environment": "staging"
  },
  "use_app_dm": true,
  "app_bot_user_id": "U0AGPDSLH0V"
}
```

## Tips

1. **Required Fields**: `request_type`, `message`, `use_app_dm` are always required
2. **Approver/Requester**: Use either `*_name` OR `*_id` (at least one is required)
3. **App Bot User ID**: Required when `use_app_dm` is `true`
4. **Request Data**: Can contain any valid JSON structure
5. **Request Types**: Choose from `deployment`, `access_request`, `code_review`, or `other`

