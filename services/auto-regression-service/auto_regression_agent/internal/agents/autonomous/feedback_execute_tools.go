package autonomous

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// TestExecutionResult represents the result of executing a single test
type TestExecutionResult struct {
	TestName        string                 `json:"test_name"`
	TestID          string                 `json:"test_id"`
	Passed          bool                   `json:"passed"`
	StatusCode      int                    `json:"status_code"`
	ExpectedStatus  int                    `json:"expected_status"`
	ResponseTimeMs  int64                  `json:"response_time_ms"`
	ResponseBody    interface{}            `json:"response_body,omitempty"`
	Error           string                 `json:"error,omitempty"`
	ContextCaptured map[string]interface{} `json:"context_captured,omitempty"`
}

// ExecuteSubsetResult represents the result of executing multiple tests
type ExecuteSubsetResult struct {
	Total      int                    `json:"total"`
	Passed     int                    `json:"passed"`
	Failed     int                    `json:"failed"`
	Skipped    int                    `json:"skipped"`
	PassRate   float64                `json:"pass_rate"`
	DurationMs int64                  `json:"duration_ms"`
	Results    []TestExecutionResult  `json:"results"`
	Context    map[string]interface{} `json:"context,omitempty"`
}

// ExecuteFailedResult represents the result of re-running failed tests
type ExecuteFailedResult struct {
	OriginallyFailed int                   `json:"originally_failed"`
	NowPassed        int                   `json:"now_passed"`
	StillFailing     int                   `json:"still_failing"`
	Results          []TestExecutionResult `json:"results"`
	Improvement      string                `json:"improvement"`
}

// TestFilter specifies criteria for filtering tests
type TestFilter struct {
	TestIDs   []string `json:"test_ids,omitempty"`
	TestNames []string `json:"test_names,omitempty"`
	Methods   []string `json:"methods,omitempty"`
	Paths     []string `json:"paths,omitempty"`
	Tags      []string `json:"tags,omitempty"`
}

// ============================================================================
// ExecuteSingleTestTool - Execute one specific test case from a suite
// ============================================================================

type ExecuteSingleTestTool struct {
	httpClient  *http.Client
	testContext *TestContext
}

func (t *ExecuteSingleTestTool) Name() string { return "execute_single_test" }

func (t *ExecuteSingleTestTool) Description() string {
	return "Execute one specific test case from the suite. Useful for testing after fixing a specific test case, debugging a single failing test, or validating a fix before running full suite."
}

func (t *ExecuteSingleTestTool) Parameters() ToolParameters {
	return ToolParameters{
		Type: "object",
		Properties: map[string]ToolProperty{
			"workflow_id": {
				Type:        "string",
				Description: "Workflow ID to find the test suite",
			},
			"file_path": {
				Type:        "string",
				Description: "Direct path to the test suite file (alternative to workflow_id)",
			},
			"test_id": {
				Type:        "string",
				Description: "ID of the test case to execute",
			},
			"test_order": {
				Type:        "integer",
				Description: "Order number (1-based) of the test to execute",
			},
			"test_name": {
				Type:        "string",
				Description: "Name of the test case to execute (partial match supported)",
			},
			"base_url": {
				Type:        "string",
				Description: "Base URL for API calls (required)",
			},
			"timeout_seconds": {
				Type:        "integer",
				Description: "Request timeout in seconds (default: 30)",
			},
			"headers": {
				Type:        "object",
				Description: "Additional headers to include in requests",
			},
		},
		Required: []string{"base_url"},
	}
}

