# Frontend Developer Guide - Scorecard API

## 🎯 Quick Start

This API automatically evaluates your service/repository against 5 predefined scorecards by fetching metrics from GitHub, SonarCloud, and Jira.

**Base URL:** `http://localhost:8085`

---

## 📡 API Endpoints

### 1. Get All Scorecard Evaluations

**GET** `/api/v2/scorecards/auto-evaluate`

**Query Parameters:**

| Parameter          | Required | Default | Description                             | Example                        |
| ------------------ | -------- | ------- | --------------------------------------- | ------------------------------ |
| `service_name`     | ✅ Yes   | -       | Repository/service name                 | `delivery-management-frontend` |
| `owner`            | ❌ No    | -       | GitHub owner/organization               | `myorg`                        |
| `jira_project_key` | ❌ No    | -       | Jira project key                        | `DM`                           |
| `summary`          | ❌ No    | `true`  | Summary view (true) or detailed (false) | `true`                         |

**Example Request:**

```javascript
const response = await fetch(
  "http://localhost:8085/api/v2/scorecards/auto-evaluate?service_name=delivery-management-frontend&owner=myorg&jira_project_key=DM",
);
const data = await response.json();
```

**Response:**

```json
{
  "evaluation": {
    "service_name": "delivery-management-frontend",
    "overall_percentage": 56.1,
    "total_rules_passed": 23,
    "total_rules": 41,
    "scorecards": [
      {
        "scorecard_name": "Code Quality",
        "achieved_level_name": "Bronze",
        "achieved_level_icon": "🥉",
        "rules_passed": 8,
        "rules_total": 11,
        "pass_percentage": 72.73
      },
      {
        "scorecard_name": "Security Maturity",
        "achieved_level_name": "Great",
        "achieved_level_icon": "⭐",
        "rules_passed": 2,
        "rules_total": 2,
        "pass_percentage": 100
      }
    ],
    "strengths": ["Security Maturity: Great"],
    "improvement_areas": ["Service Health: Needs attention"]
  },
  "fetched_metrics": {
    "service_name": "myorg/delivery-management-frontend",
    "github": {
      "repository": "delivery-management-frontend",
      "has_readme": true,
      "open_prs": 1,
      "merged_prs": 0,
      "commits_last_90_days": 4,
      "contributors": 2,
      "branches": 2
    },
    "sonar": {
      "coverage": 0,
      "bugs": 0,
      "vulnerabilities": 0,
      "code_smells": 0,
      "quality_gate_status": "ERROR"
    },
    "jira": {
      "open_bugs": 3,
      "mttr": 24.5
    }
  }
}
```

---

### 2. Get Specific Scorecard Evaluation

**GET** `/api/v2/scorecards/auto-evaluate/{scorecard_name}`

**Available Scorecards:**

- `CodeQuality` - Code quality metrics (Bronze/Silver/Gold)
- `Security_Maturity` - Security posture (Basic/Good/Great)
- `Production_Readiness` - Production readiness (Red/Yellow/Orange/Green)
- `Service_Health` - Service health (Bronze/Silver/Gold)
- `PR_Metrics` - PR quality (Bronze/Silver/Gold)

**Example:**

```javascript
const response = await fetch(
  "http://localhost:8085/api/v2/scorecards/auto-evaluate/CodeQuality?service_name=delivery-management-frontend&owner=myorg",
);
const data = await response.json();
```

---

### 3. Get Scorecard Definitions

**GET** `/api/v2/scorecards/definitions`

Returns all scorecard definitions with their levels and rules.

**Example:**

```javascript
const response = await fetch(
  "http://localhost:8085/api/v2/scorecards/definitions",
);
const definitions = await response.json();
```

---

## 🎨 React Example

