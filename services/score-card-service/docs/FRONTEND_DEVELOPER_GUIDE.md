# Frontend Developer Guide - Score Card Service Integration

## Overview

This guide explains how to fetch metrics from external APIs and send them to the Score Card Service for evaluation.

## Step-by-Step Integration

### Step 1: Fetch Metrics from External APIs

You need to fetch data from three sources:

#### 1.1 GitHub Metrics API

```javascript
const fetchGitHubMetrics = async (repoName) => {
  const response = await fetch(
    `http://localhost:8080/api/v1/github/metrics?repo=${repoName}`
  );
  const result = await response.json();
  
  if (!result.success) {
    throw new Error(result.error);
  }
  
  return result.data;
};
```

**Example Response:**
```json
{
  "success": true,
  "data": {
    "repository": "delivery-management-frontend",
    "has_readme": true,
    "default_branch": "main",
    "open_prs": 1,
    "closed_prs": 0,
    "merged_prs": 0,
    "prs_with_conflicts": 0,
    "open_issues": 0,
    "closed_issues": 0,
    "total_commits": 4,
    "commits_last_90_days": 4,
    "is_active": true,
    "last_commit_date": "2026-02-25T08:54:05Z",
    "contributors": 2,
    "branches": 2,
    "score": 52
  }
}
```

#### 1.2 SonarCloud Metrics API

```javascript
const fetchSonarMetrics = async (repoName) => {
  const response = await fetch(
    `http://localhost:8080/api/v1/sonar/metrics?repo=${repoName}&include_issues=true`
  );
  const result = await response.json();
  
  if (!result.success) {
    // SonarCloud might not be configured - this is optional
    return null;
  }
  
  return result.data;
};
```

**Example Response:**
```json
{
  "success": true,
  "data": {
    "repository": "delivery-management-frontend",
    "project_key": "teknex-poc_delivery-management-frontend",
    "quality_gate_status": "ERROR",
    "metrics": {
      "bugs": "0",
      "code_smells": "89",
      "coverage": "0.0",
      "duplicated_lines_density": "0.0",
      "ncloc": "2306",
      "reliability_rating": "1.0",
      "security_rating": "1.0",
      "sqale_rating": "1.0",
      "vulnerabilities": "0"
    },
    "issues_count": 89
  }
}
```

#### 1.3 Jira Metrics API (Optional)

```javascript
const fetchJiraMetrics = async (projectKey) => {
  if (!projectKey) return null;
  
  const response = await fetch(
    `http://localhost:8080/api/v1/jira/metrics?project=${projectKey}`
  );
  const result = await response.json();
  
  if (!result.success) {
    // Jira might not be configured - this is optional
    return null;
  }
  
  return result.data;
};
```

### Step 2: Transform Data for Score Card Service

The Score Card Service expects a flat object with specific property names. Here's how to map the API responses:

```javascript
const transformToScorecardData = (githubData, sonarData, jiraData) => {
  const serviceData = {};
  
  // GitHub Metrics Mapping
  if (githubData) {
    serviceData.has_readme = githubData.has_readme ? 1 : 0;
    serviceData.open_prs = githubData.open_prs;
    serviceData.closed_prs = githubData.closed_prs;
    serviceData.merged_prs = githubData.merged_prs;
    serviceData.prs_with_conflicts = githubData.prs_with_conflicts;
    serviceData.open_issues = githubData.open_issues;
    serviceData.closed_issues = githubData.closed_issues;
    serviceData.total_commits = githubData.total_commits;
    serviceData.commits_last_90_days = githubData.commits_last_90_days;
    serviceData.contributors = githubData.contributors;
    serviceData.branches = githubData.branches;
  }
  
  // SonarCloud Metrics Mapping
  if (sonarData && sonarData.metrics) {
    serviceData.coverage = parseFloat(sonarData.metrics.coverage) || 0;
    serviceData.bugs = parseInt(sonarData.metrics.bugs) || 0;
    serviceData.vulnerabilities = parseInt(sonarData.metrics.vulnerabilities) || 0;
    serviceData.code_smells = parseInt(sonarData.metrics.code_smells) || 0;
    serviceData.duplicated_lines_density = parseFloat(sonarData.metrics.duplicated_lines_density) || 0;
    serviceData.security_hotspots = 0; // Not in current response, default to 0
  }
  
  // Jira Metrics Mapping (if available)
  if (jiraData) {
    serviceData.open_bugs = jiraData.issue_stats?.bugs || 0;
    serviceData.mttr = jiraData.issue_stats?.avg_time_to_resolve || 0;
    serviceData.deployment_frequency = 0; // Calculate from your deployment data
  }
  
  return serviceData;
};
```

### Step 3: Send to Score Card Service

```javascript
const evaluateService = async (serviceName, serviceData) => {
  const response = await fetch(
    'http://localhost:8085/api/v2/scorecards/evaluate',
    {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        service_name: serviceName,
        service_data: serviceData
      })
    }
  );
  
  const result = await response.json();
  return result;
};
```

### Step 4: Complete Integration Example

```javascript
const getServiceScorecard = async (repoName, jiraProjectKey = null) => {
  try {
    // Step 1: Fetch all metrics
    const [githubData, sonarData, jiraData] = await Promise.all([
      fetchGitHubMetrics(repoName),
      fetchSonarMetrics(repoName).catch(() => null),
      jiraProjectKey ? fetchJiraMetrics(jiraProjectKey).catch(() => null) : null
    ]);
    
    // Step 2: Transform data
    const serviceData = transformToScorecardData(githubData, sonarData, jiraData);
    
    // Step 3: Evaluate
    const scorecard = await evaluateService(repoName, serviceData);
    
    return scorecard;
  } catch (error) {
    console.error('Failed to get scorecard:', error);
    throw error;
  }
};