func (t *ExecuteSingleTestTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	workflowID := getString(args, "workflow_id")
	filePath := getString(args, "file_path")
	testID := getString(args, "test_id")
	testOrder := getInt(args, "test_order")
	testName := getString(args, "test_name")
	baseURL := getString(args, "base_url")
	timeoutSeconds := getInt(args, "timeout_seconds")
	headers, _ := args["headers"].(map[string]interface{})

	if baseURL == "" {
		return nil, fmt.Errorf("base_url is required")
	}

	// Load test suite
	suite, suitePath, err := loadTestSuiteFromArgs(workflowID, filePath)
	if err != nil {
		return nil, err
	}

	// Find the test case
	test, err := findTestCase(suite, testID, testOrder, testName)
	if err != nil {
		return nil, err
	}

	// Initialize execution context
	if t.testContext == nil {
		t.testContext = NewTestContext()
	}
	if t.httpClient == nil {
		t.httpClient = &http.Client{Timeout: 30 * time.Second}
	}

	timeout := 30 * time.Second
	if timeoutSeconds > 0 {
		timeout = time.Duration(timeoutSeconds) * time.Second
	}
	t.httpClient.Timeout = timeout

	// Execute the test
	result := executeSingleTest(ctx, t.httpClient, t.testContext, *test, baseURL, headers)

	return map[string]interface{}{
		"success":     true,
		"suite_path":  suitePath,
		"test_result": result,
	}, nil
}

// ============================================================================
// ExecuteTestSubsetTool - Execute multiple tests matching criteria
// ============================================================================

type ExecuteTestSubsetTool struct {
	httpClient  *http.Client
	testContext *TestContext
}

func (t *ExecuteTestSubsetTool) Name() string { return "execute_test_subset" }

func (t *ExecuteTestSubsetTool) Description() string {
	return "Execute multiple tests matching filter criteria. Useful for running all tests for an endpoint, running only POST/PUT tests, or running tests by tag."
}

func (t *ExecuteTestSubsetTool) Parameters() ToolParameters {
	return ToolParameters{
		Type: "object",
		Properties: map[string]ToolProperty{
			"workflow_id": {
				Type:        "string",
				Description: "Workflow ID to find the test suite",
			},
			"file_path": {
				Type:        "string",
				Description: "Direct path to the test suite file",
			},
			"base_url": {
				Type:        "string",
				Description: "Base URL for API calls (required)",
			},
			"filter": {
				Type:        "object",
				Description: "Filter criteria: test_ids (array), test_names (array), methods (array like GET, POST), paths (array of path patterns), tags (array)",
			},
			"order": {
				Type:        "string",
				Description: "Execution order: 'sequential' or 'dependency' (default: dependency)",
				Enum:        []string{"sequential", "dependency"},
			},
			"stop_on_failure": {
				Type:        "boolean",
				Description: "Stop after first failure (default: false)",
			},
			"timeout_seconds": {
				Type:        "integer",
				Description: "Request timeout in seconds (default: 30)",
			},
			"headers": {
				Type:        "object",
				Description: "Additional headers to include in requests",
			},
		},
		Required: []string{"base_url"},
	}
}

func (t *ExecuteTestSubsetTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	workflowID := getString(args, "workflow_id")
	filePath := getString(args, "file_path")
	baseURL := getString(args, "base_url")
	filterData, _ := args["filter"].(map[string]interface{})
	orderMode := getString(args, "order")
	stopOnFailure, _ := args["stop_on_failure"].(bool)
	timeoutSeconds := getInt(args, "timeout_seconds")
	headers, _ := args["headers"].(map[string]interface{})

	if baseURL == "" {
		return nil, fmt.Errorf("base_url is required")
	}

	if orderMode == "" {
		orderMode = "dependency"
	}

	// Load test suite
	suite, suitePath, err := loadTestSuiteFromArgs(workflowID, filePath)
	if err != nil {
		return nil, err
	}

	// Parse filter
	filter := parseTestFilter(filterData)

	// Filter tests
	filteredTests := filterTests(suite.Tests, filter)
	if len(filteredTests) == 0 {
		return map[string]interface{}{
			"success":    true,
			"suite_path": suitePath,
			"message":    "No tests matched the filter criteria",
			"result":     ExecuteSubsetResult{Total: 0, PassRate: 100.0},
		}, nil
	}

	// Initialize execution context
	if t.testContext == nil {
		t.testContext = NewTestContext()
	}
	if t.httpClient == nil {
		t.httpClient = &http.Client{Timeout: 30 * time.Second}
	}

	timeout := 30 * time.Second
	if timeoutSeconds > 0 {
		timeout = time.Duration(timeoutSeconds) * time.Second
	}
	t.httpClient.Timeout = timeout

	// Execute tests
	result := executeTestSubset(ctx, t.httpClient, t.testContext, filteredTests, baseURL, headers, stopOnFailure)

	return map[string]interface{}{
		"success":    true,
		"suite_path": suitePath,
		"result":     result,
	}, nil
}

