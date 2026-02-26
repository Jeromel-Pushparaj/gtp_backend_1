# Quick Start Guide

## Option 1: Using the Run Script (Easiest)

1. **Edit `run.sh`** and fill in your credentials:
   ```bash
   export GITHUB_PAT="your_github_pat_here"
   export SONAR_TOKEN="your_sonar_token_here"
   export GITHUB_ORG="teknex-poc"
   ```

2. **Run the script:**
   ```bash
   ./run.sh
   ```

## Option 2: Using Environment Variables

1. **Set environment variables:**
   ```bash
   export GITHUB_PAT="your_github_pat_here"
   export SONAR_TOKEN="your_sonar_token_here"
   export GITHUB_ORG="teknex-poc"
   export SONAR_ORG_KEY="teknex-poc"
   ```

2. **Download dependencies:**
   ```bash
   go mod download
   ```

3. **Run the tool:**
   ```bash
   go run .
   ```

## Option 3: Build and Run

1. **Build the binary:**
   ```bash
   go build -o sonar-automation
   ```

2. **Set environment variables and run:**
   ```bash
   export GITHUB_PAT="your_github_pat_here"
   export SONAR_TOKEN="your_sonar_token_here"
   export GITHUB_ORG="teknex-poc"
   
   ./sonar-automation
   ```

## What You Need

### 1. GitHub Personal Access Token (PAT)

Create a token at: https://github.com/settings/tokens

**Required scopes:**
- ✅ `repo` (Full control of private repositories)
- ✅ `admin:org` → `read:org` (Read org and team membership)

### 2. SonarCloud Token

Get your token from: https://sonarcloud.io/account/security

### 3. Organization Name

Your GitHub organization name (e.g., `teknex-poc`)

## Expected Output

```
╔══════════════════════════════════════════════════════════╗
║   SonarCloud Automation - Organization-wide Setup       ║
╚══════════════════════════════════════════════════════════╝

Organization: teknex-poc
SonarCloud Org: teknex-poc
Default Branch: main

🔍 Fetching repositories from teknex-poc...
✅ Found 3 repositories

📦 Processing repository: repo-1
  ℹ️  Default branch: main
  ⚠️  sonar.yml not found, creating...
  ✅ sonar.yml created successfully
  🔐 Adding secrets...
    ✅ SONAR_TOKEN added
    ✅ GH_PAT added
  ✅ Repository setup complete!

📦 Processing repository: repo-2
  ℹ️  Default branch: main
  ✅ sonar.yml already exists, skipping

╔══════════════════════════════════════════════════════════╗
║                      SUMMARY                             ║
╚══════════════════════════════════════════════════════════╝
Total repositories: 3
✅ Successfully processed: 3
⏭️  Skipped (archived/existing): 0
❌ Errors: 0
```

## Troubleshooting

### "GITHUB_PAT environment variable is required"
- Make sure you've exported the environment variable
- Check for typos in the variable name

### "Failed to list repositories: 404"
- Verify your organization name is correct
- Ensure your PAT has the required scopes
- Check that your PAT hasn't expired

### "Failed to create sonar.yml: 404"
- The repository might not exist
- Your PAT might not have access to the repository
- The default branch might be different (e.g., `master` instead of `main`)

### Dependencies download is slow
- This is normal on first run
- Go is downloading all required packages
- Subsequent runs will be much faster

## Next Steps

After running the tool:

1. ✅ Check your repositories on GitHub
2. ✅ Verify the workflow files were created: `.github/workflows/sonar.yml`
3. ✅ Verify secrets were added: Go to repo → Settings → Secrets and variables → Actions
4. ✅ Push a commit to trigger the SonarCloud scan
5. ✅ View results on SonarCloud: https://sonarcloud.io/organizations/teknex-poc

## Security Best Practices

- ❌ **Never commit `run.sh` with real tokens to git**
- ✅ Use environment variables or a `.env` file (add to `.gitignore`)
- ✅ Rotate your tokens regularly
- ✅ Use tokens with minimum required scopes
- ✅ Delete tokens when no longer needed

