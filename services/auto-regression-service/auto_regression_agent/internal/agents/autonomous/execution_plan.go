package autonomous

import (
	"time"
)

// ExecutionMode defines how tests are executed
type ExecutionMode string

const (
	// ModeAutonomous - AI controls everything (current behavior)
	ModeAutonomous ExecutionMode = "autonomous"
	// ModePlanned - AI plans once, deterministic execution
	ModePlanned ExecutionMode = "planned"
)

// ExecutionPlan represents an AI-generated execution plan
type ExecutionPlan struct {
	ID          string         `json:"id"`
	SpecID      string         `json:"spec_id"`
	WorkflowID  string         `json:"workflow_id"`
	CreatedAt   time.Time      `json:"created_at"`
	AIReasoning string         `json:"ai_reasoning"`
	Tests       []PlannedTest  `json:"tests"`
	Statistics  PlanStatistics `json:"statistics"`
}

// PlanStatistics contains summary statistics for the plan
type PlanStatistics struct {
	TotalTests       int `json:"total_tests"`
	PositiveTests    int `json:"positive_tests"`
	NegativeTests    int `json:"negative_tests"`
	BoundaryTests    int `json:"boundary_tests"`
	TestsWithContext int `json:"tests_with_context"`
}

// PlannedTest represents a single test in the execution plan
type PlannedTest struct {
	Order          int                    `json:"order"`
	Name           string                 `json:"name"`
	Description    string                 `json:"description"`
	Category       string                 `json:"category"` // positive, negative, boundary
	Method         string                 `json:"method"`
	Path           string                 `json:"path"`
	PathParams     map[string]interface{} `json:"path_params,omitempty"`
	QueryParams    map[string]interface{} `json:"query_params,omitempty"`
	Headers        map[string]interface{} `json:"headers,omitempty"`
	Payload        interface{}            `json:"payload,omitempty"`
	ExpectedStatus int                    `json:"expected_status"`

	// Context handling
	ContextRequired *ContextRequired `json:"context_required,omitempty"`
	ContextCapture  *ContextCapture  `json:"context_capture,omitempty"`

	// Execution options
	SkipOnContextMissing bool   `json:"skip_on_context_missing,omitempty"`
	DependsOn            string `json:"depends_on,omitempty"` // Name of test this depends on

	// Sorting priorities (internal use, not serialized)
	MethodPriority   int `json:"-"`
	CategoryPriority int `json:"-"`
}

// ContextRequired specifies what context is needed from previous tests
type ContextRequired struct {
	Type   string   `json:"type"`   // e.g., "pet", "user", "order"
	Fields []string `json:"fields"` // e.g., ["id", "name"]
}

// ContextCapture specifies what to capture from the response
type ContextCapture struct {
	Enabled bool     `json:"enabled"`
	StoreAs string   `json:"store_as"` // e.g., "pet", "user", "order"
	Fields  []string `json:"fields"`   // e.g., ["id", "name"]
}

// ExecutionResult contains the results of a planned test execution
type ExecutionResult struct {
	Plan        *ExecutionPlan         `json:"plan"`
	Results     []PlannedTestResult    `json:"results"`
	Summary     ExecutionSummary       `json:"summary"`
	ExecutedAt  time.Time              `json:"executed_at"`
	Duration    time.Duration          `json:"duration"`
	ContextData map[string]interface{} `json:"context_data"`
}

// PlannedTestResult contains the result of a single planned test
type PlannedTestResult struct {
	Test            PlannedTest            `json:"test"`
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
}

// ExecutionSummary contains summary statistics for the execution
type ExecutionSummary struct {
	TotalTests int           `json:"total_tests"`
	Passed     int           `json:"passed"`
	Failed     int           `json:"failed"`
	Skipped    int           `json:"skipped"`
	PassRate   float64       `json:"pass_rate"`
	TotalTime  time.Duration `json:"total_time"`
	LLMCalls   int           `json:"llm_calls"` // Should be 1-2 for planned mode
}