// ============================================================================
// ExecuteFailedTestsTool - Re-run only tests that failed in a previous run
// ============================================================================

type ExecuteFailedTestsTool struct {
	httpClient  *http.Client
	testContext *TestContext
}

func (t *ExecuteFailedTestsTool) Name() string { return "execute_failed_tests" }

func (t *ExecuteFailedTestsTool) Description() string {
	return "Re-run only tests that failed in a previous run. Useful for quick validation after fixes, iterative debugging, and comparing before/after results."
}

func (t *ExecuteFailedTestsTool) Parameters() ToolParameters {
	return ToolParameters{
		Type: "object",
		Properties: map[string]ToolProperty{
			"workflow_id": {
				Type:        "string",
				Description: "Workflow ID to find the test suite and results",
			},
			"file_path": {
				Type:        "string",
				Description: "Direct path to the test suite file",
			},
			"results_path": {
				Type:        "string",
				Description: "Path to previous results file (optional - will auto-detect from workflow_id)",
			},
			"base_url": {
				Type:        "string",
				Description: "Base URL for API calls (required)",
			},
			"include_dependencies": {
				Type:        "boolean",
				Description: "Also run tests that failed tests depend on (default: true)",
			},
			"timeout_seconds": {
				Type:        "integer",
				Description: "Request timeout in seconds (default: 30)",
			},
			"headers": {
				Type:        "object",
				Description: "Additional headers to include in requests",
			},
		},
		Required: []string{"base_url"},
	}
}

func (t *ExecuteFailedTestsTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	workflowID := getString(args, "workflow_id")
	filePath := getString(args, "file_path")
	resultsPath := getString(args, "results_path")
	baseURL := getString(args, "base_url")
	includeDeps := true
	if v, ok := args["include_dependencies"].(bool); ok {
		includeDeps = v
	}
	timeoutSeconds := getInt(args, "timeout_seconds")
	headers, _ := args["headers"].(map[string]interface{})

	if baseURL == "" {
		return nil, fmt.Errorf("base_url is required")
	}

	// Load test suite
	suite, suitePath, err := loadTestSuiteFromArgs(workflowID, filePath)
	if err != nil {
		return nil, err
	}

	// Load previous results
	failedTestIDs, err := loadFailedTestIDs(workflowID, resultsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load previous results: %w", err)
	}

	if len(failedTestIDs) == 0 {
		return map[string]interface{}{
			"success":    true,
			"suite_path": suitePath,
			"message":    "No failed tests found in previous results",
			"result":     ExecuteFailedResult{OriginallyFailed: 0, Improvement: "N/A"},
		}, nil
	}

	// Get tests to execute
	testsToRun := getTestsByIDs(suite.Tests, failedTestIDs, includeDeps)

	// Initialize context
	if t.testContext == nil {
		t.testContext = NewTestContext()
	}
	if t.httpClient == nil {
		t.httpClient = &http.Client{Timeout: 30 * time.Second}
	}

	timeout := 30 * time.Second
	if timeoutSeconds > 0 {
		timeout = time.Duration(timeoutSeconds) * time.Second
	}
	t.httpClient.Timeout = timeout

	// Execute tests
	subsetResult := executeTestSubset(ctx, t.httpClient, t.testContext, testsToRun, baseURL, headers, false)

	// Calculate improvement
	nowPassed := subsetResult.Passed
	stillFailing := len(failedTestIDs) - nowPassed
	if stillFailing < 0 {
		stillFailing = 0
	}
	improvement := "0%"
	if len(failedTestIDs) > 0 {
		improvement = fmt.Sprintf("+%.1f%%", float64(nowPassed)/float64(len(failedTestIDs))*100)
	}

	result := ExecuteFailedResult{
		OriginallyFailed: len(failedTestIDs),
		NowPassed:        nowPassed,
		StillFailing:     stillFailing,
		Results:          subsetResult.Results,
		Improvement:      improvement,
	}

	return map[string]interface{}{
		"success":    true,
		"suite_path": suitePath,
		"result":     result,
	}, nil
}

