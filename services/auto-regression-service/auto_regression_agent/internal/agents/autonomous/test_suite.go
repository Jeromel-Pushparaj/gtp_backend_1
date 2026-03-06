package autonomous

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// TestSuite represents a persisted, reusable test suite
type TestSuite struct {
	// Metadata
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Version     string    `json:"version"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Source information
	SpecID   string `json:"spec_id"`
	SpecName string `json:"spec_name"`
	BaseURL  string `json:"base_url"`

	// Test configuration
	Tests []PersistedTest `json:"tests"`

	// Schema definitions for validation (extracted from OpenAPI)
	Schemas map[string]SchemaDefinition `json:"schemas,omitempty"`

	// Statistics
	Statistics SuiteStatistics `json:"statistics"`
}

// PersistedTest represents a single test case that can be saved and reloaded
type PersistedTest struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"` // positive, negative, boundary, security

	// Request definition
	Method      string                 `json:"method"`
	Path        string                 `json:"path"`
	PathParams  map[string]interface{} `json:"path_params,omitempty"`
	QueryParams map[string]interface{} `json:"query_params,omitempty"`
	Headers     map[string]interface{} `json:"headers,omitempty"`
	Payload     interface{}            `json:"payload,omitempty"`

	// Expected response
	ExpectedStatus     int                 `json:"expected_status"`
	ExpectedSchema     string              `json:"expected_schema,omitempty"`  // Reference to schema in Schemas map
	ExpectedFields     []ExpectedField     `json:"expected_fields,omitempty"`  // Specific field validations
	ExpectedHeaders    map[string]string   `json:"expected_headers,omitempty"` // Expected response headers
	ResponseValidation *ResponseValidation `json:"response_validation,omitempty"`

	// Context handling
	ContextRequired *ContextRequired `json:"context_required,omitempty"`
	ContextCapture  *ContextCapture  `json:"context_capture,omitempty"`

	// Execution options
	DependsOn  []string `json:"depends_on,omitempty"` // IDs of tests this depends on
	Timeout    int      `json:"timeout_ms,omitempty"` // Custom timeout in milliseconds
	Retries    int      `json:"retries,omitempty"`    // Number of retries on failure
	Skip       bool     `json:"skip,omitempty"`       // Skip this test
	SkipReason string   `json:"skip_reason,omitempty"`

	// Tags for filtering
	Tags []string `json:"tags,omitempty"`
}

// ExpectedField defines validation for a specific response field
type ExpectedField struct {
	Path     string      `json:"path"`               // JSON path (e.g., "data.user.id")
	Type     string      `json:"type,omitempty"`     // Expected type: string, number, boolean, array, object
	Required bool        `json:"required,omitempty"` // Field must be present
	Value    interface{} `json:"value,omitempty"`    // Exact value match
	Pattern  string      `json:"pattern,omitempty"`  // Regex pattern for string fields
	MinLen   *int        `json:"min_length,omitempty"`
	MaxLen   *int        `json:"max_length,omitempty"`
	Min      *float64    `json:"min,omitempty"` // Minimum value for numbers
	Max      *float64    `json:"max,omitempty"` // Maximum value for numbers
}

// ResponseValidation defines how to validate the response
type ResponseValidation struct {
	ValidateSchema   bool `json:"validate_schema"`    // Validate against OpenAPI schema
	ValidateRequired bool `json:"validate_required"`  // Check all required fields present
	ValidateTypes    bool `json:"validate_types"`     // Check field types match schema
	AllowExtraFields bool `json:"allow_extra_fields"` // Allow fields not in schema
}

// SchemaDefinition represents an OpenAPI schema for validation
type SchemaDefinition struct {
	Type       string                        `json:"type"`
	Properties map[string]PropertyDefinition `json:"properties,omitempty"`
	Required   []string                      `json:"required,omitempty"`
	Items      *SchemaDefinition             `json:"items,omitempty"` // For arrays
}

// PropertyDefinition represents a property in a schema
type PropertyDefinition struct {
	Type      string   `json:"type"`
	Format    string   `json:"format,omitempty"`
	Enum      []string `json:"enum,omitempty"`
	Pattern   string   `json:"pattern,omitempty"`
	MinLength *int     `json:"minLength,omitempty"`
	MaxLength *int     `json:"maxLength,omitempty"`
	Minimum   *float64 `json:"minimum,omitempty"`
	Maximum   *float64 `json:"maximum,omitempty"`
	Required  bool     `json:"required,omitempty"`
}

// SuiteStatistics contains statistics about the test suite
type SuiteStatistics struct {
	TotalTests    int `json:"total_tests"`
	PositiveTests int `json:"positive_tests"`
	NegativeTests int `json:"negative_tests"`
	BoundaryTests int `json:"boundary_tests"`
	SecurityTests int `json:"security_tests"`
}

// SuiteExecutionResult contains the results of executing a test suite
type SuiteExecutionResult struct {
	Suite       *TestSuite             `json:"suite"`
	Results     []PersistedTestResult  `json:"results"`
	Summary     SuiteExecutionSummary  `json:"summary"`
	ExecutedAt  time.Time              `json:"executed_at"`
	Duration    time.Duration          `json:"duration"`
	ContextData map[string]interface{} `json:"context_data"`
}

