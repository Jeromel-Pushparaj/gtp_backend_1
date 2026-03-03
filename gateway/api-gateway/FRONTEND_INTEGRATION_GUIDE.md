# 🎨 Frontend Integration Guide

This guide helps frontend developers integrate with the GTP Backend API Gateway.

## 🚀 Quick Start

### Base URL

**Development:** `http://localhost:8000`
**Production:** `https://api.gtp-backend.com` (update with your actual URL)

### Authentication (Optional)

If JWT authentication is enabled, include the token in your requests:

```javascript
headers: {
  'Authorization': 'Bearer YOUR_JWT_TOKEN'
}
```

## 📡 Available Services

### 1. 🎫 Jira Trigger Service

**Base Path:** `/jira`

#### Create a Jira Issue

```javascript
// POST /jira/api/create-issue
const response = await fetch('http://localhost:8000/jira/api/create-issue', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    summary: 'Fix login bug',
    description: 'Users cannot log in with valid credentials',
    issueType: 'Bug',
    priority: 'High'
  })
});

const data = await response.json();
// Response: { success: true, issueKey: "PROJ-123", issueUrl: "..." }
```

### 2. 🤖 Chat Agent Service

**Base Path:** `/chat`

#### Send a Chat Message

```javascript
// POST /chat/api/chat
const response = await fetch('http://localhost:8000/chat/api/chat', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    message: 'What services are available in the platform?',
    conversationId: 'optional-conversation-id'
  })
});

const data = await response.json();
// Response: { response: "The platform has 6 microservices...", conversationId: "..." }
```

### 3. ✅ Approval Service

**Base Path:** `/approval`

#### Create Approval Request

```javascript
// POST /approval/api/v1/approval/create
const response = await fetch('http://localhost:8000/approval/api/v1/approval/create', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    title: 'Deploy to Production',
    description: 'Approve deployment of version 2.0.0',
    requestedBy: 'john.doe@company.com'
  })
});

const data = await response.json();
// Response: { success: true, approvalId: "apr-123-456", message: "..." }
```

#### Get All Approvals

```javascript
// GET /approval/api/v1/approval/all
const response = await fetch('http://localhost:8000/approval/api/v1/approval/all');
const data = await response.json();
// Response: { success: true, data: [...] }
```

### 4. 📦 Onboarding Service

**Base Path:** `/onboarding`

#### Onboard a New Service

```javascript
// POST /onboarding/api/onboard
const response = await fetch('http://localhost:8000/onboarding/api/onboard', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    serviceName: 'payment-service',
    description: 'Handles payment processing',
    team: 'Platform Team',
    repositoryUrl: 'https://github.com/company/payment-service',
    lifecycle: 'production',
    language: 'Go',
    tags: ['payment', 'critical']
  })
});

const data = await response.json();
// Response: { success: true, serviceId: "svc-123-456", message: "..." }
```

#### Get All Services

```javascript
// GET /onboarding/api/services
const response = await fetch('http://localhost:8000/onboarding/api/services');
const data = await response.json();
// Response: { success: true, data: [...] }
```

#### Get Service by ID

```javascript
// GET /onboarding/api/services/{id}
const serviceId = 'svc-123-456';
const response = await fetch(`http://localhost:8000/onboarding/api/services/${serviceId}`);
const data = await response.json();
// Response: { success: true, data: {...} }
```


#### Evaluate Service (V2 - Advanced)

```javascript
// POST /scorecard/api/v2/scorecards/evaluate
const response = await fetch('http://localhost:8000/scorecard/api/v2/scorecards/evaluate', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    serviceName: 'payment-service',
    metrics: {
      testCoverage: 85.5,
      codeQuality: 90.0,
      documentation: true
    }
  })
});

const data = await response.json();
// Response: { level: "Gold", score: 88.5 }
```

### 6. 🔍 SonarShell Service

**Base Path:** `/sonar`

#### Get SonarCloud Metrics

```javascript
// GET /sonar/api/v1/sonar/metrics?repo=payment-service&include_issues=true
const repo = 'payment-service';
const response = await fetch(
  `http://localhost:8000/sonar/api/v1/sonar/metrics?repo=${repo}&include_issues=true`
);
const data = await response.json();
// Response: { success: true, data: { repository: "...", metrics: {...}, ... } }
```

#### Full SonarCloud Setup

```javascript
// POST /sonar/api/v1/setup/full
const response = await fetch('http://localhost:8000/sonar/api/v1/setup/full', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
  }
});

const data = await response.json();
// Response: { success: true, message: "Processed 10 repositories" }
```

## 🏥 Health Checks

### Gateway Health

```javascript
// GET /health
const response = await fetch('http://localhost:8000/health');
const data = await response.json();
// Response: { status: "healthy", service: "api-gateway", version: "1.0.0", ... }
```

### Service Health Checks

Each service has its own health endpoint:

```javascript
// Check Jira service
await fetch('http://localhost:8000/jira/health');

