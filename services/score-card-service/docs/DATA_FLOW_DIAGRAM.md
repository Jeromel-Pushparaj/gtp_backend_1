# Score Card Service - Data Flow Diagram

## Frontend Integration Flow

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         FRONTEND APPLICATION                             │
│                                                                          │
│  User Input: repo_name = "delivery-management-frontend"                 │
└────────────────────────────────┬────────────────────────────────────────┘
                                 │
                                 ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                    STEP 1: FETCH EXTERNAL METRICS                        │
└─────────────────────────────────────────────────────────────────────────┘
                                 │
                    ┌────────────┼────────────┐
                    │            │            │
                    ▼            ▼            ▼
         ┌──────────────┐ ┌──────────┐ ┌──────────┐
         │   GitHub     │ │  Sonar   │ │   Jira   │
         │     API      │ │   API    │ │   API    │
         │  (Required)  │ │(Optional)│ │(Optional)│
         └──────┬───────┘ └────┬─────┘ └────┬─────┘
                │              │            │
                ▼              ▼            ▼
    ┌───────────────────────────────────────────────┐
    │         API RESPONSES (Nested JSON)           │
    ├───────────────────────────────────────────────┤
    │ GitHub:                                       │
    │ {                                             │
    │   "success": true,                            │
    │   "data": {                                   │
    │     "has_readme": true,                       │
    │     "open_prs": 1,                            │
    │     "merged_prs": 0,                          │
    │     "total_commits": 4,                       │
    │     "contributors": 2,                        │
    │     ...                                       │
    │   }                                           │
    │ }                                             │
    │                                               │
    │ Sonar:                                        │
    │ {                                             │
    │   "success": true,                            │
    │   "data": {                                   │
    │     "metrics": {                              │
    │       "coverage": "0.0",                      │
    │       "code_smells": "89",                    │
    │       "vulnerabilities": "0",                 │
    │       "duplicated_lines_density": "0.0"       │
    │     }                                         │
    │   }                                           │
    │ }                                             │
    └───────────────┬───────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│              STEP 2: TRANSFORM DATA (Frontend Logic)                    │
│                                                                          │
│  Function: transformToScorecardData(github, sonar, jira)                │
│                                                                          │
│  Transformations:                                                        │
│  - Extract nested values: github.data.has_readme → has_readme           │
│  - Convert booleans to numbers: true → 1, false → 0                     │
│  - Parse string numbers: "89" → 89, "0.0" → 0.0                         │
│  - Flatten structure: sonar.data.metrics.coverage → coverage            │
│  - Handle missing data: null → 0 (default)                              │
└────────────────────────────────┬────────────────────────────────────────┘
                                 │
                                 ▼
                    ┌────────────────────────┐
                    │  TRANSFORMED DATA      │
                    │  (Flat Object)         │
                    ├────────────────────────┤
                    │ {                      │
                    │   has_readme: 1,       │
                    │   open_prs: 1,         │
                    │   merged_prs: 0,       │
                    │   total_commits: 4,    │
                    │   contributors: 2,     │
                    │   coverage: 0,         │
                    │   code_smells: 89,     │
                    │   vulnerabilities: 0,  │
                    │   duplicated_lines: 0  │
                    │ }                      │
                    └───────────┬────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────────────┐
│         STEP 3: SEND TO SCORE CARD SERVICE                              │
│                                                                          │
│  POST http://localhost:8085/api/v2/scorecards/evaluate                  │
│                                                                          │
│  Request Body:                                                           │
│  {                                                                       │
│    "service_name": "delivery-management-frontend",                      │
│    "service_data": {                                                     │
│      has_readme: 1,                                                      │
│      coverage: 0,                                                        │
│      code_smells: 89,                                                    │
│      vulnerabilities: 0,                                                 │
│      ...                                                                 │
│    }                                                                     │
│  }                                                                       │
└────────────────────────────────┬────────────────────────────────────────┘
                                 │
                                 ▼
┌─────────────────────────────────────────────────────────────────────────┐
│              SCORE CARD SERVICE PROCESSING                               │
│                                                                          │
│  1. Load Scorecard Definitions (Code Quality, DORA, Security, etc.)     │
│  2. For each scorecard:                                                  │
│     - Evaluate each level (Bronze, Silver, Gold)                        │
│     - For each rule in level:                                           │
│       * Get property from service_data                                  │
│       * Compare with threshold using operator                           │
│       * Record pass/fail                                                │
│     - Determine achieved level                                          │
│  3. Calculate overall percentage                                        │
│  4. Identify strengths and improvement areas                            │
└────────────────────────────────┬────────────────────────────────────────┘
                                 │
                                 ▼
                    ┌────────────────────────┐
                    │  EVALUATION RESULT     │
                    ├────────────────────────┤
                    │ {                      │
                    │   overall_percentage,  │
                    │   total_rules_passed,  │
                    │   scorecards: [        │
                    │     {                  │
                    │       scorecard_name,  │
                    │       achieved_level,  │
                    │       rule_results     │
                    │     }                  │
                    │   ],                   │
                    │   strengths,           │
                    │   improvement_areas    │
                    │ }                      │
                    └───────────┬────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                    STEP 4: DISPLAY RESULTS                               │
│                                                                          │
│  Frontend renders:                                                       │
│  - Overall score badge                                                   │
│  - Achieved levels per scorecard                                        │
│  - Strengths list                                                        │
│  - Improvement areas with specific metrics                              │
│  - Detailed rule results (expandable)                                   │
└─────────────────────────────────────────────────────────────────────────┘
```

## Key Data Transformations

### Example 1: Boolean to Number

```
GitHub API Response:        Scorecard Service Input:
has_readme: true      →     has_readme: 1
has_readme: false     →     has_readme: 0
```

### Example 2: String to Number

```
Sonar API Response:         Scorecard Service Input:
coverage: "85.5"      →     coverage: 85.5
code_smells: "89"     →     code_smells: 89
```

### Example 3: Nested to Flat

```
Sonar API Response:                    Scorecard Service Input:
{                                      {
  data: {                                coverage: 0,
    metrics: {                           code_smells: 89,
      coverage: "0.0",          →        vulnerabilities: 0
      code_smells: "89",
      vulnerabilities: "0"
    }
  }
}
```

### Example 4: Missing Data Handling

```
API Response:               Scorecard Service Input:
sonarData: null       →     coverage: 0 (default)
                            code_smells: 0 (default)
```

## Rule Evaluation Example

```
Rule: "Coverage >= 60%"

Input:
  service_data.coverage = 0

Evaluation:
  property: "coverage"
  operator: ">="
  threshold: 60
  actual_value: 0
  
  0 >= 60 ? false

Result:
  {
    rule_name: "Coverage >= 60%",
    passed: false,
    actual_value: 0,
    expected_value: 60,
    operator: ">=",
    message: "Failed"
  }
```

## Common Mistakes

| Mistake | Impact | Solution |
|---------|--------|----------|
| Sending boolean instead of number | Rule always fails | Convert: `value ? 1 : 0` |
| Sending string instead of number | Rule evaluation fails | Parse: `parseFloat()` or `parseInt()` |
| Not flattening nested objects | Property not found | Extract: `data.metrics.coverage → coverage` |
| Missing required fields | Rule fails with "not found" | Provide defaults: `value || 0` |
| Wrong property names | Property not found | Check scorecard definition |