// ============================================================================
// ExecuteWithContextTool - Execute a test with pre-populated context
// ============================================================================

type ExecuteWithContextTool struct {
	httpClient  *http.Client
	testContext *TestContext
}

func (t *ExecuteWithContextTool) Name() string { return "execute_with_context" }

func (t *ExecuteWithContextTool) Description() string {
	return "Execute a test with pre-populated context values. Useful for testing dependent operations, debugging context issues, or manually overriding captured values."
}

func (t *ExecuteWithContextTool) Parameters() ToolParameters {
	return ToolParameters{
		Type: "object",
		Properties: map[string]ToolProperty{
			"workflow_id": {
				Type:        "string",
				Description: "Workflow ID to find the test suite",
			},
			"file_path": {
				Type:        "string",
				Description: "Direct path to the test suite file",
			},
			"test_id": {
				Type:        "string",
				Description: "ID of the test case to execute (required)",
			},
			"base_url": {
				Type:        "string",
				Description: "Base URL for API calls (required)",
			},
			"context": {
				Type:        "object",
				Description: "Pre-populated context values (e.g., {\"pet\": {\"id\": 123}, \"user\": {\"id\": 456}})",
			},
			"timeout_seconds": {
				Type:        "integer",
				Description: "Request timeout in seconds (default: 30)",
			},
			"headers": {
				Type:        "object",
				Description: "Additional headers to include in requests",
			},
		},
		Required: []string{"test_id", "base_url", "context"},
	}
}

func (t *ExecuteWithContextTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	workflowID := getString(args, "workflow_id")
	filePath := getString(args, "file_path")
	testID := getString(args, "test_id")
	baseURL := getString(args, "base_url")
	preContext, _ := args["context"].(map[string]interface{})
	timeoutSeconds := getInt(args, "timeout_seconds")
	headers, _ := args["headers"].(map[string]interface{})

	if baseURL == "" {
		return nil, fmt.Errorf("base_url is required")
	}
	if testID == "" {
		return nil, fmt.Errorf("test_id is required")
	}
	if preContext == nil {
		return nil, fmt.Errorf("context is required")
	}

	// Load test suite
	suite, suitePath, err := loadTestSuiteFromArgs(workflowID, filePath)
	if err != nil {
		return nil, err
	}

	// Find the test case
	test, err := findTestCase(suite, testID, 0, "")
	if err != nil {
		return nil, err
	}

	// Initialize context with pre-populated values
	if t.testContext == nil {
		t.testContext = NewTestContext()
	}
	t.testContext.Reset()

	// Populate context from provided values
	contextUsed := make(map[string]interface{})
	for contextType, data := range preContext {
		if dataMap, ok := data.(map[string]interface{}); ok {
			t.testContext.StoreContext(contextType, dataMap)
			for k, v := range dataMap {
				contextUsed[contextType+"."+k] = v
			}
		}
	}

	if t.httpClient == nil {
		t.httpClient = &http.Client{Timeout: 30 * time.Second}
	}

	timeout := 30 * time.Second
	if timeoutSeconds > 0 {
		timeout = time.Duration(timeoutSeconds) * time.Second
	}
	t.httpClient.Timeout = timeout

	// Execute the test
	result := executeSingleTest(ctx, t.httpClient, t.testContext, *test, baseURL, headers)

	return map[string]interface{}{
		"success":      true,
		"suite_path":   suitePath,
		"context_used": contextUsed,
		"test_result":  result,
	}, nil
}