// Usage
const scorecard = await getServiceScorecard('delivery-management-frontend', 'DM');
console.log(scorecard);
```

### Step 5: Understanding the Response

The Score Card Service returns a comprehensive evaluation:

```json
{
  "service_name": "delivery-management-frontend",
  "overall_percentage": 45.5,
  "total_rules_passed": 15,
  "total_rules": 33,
  "scorecards": [
    {
      "service_name": "delivery-management-frontend",
      "scorecard_name": "Code Quality",
      "achieved_level_name": "None",
      "rules_passed": 2,
      "rules_total": 9,
      "pass_percentage": 22.2,
      "rule_results": [
        {
          "rule_name": "Coverage >= 60%",
          "passed": false,
          "actual_value": 0,
          "expected_value": 60,
          "operator": ">=",
          "message": "Failed"
        },
        {
          "rule_name": "Has README",
          "passed": true,
          "actual_value": 1,
          "expected_value": 1,
          "operator": "==",
          "message": "Passed"
        }
      ]
    }
  ],
  "strengths": [
    "Has README documentation"
  ],
  "improvement_areas": [
    "Code Quality: Needs attention",
    "Test Coverage: 0% (target: 60%)",
    "Code Smells: 89 (target: <= 50)"
  ],
  "evaluated_at": "2026-03-05T10:30:00Z"
}
```

## Complete React Component Example

```jsx
import React, { useState, useEffect } from 'react';

const ServiceScorecard = ({ repoName, jiraProjectKey }) => {
  const [scorecard, setScorecard] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    const loadScorecard = async () => {
      try {
        setLoading(true);

        // Fetch GitHub metrics
        const githubRes = await fetch(
          `http://localhost:8080/api/v1/github/metrics?repo=${repoName}`
        );
        const githubResult = await githubRes.json();

        if (!githubResult.success) {
          throw new Error('Failed to fetch GitHub metrics');
        }

        // Fetch Sonar metrics (optional)
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
          console.warn('SonarCloud not available:', err);
        }

        // Transform data
        const serviceData = {
          // GitHub data
          has_readme: githubResult.data.has_readme ? 1 : 0,
          open_prs: githubResult.data.open_prs,
          merged_prs: githubResult.data.merged_prs,
          total_commits: githubResult.data.total_commits,
          contributors: githubResult.data.contributors,

          // Sonar data (if available)
          coverage: sonarData ? parseFloat(sonarData.metrics.coverage) : 0,
          code_smells: sonarData ? parseInt(sonarData.metrics.code_smells) : 0,
          vulnerabilities: sonarData ? parseInt(sonarData.metrics.vulnerabilities) : 0,
          duplicated_lines_density: sonarData ? parseFloat(sonarData.metrics.duplicated_lines_density) : 0,
        };

        // Evaluate with Score Card Service
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

        const scorecardResult = await scorecardRes.json();
        setScorecard(scorecardResult);

      } catch (err) {
        setError(err.message);
      } finally {
        setLoading(false);
      }
    };

    loadScorecard();
  }, [repoName, jiraProjectKey]);

  if (loading) return <div>Loading scorecard...</div>;
  if (error) return <div>Error: {error}</div>;
  if (!scorecard) return null;

  return (
    <div className="scorecard">
      <h2>{scorecard.service_name} Scorecard</h2>

      <div className="overall-score">
        <h3>Overall Score: {scorecard.overall_percentage.toFixed(1)}%</h3>
        <p>{scorecard.total_rules_passed} / {scorecard.total_rules} rules passed</p>
      </div>

      <div className="strengths">
        <h4>Strengths</h4>
        <ul>
          {scorecard.strengths.map((strength, idx) => (
            <li key={idx}>{strength}</li>
          ))}
        </ul>
      </div>

      <div className="improvements">
        <h4>Areas for Improvement</h4>
        <ul>
          {scorecard.improvement_areas.map((area, idx) => (
            <li key={idx}>{area}</li>
          ))}
        </ul>
      </div>

      <div className="scorecards">
        {scorecard.scorecards.map((sc, idx) => (
          <div key={idx} className="scorecard-item">
            <h4>{sc.scorecard_name}</h4>
            <p>Level: {sc.achieved_level_name}</p>
            <p>Pass Rate: {sc.pass_percentage.toFixed(1)}%</p>

            <details>
              <summary>View Rules ({sc.rules_passed}/{sc.rules_total})</summary>
              <ul>
                {sc.rule_results.map((rule, rIdx) => (
                  <li key={rIdx} className={rule.passed ? 'passed' : 'failed'}>
                    {rule.passed ? '✓' : '✗'} {rule.rule_name}
                    {!rule.passed && (
                      <span> (Got: {rule.actual_value}, Need: {rule.operator} {rule.expected_value})</span>
                    )}
                  </li>
                ))}
              </ul>
            </details>
          </div>
        ))}
      </div>
    </div>
  );
};

