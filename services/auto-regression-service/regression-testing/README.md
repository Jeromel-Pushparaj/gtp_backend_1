# Regression Testing Middleman Service

A Go-based middleman service that aggregates test results from the regression testing API.

## Features

- Accepts GitHub URL, PAT token, and branch as request parameters
- Calls the regression testing API at `http://localhost:8080/api/v1/github/test`
- Returns aggregated results showing:
  - Unique test cases with their status codes
  - Number of test cases passed
  - Number of test cases failed
  - Number of test cases skipped
  - Pass rate and execution details

## Prerequisites

- Go 1.21 or higher
- The main regression testing service running on `http://localhost:8080`

## Installation

```bash
go mod download
```

## Running the Service

```bash
go run main.go
```

The service will start on port `8081` by default.

## API Endpoints

### POST /api/v1/test/aggregate

Triggers test execution and returns aggregated results.

**Request Body:**
```json
{
  "github_url": "https://github.com/username/repo",
  "pat_token": "your_github_pat_token",
  "branch": "main"
}
```

**Response:**
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

### GET /health

Health check endpoint.

**Response:**
```json
{
  "status": "healthy",
  "time": "2026-03-03T14:23:32Z"
}
```

## Usage Example

```bash
# Set your variables
GITHUB_URL="https://github.com/username/repo"
PAT_TOKEN="your_github_pat_token"
BRANCH="main"

# Call the middleman service
curl -X POST http://localhost:8081/api/v1/test/aggregate \
  -H "Content-Type: application/json" \
  -d "{
    \"github_url\": \"$GITHUB_URL\",
    \"pat_token\": \"$PAT_TOKEN\",
    \"branch\": \"$BRANCH\"
  }"
```

## Response Fields

- **unique_test_cases**: Array of all test cases with their results
  - **name**: Test case name
  - **status_code**: HTTP status code returned
  - **passed**: Whether the test passed
  - **skipped**: Whether the test was skipped
  - **category**: Test category (positive, negative, boundary)
  - **method**: HTTP method used
  - **path**: Resolved API path
- **total_tests**: Total number of tests executed
- **tests_passed**: Number of tests that passed
- **tests_failed**: Number of tests that failed
- **tests_skipped**: Number of tests that were skipped
- **pass_rate**: Percentage of tests that passed
- **executed_at**: Timestamp of execution
- **duration_ns**: Total execution time in nanoseconds

## Building

```bash
go build -o middleman main.go
./middleman
```

## Docker Support (Optional)

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o middleman main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/middleman .
EXPOSE 8081
CMD ["./middleman"]
```

Build and run:
```bash
docker build -t middleman .
docker run -p 8081:8081 middleman
```