```jsx
import { useState, useEffect } from "react";

function ScorecardDashboard({ serviceName, owner, jiraProjectKey }) {
  const [evaluation, setEvaluation] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    const fetchScorecard = async () => {
      try {
        const params = new URLSearchParams({
          service_name: serviceName,
          ...(owner && { owner }),
          ...(jiraProjectKey && { jira_project_key: jiraProjectKey }),
        });

        const response = await fetch(
          `http://localhost:8085/api/v2/scorecards/auto-evaluate?${params}`,
        );

        if (!response.ok) throw new Error("Failed to fetch scorecard");

        const data = await response.json();
        setEvaluation(data.evaluation);
      } catch (err) {
        setError(err.message);
      } finally {
        setLoading(false);
      }
    };

    fetchScorecard();
  }, [serviceName, owner, jiraProjectKey]);

  if (loading) return <div>Loading scorecard...</div>;
  if (error) return <div>Error: {error}</div>;

  return (
    <div className="scorecard-dashboard">
      <h2>{evaluation.service_name}</h2>
      <div className="overall-score">
        <h3>Overall Score: {evaluation.overall_percentage.toFixed(1)}%</h3>
        <p>
          {evaluation.total_rules_passed} / {evaluation.total_rules} rules
          passing
        </p>
      </div>

      <div className="scorecards">
        {evaluation.scorecards.map((scorecard) => (
          <div key={scorecard.scorecard_name} className="scorecard-card">
            <h4>{scorecard.scorecard_name}</h4>
            <p className="level">
              {scorecard.achieved_level_icon} {scorecard.achieved_level_name}
            </p>
            <p className="pass-rate">{scorecard.pass_percentage.toFixed(1)}%</p>
          </div>
        ))}
      </div>

      {evaluation.strengths.length > 0 && (
        <div className="strengths">
          <h4>✅ Strengths</h4>
          <ul>
            {evaluation.strengths.map((s, i) => (
              <li key={i}>{s}</li>
            ))}
          </ul>
        </div>
      )}

      {evaluation.improvement_areas.length > 0 && (
        <div className="improvements">
          <h4>⚠️ Areas for Improvement</h4>
          <ul>
            {evaluation.improvement_areas.map((a, i) => (
              <li key={i}>{a}</li>
            ))}
          </ul>
        </div>
      )}
    </div>
  );
}

export default ScorecardDashboard;
```

---

## 🌐 Vanilla JavaScript Example

```javascript
async function getServiceScorecard(
  serviceName,
  owner = null,
  jiraProjectKey = null,
) {
  const params = new URLSearchParams({ service_name: serviceName });
  if (owner) params.append("owner", owner);
  if (jiraProjectKey) params.append("jira_project_key", jiraProjectKey);

  const url = `http://localhost:8085/api/v2/scorecards/auto-evaluate?${params}`;
  const response = await fetch(url);

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || "Failed to fetch scorecard");
  }

  return await response.json();
}

// Usage
getServiceScorecard("delivery-management-frontend", "myorg", "DM")
  .then(({ evaluation, fetched_metrics }) => {
    console.log("Overall Score:", evaluation.overall_percentage);
    console.log("Strengths:", evaluation.strengths);
    console.log("Improvements:", evaluation.improvement_areas);
    console.log("Fetched from:", fetched_metrics.service_name);
  })
  .catch((error) => console.error("Error:", error.message));