// Check Chat service
await fetch('http://localhost:8000/chat/health');

// Check Approval service
await fetch('http://localhost:8000/approval/health');

// Check Onboarding service
await fetch('http://localhost:8000/onboarding/health');

// Check ScoreCard service
await fetch('http://localhost:8000/scorecard/health');

// Check SonarShell service
await fetch('http://localhost:8000/sonar/health');
```

## 🚨 Error Handling

### Standard Error Response Format

All errors follow this format:

```javascript
{
  "error": "Error type",
  "message": "Detailed error description"
}
```

### Common HTTP Status Codes

- **200 OK** - Request successful
- **201 Created** - Resource created successfully
- **400 Bad Request** - Invalid request parameters
- **429 Too Many Requests** - Rate limit exceeded
- **502 Bad Gateway** - Backend service unavailable
- **500 Internal Server Error** - Server error

### Example Error Handling

```javascript
async function createJiraIssue(issueData) {
  try {
    const response = await fetch('http://localhost:8000/jira/api/create-issue', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(issueData)
    });

    if (!response.ok) {
      const error = await response.json();

      if (response.status === 429) {
        console.error('Rate limit exceeded:', error.message);
        // Show user-friendly message
        alert('Too many requests. Please try again in a moment.');
      } else if (response.status === 502) {
        console.error('Service unavailable:', error.message);
        alert('The Jira service is currently unavailable. Please try again later.');
      } else {
        console.error('Error:', error);
        alert(`Error: ${error.message}`);
      }

      return null;
    }

    return await response.json();
  } catch (error) {
    console.error('Network error:', error);
    alert('Network error. Please check your connection.');
    return null;
  }
}
```

## 🔐 CORS Configuration

The gateway is configured to allow cross-origin requests. If you encounter CORS issues:

1. Ensure your origin is in the `CORS_ALLOWED_ORIGINS` environment variable
2. For development, the gateway allows all origins (`*`)
3. For production, specific origins must be configured

## ⚡ Rate Limiting

The gateway implements rate limiting:

- **Default:** 100 requests per minute per IP
- **Response:** 429 Too Many Requests when exceeded
- **Recommendation:** Implement exponential backoff in your frontend

### Example Rate Limit Handling

```javascript
async function fetchWithRetry(url, options, maxRetries = 3) {
  for (let i = 0; i < maxRetries; i++) {
    const response = await fetch(url, options);

    if (response.status !== 429) {
      return response;
    }

    // Exponential backoff: wait 1s, 2s, 4s
    const delay = Math.pow(2, i) * 1000;
    console.log(`Rate limited. Retrying in ${delay}ms...`);
    await new Promise(resolve => setTimeout(resolve, delay));
  }

  throw new Error('Max retries exceeded');
}
```

## 📚 Complete API Documentation

For complete API documentation with all endpoints, request/response schemas, and examples:

- **OpenAPI Spec:** `openapi.yaml`
- **View Online:** Import into [Swagger Editor](https://editor.swagger.io/)
- **Generate Client:** Use OpenAPI Generator to create type-safe API clients

## 🛠️ TypeScript Types (Example)

```typescript
// Jira Types
interface CreateJiraIssueRequest {
  summary: string;
  description?: string;
  issueType?: 'Task' | 'Bug' | 'Story';
  projectKey?: string;
  priority?: 'Highest' | 'High' | 'Medium' | 'Low' | 'Lowest';
  assigneeName?: string;
  assigneeId?: string;
}

interface JiraIssueResponse {
  success: boolean;
  issueKey: string;
  issueUrl: string;
  message: string;
}

// Chat Types
interface ChatRequest {
  message: string;
  conversationId?: string;
}

interface ChatResponse {
  response: string;
  conversationId: string;
}

// Onboarding Types
interface OnboardServiceRequest {
  serviceName: string;
  description?: string;
  team: string;
  repositoryUrl: string;
  lifecycle?: 'development' | 'staging' | 'production';
  language?: string;
  tags?: string[];
}

interface Service {
  id: string;
  serviceName: string;
  description: string;
  team: string;
  repositoryUrl: string;
  lifecycle: string;
  language: string;
  createdAt: string;
}

// ScoreCard Types
interface CreateScorecardRequest {
  serviceName: string;
  codeQuality?: number;
  testCoverage?: number;
  securityScore?: number;
  performanceScore?: number;
  documentationScore?: number;
}

interface Scorecard {
  id: string;
  serviceName: string;
  score: number;
  codeQuality: number;
  testCoverage: number;
  securityScore: number;
  performanceScore: number;
  documentationScore: number;
  createdAt: string;
}
```

## 📞 Support

For questions or issues:
- Check the main README.md
- Review the OpenAPI specification
- Contact the backend team

---

**Happy coding! 🚀**

