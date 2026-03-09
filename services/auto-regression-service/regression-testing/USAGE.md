# Middleman Service Usage Guide

## Overview

This Go service acts as a middleman between your client and the regression testing API. It:
1. Accepts GitHub URL, PAT token, and branch as parameters
2. Calls the test API at `http://localhost:8080/api/v1/github/test`
3. Returns aggregated results with test statistics

## Quick Start

### 1. Start the Middleman Service

```bash
# Option 1: Run directly
go run main.go

# Option 2: Build and run
go build -o middleman main.go
./middleman
```

The service will start on **port 8081**.

### 2. Make a Request

```bash
# Set your variables
export GITHUB_URL="https://github.com/yourusername/yourrepo"
export PAT_TOKEN="your_github_pat_token"
export BRANCH="main"

# Call the middleman service
curl -X POST http://localhost:8081/api/v1/test/aggregate \
  -H "Content-Type: application/json" \
  -d "{
    \"github_url\": \"$GITHUB_URL\",
    \"pat_token\": \"$PAT_TOKEN\",
    \"branch\": \"$BRANCH\"
  }"
```

## API Endpoints

### POST /api/v1/test/aggregate

Triggers test execution and returns aggregated results.

**Request:**
```json
{
  "github_url": "https://github.com/username/repo",
  "pat_token": "ghp_xxxxxxxxxxxx",
  "branch": "main"
}
```

**Success Response (200 OK):**
```json
{
  "unique_test_cases": [
    {
      "name": "Get all employees",
      "status_code": 200,
      "passed": true,
      "skipped": false,
      "category": "positive",
      "method": "GET",
      "path": "/employees"
    },
    {
      "name": "Get employee by ID",
      "status_code": 404,
      "passed": false,
      "skipped": false,
      "category": "positive",
      "method": "GET",
      "path": "/employees/1"
    }
  ],
  "total_tests": 32,
  "tests_passed": 9,
  "tests_failed": 19,
  "tests_skipped": 4,
  "pass_rate": 28.125,
  "executed_at": "2026-03-03T14:23:32.144642+05:30",
  "duration_ns": 4329208917
}
```

**Error Response (4xx/5xx):**
```json
{
  "error": "Error message here"
}
```

### GET /health

Health check endpoint.

**Response:**
```json
{
  "status": "healthy",
  "time": "2026-03-03T14:23:32Z"
}
```

## Response Fields Explained

| Field | Type | Description |
|-------|------|-------------|
| `unique_test_cases` | Array | List of all test cases with their results |
| `total_tests` | Integer | Total number of tests executed |
| `tests_passed` | Integer | Number of tests that passed (status code matched expected) |
| `tests_failed` | Integer | Number of tests that failed |
| `tests_skipped` | Integer | Number of tests that were skipped |
| `pass_rate` | Float | Percentage of tests that passed |
| `executed_at` | String | Timestamp when tests were executed |
| `duration_ns` | Integer | Total execution time in nanoseconds |

### Test Case Fields

| Field | Type | Description |
|-------|------|-------------|
| `name` | String | Name of the test case |
| `status_code` | Integer | HTTP status code returned (0 if skipped) |
| `passed` | Boolean | Whether the test passed |
| `skipped` | Boolean | Whether the test was skipped |
| `category` | String | Test category (positive, negative, boundary) |
| `method` | String | HTTP method (GET, POST, PUT, DELETE, etc.) |
| `path` | String | Resolved API path that was tested |

## Testing

Run the included tests:

```bash
go test -v
```

## Troubleshooting

### Service won't start
- Check if port 8081 is already in use
- Try: `lsof -i :8081` to see what's using the port

### "Failed to call test API" error
- Ensure the main test API is running on `http://localhost:8080`
- Check network connectivity

### "Failed to parse test response" error
- Check the logs for detailed error messages
- The service logs the first 1000 characters of the response
- Verify the test API is returning valid JSON

### View logs
The service logs important information to stdout:
- Raw API responses (first 1000 chars)
- Parsing errors with details
- Success messages with test counts

## Examples

### Using the test script

```bash
./test-middleman.sh
```

This interactive script will:
1. Check if the service is healthy
2. Prompt for GitHub URL, PAT token, and branch
3. Make the request
4. Display formatted results

### Using the example request script

Edit `example-request.sh` with your details and run:

```bash
./example-request.sh
```

## Docker Deployment

Build and run with Docker:

```bash
docker build -t middleman .
docker run -p 8081:8081 middleman
```

## Architecture

```
Client → Middleman Service (port 8081) → Test API (port 8080)
         ↓
         Aggregates results
         ↓
Client ← Formatted response
```

