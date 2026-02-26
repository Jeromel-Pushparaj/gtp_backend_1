# Project Overview

This repository contains two solutions for automating SonarCloud setup across GitHub repositories:

## 🔧 Solution 1: Bash Script (Single Repository)

**File:** `test.sh`

### Purpose
Automates SonarCloud setup for a **single repository**.

### Features
- ✅ Checks if `sonar.yml` exists in one repository
- ✅ Creates the workflow file if missing
- ✅ Adds `SONAR_TOKEN` and `GH_PAT` as secrets via API
- ✅ Uses Python for secret encryption (PyNaCl)
- ✅ Works on macOS and Linux

### Usage
```bash
# Edit credentials in test.sh
./test.sh
```

### Pros
- Simple, single file
- No compilation needed
- Easy to understand and modify
- Good for one-off setups

### Cons
- Only processes one repository at a time
- Requires Python and PyNaCl
- Less robust error handling
- Manual configuration per repo

---

## 🚀 Solution 2: Go Application (Organization-wide)

**Files:** `main.go`, `config.go`, `github.go`, `sonar.go`, `run.sh`

### Purpose
Automates SonarCloud setup for **all repositories** in a GitHub organization.

### Features
- ✅ **Discovers all repositories** in your organization automatically
- ✅ **Batch processing** - handles multiple repos in one run
- ✅ Checks if `sonar.yml` exists in each repository
- ✅ Creates workflow files where missing
- ✅ Adds `SONAR_TOKEN` and `GH_PAT` as secrets via API
- ✅ **Skips archived repositories** automatically
- ✅ **Skips repositories** that already have sonar.yml
- ✅ Uses native Go crypto (no external dependencies)
- ✅ Comprehensive error handling and reporting
- ✅ Summary report at the end

### Usage
```bash
# Option 1: Using run script
./run.sh

# Option 2: Using environment variables
export GITHUB_PAT="your_pat"
export SONAR_TOKEN="your_token"
export GITHUB_ORG="teknex-poc"
go run .
```

### Pros
- **Organization-wide automation** - processes all repos at once
- Type-safe, compiled language
- Better error handling and reporting
- No external runtime dependencies (just Go)
- Faster execution
- Production-ready code structure
- Easy to extend and maintain
- Detailed logging and progress tracking

### Cons
- Requires Go installation
- More files to manage
- Slightly more complex setup

---

## 📊 Comparison

| Feature | Bash Script | Go Application |
|---------|-------------|----------------|
| **Scope** | Single repo | All repos in org |
| **Auto-discovery** | ❌ No | ✅ Yes |
| **Batch processing** | ❌ No | ✅ Yes |
| **Skip archived repos** | ❌ No | ✅ Yes |
| **Dependencies** | Python, PyNaCl | Go only |
| **Error handling** | Basic | Comprehensive |
| **Progress tracking** | Limited | Detailed |
| **Summary report** | ❌ No | ✅ Yes |
| **Type safety** | ❌ No | ✅ Yes |
| **Performance** | Slower | Faster |
| **Maintainability** | Medium | High |

---

## 🎯 Which One Should You Use?

### Use the **Bash Script** (`test.sh`) if:
- You need to set up SonarCloud for just **one repository**
- You want a quick, simple solution
- You're comfortable with bash and Python
- You don't mind running it multiple times for multiple repos

### Use the **Go Application** if:
- You need to set up SonarCloud for **multiple repositories**
- You want to automate across an **entire organization**
- You want **production-ready** code with proper error handling
- You prefer **type-safe**, compiled languages
- You want **detailed reporting** and progress tracking
- You plan to **maintain and extend** the tool over time

---

## 📁 Project Structure

```
sonar-shell-test/
├── test.sh              # Bash script for single repo
├── config.sh            # Config for bash script
│
├── main.go              # Go app entry point
├── config.go            # Configuration management
├── github.go            # GitHub API client
├── sonar.go             # SonarCloud workflow logic
├── go.mod               # Go dependencies
├── go.sum               # Go dependency checksums
├── run.sh               # Convenience runner for Go app
│
├── README.md            # Full documentation
├── QUICKSTART.md        # Quick start guide
├── PROJECT_OVERVIEW.md  # This file
└── .env.example         # Environment variable template
```

---

## 🔐 Security Notes

Both solutions:
- ✅ Encrypt secrets using GitHub's public key before transmission
- ✅ Use official GitHub APIs
- ✅ Support environment variables for credentials
- ⚠️ **Never commit tokens to version control**

---

## 🚦 Getting Started

### For Single Repository Setup:
```bash
./test.sh
```

### For Organization-wide Setup:
```bash
./run.sh
```

See [QUICKSTART.md](QUICKSTART.md) for detailed instructions.

