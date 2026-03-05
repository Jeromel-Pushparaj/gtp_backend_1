# Scorecard System Architecture

## 🎯 Overview

This is a **Port.io-style scoring system** that evaluates services using a **rule-based, level-progression** approach. Services are automatically evaluated against 5 predefined scorecards by fetching metrics from GitHub, SonarCloud, and Jira.

---

## 🏗️ System Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         Frontend/Client                          │
│                  (React, Vue, Vanilla JS, etc.)                  │
└────────────────────────────┬────────────────────────────────────┘
                             │ HTTP GET/POST
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Scorecard Service (Port 8085)                 │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │              API Layer (Gin Framework)                   │   │
│  │  • GET/POST /api/v2/scorecards/auto-evaluate             │   │
│  │  • GET/POST /api/v2/scorecards/auto-evaluate/:name       │   │
│  │  • GET /api/v2/scorecards/definitions                    │   │
│  └────────────────────┬─────────────────────────────────────┘   │
│                       │                                          │
│  ┌────────────────────▼─────────────────────────────────────┐   │
│  │              Metrics Fetcher                             │   │
│  │  • FetchGitHubMetrics(repo)                              │   │
│  │  • FetchSonarMetrics(repo)                               │   │
│  │  • FetchJiraMetrics(projectKey)                          │   │
│  │  • FetchAllMetrics() → CombinedMetrics                   │   │
│  └────────────────────┬─────────────────────────────────────┘   │
│                       │                                          │
│  ┌────────────────────▼─────────────────────────────────────┐   │
│  │              Rule Engine                                 │   │
│  │  • EvaluateAllScorecards()                               │   │
│  │  • EvaluateScorecard()                                   │   │
│  │  • Progressive level evaluation (Bronze→Silver→Gold)     │   │
│  └──────────────────────────────────────────────────────────┘   │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │         Scorecard Definitions (In-Memory)                │   │
│  │  • Code Quality (Bronze/Silver/Gold)                     │   │
│  │  • Security Maturity (Basic/Good/Great)                  │   │
│  │  • Production Readiness (Red/Yellow/Orange/Green)        │   │
│  │  • Service Health (Bronze/Silver/Gold)                   │   │
│  │  • PR Metrics (Bronze/Silver/Gold)                       │   │
│  └──────────────────────────────────────────────────────────┘   │
└────────────────────────────┬────────────────────────────────────┘
                             │ HTTP Requests
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                  Shell Test Service (Port 8080)                  │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  GitHub API Handler                                      │   │
│  │  GET /api/v1/github/metrics?repo={name}                  │   │
│  │  → Fetches: PRs, commits, contributors, README, etc.     │   │
│  └──────────────────────────────────────────────────────────┘   │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  SonarCloud API Handler                                  │   │
│  │  GET /api/v1/sonar/metrics?repo={name}                   │   │
│  │  → Fetches: coverage, bugs, vulnerabilities, etc.        │   │
│  └──────────────────────────────────────────────────────────┘   │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  Jira API Handler                                        │   │
│  │  GET /api/v1/jira/metrics?project={key}                  │   │
│  │  → Fetches: bugs, MTTR, sprint metrics, etc.            │   │
│  └──────────────────────────────────────────────────────────┘   │
└────────────────────────────┬────────────────────────────────────┘
                             │ External API Calls
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                      External Services                           │
│  • GitHub API (api.github.com)                                   │
│  • SonarCloud API (sonarcloud.io)                                │
│  • Jira API (*.atlassian.net)                                    │
└─────────────────────────────────────────────────────────────────┘
```

---

## 🔄 Request Flow

### Example: Evaluating "delivery-management-frontend"

**1. Frontend Request**

```javascript
GET /api/v2/scorecards/auto-evaluate?service_name=delivery-management-frontend&owner=myorg&jira_project_key=DM
```

**2. Scorecard Service (Port 8085)**

- Receives request with `service_name`, `owner`, `jira_project_key`
- Calls Metrics Fetcher

**3. Metrics Fetcher**

Fetches metrics from Shell Test Service (Port 8080):

```
→ GET http://localhost:8080/api/v1/github/metrics?repo=delivery-management-frontend
→ GET http://localhost:8080/api/v1/sonar/metrics?repo=delivery-management-frontend
→ GET http://localhost:8080/api/v1/jira/metrics?project=DM
```

**4. Shell Test Service (Port 8080)**

Makes external API calls:

```
→ GET https://api.github.com/repos/myorg/delivery-management-frontend
→ GET https://sonarcloud.io/api/measures/component?component=...
→ GET https://mycompany.atlassian.net/rest/api/3/search?jql=...
```

**5. Metrics Combination**

Combines all fetched metrics into a single object:

```json
{
  "github": { "has_readme": true, "open_prs": 1, "commits_last_90_days": 4, ... },
  "sonar": { "coverage": 0, "bugs": 0, "vulnerabilities": 0, ... },
  "jira": { "open_bugs": 3, "mttr": 24.5, ... }
}
```

**6. Rule Engine Evaluation**

Evaluates against all 5 scorecards using progressive level evaluation:

```
For each scorecard:
  For each level (Bronze → Silver → Gold):
    Evaluate all rules at this level
    If ALL pass → Achieved, continue to next level
    If ANY fail → Stop, return previous level
