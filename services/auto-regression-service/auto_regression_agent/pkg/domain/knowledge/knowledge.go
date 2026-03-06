package knowledge

import (
	"time"
)

// KnowledgeBase represents the shared knowledge repository for all agents
type KnowledgeBase struct {
	SuccessPatterns  []SuccessPattern  `json:"success_patterns"`
	FailurePatterns  []FailurePattern  `json:"failure_patterns"`
	PerformanceData  []PerformanceData `json:"performance_data"`
	CoverageData     CoverageData      `json:"coverage_data"`
	AgentFeedback    []AgentFeedback   `json:"agent_feedback"`
	AdaptiveSettings AdaptiveSettings  `json:"adaptive_settings"`
	LastUpdated      time.Time         `json:"last_updated"`
}

// SuccessPattern represents a learned pattern of successful test execution
type SuccessPattern struct {
	ID              string                 `json:"id"`
	EndpointPattern string                 `json:"endpoint_pattern"` // e.g., "POST /users"
	Method          string                 `json:"method"`
	PayloadTemplate map[string]interface{} `json:"payload_template"`
	StatusCode      int                    `json:"status_code"`
	ResponseSchema  map[string]interface{} `json:"response_schema"`
	SuccessCount    int                    `json:"success_count"`
	Confidence      float64                `json:"confidence"` // 0.0 to 1.0
	FirstSeen       time.Time              `json:"first_seen"`
	LastSeen        time.Time              `json:"last_seen"`
	Tags            []string               `json:"tags"`
}

// FailurePattern represents a learned pattern of test failures
type FailurePattern struct {
	ID              string                 `json:"id"`
	EndpointPattern string                 `json:"endpoint_pattern"`
	Method          string                 `json:"method"`
	PayloadPattern  map[string]interface{} `json:"payload_pattern"`
	ErrorType       string                 `json:"error_type"` // e.g., "validation_error", "auth_error"
	ErrorMessage    string                 `json:"error_message"`
	StatusCode      int                    `json:"status_code"`
	FailureCount    int                    `json:"failure_count"`
	Confidence      float64                `json:"confidence"`
	FirstSeen       time.Time              `json:"first_seen"`
	LastSeen        time.Time              `json:"last_seen"`
	RootCause       string                 `json:"root_cause"`
	AvoidStrategy   string                 `json:"avoid_strategy"` // How to avoid this failure
	Tags            []string               `json:"tags"`
}

// PerformanceData represents performance metrics for endpoints
type PerformanceData struct {
	ID              string        `json:"id"`
	EndpointPattern string        `json:"endpoint_pattern"`
	Method          string        `json:"method"`
	AvgResponseTime time.Duration `json:"avg_response_time"`
	MinResponseTime time.Duration `json:"min_response_time"`
	MaxResponseTime time.Duration `json:"max_response_time"`
	P95ResponseTime time.Duration `json:"p95_response_time"`
	P99ResponseTime time.Duration `json:"p99_response_time"`
	ExecutionCount  int           `json:"execution_count"`
	LastUpdated     time.Time     `json:"last_updated"`
	IsSlow          bool          `json:"is_slow"` // Flagged as slow endpoint
}

// CoverageData represents test coverage information
type CoverageData struct {
	TotalEndpoints     int                      `json:"total_endpoints"`
	TestedEndpoints    int                      `json:"tested_endpoints"`
	CoveragePercentage float64                  `json:"coverage_percentage"`
	EndpointCoverage   map[string]EndpointCover `json:"endpoint_coverage"`
	StatusCodeCoverage map[int]int              `json:"status_code_coverage"`
	MethodCoverage     map[string]int           `json:"method_coverage"`
	UncoveredEndpoints []string                 `json:"uncovered_endpoints"`
	LastUpdated        time.Time                `json:"last_updated"`
}

// EndpointCover represents coverage for a specific endpoint
type EndpointCover struct {
	Endpoint        string    `json:"endpoint"`
	Method          string    `json:"method"`
	TestCount       int       `json:"test_count"`
	StatusCodesSeen []int     `json:"status_codes_seen"`
	LastTested      time.Time `json:"last_tested"`
	IsStable        bool      `json:"is_stable"` // No failures in last N runs
	StabilityScore  float64   `json:"stability_score"`
}

// AgentFeedback represents feedback from one agent to another
type AgentFeedback struct {
	ID           string                 `json:"id"`
	FromAgent    string                 `json:"from_agent"`    // e.g., "executor", "payload"
	ToAgent      string                 `json:"to_agent"`      // e.g., "payload", "seeder"
	FeedbackType string                 `json:"feedback_type"` // e.g., "improvement", "error", "success"
	Message      string                 `json:"message"`
	Data         map[string]interface{} `json:"data"`
	Priority     int                    `json:"priority"` // 1-5, 5 being highest
	Timestamp    time.Time              `json:"timestamp"`
	Acknowledged bool                   `json:"acknowledged"`
	ActionTaken  string                 `json:"action_taken"`
}

// AdaptiveSettings represents adaptive testing configuration
type AdaptiveSettings struct {
	FocusOnFailures        bool           `json:"focus_on_failures"`
	SkipStableTests        bool           `json:"skip_stable_tests"`
	StabilityThreshold     int            `json:"stability_threshold"`      // Number of consecutive successes
	FailureFocusMultiplier float64        `json:"failure_focus_multiplier"` // Multiply test count for failing endpoints
	RiskBasedPriority      bool           `json:"risk_based_priority"`
	EndpointPriorities     map[string]int `json:"endpoint_priorities"` // endpoint -> priority (1-10)
	LastUpdated            time.Time      `json:"last_updated"`
}

// LearningEvent represents an event that agents can learn from
type LearningEvent struct {
	ID           string                 `json:"id"`
	EventType    string                 `json:"event_type"` // "test_success", "test_failure", "performance_issue"
	SpecID       string                 `json:"spec_id"`
	Endpoint     string                 `json:"endpoint"`
	Method       string                 `json:"method"`
	Payload      map[string]interface{} `json:"payload"`
	Response     map[string]interface{} `json:"response"`
	StatusCode   int                    `json:"status_code"`
	ResponseTime time.Duration          `json:"response_time"`
	Success      bool                   `json:"success"`
	ErrorMessage string                 `json:"error_message"`
	Timestamp    time.Time              `json:"timestamp"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// ConsensusDecision represents a decision made by multiple agents
type ConsensusDecision struct {
	ID           string                 `json:"id"`
	DecisionType string                 `json:"decision_type"` // e.g., "payload_selection", "retry_strategy"
	Context      map[string]interface{} `json:"context"`
	Votes        []AgentVote            `json:"votes"`
	Decision     string                 `json:"decision"`
	Confidence   float64                `json:"confidence"`
	Timestamp    time.Time              `json:"timestamp"`
}

// AgentVote represents a single agent's vote in consensus
type AgentVote struct {
	AgentName  string                 `json:"agent_name"`
	Vote       string                 `json:"vote"`
	Confidence float64                `json:"confidence"`
	Reasoning  string                 `json:"reasoning"`
	Data       map[string]interface{} `json:"data"`
}
