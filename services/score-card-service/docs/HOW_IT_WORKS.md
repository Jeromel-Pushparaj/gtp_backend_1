# How the Port.io-Style Scoring System Works

## 🎯 Overview

This scoring system evaluates services using a **rule-based, level-progression** approach similar to Port.io. Services must pass **ALL rules** at a level to achieve it, and levels progress from lowest to highest (e.g., Bronze → Silver → Gold).

---

## 📊 Example: Evaluating the "Authentication" Service

### Step 1: Collect Metrics

First, we fetch metrics from port 8080:

```json
{
  "github": {
    "merged_prs": 42,
    "contributors": 4,
    "commits_last_90_days": 45,
    "has_readme": true,
    "days_since_last_commit": 5
  },
  "sonar": {
    "coverage": 78.5,
    "vulnerabilities": 3,
    "code_smells": 35,
    "bugs": 8,
    "duplicated_lines_density": 4.2,
    "security_hotspots": 2
  },
  "jira": {
    "mttr": 18.5,
    "open_bugs": 6
  }
}
```

**Derived Metrics:**
- `deployment_frequency` = 45 commits / 13 weeks = **3.5 per week**

---

### Step 2: Evaluate Code Quality Scorecard

**Scorecard:** Code Quality (Bronze/Silver/Gold)

#### 🥉 Bronze Level - Evaluating...

| Rule | Threshold | Actual | Pass? |
|------|-----------|--------|-------|
| Coverage >= 60% | 60 | 78.5 | ✅ |
| Vulnerabilities <= 10 | 10 | 3 | ✅ |
| Duplications <= 5% | 5 | 4.2 | ✅ |
| Has README | 1 | 1 | ✅ |

**Result:** ✅ **ALL rules passed** → Bronze level achieved!

#### 🥈 Silver Level - Evaluating...

| Rule | Threshold | Actual | Pass? |
|------|-----------|--------|-------|
| Coverage >= 80% | 80 | 78.5 | ❌ |
| Code Smells <= 50 | 50 | 35 | ✅ |
| Vulnerabilities <= 5 | 5 | 3 | ✅ |

**Result:** ❌ **Failed coverage rule** → Silver level NOT achieved

**Stopped here** - Can't achieve Gold if Silver failed

**Final Result:** Code Quality = **🥉 Bronze** (4/7 rules passing = 57%)

---

### Step 3: Evaluate DORA Metrics Scorecard

**Scorecard:** DORA Metrics (Low/Medium/High/Elite)

#### 📉 Low Level - Evaluating...

| Rule | Threshold | Actual | Pass? |
|------|-----------|--------|-------|
| Deployment Frequency >= 1/week | 1 | 3.5 | ✅ |
| MTTR < 24 hours | 24 | 18.5 | ✅ |

**Result:** ✅ **ALL rules passed** → Low level achieved!

#### 📊 Medium Level - Evaluating...

| Rule | Threshold | Actual | Pass? |
|------|-----------|--------|-------|
| Deployment Frequency >= 3/week | 3 | 3.5 | ✅ |
| MTTR < 12 hours | 12 | 18.5 | ❌ |

**Result:** ❌ **Failed MTTR rule** → Medium level NOT achieved

**Final Result:** DORA Metrics = **📉 Low** (2/8 rules passing = 25%)

---

### Step 4: Evaluate Security Maturity Scorecard

**Scorecard:** Security Maturity (Basic/Good/Great)

#### ⚪ Basic Level - Evaluating...

| Rule | Threshold | Actual | Pass? |
|------|-----------|--------|-------|
| Vulnerabilities <= 20 | 20 | 3 | ✅ |
| Security Hotspots <= 10 | 10 | 2 | ✅ |

**Result:** ✅ **ALL rules passed** → Basic level achieved!

#### ✅ Good Level - Evaluating...