```

**7. Response**

Returns evaluation results + fetched metrics to frontend.

---

## 📊 How Scoring Works

### Progressive Level Evaluation

Services must pass **ALL rules** at a level to achieve it. Levels progress from lowest to highest.

**Example: Code Quality Scorecard**

```
Metrics: coverage=78.5, vulnerabilities=3, code_smells=35, duplications=4.2

🥉 Bronze Level:
  ✅ Coverage >= 60%       (78.5 >= 60)
  ✅ Vulnerabilities <= 10 (3 <= 10)
  ✅ Duplications <= 5%    (4.2 <= 5)
  ✅ Has README            (true)
  → ALL PASS → Bronze Achieved!

🥈 Silver Level:
  ❌ Coverage >= 80%       (78.5 < 80) FAILED
  ✅ Code Smells <= 50     (35 <= 50)
  ✅ Vulnerabilities <= 5  (3 <= 5)
  → FAILED → Stop here

Final Result: 🥉 Bronze (4/7 rules = 57%)
```

### All-or-Nothing Rule

- Must pass **100% of rules** at a level to achieve it
- Passing 3 out of 4 rules = level **NOT** achieved
- Can't skip levels (must achieve Bronze before Silver)

---

## 🎯 The 5 Scorecards

### 1. Code Quality (⚪🥉🥈🥇)

**Pattern:** Starter → Bronze → Silver → Gold

| Level      | Rules                                                                |
| ---------- | -------------------------------------------------------------------- |
| ⚪ Starter | Has README, Coverage ≥30%                                            |
| 🥉 Bronze  | Coverage ≥60%, Vulnerabilities ≤10, Duplications ≤5%, Has README     |
| 🥈 Silver  | Coverage ≥80%, Code Smells ≤50, Vulnerabilities ≤5                   |
| 🥇 Gold    | Coverage ≥90%, Code Smells ≤10, Vulnerabilities =0, Duplications ≤3% |

**Total Rules:** 13

---

### 2. Security Maturity (⚪⚪✅⭐)

**Pattern:** Starter → Basic → Good → Great

| Level      | Rules                                      |
| ---------- | ------------------------------------------ |
| ⚪ Starter | (Minimal requirements - always passes)     |
| ⚪ Basic   | Vulnerabilities ≤20, Security Hotspots ≤10 |
| ✅ Good    | Vulnerabilities ≤5, Security Hotspots ≤3   |
| ⭐ Great   | Vulnerabilities =0, Security Hotspots =0   |

**Total Rules:** 6

---

### 3. Production Readiness (🔴🟡🟠🟢)

**Pattern:** Red → Yellow → Orange → Green

| Level     | Rules                                                        |
| --------- | ------------------------------------------------------------ |
| 🔴 Red    | Has README, Active in last 90 days                           |
| 🟡 Yellow | Active in last 30 days, ≥2 contributors                      |
| 🟠 Orange | Active in last 14 days, ≥3 contributors, Quality gate passed |
| 🟢 Green  | Active in last 7 days, ≥5 contributors, Coverage ≥80%        |

**Total Rules:** 8

---

### 4. Service Health (⚪🥉🥈🥇)

**Pattern:** Starter → Bronze → Silver → Gold

| Level      | Rules                              |
| ---------- | ---------------------------------- |
| ⚪ Starter | Bugs ≤100, Open Bugs ≤50           |
| 🥉 Bronze  | Bugs ≤50, Open Bugs ≤20, MTTR <48h |
| 🥈 Silver  | Bugs ≤20, Open Bugs ≤10, MTTR <24h |
| 🥇 Gold    | Bugs ≤5, Open Bugs ≤3, MTTR <12h   |

**Total Rules:** 11

---

### 5. PR Metrics (⚪🥉🥈🥇)

**Pattern:** Starter → Bronze → Silver → Gold

| Level      | Rules                                                        |
| ---------- | ------------------------------------------------------------ |
| ⚪ Starter | Merged PRs ≥1, Open PRs ≤20                                  |
| 🥉 Bronze  | Merged PRs ≥5, Conflicts ≤30%, Open PRs ≤10                  |
| 🥈 Silver  | Merged PRs ≥20, Conflicts ≤10%, Open PRs ≤5, ≥3 contributors |
| 🥇 Gold    | Merged PRs ≥50, Conflicts ≤5%, Open PRs ≤3, ≥5 contributors  |

**Total Rules:** 13

---

## 📈 Overall Score Calculation

```
Overall Score = (Total Rules Passed / Total Rules) × 100