// ============================================================================
// Helper functions for execute tools
// ============================================================================

// loadTestSuiteFromArgs loads a test suite from workflow_id or file_path
func loadTestSuiteFromArgs(workflowID, filePath string) (*TestSuite, string, error) {
	var suitePath string
	if filePath != "" {
		suitePath = filePath
	} else if workflowID != "" {
		suitePath = findOutputFile(SuitesDir, workflowID, "test-suite")
		if suitePath == "" {
			suitePath = findOutputFile(PlansDir, workflowID, "plan")
		}
	} else {
		return nil, "", fmt.Errorf("either workflow_id or file_path is required")
	}

	if suitePath == "" {
		return nil, "", fmt.Errorf("test suite not found for workflow_id: %s", workflowID)
	}

	suite, err := LoadTestSuite(suitePath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to load test suite: %w", err)
	}

	return suite, suitePath, nil
}

// findTestCase finds a test case by ID, order, or name
func findTestCase(suite *TestSuite, testID string, testOrder int, testName string) (*PersistedTest, error) {
	for i := range suite.Tests {
		test := &suite.Tests[i]
		if testID != "" && test.ID == testID {
			return test, nil
		}
		if testOrder > 0 && i+1 == testOrder {
			return test, nil
		}
		if testName != "" && strings.Contains(strings.ToLower(test.Name), strings.ToLower(testName)) {
			return test, nil
		}
	}
	return nil, fmt.Errorf("test case not found")
}

// parseTestFilter parses filter criteria from arguments
func parseTestFilter(filterData map[string]interface{}) TestFilter {
	filter := TestFilter{}
	if filterData == nil {
		return filter
	}

	if ids, ok := filterData["test_ids"].([]interface{}); ok {
		for _, id := range ids {
			if s, ok := id.(string); ok {
				filter.TestIDs = append(filter.TestIDs, s)
			}
		}
	}
	if names, ok := filterData["test_names"].([]interface{}); ok {
		for _, name := range names {
			if s, ok := name.(string); ok {
				filter.TestNames = append(filter.TestNames, s)
			}
		}
	}
	if methods, ok := filterData["methods"].([]interface{}); ok {
		for _, method := range methods {
			if s, ok := method.(string); ok {
				filter.Methods = append(filter.Methods, strings.ToUpper(s))
			}
		}
	}
	if paths, ok := filterData["paths"].([]interface{}); ok {
		for _, path := range paths {
			if s, ok := path.(string); ok {
				filter.Paths = append(filter.Paths, s)
			}
		}
	}
	if tags, ok := filterData["tags"].([]interface{}); ok {
		for _, tag := range tags {
			if s, ok := tag.(string); ok {
				filter.Tags = append(filter.Tags, s)
			}
		}
	}
	return filter
}

// filterTests filters tests based on criteria
func filterTests(tests []PersistedTest, filter TestFilter) []PersistedTest {
	// If no filter criteria, return all tests
	if len(filter.TestIDs) == 0 && len(filter.TestNames) == 0 &&
		len(filter.Methods) == 0 && len(filter.Paths) == 0 && len(filter.Tags) == 0 {
		return tests
	}

	var filtered []PersistedTest
	for _, test := range tests {
		if matchesFilter(test, filter) {
			filtered = append(filtered, test)
		}
	}
	return filtered
}

