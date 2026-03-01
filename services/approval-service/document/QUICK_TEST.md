# Quick Test - IT Approval Workflow

## Pick One and Test!

### 🚀 Production Deployment (Most Common)
```bash
curl -X POST http://localhost:8083/api/v1/approval/request \
  -H "Content-Type: application/json" \
  -d '{
    "channel_name": "channel_approval_service",
    "approver_name": "Sarumathi S",
    "requester_name": "Jeromel Pushparaj",
    "request_type": "production_deployment",
    "message": "Please approve production deployment for Payment Gateway v2.3.1",
    "request_data": {
      "department": "DevOps",
      "service": "payment-gateway",
      "version": "v2.3.1",
      "environment": "production",
      "scheduled_time": "2026-03-02 02:00 AM"
    }
  }'
```

### 🔥 Emergency Hotfix (Urgent)
```bash
curl -X POST http://localhost:8083/api/v1/approval/request \
  -H "Content-Type: application/json" \
  -d '{
    "channel_name": "channel_approval_service",
    "approver_name": "Sarumathi S",
    "requester_name": "Jeromel Pushparaj",
    "request_type": "emergency_hotfix",
    "message": "URGENT: Please approve emergency hotfix for critical bug in checkout flow",
    "request_data": {
      "department": "Support",
      "severity": "critical",
      "issue": "Users unable to complete checkout",
      "affected_users": "~500 users",
      "estimated_time": "15 minutes"
    }
  }'
```

### 🔐 Server Access Request
```bash
curl -X POST http://localhost:8083/api/v1/approval/request \
  -H "Content-Type: application/json" \
  -d '{
    "channel_name": "channel_approval_service",
    "approver_name": "Sarumathi S",
    "requester_name": "Jeromel Pushparaj",
    "request_type": "server_access",
    "message": "Please approve SSH access to production database server",
    "request_data": {
      "department": "Support",
      "server": "prod-db-01.company.com",
      "access_level": "read-only",
      "duration": "2 hours",
      "reason": "Investigate slow queries"
    }
  }'
```

### ☁️ Cloud Resource Request
```bash
curl -X POST http://localhost:8083/api/v1/approval/request \
  -H "Content-Type: application/json" \
  -d '{
    "channel_name": "channel_approval_service",
    "approver_name": "Sarumathi S",
    "requester_name": "Jeromel Pushparaj",
    "request_type": "cloud_provisioning",
    "message": "Please approve new AWS RDS instance for analytics",
    "request_data": {
      "department": "Data Engineering",
      "cloud_provider": "AWS",
      "resource_type": "RDS PostgreSQL",
      "estimated_monthly_cost": "$320"
    }
  }'
```

---

## What Happens After You Run This?

1. ✅ **Kafka** receives the message on `approval.requested` topic
2. ✅ **Approval Service** consumes it
3. ✅ **Slack message** appears in #channel_approval_service with:
   - @Sarumathi S mention
   - Your message
   - Request details
   - **[Approve]** and **[Reject]** buttons
4. ✅ Manager clicks **[Approve]** or **[Reject]**
5. ✅ **Database** gets updated
6. ✅ **Kafka** publishes to `approval.completed` topic
7. ✅ **Slack message** updates to show approval status

---

## IT Departments You Can Use:

- **DevOps** - Deployments, infrastructure, CI/CD
- **Support** - Access requests, troubleshooting
- **Security** - Firewall rules, policy exceptions
- **Backend Development** - Database changes, API updates
- **Data Engineering** - Analytics, data pipelines
- **Frontend Development** - UI deployments
- **QA** - Test environment changes
- **SRE** - Site reliability, monitoring

---

## Request Types You Can Use:

- `production_deployment`
- `emergency_hotfix`
- `server_access`
- `database_change`
- `infrastructure_change`
- `security_exception`
- `cloud_provisioning`
- `permission_change`
- `config_change`
- `api_key_rotation`

---

## 🎯 START HERE - Simple Test:

```bash
curl -X POST http://localhost:8083/api/v1/approval/request \
  -H "Content-Type: application/json" \
  -d '{
    "channel_name": "channel_approval_service",
    "approver_name": "Sarumathi S",
    "requester_name": "Jeromel Pushparaj",
    "request_type": "production_deployment",
    "message": "Please approve deployment to production",
    "request_data": {
      "department": "DevOps",
      "service": "api-gateway",
      "version": "v1.2.3"
    }
  }'
```

**Then check Slack!** You'll see the message with buttons!