Example:
  Code Quality: 4/11 rules passed
  Security Maturity: 2/2 rules passed
  Production Readiness: 4/8 rules passed
  Service Health: 3/9 rules passed
  PR Metrics: 6/11 rules passed

  Total: 19/41 rules passed = 46.34%
```

### Score Interpretation

| Overall Score | Status            | Icon |
| ------------- | ----------------- | ---- |
| 90-100%       | Excellent         | 🟢   |
| 70-89%        | Good              | 🟡   |
| 50-69%        | Fair              | 🟠   |
| 0-49%         | Needs Improvement | 🔴   |

---

## 🔧 Technology Stack

### Scorecard Service (Port 8085)

- **Language:** Go (Golang)
- **Framework:** Gin (HTTP router)
- **Database:** PostgreSQL (for V1 API - not used in V2)
- **Architecture:** Layered (API → Service → Repository)

**Key Components:**

```
services/score-card-service/
├── api/v1/
│   ├── scorecard_v2_handler.go    # V2 API handlers
│   └── routes.go                   # Route registration
├── internal/
│   ├── engine/
│   │   └── rule_engine.go          # Rule evaluation logic
│   ├── fetcher/
│   │   └── metrics_fetcher.go      # Fetches from port 8080
│   ├── scorecards/
│   │   └── definitions.go          # 5 scorecard definitions
│   └── models/
│       ├── scorecard_v2.go         # Data structures
│       └── metrics.go              # Metrics models
└── cmd/
    └── main.go                     # Entry point
```

### Shell Test Service (Port 8080)

- **Language:** Go (Golang)
- **Purpose:** Proxy to external APIs (GitHub, SonarCloud, Jira)
- **Endpoints:**
  - `GET /api/v1/github/metrics?repo={name}`
  - `GET /api/v1/sonar/metrics?repo={name}`
  - `GET /api/v1/jira/metrics?project={key}`

---

## 🚀 Running the System

### Prerequisites

1. **GitHub Personal Access Token** (for GitHub API)
2. **SonarCloud Token** (for SonarCloud API)
3. **Jira API Token** (for Jira API)

### Step 1: Start Shell Test Service (Port 8080)

```bash
cd sonar-shell-test

# Set environment variables
export GITHUB_PAT="your_github_token"
export GITHUB_ORG="your-org"
export SONAR_TOKEN="your_sonar_token"
export SONAR_ORG_KEY="your-org"
export JIRA_TOKEN="your_jira_token"
export JIRA_DOMAIN="your-domain.atlassian.net"
export JIRA_EMAIL="your-email@example.com"

# Start service
go run main.go -server -port 8080
```

### Step 2: Start Scorecard Service (Port 8085)

```bash
cd services/score-card-service

