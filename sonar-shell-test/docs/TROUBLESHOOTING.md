# Troubleshooting Guide

## Common Issues and Solutions

### 1. "You are running CI analysis while Automatic Analysis is enabled"

**Error:**
```
ERROR You are running CI analysis while Automatic Analysis is enabled. 
Please consider disabling one or the other.
```

**Cause:** SonarCloud has Automatic Analysis enabled, which conflicts with CI-based analysis.

**Solution:**

1. Go to your SonarCloud project:
   ```
   https://sonarcloud.io/project/configuration?id=teknex-poc_delivery-management-frontend
   ```

2. Navigate to: **Administration** → **Analysis Method**

3. Disable **Automatic Analysis**

4. Enable **GitHub Actions** (or your CI/CD method)

5. Re-run your GitHub Actions workflow

**Quick Fix for All Projects:**

For each project in your organization:
- https://sonarcloud.io/project/configuration?id=teknex-poc_delivery-management-frontend
- https://sonarcloud.io/project/configuration?id=teknex-poc_sonarqube

Go to **Administration** → **Analysis Method** → Turn OFF "Automatic Analysis"

---

### 2. Environment Secrets Not Found (404)

**Error:**
```
⚠️  No environment 'production' or error: 404 Not Found
```

**Cause:** The environment doesn't exist in the repository.

**Solution:**

Run the Go tool to add environment secrets:
```bash
./add-env-secrets-go.sh
```

Or manually:
```bash
export GITHUB_PAT="your_token"
export SONAR_TOKEN="your_token"
export GITHUB_ORG="teknex-poc"
export ENVIRONMENT_NAME="production"

./sonar-automation --add-env-secrets
```

---

### 3. GitHub PAT Permission Issues

**Error:**
```
❌ Failed to create environment: 403 Forbidden
```

**Cause:** GitHub PAT doesn't have sufficient permissions.

**Required Scopes:**
- ✅ `repo` (Full control of private repositories)
- ✅ `admin:org` → `read:org` (Read org and team membership)
- ✅ `workflow` (Update GitHub Action workflows)

**Solution:**

1. Go to: https://github.com/settings/tokens
2. Create a new token or update existing one
3. Ensure all required scopes are checked
4. Update your token in the scripts

---

### 4. Secret Names Starting with GITHUB_

**Error:**
```
422 Secret names must not start with GITHUB_
```

**Cause:** GitHub reserves the `GITHUB_` prefix for system secrets.

**Solution:**

The scripts now use `GH_PAT` instead of `GITHUB_PAT` for the secret name.
This is already fixed in the current version.

---

### 5. Workflow Not Using Environment Secrets

**Issue:** Workflow runs but doesn't use environment secrets.

**Check:**

1. Verify the workflow file has `environment:` specified:
   ```yaml
   jobs:
     sonarcloud:
       environment: production
   ```

2. Update workflows to use environment:
   ```bash
   ./update-workflows-go.sh
   ```

---

### 6. Repository vs Environment Secrets

**Understanding the difference:**

- **Repository Secrets** (`/repos/{org}/{repo}/actions/secrets`)
  - Available to all workflows
  - No environment protection rules
  - Simpler setup

- **Environment Secrets** (`/repos/{org}/{repo}/environments/{env}/secrets`)
  - Scoped to specific environment
  - Can have protection rules (approvals, wait timers)
  - Better for production deployments
  - **This is what we're using now**

**Current Setup:**

Both repositories have:
- 📋 Repository-level: `SONAR_TOKEN` (old)
- 🌍 Environment 'production': `GH_PAT`, `SONAR_TOKEN` (new)

The workflows now use the environment secrets.

---

## Verification Commands

### List All Secrets
```bash
./list-secrets.sh
```

### Add Environment Secrets
```bash
./add-env-secrets-go.sh
```

### Update Workflows to Use Environment
```bash
./update-workflows-go.sh
```

### Run Full Setup (New Repos)
```bash
./run.sh
```

---

## Manual Verification

### Check Environment in GitHub UI

1. Go to repository settings:
   ```
   https://github.com/teknex-poc/{repo}/settings/environments
   ```

2. Click on `production` environment

3. Verify secrets exist:
   - `GH_PAT`
   - `SONAR_TOKEN`

### Check Workflow File

View the workflow:
```
https://github.com/teknex-poc/{repo}/blob/main/.github/workflows/sonar.yml
```

Should contain:
```yaml
jobs:
  sonarcloud:
    environment: production
```

---

## Getting Help

If you encounter other issues:

1. Check GitHub Actions logs
2. Verify all environment variables are set
3. Ensure tokens haven't expired
4. Check repository permissions
5. Review SonarCloud project settings