export default ServiceScorecard;
```

## Data Mapping Reference

### Complete Field Mapping Table

| Scorecard Property | Source API | Source Field | Transformation | Notes |
|-------------------|------------|--------------|----------------|-------|
| `has_readme` | GitHub | `has_readme` | `value ? 1 : 0` | Boolean to number |
| `open_prs` | GitHub | `open_prs` | Direct | Number of open PRs |
| `closed_prs` | GitHub | `closed_prs` | Direct | Number of closed PRs |
| `merged_prs` | GitHub | `merged_prs` | Direct | Number of merged PRs |
| `prs_with_conflicts` | GitHub | `prs_with_conflicts` | Direct | PRs with merge conflicts |
| `open_issues` | GitHub | `open_issues` | Direct | Number of open issues |
| `closed_issues` | GitHub | `closed_issues` | Direct | Number of closed issues |
| `total_commits` | GitHub | `total_commits` | Direct | Total commits in repo |
| `commits_last_90_days` | GitHub | `commits_last_90_days` | Direct | Recent activity |
| `contributors` | GitHub | `contributors` | Direct | Number of contributors |
| `branches` | GitHub | `branches` | Direct | Number of branches |
| `coverage` | Sonar | `metrics.coverage` | `parseFloat()` | Test coverage percentage |
| `bugs` | Sonar | `metrics.bugs` | `parseInt()` | Number of bugs |
| `vulnerabilities` | Sonar | `metrics.vulnerabilities` | `parseInt()` | Security vulnerabilities |
| `code_smells` | Sonar | `metrics.code_smells` | `parseInt()` | Code quality issues |
| `duplicated_lines_density` | Sonar | `metrics.duplicated_lines_density` | `parseFloat()` | Duplication percentage |
| `security_rating` | Sonar | `metrics.security_rating` | `parseFloat()` | 1.0 = A, 5.0 = E |
| `reliability_rating` | Sonar | `metrics.reliability_rating` | `parseFloat()` | 1.0 = A, 5.0 = E |
| `sqale_rating` | Sonar | `metrics.sqale_rating` | `parseFloat()` | Maintainability rating |
| `ncloc` | Sonar | `metrics.ncloc` | `parseInt()` | Lines of code |
| `deployment_frequency` | Custom | - | Calculate | Deployments per week |
| `mttr` | Jira/Custom | `avg_time_to_resolve` | Direct | Mean time to resolve (hours) |
| `lead_time` | Custom | - | Calculate | Lead time for changes |
| `change_failure_rate` | Custom | - | Calculate | Failed deployment rate |

### Scorecard-Specific Required Fields

#### Code Quality Scorecard
Required fields for evaluation:
- `coverage` - Test coverage percentage
- `code_smells` - Number of code smells
- `vulnerabilities` - Number of vulnerabilities
- `duplicated_lines_density` - Duplication percentage
- `has_readme` - README exists (0 or 1)

#### DORA Metrics Scorecard
Required fields for evaluation:
- `deployment_frequency` - Deployments per week
- `mttr` - Mean time to resolve (hours)
- `lead_time` - Lead time for changes (hours)
- `change_failure_rate` - Percentage of failed changes

#### Security Maturity Scorecard
Required fields for evaluation:
- `vulnerabilities` - Number of vulnerabilities
- `security_hotspots` - Number of security hotspots
- `security_rating` - Security rating (1-5)
- `has_security_policy` - Security policy exists (0 or 1)

#### Production Readiness Scorecard
Required fields for evaluation:
- `has_readme` - README exists (0 or 1)
- `has_monitoring` - Monitoring configured (0 or 1)
- `has_logging` - Logging configured (0 or 1)
- `has_health_check` - Health check exists (0 or 1)

## Common Issues and Solutions

### Issue 1: "Metric not found in data"

**Problem:** Rule evaluation fails with message "Metric 'coverage' not found in data"

**Solution:** Ensure you're sending the exact property name expected by the scorecard. Check the scorecard definition to see which properties are required.

```javascript
// Get scorecard definition to see required fields
const response = await fetch(
  'http://localhost:8085/api/v2/scorecards/definitions/CodeQuality'
);
const definition = await response.json();
console.log('Required fields:', definition.levels.flatMap(l => l.rules.map(r => r.property)));
```

### Issue 2: Boolean values not working

**Problem:** Rules like "Has README" always fail

**Solution:** Convert booleans to numbers (0 or 1)

```javascript
// Wrong
serviceData.has_readme = githubData.has_readme; // true/false