| Rule | Threshold | Actual | Pass? |
|------|-----------|--------|-------|
| Vulnerabilities <= 5 | 5 | 3 | ✅ |
| Security Hotspots <= 3 | 3 | 2 | ✅ |

**Result:** ✅ **ALL rules passed** → Good level achieved!

#### ⭐ Great Level - Evaluating...

| Rule | Threshold | Actual | Pass? |
|------|-----------|--------|-------|
| Vulnerabilities == 0 | 0 | 3 | ❌ |
| Security Hotspots == 0 | 0 | 2 | ❌ |

**Result:** ❌ **Failed both rules** → Great level NOT achieved

**Final Result:** Security Maturity = **✅ Good** (4/6 rules passing = 67%)

---

### Step 5: Calculate Overall Score

After evaluating all 6 scorecards:

| Scorecard | Level Achieved | Rules Passed | Rules Total | Pass % |
|-----------|----------------|--------------|-------------|--------|
| Code Quality | 🥉 Bronze | 4 | 7 | 57% |
| DORA Metrics | 📉 Low | 2 | 8 | 25% |
| Security Maturity | ✅ Good | 4 | 6 | 67% |
| Production Readiness | 🟡 Yellow | 5 | 10 | 50% |
| Service Health | 🥉 Bronze | 3 | 9 | 33% |
| PR Metrics | 🥈 Silver | 7 | 11 | 64% |

**Overall Score:** 25/51 rules passing = **49%**

---

### Step 6: Identify Strengths & Improvements

**Strengths** (>90% pass rate):
- None yet - keep improving!

**Improvement Areas** (<60% pass rate):
- ⚠️ DORA Metrics: Only 2/8 rules passing
- ⚠️ Service Health: Only 3/9 rules passing
- ⚠️ Code Quality: Only 4/7 rules passing
- ⚠️ Production Readiness: Only 5/10 rules passing

---

## 🔑 Key Concepts

### 1. Progressive Evaluation
- Start at lowest level (Bronze, Low, Red, Basic)
- Evaluate all rules at that level
- If **ALL pass**, move to next level
- If **ANY fail**, stop (can't achieve higher levels)

### 2. All-or-Nothing per Level
- Must pass **100% of rules** at a level to achieve it
- Passing 2 out of 3 rules = level NOT achieved

### 3. Overall Score
- Aggregate across all scorecards
- Total rules passed / Total rules = Overall %
- Example: 25/51 = 49%

### 4. Different Level Patterns
- **Metal:** Bronze 🥉 → Silver 🥈 → Gold 🥇
- **Performance:** Low 📉 → Medium 📊 → High 📈 → Elite 🏆
- **Traffic Light:** Red 🔴 → Yellow 🟡 → Orange 🟠 → Green 🟢
- **Descriptive:** Basic ⚪ → Good ✅ → Great ⭐

---

## 🚀 How to Use

```go
// 1. Fetch metrics from port 8080
fetcher := fetcher.NewMetricsFetcher("http://localhost:8080")
metrics, _ := fetcher.FetchAllMetrics("authentication", "AUTH")

// 2. Convert to map
metricsMap := metrics.ToMap()

// 3. Get scorecard definitions
scorecards := scorecards.GetAllScorecardDefinitions()

// 4. Evaluate
engine := engine.NewRuleEngine()
result := engine.EvaluateAllScorecards(scorecards, metricsMap, "authentication")

// 5. Use results
fmt.Printf("Overall: %.1f%%\n", result.OverallPercentage)
```

---

## 📈 Example Output

```
=== Authentication Service Score ===

Overall Score: 49% (25/51 rules passing)

Code Quality: 🥉 Bronze (57%)
DORA Metrics: 📉 Low (25%)
Security Maturity: ✅ Good (67%)
Production Readiness: 🟡 Yellow (50%)
Service Health: 🥉 Bronze (33%)
PR Metrics: 🥈 Silver (64%)

Strengths: None yet
Improvements: Focus on DORA Metrics, Service Health
```

