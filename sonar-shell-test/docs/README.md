# SonarCloud Automation Tool

A Go-based automation tool that sets up SonarCloud scanning across all repositories in a GitHub organization.

## Features

✅ **Automatic Discovery** - Lists all repositories in your GitHub organization
✅ **Smart Detection** - Checks if `sonar.yml` already exists in `.github/workflows`
✅ **Workflow Creation** - Automatically creates SonarCloud workflow if missing
✅ **Environment Secrets** - Encrypts and adds `GH_PAT` and `SONAR_TOKEN` to GitHub Environments
✅ **Secret Management** - Uses environment-based secrets for better security and control
✅ **Skip Archived** - Automatically skips archived repositories
✅ **Batch Processing** - Processes all repositories in one run
✅ **Multiple Commands** - Add secrets, update workflows, list secrets via CLI flags

## Prerequisites

- Go 1.21 or higher
- GitHub Personal Access Token (PAT) with `repo` and `admin:org` scopes
- SonarCloud Token

## Setup

### 1. Set Environment Variables

```bash
export GITHUB_PAT="your_github_personal_access_token"
export SONAR_TOKEN="your_sonarcloud_token"
export GITHUB_ORG="teknex-poc"
export SONAR_ORG_KEY="teknex-poc"  # Optional, defaults to GITHUB_ORG
export DEFAULT_BRANCH="main"        # Optional, defaults to "main"
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Run the Tool

#### Main Automation (New Repositories)
```bash
go run .
```

Or build and run:
```bash
go build -o sonar-automation
./sonar-automation
```

#### Add Environment Secrets (Existing Repositories)
```bash
./sonar-automation --add-env-secrets
```

#### Update Workflows to Use Environment
```bash
./sonar-automation --update-workflows
```

#### List All Secrets
```bash
./sonar-automation --list-secrets
```

## What It Does

For each repository in your organization, the tool:

1. **Checks** if `.github/workflows/sonar.yml` exists
2. **Creates** the workflow file if missing with proper SonarCloud configuration
3. **Encrypts** and adds these secrets to the repository:
   - `SONAR_TOKEN` - Your SonarCloud authentication token
   - `GH_PAT` - Your GitHub Personal Access Token
4. **Reports** success or failure for each repository

## Generated Workflow

The tool creates a GitHub Actions workflow that:

- Runs on every push to `main` branch
- Runs on every pull request
- Uses SonarCloud's official action
- Automatically configures project key and organization

## Example Output

```
╔══════════════════════════════════════════════════════════╗
║   SonarCloud Automation - Organization-wide Setup       ║
╚══════════════════════════════════════════════════════════╝

Organization: teknex-poc
SonarCloud Org: teknex-poc
Default Branch: main

🔍 Fetching repositories from teknex-poc...
✅ Found 5 repositories

📦 Processing repository: delivery-management-frontend
  ℹ️  Default branch: main
  ⚠️  sonar.yml not found, creating...
  ✅ sonar.yml created successfully
  🔐 Adding secrets...
    ✅ SONAR_TOKEN added
    ✅ GH_PAT added
  ✅ Repository setup complete!

╔══════════════════════════════════════════════════════════╗
║                      SUMMARY                             ║
╚══════════════════════════════════════════════════════════╝
Total repositories: 5
✅ Successfully processed: 5
⏭️  Skipped (archived/existing): 0
❌ Errors: 0
```

## Troubleshooting

### Authentication Errors
- Ensure your `GITHUB_PAT` has `repo` and `admin:org` scopes
- Check that the token hasn't expired

### 404 Errors
- Verify the organization name is correct
- Ensure your PAT has access to the organization

### Secret Creation Failures
- Confirm your PAT has `repo` scope (required for secrets)
- Check that the repository isn't archived

## Security Notes

⚠️ **Never commit tokens to version control**  
✅ Always use environment variables for sensitive data  
✅ Secrets are encrypted using GitHub's public key before transmission  
✅ The tool uses official GitHub API libraries with proper authentication  

## License

MIT