// Correct
serviceData.has_readme = githubData.has_readme ? 1 : 0; // 1/0
```

### Issue 3: String numbers not evaluated correctly

**Problem:** Sonar metrics are strings but rules expect numbers

**Solution:** Always parse string values

```javascript
// Wrong
serviceData.coverage = sonarData.metrics.coverage; // "85.5" (string)

// Correct
serviceData.coverage = parseFloat(sonarData.metrics.coverage); // 85.5 (number)
```

### Issue 4: Missing optional data

**Problem:** SonarCloud or Jira not configured for all repos

**Solution:** Handle missing data gracefully with defaults

```javascript
const serviceData = {
  // Required GitHub data
  has_readme: githubData.has_readme ? 1 : 0,

  // Optional Sonar data with defaults
  coverage: sonarData?.metrics?.coverage ? parseFloat(sonarData.metrics.coverage) : 0,
  code_smells: sonarData?.metrics?.code_smells ? parseInt(sonarData.metrics.code_smells) : 0,

  // Optional Jira data with defaults
  mttr: jiraData?.avg_time_to_resolve || 0,
};
```

## Testing Your Integration

### Test with Real Data

```bash
# 1. Fetch GitHub metrics
curl 'http://localhost:8080/api/v1/github/metrics?repo=delivery-management-frontend'

# 2. Fetch Sonar metrics
curl 'http://localhost:8080/api/v1/sonar/metrics?repo=delivery-management-frontend&include_issues=true'

# 3. Manually construct service_data from the responses

# 4. Test scorecard evaluation
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

### Validate Your Data Transformation

```javascript
const validateServiceData = (serviceData) => {
  const issues = [];

  // Check for required fields
  const requiredFields = ['has_readme', 'coverage', 'code_smells', 'vulnerabilities'];
  requiredFields.forEach(field => {
    if (!(field in serviceData)) {
      issues.push(`Missing required field: ${field}`);
    }
  });

  // Check data types
  Object.entries(serviceData).forEach(([key, value]) => {
    if (typeof value !== 'number') {
      issues.push(`Field ${key} should be a number, got ${typeof value}`);
    }
  });

  return issues;
};

// Usage
const issues = validateServiceData(serviceData);
if (issues.length > 0) {
  console.error('Data validation failed:', issues);
}
```

## Summary

1. Fetch metrics from GitHub API (required) and Sonar API (optional)
2. Transform the nested response objects into a flat `service_data` object
3. Convert booleans to numbers (0 or 1)
4. Parse string numbers to actual numbers
5. Handle missing optional data with defaults
6. Send to Score Card Service for evaluation
7. Display the comprehensive results to users

For more details, see:
- [API_ENDPOINTS.md](./API_ENDPOINTS.md) - Complete API documentation
- [openapi.yaml](./openapi.yaml) - OpenAPI specification
- [QUICK_API_REFERENCE.md](./QUICK_API_REFERENCE.md) - Quick reference guide


