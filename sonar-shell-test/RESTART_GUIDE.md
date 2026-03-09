# How to Restart the Shell Service with Correct Configuration

## 🎯 Quick Fix

Run these commands in the `sonar-shell-test` directory:

```bash
# 1. Set environment variables
export GITHUB_PAT="your_github_personal_access_token_here"
export GITHUB_ORG="your-github-org"
export SONAR_TOKEN="your_sonar_token_here"
export SONAR_ORG_KEY="your-sonar-org-key"

# 2. Start the server
go run main.go -server -port 8080
```

## ✅ Verify It's Working

After starting, test in a **new terminal**:

```bash
# Should return actual data (not zeros)
curl "http://localhost:8080/api/v1/github/metrics?repo=delivery-management-frontend" | jq '.data.open_issues'

# Should return a number > 0 (you have 1 open issue)
```

## 📋 What Was Wrong

The service was running **without** the environment variables, so:

- `GITHUB_PAT` was empty → couldn't authenticate with GitHub
- `GITHUB_ORG` was empty → didn't know which organization to query

This caused all API calls to fail silently and return zeros.

## 🔧 Alternative: Use .env File

Create a `.env` file in `sonar-shell-test/`:

```bash
GITHUB_PAT=your_github_personal_access_token_here
GITHUB_ORG=your-github-org
SONAR_TOKEN=your_sonar_token_here
SONAR_ORG_KEY=your-sonar-org-key
```

Then just run:

```bash
go run main.go -server -port 8080
```

The service will automatically load the `.env` file!

## 🚀 Expected Output

When you start the service, you should see:

```
✅ Database initialized at: ./data/metrics.db
🚀 Starting server on :8080
```

When you test the endpoint, you should see:

```json
{
  "success": true,
  "data": {
    "repository": "delivery-management-frontend",
    "has_readme": true,
    "default_branch": "main",
    "open_issues": 1,
    "contributors": 3
    // ... actual data, not zeros!
  }
}
```
