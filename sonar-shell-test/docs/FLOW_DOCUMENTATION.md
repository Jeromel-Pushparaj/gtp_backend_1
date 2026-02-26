# SonarCloud Automation Flow Documentation

## 🔄 Complete Automation Flow

This document describes the exact flow implemented in the Go backend for automating SonarCloud setup across GitHub repositories.

---

## 📋 Flow Overview

```
1. Fetch GitHub Repositories
   ↓
2. Create SonarCloud Project
   ↓
3. Check/Create Environment & Secrets
   ↓
4. Check/Push sonar.yml Workflow
   ↓
5. Fetch Analysis Results from SonarCloud API
```

---

## 🔍 Detailed Step-by-Step Flow

### **Step 1: Fetch GitHub Repositories**

**What happens:**
- Connects to GitHub API using `GITHUB_PAT`
- Lists all repositories in the organization (`teknex-poc`)
- Filters out archived repositories
- Gets default branch for each repository

**API Endpoint:**
```
GET https://api.github.com/orgs/{org}/repos
```

**Output:**
```
✅ Found 2 repositories
```

---

### **Step 2: Create SonarCloud Project**

**What happens:**
- Checks if project exists in SonarCloud
- Project key format: `{org}_{repo}` (e.g., `teknex-poc_delivery-management-frontend`)
- Creates project if it doesn't exist
- Sets the main branch for the project

**API Endpoints:**
```
GET  https://sonarcloud.io/api/projects/search?projects={projectKey}
POST https://sonarcloud.io/api/projects/create
POST https://sonarcloud.io/api/project_branches/rename
```

**Authentication:**
- Uses `SONAR_TOKEN` with Basic Auth
- Format: `{token}:` (token as username, empty password)

**Output:**
```
[1/4] Creating SonarCloud project...
      ✅ SonarCloud project created successfully
      ✅ Main branch set to: main
```

---

### **Step 3: Check and Create Environment & Secrets**

**What happens:**
- Checks if GitHub environment exists (default: `production`)
- Lists existing secrets in the environment
- If environment doesn't exist or secrets are missing:
  - Creates the environment
  - Gets the environment's public key for encryption
  - Encrypts and adds `GH_PAT` secret
  - Encrypts and adds `SONAR_TOKEN` secret
  - Verifies secrets were added successfully

**API Endpoints:**
```
GET  https://api.github.com/repos/{org}/{repo}/environments/{env}/secrets
PUT  https://api.github.com/repos/{org}/{repo}/environments/{env}
GET  https://api.github.com/repos/{org}/{repo}/environments/{env}/secrets/public-key
PUT  https://api.github.com/repos/{org}/{repo}/environments/{env}/secrets/{secret_name}
```

**Encryption:**
- Uses NaCl/libsodium (golang.org/x/crypto/nacl/box)
- Encrypts secret value with environment's public key
- Sends base64-encoded encrypted value

**Output:**
```
[2/4] Checking environment and secrets...
      📝 Setting up environment 'production' and secrets...
      ✅ Environment and secrets configured
```

---

### **Step 4: Check and Push sonar.yml**

**What happens:**
- Checks if `.github/workflows/sonar.yml` exists
- If it doesn't exist:
  - Generates sonar.yml content with:
    - Trigger on push to main branch
    - Trigger on pull requests
    - Reference to environment for secrets
    - SonarCloud scan action
  - Pushes the file to the repository

**Generated Workflow:**
```yaml
name: SonarCloud Analysis

on:
  push:
    branches:
      - main
  pull_request:
    types: [opened, synchronize, reopened]

jobs:
  sonarcloud:
    name: SonarCloud Scan
    runs-on: ubuntu-latest
    environment: production  # References the environment with secrets
    
    steps:
      - name: Checkout Repository
        uses: actions/checkout@v4
        with:
          fetch-depth: 0
      
      - name: SonarCloud Scan
        uses: SonarSource/sonarqube-scan-action@master
        env:
          GITHUB_TOKEN: ${{ secrets.GH_PAT }}
          SONAR_TOKEN: ${{ secrets.SONAR_TOKEN }}
        with:
          args: >
            -Dsonar.organization={org}
            -Dsonar.projectKey={projectKey}
```

**API Endpoints:**
```
GET  https://api.github.com/repos/{org}/{repo}/contents/{path}?ref={branch}
PUT  https://api.github.com/repos/{org}/{repo}/contents/{path}
```