```

---

## 📊 The 5 Scorecards

### 1. Code Quality (⚪ Starter → 🥉 Bronze → 🥈 Silver → 🥇 Gold)

Evaluates code quality based on test coverage, vulnerabilities, code smells, and duplications.

**Levels:**

- **⚪ Starter**: Has README, Coverage ≥30%
- **🥉 Bronze**: Coverage ≥60%, Vulnerabilities ≤10, Duplications ≤5%, Has README
- **🥈 Silver**: Coverage ≥80%, Code Smells ≤50, Vulnerabilities ≤5
- **🥇 Gold**: Coverage ≥90%, Code Smells ≤10, Vulnerabilities =0, Duplications ≤3%

**Key Metrics:** `coverage`, `vulnerabilities`, `code_smells`, `duplicated_lines_density`, `has_readme`

### 2. Security Maturity (⚪ Starter → ⚪ Basic → ✅ Good → ⭐ Great)

Evaluates security posture based on vulnerabilities and security hotspots.

**Key Metrics:** `vulnerabilities`, `security_hotspots`

### 3. Production Readiness (🔴 Red → 🟡 Yellow → 🟠 Orange → 🟢 Green)

Evaluates production readiness based on freshness, documentation, and team collaboration.

**Key Metrics:** `has_readme`, `days_since_last_commit`, `contributors`, `quality_gate_status`, `coverage`

### 4. Service Health (⚪ Starter → 🥉 Bronze → 🥈 Silver → 🥇 Gold)

Evaluates service health based on bugs and mean time to resolve.

**Levels:**

- **⚪ Starter**: Bugs ≤100, Open Bugs ≤50
- **🥉 Bronze**: Bugs ≤50, Open Bugs ≤20, MTTR <48 hours
- **🥈 Silver**: Bugs ≤20, Open Bugs ≤10, MTTR <24 hours
- **🥇 Gold**: Bugs ≤5, Open Bugs ≤3, MTTR <12 hours

**Key Metrics:** `bugs`, `open_bugs`, `mttr` (mean time to resolve)

### 5. PR Metrics (⚪ Starter → 🥉 Bronze → 🥈 Silver → 🥇 Gold)

Evaluates PR quality and velocity based on merged PRs, conflicts, and collaboration.

**Levels:**

- **⚪ Starter**: Merged PRs ≥1, Open PRs ≤20
- **🥉 Bronze**: Merged PRs ≥5, PRs with conflicts ≤30%, Open PRs ≤10
- **🥈 Silver**: Merged PRs ≥20, PRs with conflicts ≤10%, Open PRs ≤5, Contributors ≥3
- **🥇 Gold**: Merged PRs ≥50, PRs with conflicts ≤5%, Open PRs ≤3, Contributors ≥5

**Key Metrics:** `merged_prs`, `prs_with_conflicts`, `open_prs`, `contributors`

---

## ⚙️ Summary vs Detailed View

### Summary View (Default - Lightweight ~2KB)

```javascript
// Returns scorecard results WITHOUT detailed rule_results arrays
const response = await fetch(
  "http://localhost:8085/api/v2/scorecards/auto-evaluate?service_name=myapp&summary=true",
);
```

**Use when:** Displaying dashboard, listing multiple services, mobile apps

### Detailed View (Complete ~10KB+)

```javascript
// Returns scorecard results WITH detailed rule_results arrays
const response = await fetch(
  "http://localhost:8085/api/v2/scorecards/auto-evaluate?service_name=myapp&summary=false",
);
```

**Use when:** Showing detailed rule breakdown, debugging, admin views

---

## 🔧 Error Handling

```javascript
async function fetchScorecardSafely(serviceName) {
  try {
    const response = await fetch(
      `http://localhost:8085/api/v2/scorecards/auto-evaluate?service_name=${serviceName}`,
    );

    if (!response.ok) {
      const error = await response.json();
      console.error("API Error:", error.error);
      return null;
    }

    return await response.json();
  } catch (error) {
    console.error("Network Error:", error.message);
    return null;
  }
}
```

---

## 📦 What Gets Fetched Automatically

The API automatically fetches metrics from:

- ✅ **GitHub** - README, PRs, commits, contributors, branches
- ✅ **SonarCloud** - Coverage, bugs, vulnerabilities, code smells, quality gate
- ✅ **Jira** - Open bugs, MTTR, sprint metrics (if project key provided)

**No need to fetch these manually!** Just provide the service name and optional owner/project key.

---

## 🚀 POST Alternative (if needed)

If you prefer POST requests:

```javascript
const response = await fetch(
  "http://localhost:8085/api/v2/scorecards/auto-evaluate",
  {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      service_name: "delivery-management-frontend",
      jira_project_key: "DM",
    }),
  },
);
```

Both GET and POST endpoints are available for all auto-evaluate operations.
