# Quick Start Guide

## What Was Created

A Go-based middleman service that:
- ✅ Accepts GitHub URL, PAT token, and branch via HTTP POST
- ✅ Calls the test API at `http://localhost:8080/api/v1/github/test`
- ✅ Returns aggregated results showing:
  - Unique test cases with status codes
  - Number of tests passed
  - Number of tests failed
  - Number of tests skipped
  - Pass rate and execution details

## Files Created

```
├── main.go                 # Main service code
├── main_test.go           # Unit tests
├── go.mod                 # Go module file
├── Dockerfile             # Docker configuration
├── USAGE.md              # Detailed usage guide
├── test-middleman.sh     # Interactive test script
├── example-request.sh    # Example curl request
└── middleman             # Compiled binary
```

## Run in 3 Steps

### Step 1: Start the Middleman Service

```bash
go run main.go
```

Output:
```
2026/03/03 14:23:32 Starting middleman service on port :8081
2026/03/03 14:23:32 Endpoint: POST http://localhost:8081/api/v1/test/aggregate
2026/03/03 14:23:32 Health check: GET http://localhost:8081/health
```

### Step 2: Set Your Variables

```bash
export GITHUB_URL="https://github.com/yourusername/yourrepo"
export PAT_TOKEN="your_github_pat_token"
export BRANCH="main"
```

### Step 3: Make a Request

```bash
curl -X POST http://localhost:8081/api/v1/test/aggregate \
  -H "Content-Type: application/json" \
  -d "{
    \"github_url\": \"$GITHUB_URL\",
    \"pat_token\": \"$PAT_TOKEN\",
    \"branch\": \"$BRANCH\"
  }" | jq '.'
```

## Expected Response

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

## Key Features

### 1. Unique Test Cases
Each test case shows:
- Name and description
- HTTP status code received
- Pass/fail status
- Whether it was skipped
- Test category (positive/negative/boundary)
- HTTP method and path

### 2. Aggregated Statistics
- **total_tests**: Total number of tests
- **tests_passed**: Tests where actual status matched expected
- **tests_failed**: Tests where status didn't match
- **tests_skipped**: Tests that were skipped (e.g., missing context)
- **pass_rate**: Percentage of successful tests

### 3. Enhanced Logging
The service logs:
- Raw API responses (first 1000 chars)
- Parsing errors with details
- Success/failure counts
- Helpful debugging information

## Troubleshooting

### Error: "Failed to parse test response"

**Check the logs** - The service now logs:
1. The raw response from the API
2. The structure of the response
3. Specific parsing errors

**Common causes:**
- Test API returned an error instead of results
- Test API is not running on port 8080
- Response format doesn't match expected structure

**Solution:**
Look at the service logs (stdout) for detailed error messages showing:
- What the API actually returned
- What type of data was in the "results" field
- Specific JSON parsing errors

### Error: "Failed to call test API"

**Cause:** The test API at `http://localhost:8080` is not reachable

**Solution:**
1. Start the main test API service first
2. Verify it's running: `curl http://localhost:8080/health`

### Port Already in Use

**Solution:**
```bash
# Find what's using port 8081
lsof -i :8081

# Kill the process or change the port in main.go
```

## Testing

Run the unit tests:

```bash
go test -v
```

Expected output:
```
=== RUN   TestParseExampleResponse
    main_test.go:22: ✓ Successfully parsed test response
    main_test.go:23: Total tests: 32
    main_test.go:24: Passed: 9
    main_test.go:25: Failed: 19
    main_test.go:26: Skipped: 4
--- PASS: TestParseExampleResponse (0.00s)
=== RUN   TestAggregateResults
    main_test.go:54: ✓ Successfully aggregated results
--- PASS: TestAggregateResults (0.00s)
PASS
```

## Next Steps

1. **Read USAGE.md** for detailed API documentation
2. **Use test-middleman.sh** for interactive testing
3. **Check logs** if you encounter errors - they're very detailed now!

## Support

If you see parsing errors, the service will now log:
- The exact response received
- The type of data in problematic fields
- Suggestions for what might be wrong

Check the terminal where you ran `go run main.go` for these logs.

