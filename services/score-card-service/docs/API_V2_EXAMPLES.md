# Score Card Service - API v2 Examples

## 🎯 Advanced Scorecard System with Levels

API v2 provides advanced scorecard evaluation with **Gold/Silver/Bronze** levels, **Traffic Light** scoring (Red/Yellow/Orange/Green), and rule-based evaluation.

---

## 📋 Available Scorecard Types

1. **Code Quality** - 🥉 Bronze, 🥈 Silver, 🥇 Gold
2. **DORA Metrics** - Low, Medium, High, 🏆 Elite
3. **Security Maturity** - Basic, Good, Great
4. **Production Readiness** - 🔴 Red, 🟡 Yellow, 🟠 Orange, 🟢 Green
5. **Service Health** - 🥉 Bronze, 🥈 Silver, 🥇 Gold
6. **PR Metrics** - 🥉 Bronze, 🥈 Silver, 🥇 Gold

---

## 🔌 API Endpoints

### 1. Get All Scorecard Definitions

```bash
GET http://localhost:8085/api/v2/scorecards/definitions
```

**Response:**
```json
{
  "count": 6,
  "scorecards": [
    {
      "name": "CodeQuality",
      "display_name": "Code Quality",
      "category": "code_quality",
      "level_pattern": "metal",
      "levels": [
        {
          "name": "Bronze",
          "display_name": "🥉 Bronze",
          "color": "#CD7F32",
          "icon": "🥉",
          "rules": [...]
        }
      ]
    }
  ]
}
```

### 2. Get Specific Scorecard Definition

```bash
GET http://localhost:8085/api/v2/scorecards/definitions/CodeQuality
```

Available names: `CodeQuality`, `DORA_Metrics`, `Security_Maturity`, `Production_Readiness`, `Service_Health`, `PR_Metrics`

### 3. Evaluate Service Against All Scorecards

```bash
POST http://localhost:8085/api/v2/scorecards/evaluate
Content-Type: application/json

{
  "service_name": "my-awesome-service",
  "service_data": {
    "coverage": 85.5,
    "vulnerabilities": 2,
    "code_smells": 15,
    "duplicated_lines_density": 2.5,
    "has_readme": 1,
    "bugs": 8,
    "open_bugs": 3,
    "mttr": 18,
    "deployment_frequency": 5,
    "merged_prs": 25,
    "prs_with_conflicts": 1,
    "open_prs": 4,
    "contributors": 4,
    "days_since_last_commit": 2,
    "quality_gate_passed": 1,
    "security_hotspots": 1
  }
}
```

**Response:**
```json
{
  "service_name": "my-awesome-service",
  "overall_percentage": 75.5,
  "total_rules_passed": 18,
  "total_rules": 24,
  "scorecards": [
    {
      "service_name": "my-awesome-service",
      "scorecard_name": "Code Quality",
      "achieved_level_name": "🥈 Silver",
      "rules_passed": 3,
      "rules_total": 3,
      "pass_percentage": 100,
      "rule_results": [
        {
          "rule_name": "Coverage >= 80%",
          "passed": true,
          "actual_value": 85.5,
          "expected_value": 80,
          "message": "✅ Passed"
        }
      ]
    }
  ],
  "strengths": [
    "Code Quality: 🥈 Silver",
    "Production Readiness: 🟢 Green"
  ],
  "improvement_areas": [
    "Security Maturity: Needs attention"
  ]
}
```

### 4. Evaluate Service Against Specific Scorecard

```bash
POST http://localhost:8085/api/v2/scorecards/evaluate/Production_Readiness
Content-Type: application/json

{
  "service_name": "my-awesome-service",
  "service_data": {
    "has_readme": 1,
    "days_since_last_commit": 3,
    "contributors": 6,
    "coverage": 85,
    "quality_gate_passed": 1
  }
}
```

**Response:**
```json
{
  "service_name": "my-awesome-service",
  "scorecard_name": "Production Readiness",
  "achieved_level_name": "🟢 Green",
  "rules_passed": 3,
  "rules_total": 3,
  "pass_percentage": 100,
  "rule_results": [
    {
      "rule_name": "Active in last 7 days",
      "passed": true,
      "actual_value": 3,
      "expected_value": 7,
      "message": "✅ Passed"
    },
    {
      "rule_name": "Strong team",
      "passed": true,
      "actual_value": 6,
      "expected_value": 5,
      "message": "✅ Passed"
    },
    {
      "rule_name": "High coverage",
      "passed": true,
      "actual_value": 85,
      "expected_value": 80,
      "message": "✅ Passed"
    }
  ]
}
```

---

## 🎨 Level Patterns

### Metal Pattern (Bronze → Silver → Gold)
- Used by: Code Quality, Service Health, PR Metrics
- Levels: 🥉 Bronze → 🥈 Silver → 🥇 Gold

### Traffic Light Pattern (Red → Yellow → Orange → Green)
- Used by: Production Readiness
- Levels: 🔴 Red → 🟡 Yellow → 🟠 Orange → 🟢 Green

### Performance Pattern (Low → Medium → High → Elite)
- Used by: DORA Metrics
- Levels: Low → Medium → High → 🏆 Elite

### Descriptive Pattern (Basic → Good → Great)
- Used by: Security Maturity
- Levels: Basic → Good → Great

---

## 📊 Service Data Fields Reference

| Field | Description | Example |
|-------|-------------|---------|
| `coverage` | Test coverage percentage | 85.5 |
| `vulnerabilities` | Number of vulnerabilities | 2 |
| `code_smells` | Number of code smells | 15 |
| `duplicated_lines_density` | Code duplication % | 2.5 |
| `has_readme` | Has README (1=yes, 0=no) | 1 |
| `bugs` | Total bugs | 8 |
| `open_bugs` | Open bugs | 3 |
| `mttr` | Mean time to resolve (hours) | 18 |
| `deployment_frequency` | Deployments per week | 5 |
| `merged_prs` | Merged pull requests | 25 |
| `prs_with_conflicts` | PRs with conflicts | 1 |
| `open_prs` | Open pull requests | 4 |
| `contributors` | Number of contributors | 4 |
| `days_since_last_commit` | Days since last commit | 2 |
| `quality_gate_passed` | Quality gate passed (1=yes) | 1 |
| `security_hotspots` | Security hotspots | 1 |