**Output:**
```
[3/4] Checking and pushing sonar.yml...
      📝 Creating sonar.yml workflow...
      ✅ sonar.yml created and pushed to repository
```

---

### **Step 5: Fetch Analysis Results from SonarCloud API**

**What happens:**
- Fetches quality gate status
- Fetches project measures (bugs, vulnerabilities, coverage, etc.)
- Fetches recent issues
- Displays formatted results

**API Endpoints:**
```
GET https://sonarcloud.io/api/qualitygates/project_status?projectKey={projectKey}
GET https://sonarcloud.io/api/measures/component?component={projectKey}&metricKeys=...
GET https://sonarcloud.io/api/issues/search?componentKeys={projectKey}
```

**Metrics Retrieved:**
- Lines of Code (ncloc)
- Bugs
- Vulnerabilities
- Code Smells
- Coverage
- Duplications
- Maintainability Rating
- Reliability Rating
- Security Rating

**Output:**
```
[4/4] Fetching analysis results from SonarCloud...
   ╔════════════════════════════════════════════╗
   ║     SonarCloud Analysis Results            ║
   ╚════════════════════════════════════════════╝
   ✅ Quality Gate: OK
   
   📊 Metrics:
      Lines of Code:            1234
      Bugs:                     0
      Vulnerabilities:          0
      Code Smells:              5
      Coverage:                 85.5%
      Duplications:             2.1%
   
   ⭐ Ratings:
      Maintainability:          A ✅
      Reliability:              A ✅
      Security:                 A ✅
   
   🐛 Issues: 5 total
```

**Note:** If no analysis has run yet:
```
⚠️  Could not fetch results: ...
ℹ️  Note: Results will be available after the first workflow run
ℹ️  Push a commit to trigger the analysis
```

---

## 🚀 Running the Full Setup

### Command:
```bash
./full-setup.sh
```

### What it does:
Processes all repositories in the organization following the 5-step flow above.

### Expected Output:
```
╔══════════════════════════════════════════════════════════╗
║          Full SonarCloud & GitHub Setup                 ║
╚══════════════════════════════════════════════════════════╝

Organization: teknex-poc
SonarCloud Org: teknex-poc
Environment: production

🔍 Fetching repositories from teknex-poc...
✅ Found 2 repositories

📦 Processing repository: delivery-management-frontend
  ℹ️  Default branch: main
  [1/4] Creating SonarCloud project...
        ✅ SonarCloud project created successfully
  [2/4] Checking environment and secrets...
        ✅ Environment and secrets configured
  [3/4] Checking and pushing sonar.yml...
        ✅ sonar.yml created and pushed to repository
  [4/4] Fetching analysis results from SonarCloud...
        ℹ️  Results will be available after first workflow run
  ✅ Repository setup complete!

╔══════════════════════════════════════════════════════════╗
║                      SUMMARY                             ║
╚══════════════════════════════════════════════════════════╝
Total repositories: 2
✅ Successfully processed: 2
⏭️  Skipped (archived): 0
❌ Errors: 0
```

---

## 📊 Fetching Results Later

After workflows have run and analysis is complete:

### Command:
```bash
./fetch-results.sh
```

### What it does:
Fetches and displays analysis results for all projects without making any changes.

---

## 🔑 Environment Variables

All scripts use these environment variables:

```bash
GITHUB_PAT="ghp_..."           # GitHub Personal Access Token
SONAR_TOKEN="..."              # SonarCloud Token (without sqp_ prefix)
GITHUB_ORG="teknex-poc"        # GitHub Organization
SONAR_ORG_KEY="teknex-poc"     # SonarCloud Organization
DEFAULT_BRANCH="main"          # Default branch name
ENVIRONMENT_NAME="production"  # GitHub Environment name
```

---

## ✅ Success Criteria

For each repository, the automation ensures:

1. ✅ SonarCloud project exists
2. ✅ GitHub environment exists with required secrets
3. ✅ sonar.yml workflow file is in place
4. ✅ Analysis results are accessible (after first run)

---

## 🎯 Next Steps After Setup

1. **Push a commit** to trigger the first analysis
2. **Check GitHub Actions** tab to see workflow running
3. **View results** in SonarCloud dashboard
4. **Run `./fetch-results.sh`** to see metrics in terminal

---

For more information, see:
- **README.md** - Full documentation
- **QUICKSTART.md** - Quick start guide
- **TROUBLESHOOTING.md** - Common issues