// PersistedTestResult contains the result of a single persisted test
type PersistedTestResult struct {
	Test            PersistedTest          `json:"test"`
	ResolvedPath    string                 `json:"resolved_path,omitempty"`    // Actual path after variable substitution
	ResolvedPayload interface{}            `json:"resolved_payload,omitempty"` // Actual payload after context resolution
	Passed          bool                   `json:"passed"`
	StatusCode      int                    `json:"status_code"`
	ResponseBody    interface{}            `json:"response_body,omitempty"`
	ResponseTime    time.Duration          `json:"response_time"`
	Error           string                 `json:"error,omitempty"`
	ContextCaptured map[string]interface{} `json:"context_captured,omitempty"`
	Skipped         bool                   `json:"skipped"`
	SkipReason      string                 `json:"skip_reason,omitempty"`
	Validation      *ValidationResult      `json:"validation,omitempty"`
}

// SuiteExecutionSummary contains summary statistics for suite execution
type SuiteExecutionSummary struct {
	TotalTests       int           `json:"total_tests"`
	Passed           int           `json:"passed"`
	Failed           int           `json:"failed"`
	Skipped          int           `json:"skipped"`
	ValidationErrors int           `json:"validation_errors"`
	PassRate         float64       `json:"pass_rate"`
	TotalTime        time.Duration `json:"total_time"`
}

// SaveTestSuite saves a test suite to a JSON file
func SaveTestSuite(suite *TestSuite, outputDir string) (string, error) {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	filename := fmt.Sprintf("%s-test-suite.json", suite.ID)
	filepath := filepath.Join(outputDir, filename)

	suite.UpdatedAt = time.Now()

	data, err := json.MarshalIndent(suite, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal test suite: %w", err)
	}

	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write test suite file: %w", err)
	}

	return filepath, nil
}

// LoadTestSuite loads a test suite from a JSON file
func LoadTestSuite(filepath string) (*TestSuite, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read test suite file: %w", err)
	}

	var suite TestSuite
	if err := json.Unmarshal(data, &suite); err != nil {
		return nil, fmt.Errorf("failed to unmarshal test suite: %w", err)
	}

	return &suite, nil
}

// ConvertFromExecutionPlan converts an ExecutionPlan to a TestSuite
func ConvertFromExecutionPlan(plan *ExecutionPlan, specName, baseURL string) *TestSuite {
	suite := &TestSuite{
		ID:          plan.ID,
		Name:        fmt.Sprintf("Test Suite - %s", specName),
		Description: plan.AIReasoning,
		Version:     "1.0.0",
		CreatedAt:   plan.CreatedAt,
		UpdatedAt:   time.Now(),
		SpecID:      plan.SpecID,
		SpecName:    specName,
		BaseURL:     baseURL,
		Tests:       make([]PersistedTest, len(plan.Tests)),
		Schemas:     make(map[string]SchemaDefinition),
	}

	// Convert tests
	for i, pt := range plan.Tests {
		suite.Tests[i] = PersistedTest{
			ID:              fmt.Sprintf("test-%d", i+1),
			Name:            pt.Name,
			Description:     pt.Description,
			Category:        pt.Category,
			Method:          pt.Method,
			Path:            pt.Path,
			PathParams:      pt.PathParams,
			QueryParams:     pt.QueryParams,
			Headers:         pt.Headers,
			Payload:         pt.Payload,
			ExpectedStatus:  pt.ExpectedStatus,
			ContextRequired: pt.ContextRequired,
			ContextCapture:  pt.ContextCapture,
			ResponseValidation: &ResponseValidation{
				ValidateSchema:   true,
				ValidateRequired: true,
				ValidateTypes:    true,
				AllowExtraFields: true,
			},
		}

		// Count by category
		switch pt.Category {
		case "positive":
			suite.Statistics.PositiveTests++
		case "negative":
			suite.Statistics.NegativeTests++
		case "boundary":
			suite.Statistics.BoundaryTests++
		case "security":
			suite.Statistics.SecurityTests++
		}
	}

	suite.Statistics.TotalTests = len(plan.Tests)
	return suite
}

// ConvertToExecutionPlan converts a TestSuite back to an ExecutionPlan for execution
func ConvertToExecutionPlan(suite *TestSuite) *ExecutionPlan {
	plan := &ExecutionPlan{
		ID:          suite.ID,
		SpecID:      suite.SpecID,
		WorkflowID:  "",
		CreatedAt:   suite.CreatedAt,
		AIReasoning: suite.Description,
		Tests:       make([]PlannedTest, 0, len(suite.Tests)),
	}

	for i, pt := range suite.Tests {
		if pt.Skip {
			continue // Skip tests marked as skip
		}

		plan.Tests = append(plan.Tests, PlannedTest{
			Order:           i + 1,
			Name:            pt.Name,
			Description:     pt.Description,
			Category:        pt.Category,
			Method:          pt.Method,
			Path:            pt.Path,
			PathParams:      pt.PathParams,
			QueryParams:     pt.QueryParams,
			Headers:         pt.Headers,
			Payload:         pt.Payload,
			ExpectedStatus:  pt.ExpectedStatus,
			ContextRequired: pt.ContextRequired,
			ContextCapture:  pt.ContextCapture,
		})
	}

	plan.Statistics = PlanStatistics{
		TotalTests:    len(plan.Tests),
		PositiveTests: suite.Statistics.PositiveTests,
		NegativeTests: suite.Statistics.NegativeTests,
		BoundaryTests: suite.Statistics.BoundaryTests,
	}

	return plan
}