// matchesFilter checks if a test matches filter criteria
func matchesFilter(test PersistedTest, filter TestFilter) bool {
	// Check test IDs
	if len(filter.TestIDs) > 0 {
		found := false
		for _, id := range filter.TestIDs {
			if test.ID == id {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check test names (partial match)
	if len(filter.TestNames) > 0 {
		found := false
		for _, name := range filter.TestNames {
			if strings.Contains(strings.ToLower(test.Name), strings.ToLower(name)) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check methods
	if len(filter.Methods) > 0 {
		found := false
		for _, method := range filter.Methods {
			if test.Method == method {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check paths (pattern match)
	if len(filter.Paths) > 0 {
		found := false
		for _, pattern := range filter.Paths {
			if matchPath(test.Path, pattern) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check tags
	if len(filter.Tags) > 0 {
		found := false
		for _, filterTag := range filter.Tags {
			for _, testTag := range test.Tags {
				if testTag == filterTag {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// matchPath matches a path against a pattern (supports * wildcard)
func matchPath(path, pattern string) bool {
	if pattern == "*" {
		return true
	}
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(path, prefix)
	}
	if strings.HasPrefix(pattern, "*") {
		suffix := strings.TrimPrefix(pattern, "*")
		return strings.HasSuffix(path, suffix)
	}
	return path == pattern
}

// loadFailedTestIDs loads IDs of failed tests from previous results
func loadFailedTestIDs(workflowID, resultsPath string) ([]string, error) {
	var path string
	if resultsPath != "" {
		path = resultsPath
	} else if workflowID != "" {
		path = findOutputFile(ResultsDir, workflowID, "results")
	}

	if path == "" {
		return nil, fmt.Errorf("no results file found")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Try to parse as SuiteExecutionResult
	var suiteResult SuiteExecutionResult
	if err := json.Unmarshal(data, &suiteResult); err == nil {
		var failedIDs []string
		for _, r := range suiteResult.Results {
			if !r.Passed && !r.Skipped {
				failedIDs = append(failedIDs, r.Test.ID)
			}
		}
		return failedIDs, nil
	}

	// Try to parse as ExecutionResult (planned mode)
	var planResult ExecutionResult
	if err := json.Unmarshal(data, &planResult); err == nil {
		var failedIDs []string
		for _, r := range planResult.Results {
			if !r.Passed && !r.Skipped {
				failedIDs = append(failedIDs, fmt.Sprintf("test-%d", r.Test.Order))
			}
		}
		return failedIDs, nil
	}

	return nil, fmt.Errorf("unable to parse results file")
}

// getTestsByIDs returns tests matching the given IDs, optionally including dependencies
func getTestsByIDs(tests []PersistedTest, ids []string, includeDeps bool) []PersistedTest {
	idSet := make(map[string]bool)
	for _, id := range ids {
		idSet[id] = true
	}

	// Build dependency map if needed
	depMap := make(map[string][]string)
	if includeDeps {
		for _, test := range tests {
			if len(test.DependsOn) > 0 {
				depMap[test.ID] = test.DependsOn
			}
		}
	}

	// Add dependencies recursively
	if includeDeps {
		for id := range idSet {
			addDependencies(id, depMap, idSet)
		}
	}

	// Collect matching tests in order
	var result []PersistedTest
	for _, test := range tests {
		if idSet[test.ID] {
			result = append(result, test)
		}
	}
	return result
}

// addDependencies recursively adds dependencies to the ID set
func addDependencies(id string, depMap map[string][]string, idSet map[string]bool) {
	deps, ok := depMap[id]
	if !ok {
		return
	}
	for _, dep := range deps {
		if !idSet[dep] {
			idSet[dep] = true
			addDependencies(dep, depMap, idSet)
		}
	}
}

// executeSingleTest executes a single test and returns the result
func executeSingleTest(ctx context.Context, client *http.Client, testCtx *TestContext, test PersistedTest, baseURL string, extraHeaders map[string]interface{}) TestExecutionResult {
	result := TestExecutionResult{
		TestName:       test.Name,
		TestID:         test.ID,
		ExpectedStatus: test.ExpectedStatus,
	}

	// Check if test should be skipped
	if test.Skip {
		result.Error = "Test skipped: " + test.SkipReason
		return result
	}

	// Resolve path with context - replace placeholders like {id} with context values
	resolvedPath := test.Path
	if test.ContextRequired != nil {
		ctxData := testCtx.GetContext(test.ContextRequired.Type)
		if ctxData != nil {
			for _, field := range test.ContextRequired.Fields {
				if val, ok := ctxData[field]; ok {
					placeholder := fmt.Sprintf("{%s}", field)
					resolvedPath = strings.ReplaceAll(resolvedPath, placeholder, fmt.Sprintf("%v", val))
				}
			}
		}
	}
	// Also resolve any path params
	for k, v := range test.PathParams {
		placeholder := fmt.Sprintf("{%s}", k)
		resolvedPath = strings.ReplaceAll(resolvedPath, placeholder, fmt.Sprintf("%v", v))
	}

	// Resolve payload with context
	resolvedPayload := test.Payload
	if test.ContextRequired != nil && test.Payload != nil {
		resolvedPayload = testCtx.ResolveVariables(test.Payload)
	}

	// Build full URL
	fullURL := strings.TrimSuffix(baseURL, "/") + resolvedPath

	// Add query params
	if len(test.QueryParams) > 0 {
		params := make([]string, 0, len(test.QueryParams))
		for k, v := range test.QueryParams {
			params = append(params, fmt.Sprintf("%s=%v", k, v))
		}
		fullURL += "?" + strings.Join(params, "&")
	}

	// Create request
	var bodyReader *strings.Reader
	if resolvedPayload != nil {
		bodyBytes, _ := json.Marshal(resolvedPayload)
		bodyReader = strings.NewReader(string(bodyBytes))
	} else {
		bodyReader = strings.NewReader("")
	}

	req, err := http.NewRequestWithContext(ctx, test.Method, fullURL, bodyReader)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to create request: %v", err)
		return result
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	for k, v := range test.Headers {
		if s, ok := v.(string); ok {
			req.Header.Set(k, s)
		}
	}
	for k, v := range extraHeaders {
		if s, ok := v.(string); ok {
			req.Header.Set(k, s)
		}
	}

	// Execute request
	start := time.Now()
	resp, err := client.Do(req)
	result.ResponseTimeMs = time.Since(start).Milliseconds()

	if err != nil {
		result.Error = fmt.Sprintf("Request failed: %v", err)
		return result
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode

	// Parse response body
	var respBody interface{}
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err == nil {
		result.ResponseBody = respBody
	}

	// Check if passed
	result.Passed = result.StatusCode == test.ExpectedStatus

	// Capture context if configured
	if test.ContextCapture != nil && test.ContextCapture.Enabled && result.Passed {
		if respMap, ok := respBody.(map[string]interface{}); ok {
			captured := make(map[string]interface{})
			for _, field := range test.ContextCapture.Fields {
				if val, exists := respMap[field]; exists {
					captured[field] = val
				}
			}
			if len(captured) > 0 {
				testCtx.StoreContext(test.ContextCapture.StoreAs, captured)
				result.ContextCaptured = captured
			}
		}
	}

	return result
}

// executeTestSubset executes multiple tests and returns aggregated results
func executeTestSubset(ctx context.Context, client *http.Client, testCtx *TestContext, tests []PersistedTest, baseURL string, extraHeaders map[string]interface{}, stopOnFailure bool) ExecuteSubsetResult {
	start := time.Now()
	result := ExecuteSubsetResult{
		Total:   len(tests),
		Results: make([]TestExecutionResult, 0, len(tests)),
		Context: make(map[string]interface{}),
	}

	for _, test := range tests {
		testResult := executeSingleTest(ctx, client, testCtx, test, baseURL, extraHeaders)
		result.Results = append(result.Results, testResult)

		if testResult.Passed {
			result.Passed++
		} else if testResult.Error != "" && strings.Contains(testResult.Error, "skipped") {
			result.Skipped++
		} else {
			result.Failed++
			if stopOnFailure {
				break
			}
		}
	}

	result.DurationMs = time.Since(start).Milliseconds()
	if result.Total > 0 {
		result.PassRate = float64(result.Passed) / float64(result.Total) * 100
	}

	// Include captured context
	result.Context = testCtx.GetAllContext()

	return result
}
