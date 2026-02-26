# Quick Reference Guide

## 🚀 Available Commands

### 1. Full Setup (Recommended for New Repositories)
```bash
./full-setup.sh
```
**What it does:**
- ✅ Fetches all GitHub repositories
- ✅ Creates SonarCloud projects
- ✅ Sets up environment and secrets
- ✅ Pushes sonar.yml workflow
- ✅ Fetches initial analysis results

**When to use:** First-time setup or adding new repositories

---

### 2. Fetch Results Only
```bash
./fetch-results.sh
```
**What it does:**
- 📊 Fetches analysis results from SonarCloud
- 📊 Displays quality gate status
- 📊 Shows metrics and issues

**When to use:** After workflows have run and you want to see results

---

### 3. Add Environment Secrets
```bash
./add-env-secrets-go.sh
```
**What it does:**
- 🔐 Creates GitHub environment (if needed)
- 🔐 Adds GH_PAT secret
- 🔐 Adds SONAR_TOKEN secret

**When to use:** When secrets are missing or need to be updated

---

### 4. Update Workflows
```bash
./update-workflows-go.sh
```
**What it does:**
- 📝 Updates existing sonar.yml files
- 📝 Adds environment reference to workflows

**When to use:** When you need to update existing workflows to use environments

---

### 5. List Secrets
```bash
./list-secrets.sh
```
**What it does:**
- 📋 Lists all environment secrets for each repository

**When to use:** To verify which secrets are configured

---

### 6. Original Automation (Legacy)
```bash
./run.sh
```
**What it does:**
- Runs the original automation (without SonarCloud API integration)

**When to use:** For basic workflow setup without SonarCloud project creation

---

## 📊 Flow Diagram

```
┌─────────────────────────────────────────────────────────┐
│                    FULL SETUP FLOW                      │
└─────────────────────────────────────────────────────────┘

Step 1: Fetch GitHub Repos
   ↓
Step 2: Create SonarCloud Project
   ↓
Step 3: Check/Create Environment & Secrets
   ↓
Step 4: Check/Push sonar.yml
   ↓
Step 5: Fetch Analysis Results
```

---

## 🔑 Environment Variables

All scripts use these variables (configured in each script):

| Variable | Value | Description |
|----------|-------|-------------|
| `GITHUB_PAT` | `ghp_...` | GitHub Personal Access Token |
| `SONAR_TOKEN` | `...` | SonarCloud Token (no `sqp_` prefix) |
| `GITHUB_ORG` | `teknex-poc` | GitHub Organization |
| `SONAR_ORG_KEY` | `teknex-poc` | SonarCloud Organization |
| `DEFAULT_BRANCH` | `main` | Default branch name |
| `ENVIRONMENT_NAME` | `production` | GitHub Environment name |

---

## 📁 Key Files

| File | Purpose |
|------|---------|
| `main.go` | Entry point with CLI flags |
| `config.go` | Configuration management |
| `github.go` | GitHub API client |
| `sonarcloud.go` | SonarCloud API client |
| `sonar.go` | Workflow generation and orchestration |
| `go.mod` | Go dependencies |

---

## 🔍 Checking Results

### In GitHub:
1. Go to repository → **Settings** → **Environments** → **production**
2. Verify secrets: `GH_PAT`, `SONAR_TOKEN`
3. Go to **Actions** tab to see workflow runs
4. Check `.github/workflows/sonar.yml` file

### In SonarCloud:
1. Visit: https://sonarcloud.io/organizations/teknex-poc/projects
2. Click on your project
3. View dashboard with metrics

### In Terminal:
```bash
./fetch-results.sh
```

---

## 🐛 Troubleshooting

### Issue: 401 Unauthorized from SonarCloud
**Solution:** Check that `SONAR_TOKEN` doesn't have `sqp_` prefix

### Issue: Secrets not found
**Solution:** Run `./add-env-secrets-go.sh`

### Issue: No analysis results
**Solution:** Push a commit to trigger the workflow, then run `./fetch-results.sh`

### Issue: Workflow not running
**Solution:** Check that sonar.yml exists and environment secrets are configured

---

## 📚 Documentation Files

- **README.md** - Complete documentation
- **QUICKSTART.md** - Quick start guide
- **FLOW_DOCUMENTATION.md** - Detailed flow explanation
- **EXECUTION_ORDER.md** - Execution order details
- **TROUBLESHOOTING.md** - Common issues
- **PROJECT_OVERVIEW.md** - Project overview
- **SUMMARY.md** - Project summary

---

## 🎯 Common Workflows

### First-Time Setup:
```bash
./full-setup.sh
```

### Check Results After Workflow Runs:
```bash
./fetch-results.sh
```

### Update Secrets:
```bash
./add-env-secrets-go.sh
```

### Verify Configuration:
```bash
./list-secrets.sh
```

---

## ✅ Success Indicators

After running `./full-setup.sh`, you should see:

```
✅ SonarCloud project created successfully
✅ Environment and secrets configured
✅ sonar.yml created and pushed to repository
✅ Repository setup complete!
```

---

## 🔗 Useful Links

- **GitHub Organization:** https://github.com/teknex-poc
- **SonarCloud Organization:** https://sonarcloud.io/organizations/teknex-poc
- **SonarCloud API Docs:** https://sonarcloud.io/web_api
- **GitHub API Docs:** https://docs.github.com/en/rest

---

## 💡 Tips

1. **Run full-setup.sh once** for initial setup
2. **Use fetch-results.sh** to check analysis results anytime
3. **Check GitHub Actions** tab to see workflow execution
4. **Disable Automatic Analysis** in SonarCloud to avoid conflicts
5. **Token format:** SonarCloud token should NOT include `sqp_` prefix in scripts

---

For detailed information, see **FLOW_DOCUMENTATION.md**