# Set environment variable
export METRICS_API_BASE_URL="http://localhost:8080"

# Start service
go run cmd/main.go
```

### Step 3: Test the API

```bash
# Get all scorecard evaluations
curl "http://localhost:8085/api/v2/scorecards/auto-evaluate?service_name=delivery-management-frontend&owner=myorg" | jq '.'

# Get specific scorecard
curl "http://localhost:8085/api/v2/scorecards/auto-evaluate/CodeQuality?service_name=delivery-management-frontend" | jq '.'

# Get scorecard definitions
curl "http://localhost:8085/api/v2/scorecards/definitions" | jq '.'
```

---

## 📊 Metrics Sources

### From GitHub

- `has_readme`, `default_branch`
- `open_prs`, `closed_prs`, `merged_prs`, `prs_with_conflicts`
- `open_issues`, `closed_issues`
- `total_commits`, `commits_last_90_days`
- `contributors`, `branches`
- `last_commit_date`

### From SonarCloud

- `coverage`, `bugs`, `vulnerabilities`, `security_hotspots`
- `code_smells`, `duplicated_lines_density`
- `quality_gate_status`
- `security_rating`, `reliability_rating`, `maintainability_rating`

### From Jira

- `open_bugs`, `closed_bugs`
- `mttr` (mean time to resolve)
- `total_story_points`, `completed_story_points`
- `active_sprints`

### Derived Metrics

Calculated by the system:

- `deployment_frequency` = `commits_last_90_days / 13` (weeks)
- `days_since_last_commit` = time since `last_commit_date`
- `quality_gate_passed` = `quality_gate_status == "OK" ? 1 : 0`

---

## 🔑 Key Design Decisions

### 1. Why Progressive Evaluation?

**Port.io-style approach:** Services must achieve lower levels before higher ones. This ensures:

- Clear progression path
- Meaningful level achievements
- No "gaming" the system by cherry-picking rules

### 2. Why All-or-Nothing per Level?

Ensures quality standards are met completely at each level. Partial achievement doesn't count.

### 3. Why Auto-Fetch Metrics?

- **Consistency:** Same metrics for all services
- **Freshness:** Always up-to-date data
- **Simplicity:** Frontend doesn't need to fetch from multiple sources

### 4. Why Summary vs Detailed Views?

- **Summary:** Fast, lightweight for dashboards (~2KB)
- **Detailed:** Complete rule breakdown for debugging (~10KB+)

### 5. Why Separate Shell Service?

- **Separation of concerns:** Metrics fetching vs evaluation
- **Reusability:** Shell service can be used by other systems
- **Security:** API tokens isolated in one service

---

## 💡 Example Evaluation

**Service:** `delivery-management-frontend`

**Fetched Metrics:**

```json
{
  "github": {
    "coverage": 0,
    "commits_last_90_days": 4,
    "contributors": 2,
    "has_readme": true
  },
  "sonar": { "bugs": 0, "vulnerabilities": 0, "security_hotspots": 0 },
  "jira": { "open_bugs": 3, "mttr": 24.5 }
}
```

**Evaluation Results:**

| Scorecard            | Level     | Rules Passed | Pass % |
| -------------------- | --------- | ------------ | ------ |
| Code Quality         | None      | 8/11         | 72.73% |
| Security Maturity    | ⭐ Great  | 2/2          | 100%   |
| Production Readiness | 🟡 Yellow | 4/8          | 50%    |
| Service Health       | None      | 3/9          | 33.33% |
| PR Metrics           | None      | 6/11         | 54.55% |

**Overall:** 56.10% (23/41 rules passed)

**Strengths:** Security Maturity: Great

**Improvements:** Service Health, Code Quality

---

## 🎓 Summary

This system provides:

- ✅ **Automated evaluation** of services against quality standards
- ✅ **Progressive achievement** system (Bronze → Silver → Gold)
- ✅ **Real-time metrics** from GitHub, SonarCloud, Jira
- ✅ **Clear improvement paths** with strengths and weaknesses
- ✅ **Frontend-friendly API** with GET endpoints and summary views

Perfect for engineering teams to track and improve service quality over time!
