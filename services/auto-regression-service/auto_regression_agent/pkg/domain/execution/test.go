package execution

import "time"

// TestManifest represents a test manifest loaded from JSON
type TestManifest struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	BaseURL     string                 `json:"base_url"`
	Tests       []TestCase             `json:"tests"`
	Setup       []SetupStep            `json:"setup,omitempty"`
	Teardown    []TeardownStep         `json:"teardown,omitempty"`
	Variables   map[string]interface{} `json:"variables,omitempty"`
	Config      TestConfig             `json:"config,omitempty"`
}

// TestCase represents a single test case
type TestCase struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Endpoint    string                 `json:"endpoint"`
	Method      string                 `json:"method"`
	Headers     map[string]string      `json:"headers,omitempty"`
	PathParams  map[string]string      `json:"path_params,omitempty"`
	QueryParams map[string]string      `json:"query_params,omitempty"`
	Payload     map[string]interface{} `json:"payload,omitempty"`
	Assertions  Assertions             `json:"assertions"`
	Timeout     int                    `json:"timeout,omitempty"` // seconds
	RetryCount  int                    `json:"retry_count,omitempty"`
	DependsOn   []string               `json:"depends_on,omitempty"`
}

// SetupStep represents a setup step
type SetupStep struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Endpoint string                 `json:"endpoint"`
	Method   string                 `json:"method"`
	Payload  map[string]interface{} `json:"payload,omitempty"`
	Extract  map[string]string      `json:"extract,omitempty"` // variable_name: json_path
}

// TeardownStep represents a teardown step
type TeardownStep struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Endpoint string `json:"endpoint"`
	Method   string `json:"method"`
}

// Assertions represents test assertions
type Assertions struct {
	StatusCode     int                    `json:"status_code"`
	ResponseTime   int                    `json:"response_time,omitempty"` // max milliseconds
	Headers        map[string]string      `json:"headers,omitempty"`
	JSONSchema     map[string]interface{} `json:"json_schema,omitempty"`
	JSONPath       map[string]interface{} `json:"json_path,omitempty"` // path: expected_value
	ContainsFields []string               `json:"contains_fields,omitempty"`
}

// TestConfig represents test configuration
type TestConfig struct {
	Parallel       bool `json:"parallel"`
	MaxConcurrency int  `json:"max_concurrency,omitempty"`
	Timeout        int  `json:"timeout,omitempty"` // seconds
	StopOnFailure  bool `json:"stop_on_failure,omitempty"`
}

// TestResult represents the result of a test execution
type TestResult struct {
	TestID      string            `json:"test_id"`
	TestName    string            `json:"test_name"`
	Status      TestStatus        `json:"status"`
	StartedAt   time.Time         `json:"started_at"`
	CompletedAt time.Time         `json:"completed_at"`
	Duration    int64             `json:"duration_ms"`
	Request     RequestDetails    `json:"request"`
	Response    ResponseDetails   `json:"response"`
	Assertions  []AssertionResult `json:"assertions"`
	Error       string            `json:"error,omitempty"`
}

// TestStatus represents test execution status
type TestStatus string

const (
	TestStatusPassed  TestStatus = "passed"
	TestStatusFailed  TestStatus = "failed"
	TestStatusSkipped TestStatus = "skipped"
	TestStatusError   TestStatus = "error"
)

// RequestDetails represents HTTP request details
type RequestDetails struct {
	Method      string                 `json:"method"`
	URL         string                 `json:"url"`
	Curl        string                 `json:"curl,omitempty"`
	Headers     map[string]string      `json:"headers,omitempty"`
	QueryParams map[string]string      `json:"query_params,omitempty"`
	Body        map[string]interface{} `json:"body,omitempty"`
}

// ResponseDetails represents HTTP response details
type ResponseDetails struct {
	StatusCode   int                    `json:"status_code"`
	Headers      map[string]string      `json:"headers,omitempty"`
	Body         map[string]interface{} `json:"body,omitempty"`
	RawBody      string                 `json:"raw_body,omitempty"`
	ResponseTime int64                  `json:"response_time_ms"`
}

// AssertionResult represents the result of a single assertion
type AssertionResult struct {
	Type     string      `json:"type"` // status_code, response_time, json_schema, json_path, etc.
	Expected interface{} `json:"expected"`
	Actual   interface{} `json:"actual"`
	Passed   bool        `json:"passed"`
	Message  string      `json:"message,omitempty"`
}

// ExecutionReport represents the complete test execution report
type ExecutionReport struct {
	ManifestID   string                 `json:"manifest_id"`
	ManifestName string                 `json:"manifest_name"`
	StartedAt    time.Time              `json:"started_at"`
	CompletedAt  time.Time              `json:"completed_at"`
	Duration     int64                  `json:"duration_ms"`
	Summary      ExecutionSummary       `json:"summary"`
	Results      []TestResult           `json:"results"`
	Environment  map[string]interface{} `json:"environment,omitempty"`
}

// ExecutionSummary represents execution summary statistics
type ExecutionSummary struct {
	Total   int `json:"total"`
	Passed  int `json:"passed"`
	Failed  int `json:"failed"`
	Skipped int `json:"skipped"`
	Errors  int `json:"errors"`
}
