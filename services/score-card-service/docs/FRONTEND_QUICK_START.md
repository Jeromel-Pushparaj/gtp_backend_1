# Frontend Quick Start - Score Card Integration

## TL;DR - Copy & Paste Solution

### Step 1: Fetch and Transform (Copy this function)

```javascript
async function getServiceScorecard(repoName) {
  // 1. Fetch GitHub metrics (required)
  const githubRes = await fetch(
    `http://localhost:8080/api/v1/github/metrics?repo=${repoName}`
  );
  const githubData = await githubRes.json();
  
  if (!githubData.success) {
    throw new Error('Failed to fetch GitHub metrics');
  }
  
  // 2. Fetch Sonar metrics (optional)
  let sonarData = null;
  try {
    const sonarRes = await fetch(
      `http://localhost:8080/api/v1/sonar/metrics?repo=${repoName}&include_issues=true`
    );
    const sonarResult = await sonarRes.json();
    if (sonarResult.success) {
      sonarData = sonarResult.data;
    }
  } catch (err) {
    console.warn('Sonar not available:', err);
  }
  
  // 3. Transform to flat object
  const serviceData = {
    // GitHub metrics
    has_readme: githubData.data.has_readme ? 1 : 0,
    open_prs: githubData.data.open_prs || 0,
    merged_prs: githubData.data.merged_prs || 0,
    total_commits: githubData.data.total_commits || 0,
    contributors: githubData.data.contributors || 0,
    
    // Sonar metrics (with defaults)
    coverage: sonarData ? parseFloat(sonarData.metrics.coverage) || 0 : 0,
    code_smells: sonarData ? parseInt(sonarData.metrics.code_smells) || 0 : 0,
    vulnerabilities: sonarData ? parseInt(sonarData.metrics.vulnerabilities) || 0 : 0,
    duplicated_lines_density: sonarData ? parseFloat(sonarData.metrics.duplicated_lines_density) || 0 : 0,
  };
  
  // 4. Evaluate with Score Card Service
  const scorecardRes = await fetch(
    'http://localhost:8085/api/v2/scorecards/evaluate',
    {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        service_name: repoName,
        service_data: serviceData
      })
    }
  );
  
  return await scorecardRes.json();
}
```

### Step 2: Use it

```javascript
// Usage
const scorecard = await getServiceScorecard('delivery-management-frontend');
console.log('Overall Score:', scorecard.overall_percentage);
console.log('Strengths:', scorecard.strengths);
console.log('Improvements:', scorecard.improvement_areas);
```

## Real Example with Actual Data

### Input (What you fetch from APIs)

```javascript
// GitHub API Response
{
  "success": true,
  "data": {
    "repository": "delivery-management-frontend",
    "has_readme": true,              // ← Boolean
    "open_prs": 1,
    "merged_prs": 0,
    "total_commits": 4,
    "contributors": 2
  }
}

// Sonar API Response
{
  "success": true,
  "data": {
    "metrics": {
      "coverage": "0.0",             // ← String
      "code_smells": "89",           // ← String
      "vulnerabilities": "0"         // ← String
    }
  }
}
```

### Transform (What you send to Score Card Service)

```javascript
{
  "service_name": "delivery-management-frontend",
  "service_data": {
    "has_readme": 1,                 // ← Number (was boolean)
    "open_prs": 1,
    "merged_prs": 0,
    "total_commits": 4,
    "contributors": 2,
    "coverage": 0,                   // ← Number (was string)
    "code_smells": 89,               // ← Number (was string)
    "vulnerabilities": 0             // ← Number (was string)
  }
}
```

### Output (What you get back)

```javascript
{
  "service_name": "delivery-management-frontend",
  "overall_percentage": 45.5,
  "total_rules_passed": 15,
  "total_rules": 33,
  "strengths": [
    "Has README documentation"
  ],
  "improvement_areas": [
    "Code Quality: Needs attention",
    "Test Coverage: 0% (target: 60%)",
    "Code Smells: 89 (target: <= 50)"
  ],
  "scorecards": [
    {
      "scorecard_name": "Code Quality",
      "achieved_level_name": "None",
      "pass_percentage": 22.2,
      "rule_results": [...]
    }
  ]
}
```

## Critical Transformations

### 1. Boolean to Number
```javascript
// WRONG
has_readme: githubData.data.has_readme  // true/false

// CORRECT
has_readme: githubData.data.has_readme ? 1 : 0  // 1/0
```

### 2. String to Number
```javascript
// WRONG
coverage: sonarData.metrics.coverage  // "0.0" (string)

// CORRECT
coverage: parseFloat(sonarData.metrics.coverage)  // 0.0 (number)
```

### 3. Nested to Flat
```javascript
// WRONG
metrics: sonarData.metrics  // Nested object

// CORRECT
coverage: parseFloat(sonarData.metrics.coverage),
code_smells: parseInt(sonarData.metrics.code_smells)
```

## Common Errors and Fixes

| Error Message | Cause | Fix |
|--------------|-------|-----|
| "Metric 'coverage' not found in data" | Property name mismatch | Use exact property name from scorecard definition |
| Rule always fails for boolean | Sent `true` instead of `1` | Convert: `value ? 1 : 0` |
| "Invalid request body" | Missing `service_name` or `service_data` | Check request structure |
| All rules fail | Sent strings instead of numbers | Parse: `parseFloat()` or `parseInt()` |

## Testing Your Integration

```bash
# Test with curl
curl -X POST 'http://localhost:8085/api/v2/scorecards/evaluate' \
  -H 'Content-Type: application/json' \
  -d '{
    "service_name": "delivery-management-frontend",
    "service_data": {
      "has_readme": 1,
      "coverage": 0,
      "code_smells": 89,
      "vulnerabilities": 0,
      "duplicated_lines_density": 0
    }
  }'
```

## Need More Details?

- **Complete Guide**: [FRONTEND_DEVELOPER_GUIDE.md](./FRONTEND_DEVELOPER_GUIDE.md)
- **Data Flow**: [DATA_FLOW_DIAGRAM.md](./DATA_FLOW_DIAGRAM.md)
- **API Docs**: [API_ENDPOINTS.md](./API_ENDPOINTS.md)
- **OpenAPI Spec**: [openapi.yaml](./openapi.yaml)
