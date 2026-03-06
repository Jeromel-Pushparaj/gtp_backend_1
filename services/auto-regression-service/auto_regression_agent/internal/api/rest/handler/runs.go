package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gitlab.com/tekion/development/toc/poc/opentest/internal/orchestration"
	"gitlab.com/tekion/development/toc/poc/opentest/pkg/domain/workflow"
)

// RunsHandler handles test run related endpoints
type RunsHandler struct {
	orchestrator *orchestration.Orchestrator
}

// NewRunsHandler creates a new runs handler
func NewRunsHandler(orch *orchestration.Orchestrator) *RunsHandler {
	return &RunsHandler{
		orchestrator: orch,
	}
}

// getMetadataKeys returns the keys of a metadata map as a slice
func getMetadataKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// RunSummary represents a test run summary
type RunSummary struct {
	ID          string     `json:"id"`
	SpecName    string     `json:"spec_name"`
	SpecVersion string     `json:"spec_version"`
	Status      string     `json:"status"`   // "pending", "running", "completed", "failed"
	Phase       string     `json:"phase"`    // "phase_1", "phase_2", "phase_3"
	Progress    int        `json:"progress"` // 0-100
	TotalTests  int        `json:"total_tests"`
	PassedTests int        `json:"passed_tests"`
	FailedTests int        `json:"failed_tests"`
	StartedAt   time.Time  `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Duration    int64      `json:"duration_ms,omitempty"`
}

// RunDetails represents detailed run information
type RunDetails struct {
	RunSummary
	Endpoints        []EndpointInfo    `json:"endpoints"`
	AgentActivities  []AgentActivity   `json:"agent_activities"`
	SchemaDrifts     []SchemaDrift     `json:"schema_drifts"`
	SecurityFindings []SecurityFinding `json:"security_findings"`
}

// EndpointInfo represents endpoint information
type EndpointInfo struct {
	Method      string `json:"method"`
	Path        string `json:"path"`
	TestsCount  int    `json:"tests_count"`
	PassedCount int    `json:"passed_count"`
	FailedCount int    `json:"failed_count"`
}

// AgentActivity represents AI agent activity
type AgentActivity struct {
	Agent     string                 `json:"agent"`
	Status    string                 `json:"status"` // "pending", "active", "completed", "failed"
	Message   string                 `json:"message"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// SchemaDrift represents detected schema drift
type SchemaDrift struct {
	Endpoint   string  `json:"endpoint"`
	DriftType  string  `json:"drift_type"`
	FieldPath  string  `json:"field_path"`
	OldValue   string  `json:"old_value,omitempty"`
	NewValue   string  `json:"new_value,omitempty"`
	Confidence float64 `json:"confidence"`
	AutoFixed  bool    `json:"auto_fixed"`
}

// SecurityFinding represents security test finding
type SecurityFinding struct {
	Endpoint    string `json:"endpoint"`
	Severity    string `json:"severity"` // "critical", "high", "medium", "low"
	Category    string `json:"category"` // "sql_injection", "xss", etc.
	Description string `json:"description"`
	Payload     string `json:"payload"`
	Response    string `json:"response,omitempty"`
}

// TestResult represents a single test result
type TestResult struct {
	ID          string                 `json:"id"`
	Endpoint    string                 `json:"endpoint"`
	Method      string                 `json:"method"`
	TestType    string                 `json:"test_type"` // "smart_data", "mutation", "security", "performance"
	Status      string                 `json:"status"`    // "passed", "failed", "skipped"
	DurationMS  int64                  `json:"duration_ms"`
	Request     map[string]interface{} `json:"request"`
	Response    map[string]interface{} `json:"response"`
	Error       string                 `json:"error,omitempty"`
	Validations []ValidationResult     `json:"validations"`
}

// ValidationResult represents validation result
type ValidationResult struct {
	Type     string `json:"type"`
	Expected string `json:"expected"`
	Actual   string `json:"actual"`
	Passed   bool   `json:"passed"`
	Message  string `json:"message,omitempty"`
}

// LogEntry represents a log entry
type LogEntry struct {
	Level     string                 `json:"level"` // "debug", "info", "warn", "error"
	Message   string                 `json:"message"`
	Timestamp time.Time              `json:"timestamp"`
	Agent     string                 `json:"agent,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// ListRuns handles GET /api/v1/runs
func (h *RunsHandler) ListRuns(c *gin.Context) {
	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")

	// Fetch workflows from store
	wfStore := h.orchestrator.GetWorkflowStore()
	offset := (page - 1) * pageSize
	workflows, total, err := wfStore.List(c.Request.Context(), "", pageSize, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch runs",
		})
		return
	}

	// Convert workflows to run summaries
	runs := make([]RunSummary, 0, len(workflows))
	for _, wf := range workflows {
		// Map workflow state to status
		runStatus := mapWorkflowStateToStatus(string(wf.State))

		// Skip if status filter doesn't match
		if status != "" && runStatus != status {
			continue
		}

		// Get spec name from metadata
		specName := "unknown"
		if name, ok := wf.Metadata["spec_name"].(string); ok {
			specName = name
		}

		// Calculate ACTUAL progress from job completion (not state-based)
		progress, err := wfStore.CalculateProgress(c.Request.Context(), wf.ID)
		if err != nil {
			log.Printf("Failed to calculate progress for workflow %s: %v", wf.ID, err)
			progress = 0
		}

		run := RunSummary{
			ID:          wf.ID,
			SpecName:    specName,
			SpecVersion: "1.0.0", // TODO: Get from metadata
			Status:      runStatus,
			Phase:       mapWorkflowStateToPhase(wf.State),
			Progress:    progress,
			TotalTests:  0, // TODO: Calculate from jobs
			PassedTests: 0,
			FailedTests: 0,
			StartedAt:   wf.CreatedAt,
			CompletedAt: wf.CompletedAt,
		}

		if wf.CompletedAt != nil {
			duration := wf.CompletedAt.Sub(wf.CreatedAt)
			run.Duration = duration.Milliseconds()
		}

		runs = append(runs, run)
	}

	hasMore := (page * pageSize) < total

	c.JSON(http.StatusOK, gin.H{
		"runs":      runs,
		"page":      page,
		"page_size": pageSize,
		"total":     total,
		"has_more":  hasMore,
	})
}

func timePtr(t time.Time) *time.Time {
	return &t
}

// mapWorkflowStateToStatus maps workflow state to UI status
func mapWorkflowStateToStatus(state string) string {
	switch state {
	case "created", "queued":
		return "pending"
	case "analyzing", "generating", "validating", "executing", "comparing", "reporting":
		return "running"
	case "completed":
		return "completed"
	case "failed":
		return "failed"
	default:
		return "pending"
	}
}

// mapWorkflowStateToPhase maps workflow state to phase
func mapWorkflowStateToPhase(state interface{}) string {
	stateStr := fmt.Sprintf("%v", state)
	switch stateStr {
	case "created", "queued", "analyzing":
		return "phase_1"
	case "generating", "validating":
		return "phase_2"
	case "ready", "executing", "comparing", "reporting":
		return "phase_3"
	case "completed", "failed":
		return "phase_3"
	default:
		return "phase_1"
	}
}

// calculateProgress calculates progress percentage based on workflow state
func calculateProgress(state interface{}) int {
	stateStr := fmt.Sprintf("%v", state)
	switch stateStr {
	case "created":
		return 0
	case "queued":
		return 5
	case "analyzing":
		return 15
	case "generating":
		return 35
	case "validating":
		return 50
	case "ready":
		return 60
	case "executing":
		return 75
	case "comparing":
		return 85
	case "reporting":
		return 95
	case "completed":
		return 100
	case "failed":
		return 100
	default:
		return 0
	}
}

// GetRun handles GET /api/v1/runs/:runId
func (h *RunsHandler) GetRun(c *gin.Context) {
	runID := c.Param("runId")

	// Fetch workflow from store
	wfStore := h.orchestrator.GetWorkflowStore()
	wf, err := wfStore.Get(c.Request.Context(), runID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":  "Run not found",
			"run_id": runID,
		})
		return
	}

	// Get spec name from metadata
	specName := "unknown"
	if name, ok := wf.Metadata["spec_name"].(string); ok {
		specName = name
	}

	specVersion := "1.0.0"
	if version, ok := wf.Metadata["spec_version"].(string); ok {
		specVersion = version
	}

	// Map workflow state to status
	runStatus := mapWorkflowStateToStatus(string(wf.State))

	// Calculate ACTUAL progress from job completion (not state-based)
	progress, err := wfStore.CalculateProgress(c.Request.Context(), runID)
	if err != nil {
		log.Printf("Failed to calculate progress for workflow %s: %v", runID, err)
		progress = 0
	}

	// Create run summary
	runSummary := RunSummary{
		ID:          wf.ID,
		SpecName:    specName,
		SpecVersion: specVersion,
		Status:      runStatus,
		Phase:       mapWorkflowStateToPhase(wf.State),
		Progress:    progress,
		TotalTests:  0, // TODO: Calculate from jobs
		PassedTests: 0,
		FailedTests: 0,
		StartedAt:   wf.CreatedAt,
		CompletedAt: wf.CompletedAt,
	}

	if wf.CompletedAt != nil {
		duration := wf.CompletedAt.Sub(wf.CreatedAt)
		runSummary.Duration = duration.Milliseconds()
	}

	// Build agent activities from job_statuses and job_types metadata
	agentActivities := []AgentActivity{}
	if jobStatuses, ok := wf.Metadata["job_statuses"].(map[string]string); ok {
		// Get job types map
		jobTypes, _ := wf.Metadata["job_types"].(map[string]string)

		for jobID, status := range jobStatuses {
			// Determine agent type from job type metadata
			agent := "unknown_agent"
			message := fmt.Sprintf("Job %s: %s", jobID, status)

			// Get job type from metadata
			jobType := ""
			if jobTypes != nil {
				jobType = jobTypes[jobID]
			}

			// Map job type to agent name and create meaningful message
			switch jobType {
			case "spec_analysis":
				agent = "discovery_agent"
				if status == "completed" {
					message = fmt.Sprintf("Analyzed spec: %s", specName)
				} else if status == "running" {
					message = "Analyzing OpenAPI specification..."
				} else {
					message = "Waiting to analyze spec"
				}
			case "test_generation":
				agent = "payload_generator"
				if status == "completed" {
					message = "Generated test cases for endpoint"
				} else if status == "running" {
					message = "Generating test cases..."
				} else {
					message = "Waiting to generate tests"
				}
			case "test_execution":
				agent = "test_executor"
				if status == "completed" {
					message = "Executed test cases"
				} else if status == "running" {
					message = "Executing test cases..."
				} else {
					message = "Waiting to execute tests"
				}
			case "result_analysis":
				agent = "result_analyzer"
				if status == "completed" {
					message = "Analyzed test results"
				} else if status == "running" {
					message = "Analyzing test results..."
				} else {
					message = "Waiting to analyze results"
				}
			default:
				// Fallback to job ID pattern matching for backward compatibility
				if strings.Contains(jobID, "spec-analysis") {
					agent = "discovery_agent"
					message = fmt.Sprintf("Analyzed spec: %s", specName)
				} else if strings.Contains(jobID, "test-gen") {
					agent = "payload_generator"
					message = "Generated test cases for endpoint"
				} else if strings.Contains(jobID, "test-exec") {
					agent = "test_executor"
					message = "Executed test cases"
				}
			}

			agentActivities = append(agentActivities, AgentActivity{
				Agent:     agent,
				Status:    status,
				Message:   message,
				Details:   map[string]interface{}{"job_id": jobID, "job_type": jobType},
				Timestamp: wf.UpdatedAt,
			})
		}
	}

	// If no activities yet, add a placeholder
	if len(agentActivities) == 0 {
		agentActivities = []AgentActivity{
			{
				Agent:     "system",
				Status:    "pending",
				Message:   "Workflow created, waiting for jobs to start",
				Details:   map[string]interface{}{"endpoints_count": wf.Metadata["endpoints_count"]},
				Timestamp: wf.CreatedAt,
			},
		}
	}

	// Create run details
	run := RunDetails{
		RunSummary:       runSummary,
		Endpoints:        []EndpointInfo{}, // TODO: Fetch from jobs
		AgentActivities:  agentActivities,
		SchemaDrifts:     []SchemaDrift{},
		SecurityFindings: []SecurityFinding{},
	}

	c.JSON(http.StatusOK, run)
}

// GetRunReport handles GET /api/v1/runs/:runId/report
func (h *RunsHandler) GetRunReport(c *gin.Context) {
	runID := c.Param("runId")

	// Fetch workflow from store
	wfStore := h.orchestrator.GetWorkflowStore()
	wf, err := wfStore.Get(c.Request.Context(), runID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":  "Run not found",
			"run_id": runID,
		})
		return
	}

	// Check if workflow is completed
	if wf.State != workflow.WorkflowStateCompleted && wf.State != workflow.WorkflowStateFailed {
		c.JSON(http.StatusAccepted, gin.H{
			"error":   "Report not ready",
			"message": "Workflow is still in progress. Please wait for completion.",
			"run_id":  runID,
			"state":   string(wf.State),
			"progress": func() int {
				p, _ := wfStore.CalculateProgress(c.Request.Context(), runID)
				return p
			}(),
		})
		return
	}

	// Get spec name from metadata
	specName := "unknown"
	if name, ok := wf.Metadata["spec_name"].(string); ok {
		specName = name
	}

	specVersion := "1.0.0"
	if version, ok := wf.Metadata["spec_version"].(string); ok {
		specVersion = version
	}

	// Load actual test results from job outputs
	summary, testResults, err := h.loadTestResults(wf)
	if err != nil {
		log.Printf("Failed to load test results: %v", err)
		// Return error response
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to load test results",
			"message": err.Error(),
		})
		return
	}

	report := gin.H{
		"run_id":            runID,
		"spec_name":         specName,
		"spec_version":      specVersion,
		"summary":           summary,
		"test_results":      testResults,
		"schema_drifts":     []SchemaDrift{},     // TODO: Implement schema drift detection
		"security_findings": []SecurityFinding{}, // TODO: Implement security findings
	}

	c.JSON(http.StatusOK, report)
}

// loadTestResults loads actual test results from job outputs
func (h *RunsHandler) loadTestResults(wf *workflow.Workflow) (gin.H, []TestResult, error) {
	// Initialize summary
	summary := gin.H{
		"total_tests":   0,
		"passed_tests":  0,
		"failed_tests":  0,
		"skipped_tests": 0,
		"duration_ms":   int64(0),
	}

	var testResults []TestResult

	// Get job types from metadata to find test execution jobs
	log.Printf("DEBUG loadTestResults: workflow=%s metadata_keys=%v", wf.ID, getMetadataKeys(wf.Metadata))

	// Check the actual type of job_types
	jobTypesRaw := wf.Metadata["job_types"]
	log.Printf("DEBUG loadTestResults: job_types type=%T value=%v", jobTypesRaw, jobTypesRaw)

	jobTypes, ok := jobTypesRaw.(map[string]string)
	if !ok {
		// Try map[string]interface{} as fallback
		if jobTypesInterface, ok2 := jobTypesRaw.(map[string]interface{}); ok2 {
			log.Printf("DEBUG loadTestResults: job_types is map[string]interface{}, converting...")
			jobTypes = make(map[string]string)
			for k, v := range jobTypesInterface {
				if strVal, ok3 := v.(string); ok3 {
					jobTypes[k] = strVal
				}
			}
		} else {
			// No job types tracked yet
			log.Printf("DEBUG loadTestResults: No job_types in metadata for workflow=%s (type=%T)", wf.ID, jobTypesRaw)
			return summary, testResults, nil
		}
	}
	log.Printf("DEBUG loadTestResults: Found %d job types for workflow=%s", len(jobTypes), wf.ID)
	log.Printf("DEBUG loadTestResults: About to iterate over job types map")

	// Find all test execution jobs
	testExecCount := 0
	for jobID, jobType := range jobTypes {
		log.Printf("DEBUG loadTestResults: LOOP ITERATION - Checking job %s with type %s", jobID, jobType)

		if jobType != string(workflow.JobTypeTestExecution) {
			log.Printf("DEBUG loadTestResults: Skipping job %s (type=%s, not test_execution)", jobID, jobType)
			continue
		}

		testExecCount++
		log.Printf("DEBUG loadTestResults: Found test execution job %s (count=%d)", jobID, testExecCount)

		// Get job results from metadata
		jobResults, ok := wf.Metadata["job_results"].(map[string]interface{})
		if !ok {
			log.Printf("DEBUG loadTestResults: No job_results in workflow metadata for workflow=%s", wf.ID)
			continue
		}

		jobResult, ok := jobResults[jobID].(map[string]interface{})
		if !ok {
			log.Printf("DEBUG loadTestResults: Job %s has no result in metadata for workflow=%s", jobID, wf.ID)
			continue
		}

		log.Printf("DEBUG loadTestResults: Job %s result keys: %v", jobID, getMetadataKeys(jobResult))

		// Get report path directly from job result (worker already provides this)
		reportPath, ok := jobResult["report_path"].(string)
		if !ok {
			log.Printf("DEBUG loadTestResults: Job %s has no report_path in result for workflow=%s", jobID, wf.ID)
			continue
		}

		log.Printf("DEBUG loadTestResults: Job %s report_path=%s", jobID, reportPath)

		// Load execution report
		report, err := h.loadExecutionReport(reportPath)
		if err != nil {
			log.Printf("Failed to load report from %s: %v", reportPath, err)
			continue
		}

		// Update summary
		summary["total_tests"] = summary["total_tests"].(int) + report.Summary.Total
		summary["passed_tests"] = summary["passed_tests"].(int) + report.Summary.Passed
		summary["failed_tests"] = summary["failed_tests"].(int) + report.Summary.Failed
		summary["skipped_tests"] = summary["skipped_tests"].(int) + report.Summary.Skipped
		summary["duration_ms"] = summary["duration_ms"].(int64) + report.Duration

		// Convert execution results to test results
		for _, result := range report.Results {
			// Build request data with all available fields
			requestData := map[string]interface{}{
				"method": result.Request.Method,
				"url":    result.Request.URL,
			}

			// Add headers if present
			if len(result.Request.Headers) > 0 {
				requestData["headers"] = result.Request.Headers
			}

			// Add body if present
			if result.Request.Body != nil {
				requestData["body"] = result.Request.Body
			}

			// Add raw body if present
			if result.Request.RawBody != "" {
				requestData["raw_body"] = result.Request.RawBody
			}

			// Add path params if present
			if len(result.Request.PathParams) > 0 {
				requestData["path_params"] = result.Request.PathParams
			}

			// Add query params if present
			if len(result.Request.QueryParams) > 0 {
				requestData["query_params"] = result.Request.QueryParams
			}

			// Build response data with all available fields
			responseData := map[string]interface{}{
				"status_code":      result.Response.StatusCode,
				"response_time_ms": result.Response.ResponseTime,
			}

			// Add headers if present
			if len(result.Response.Headers) > 0 {
				responseData["headers"] = result.Response.Headers
			}

			// Add body if present
			if result.Response.Body != nil {
				responseData["body"] = result.Response.Body
			}

			// Add raw body if present
			if result.Response.RawBody != "" {
				responseData["raw_body"] = result.Response.RawBody
			}

			testResult := TestResult{
				ID:          result.TestID,
				Endpoint:    fmt.Sprintf("%s %s", result.Request.Method, result.Request.URL),
				Method:      result.Request.Method,
				TestType:    "api_test", // TODO: Get from test metadata
				Status:      string(result.Status),
				DurationMS:  result.Duration,
				Request:     requestData,
				Response:    responseData,
				Error:       result.Error,
				Validations: []ValidationResult{},
			}

			// Convert assertions to validations
			for _, assertion := range result.Assertions {
				validation := ValidationResult{
					Type:     assertion.Type,
					Expected: fmt.Sprintf("%v", assertion.Expected),
					Actual:   fmt.Sprintf("%v", assertion.Actual),
					Passed:   assertion.Passed,
					Message:  assertion.Message,
				}
				testResult.Validations = append(testResult.Validations, validation)
			}

			testResults = append(testResults, testResult)
		}
	}

	return summary, testResults, nil
}

// ExecutionReport represents the structure of execution report JSON files
type ExecutionReport struct {
	ManifestID   string `json:"manifest_id"`
	ManifestName string `json:"manifest_name"`
	StartedAt    string `json:"started_at"`
	CompletedAt  string `json:"completed_at"`
	Duration     int64  `json:"duration_ms"`
	Summary      struct {
		Total   int `json:"total"`
		Passed  int `json:"passed"`
		Failed  int `json:"failed"`
		Skipped int `json:"skipped"`
		Errors  int `json:"errors"`
	} `json:"summary"`
	Results []struct {
		TestID      string `json:"test_id"`
		TestName    string `json:"test_name"`
		Status      string `json:"status"`
		StartedAt   string `json:"started_at"`
		CompletedAt string `json:"completed_at"`
		Duration    int64  `json:"duration_ms"`
		Request     struct {
			Method      string                 `json:"method"`
			URL         string                 `json:"url"`
			Headers     map[string]string      `json:"headers,omitempty"`
			Body        map[string]interface{} `json:"body,omitempty"`
			RawBody     string                 `json:"raw_body,omitempty"`
			PathParams  map[string]string      `json:"path_params,omitempty"`
			QueryParams map[string]string      `json:"query_params,omitempty"`
		} `json:"request"`
		Response struct {
			StatusCode   int                    `json:"status_code"`
			Headers      map[string]string      `json:"headers,omitempty"`
			Body         map[string]interface{} `json:"body,omitempty"`
			RawBody      string                 `json:"raw_body,omitempty"`
			ResponseTime int64                  `json:"response_time_ms"`
		} `json:"response"`
		Assertions []struct {
			Type     string      `json:"type"`
			Expected interface{} `json:"expected"`
			Actual   interface{} `json:"actual"`
			Passed   bool        `json:"passed"`
			Message  string      `json:"message,omitempty"`
		} `json:"assertions"`
		Error string `json:"error,omitempty"`
	} `json:"results"`
}

// loadExecutionReport loads an execution report from a JSON file
func (h *RunsHandler) loadExecutionReport(path string) (*ExecutionReport, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read report file: %w", err)
	}

	var report ExecutionReport
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, fmt.Errorf("failed to unmarshal report: %w", err)
	}

	return &report, nil
}

// GetRunLogs handles GET /api/v1/runs/:runId/logs
func (h *RunsHandler) GetRunLogs(c *gin.Context) {
	runID := c.Param("runId")

	// TODO: Fetch from database/log storage
	// For now, return mock data
	logs := []LogEntry{
		{
			Level:     "info",
			Message:   "Starting Phase 1: Spec Analysis",
			Timestamp: time.Now().Add(-3 * time.Minute),
			Agent:     "spec_analyzer",
		},
		{
			Level:     "info",
			Message:   "Parsed OpenAPI spec successfully",
			Timestamp: time.Now().Add(-2*time.Minute - 50*time.Second),
			Agent:     "spec_analyzer",
			Details:   map[string]interface{}{"endpoints_found": 3},
		},
		{
			Level:     "info",
			Message:   "Starting Phase 2: Test Generation",
			Timestamp: time.Now().Add(-2 * time.Minute),
			Agent:     "payload_generator",
		},
		{
			Level:     "info",
			Message:   "Smart Data Generator: Generating realistic payloads",
			Timestamp: time.Now().Add(-1*time.Minute - 30*time.Second),
			Agent:     "smart_data_generator",
		},
		{
			Level:     "warn",
			Message:   "Schema drift detected: field 'created_at' added",
			Timestamp: time.Now().Add(-1 * time.Minute),
			Agent:     "schema_drift_detector",
			Details:   map[string]interface{}{"endpoint": "GET /users", "drift_type": "field_added"},
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"run_id": runID,
		"logs":   logs,
		"total":  len(logs),
	})
}

// DownloadReport handles GET /api/v1/runs/:runId/download
func (h *RunsHandler) DownloadReport(c *gin.Context) {
	runID := c.Param("runId")
	format := c.DefaultQuery("format", "json")

	// TODO: Generate actual report file
	// For now, return JSON

	if format == "json" {
		c.Header("Content-Disposition", "attachment; filename=report-"+runID+".json")
		c.Header("Content-Type", "application/json")

		// Reuse GetRunReport logic
		h.GetRunReport(c)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported format. Use 'json'"})
	}
}

// GetTestCases handles GET /api/v1/runs/:runId/test-cases
func (h *RunsHandler) GetTestCases(c *gin.Context) {
	runID := c.Param("runId")

	// Fetch workflow from store
	wfStore := h.orchestrator.GetWorkflowStore()
	wf, err := wfStore.Get(c.Request.Context(), runID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":  "Run not found",
			"run_id": runID,
		})
		return
	}

	// Get spec name from metadata
	specName := "unknown"
	if name, ok := wf.Metadata["spec_name"].(string); ok {
		specName = name
	}

	// Generate mock test cases for now
	// TODO: Fetch real test cases from job results
	testCases := h.generateMockTestCases(wf)

	response := gin.H{
		"run_id":           runID,
		"spec_name":        specName,
		"total_test_cases": len(testCases),
		"test_cases_by_type": gin.H{
			"smart_data":  countTestsByType(testCases, "smart_data"),
			"mutation":    countTestsByType(testCases, "mutation"),
			"security":    countTestsByType(testCases, "security"),
			"performance": countTestsByType(testCases, "performance"),
		},
		"test_cases": testCases,
	}

	c.JSON(http.StatusOK, response)
}

func countTestsByType(tests []gin.H, testType string) int {
	count := 0
	for _, test := range tests {
		if test["test_type"] == testType {
			count++
		}
	}
	return count
}

func (h *RunsHandler) generateMockTestCases(wf interface{}) []gin.H {
	// Generate realistic mock test cases
	testCases := []gin.H{
		{
			"id":          "test-001",
			"endpoint":    "/api/v3/resources",
			"method":      "POST",
			"test_type":   "smart_data",
			"category":    "happy_path",
			"description": "Create resource with valid realistic data",
			"payload": gin.H{
				"name":   "Sample Resource",
				"status": "active",
				"category": gin.H{
					"id":   1,
					"name": "Category A",
				},
				"metadata": gin.H{
					"key": "value",
				},
			},
			"expected_status": 200,
			"validations": []string{
				"Response contains resource ID",
				"Resource name matches request",
				"Status is 'active'",
			},
			"generated_by": "smart_data_generator",
			"created_at":   "2025-12-19T00:00:00Z",
		},
		{
			"id":              "test-002",
			"endpoint":        "/api/v3/resources/{id}",
			"method":          "GET",
			"test_type":       "smart_data",
			"category":        "happy_path",
			"description":     "Retrieve existing resource by ID",
			"payload":         gin.H{},
			"expected_status": 200,
			"validations": []string{
				"Response contains resource details",
				"Resource ID matches request",
			},
			"generated_by": "smart_data_generator",
			"created_at":   "2025-12-19T00:00:01Z",
		},
		{
			"id":          "test-003",
			"endpoint":    "/api/v3/resources",
			"method":      "POST",
			"test_type":   "mutation",
			"category":    "boundary_test",
			"description": "Create resource with missing required field (name)",
			"payload": gin.H{
				"status": "active",
			},
			"expected_status": 400,
			"validations": []string{
				"Response indicates missing required field",
				"Error message mentions 'name'",
			},
			"generated_by": "mutation_generator",
			"created_at":   "2025-12-19T00:00:02Z",
		},
		{
			"id":          "test-004",
			"endpoint":    "/api/v3/resources",
			"method":      "POST",
			"test_type":   "mutation",
			"category":    "type_mismatch",
			"description": "Create resource with invalid data type for ID",
			"payload": gin.H{
				"id":     "invalid_string",
				"name":   "Test Resource",
				"status": "active",
			},
			"expected_status": 400,
			"validations": []string{
				"Response indicates type error",
				"Error message mentions 'id' field",
			},
			"generated_by": "mutation_generator",
			"created_at":   "2025-12-19T00:00:03Z",
		},
		{
			"id":          "test-005",
			"endpoint":    "/api/v3/resources",
			"method":      "POST",
			"test_type":   "security",
			"category":    "sql_injection",
			"description": "Test SQL injection in resource name field",
			"payload": gin.H{
				"name":   "'; DROP TABLE resources; --",
				"status": "active",
			},
			"expected_status": 400,
			"validations": []string{
				"Request is rejected or sanitized",
				"No SQL execution occurs",
				"Response indicates invalid input",
			},
			"generated_by": "security_generator",
			"created_at":   "2025-12-19T00:00:04Z",
		},
		{
			"id":          "test-006",
			"endpoint":    "/api/v3/resources",
			"method":      "POST",
			"test_type":   "security",
			"category":    "xss_attack",
			"description": "Test XSS attack in resource name field",
			"payload": gin.H{
				"name":   "<script>alert('XSS')</script>",
				"status": "active",
			},
			"expected_status": 400,
			"validations": []string{
				"Script tags are escaped or rejected",
				"Response is safe from XSS",
			},
			"generated_by": "security_generator",
			"created_at":   "2025-12-19T00:00:05Z",
		},
		{
			"id":              "test-007",
			"endpoint":        "/api/v3/resources/search",
			"method":          "GET",
			"test_type":       "performance",
			"category":        "load_test",
			"description":     "Test response time for searching resources",
			"payload":         gin.H{},
			"expected_status": 200,
			"validations": []string{
				"Response time < 500ms",
				"Returns valid resource list",
			},
			"generated_by": "performance_generator",
			"created_at":   "2025-12-19T00:00:06Z",
		},
		{
			"id":          "test-008",
			"endpoint":    "/api/v3/orders",
			"method":      "POST",
			"test_type":   "smart_data",
			"category":    "happy_path",
			"description": "Create order with valid data",
			"payload": gin.H{
				"resourceId": 1,
				"quantity":   1,
				"orderDate":  "2025-12-20T00:00:00Z",
				"status":     "placed",
			},
			"expected_status": 200,
			"validations": []string{
				"Order is created successfully",
				"Order ID is returned",
				"Status is 'placed'",
			},
			"generated_by": "smart_data_generator",
			"created_at":   "2025-12-19T00:00:07Z",
		},
	}

	return testCases
}

// GetAgentCollaboration returns agent collaboration data for a run
func (h *RunsHandler) GetAgentCollaboration(c *gin.Context) {
	runID := c.Param("runId")

	// TODO: Implement actual collaboration data retrieval from knowledge store
	// For now, return mock data

	type AgentFeedback struct {
		ID           string                 `json:"id"`
		FromAgent    string                 `json:"from_agent"`
		ToAgent      string                 `json:"to_agent"`
		FeedbackType string                 `json:"feedback_type"`
		Message      string                 `json:"message"`
		Priority     int                    `json:"priority"`
		Timestamp    time.Time              `json:"timestamp"`
		Acknowledged bool                   `json:"acknowledged"`
		Data         map[string]interface{} `json:"data,omitempty"`
	}

	type ConsensusDecision struct {
		ID                  string                            `json:"id"`
		DecisionType        string                            `json:"decision_type"`
		ParticipatingAgents []string                          `json:"participating_agents"`
		Votes               map[string]map[string]interface{} `json:"votes"`
		FinalDecision       string                            `json:"final_decision"`
		Confidence          float64                           `json:"confidence"`
		Timestamp           time.Time                         `json:"timestamp"`
	}

	type AgentNode struct {
		ID               string `json:"id"`
		Name             string `json:"name"`
		Status           string `json:"status"`
		MessagesSent     int    `json:"messages_sent"`
		MessagesReceived int    `json:"messages_received"`
	}

	feedbacks := []AgentFeedback{
		{
			ID:           "fb-001",
			FromAgent:    "smart_data_generator",
			ToAgent:      "mutation_generator",
			FeedbackType: "data_pattern",
			Message:      "Found common pattern: email validation in 5 endpoints",
			Priority:     2,
			Timestamp:    time.Now().Add(-2 * time.Minute),
			Acknowledged: true,
		},
		{
			ID:           "fb-002",
			FromAgent:    "schema_drift_detector",
			ToAgent:      "automated_fix_generator",
			FeedbackType: "drift_detected",
			Message:      "Field type changed: user.age (string → number)",
			Priority:     3,
			Timestamp:    time.Now().Add(-1 * time.Minute),
			Acknowledged: true,
		},
		{
			ID:           "fb-003",
			FromAgent:    "security_generator",
			ToAgent:      "test_executor",
			FeedbackType: "security_concern",
			Message:      "SQL injection vulnerability detected in /api/search",
			Priority:     5,
			Timestamp:    time.Now().Add(-30 * time.Second),
			Acknowledged: false,
		},
	}

	decisions := []ConsensusDecision{
		{
			ID:                  "dec-001",
			DecisionType:        "test_priority",
			ParticipatingAgents: []string{"smart_data_generator", "security_generator", "performance_generator"},
			Votes: map[string]map[string]interface{}{
				"smart_data_generator":  {"option": "security_first", "confidence": 0.85},
				"security_generator":    {"option": "security_first", "confidence": 0.95},
				"performance_generator": {"option": "balanced", "confidence": 0.70},
			},
			FinalDecision: "security_first",
			Confidence:    0.83,
			Timestamp:     time.Now().Add(-3 * time.Minute),
		},
	}

	agents := []AgentNode{
		{ID: "1", Name: "smart_data_generator", Status: "active", MessagesSent: 5, MessagesReceived: 2},
		{ID: "2", Name: "mutation_generator", Status: "active", MessagesSent: 3, MessagesReceived: 4},
		{ID: "3", Name: "security_generator", Status: "active", MessagesSent: 8, MessagesReceived: 1},
		{ID: "4", Name: "schema_drift_detector", Status: "completed", MessagesSent: 4, MessagesReceived: 0},
		{ID: "5", Name: "automated_fix_generator", Status: "active", MessagesSent: 2, MessagesReceived: 3},
		{ID: "6", Name: "test_executor", Status: "idle", MessagesSent: 0, MessagesReceived: 5},
	}

	c.JSON(http.StatusOK, gin.H{
		"run_id":    runID,
		"feedbacks": feedbacks,
		"decisions": decisions,
		"agents":    agents,
	})
}
