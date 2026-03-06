package autonomous

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"gitlab.com/tekion/development/toc/poc/opentest/pkg/vectordb"
)

// Output directories for file operations
const (
	OutputBaseDir = "./output"
	SpecsDir      = "./output/specs"
	SuitesDir     = "./output/suites"
	ResultsDir    = "./output/results"
	PayloadsDir   = "./output/payloads"
	DiscoveryDir  = "./output/discovery"
	StrategyDir   = "./output/strategy"
	PlansDir      = "./output/plans"
	BackupDir     = "./output/backups"
)

// FeedbackTool defines the interface for feedback agent tools
type FeedbackTool interface {
	Name() string
	Description() string
	Parameters() ToolParameters
	Execute(ctx context.Context, args map[string]interface{}) (interface{}, error)
}

// FeedbackToolExecutor manages and executes feedback-specific tools
type FeedbackToolExecutor struct {
	vectorStore *vectordb.Store
	tools       map[string]FeedbackTool
}

// NewFeedbackToolExecutor creates a new feedback tool executor
func NewFeedbackToolExecutor(vectorStore *vectordb.Store) *FeedbackToolExecutor {
	executor := &FeedbackToolExecutor{
		vectorStore: vectorStore,
		tools:       make(map[string]FeedbackTool),
	}

	// Register vector DB tools (require vectorStore)
	executor.registerTool(&StoreLearnedPatternTool{store: vectorStore})
	executor.registerTool(&SearchSimilarPatternsTool{store: vectorStore})
	executor.registerTool(&StoreFailurePatternTool{store: vectorStore})
	executor.registerTool(&SearchFailureFixesTool{store: vectorStore})
	executor.registerTool(&StoreSuccessfulStrategyTool{store: vectorStore})
	executor.registerTool(&SearchStrategiesTool{store: vectorStore})
	executor.registerTool(&UpdateLearningConfidenceTool{store: vectorStore})

	// Register file read tools (no dependencies)
	executor.registerTool(&ReadTestSuiteTool{})
	executor.registerTool(&ReadTestResultsTool{})
	executor.registerTool(&ReadSpecTool{})
	executor.registerTool(&ListOutputFilesTool{})

	// Register file write tools (no dependencies)
	executor.registerTool(&EditTestCaseTool{})
	executor.registerTool(&AddTestCaseTool{})
	executor.registerTool(&RemoveTestCaseTool{})
	executor.registerTool(&UpdateTestContextTool{})

	// Register analysis tools
	executor.registerTool(&GenerateRecommendationsTool{store: vectorStore})

	// Register execute tools (for running tests)
	executor.registerTool(&ExecuteSingleTestTool{})
	executor.registerTool(&ExecuteTestSubsetTool{})
	executor.registerTool(&ExecuteFailedTestsTool{})
	executor.registerTool(&ExecuteWithContextTool{})

	return executor
}

// registerTool registers a tool with the executor
func (fte *FeedbackToolExecutor) registerTool(tool FeedbackTool) {
	fte.tools[tool.Name()] = tool
}

// GetToolDefinitions returns all tool definitions for LLM function calling
func (fte *FeedbackToolExecutor) GetToolDefinitions() []ToolDefinition {
	definitions := make([]ToolDefinition, 0, len(fte.tools))
	for _, tool := range fte.tools {
		definitions = append(definitions, ToolDefinition{
			Name:        tool.Name(),
			Description: tool.Description(),
			Parameters:  tool.Parameters(),
		})
	}
	return definitions
}

// ExecuteTool executes a feedback tool by name
func (fte *FeedbackToolExecutor) ExecuteTool(ctx context.Context, call ToolCall) ToolResponse {
	tool, exists := fte.tools[call.Name]
	if !exists {
		return ToolResponse{
			Name:  call.Name,
			Error: fmt.Sprintf("unknown feedback tool: %s", call.Name),
		}
	}

	result, err := tool.Execute(ctx, call.Args)
	if err != nil {
		return ToolResponse{
			Name:  call.Name,
			Error: err.Error(),
		}
	}

	return ToolResponse{
		Name:   call.Name,
		Result: result,
	}
}

// HasTool checks if a tool exists
func (fte *FeedbackToolExecutor) HasTool(name string) bool {
	_, exists := fte.tools[name]
	return exists
}

// ============================================================================
// StoreLearnedPatternTool - Store a learned pattern in the vector database
// ============================================================================

type StoreLearnedPatternTool struct {
	store *vectordb.Store
}

func (t *StoreLearnedPatternTool) Name() string { return "store_learned_pattern" }

func (t *StoreLearnedPatternTool) Description() string {
	return "Store a learned pattern from user feedback or test analysis. Use this to remember auth patterns, edge cases, business rules, or API patterns for future use."
}

func (t *StoreLearnedPatternTool) Parameters() ToolParameters {
	return ToolParameters{
		Type: "object",
		Properties: map[string]ToolProperty{
			"category": {
				Type:        "string",
				Description: "Category of the learning",
				Enum:        []string{"auth_pattern", "edge_case", "business_rule", "api_pattern", "validation_rule"},
			},
			"content": {
				Type:        "string",
				Description: "The actual learning or pattern description",
			},
			"source_api": {
				Type:        "string",
				Description: "The API or spec this was learned from (optional)",
			},
			"confidence": {
				Type:        "number",
				Description: "Confidence score from 0.0 to 1.0 (default: 0.7)",
			},
			"context": {
				Type:        "object",
				Description: "Additional context like endpoint, method, etc.",
			},
		},
		Required: []string{"category", "content"},
	}
}

func (t *StoreLearnedPatternTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	category, _ := args["category"].(string)
	content, _ := args["content"].(string)
	sourceAPI, _ := args["source_api"].(string)
	confidence := 0.7
	if c, ok := args["confidence"].(float64); ok {
		confidence = c
	}

	var contextData map[string]interface{}
	if c, ok := args["context"].(map[string]interface{}); ok {
		contextData = c
	}

	learning := &vectordb.Learning{
		Category:   category,
		Content:    content,
		SourceAPI:  sourceAPI,
		Confidence: confidence,
		Context:    contextData,
	}

	// Check if vector store is available
	if t.store == nil {
		return map[string]interface{}{
			"success": false,
			"message": "Vector store not available (disabled in config)",
		}, nil
	}

	if err := t.store.StoreLearning(ctx, learning); err != nil {
		return nil, fmt.Errorf("failed to store learning: %w", err)
	}

	return map[string]interface{}{
		"success": true,
		"id":      learning.ID.String(),
		"message": fmt.Sprintf("Stored learning: %s", category),
	}, nil
}

// ============================================================================
// SearchSimilarPatternsTool - Search for similar learned patterns
// ============================================================================

type SearchSimilarPatternsTool struct {
	store *vectordb.Store
}

func (t *SearchSimilarPatternsTool) Name() string { return "search_similar_patterns" }

func (t *SearchSimilarPatternsTool) Description() string {
	return "Search for previously learned patterns similar to the query. Use this to find relevant learnings for a new situation."
}

func (t *SearchSimilarPatternsTool) Parameters() ToolParameters {
	return ToolParameters{
		Type: "object",
		Properties: map[string]ToolProperty{
			"query": {
				Type:        "string",
				Description: "Natural language description of what you're looking for",
			},
			"category": {
				Type:        "string",
				Description: "Filter by category (optional)",
				Enum:        []string{"auth_pattern", "edge_case", "business_rule", "api_pattern", "validation_rule"},
			},
			"limit": {
				Type:        "integer",
				Description: "Maximum number of results (default: 5)",
			},
		},
		Required: []string{"query"},
	}
}

func (t *SearchSimilarPatternsTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	query, _ := args["query"].(string)
	category, _ := args["category"].(string)
	limit := 5
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}

	opts := vectordb.SearchOptions{
		Limit:    limit,
		MinScore: 0.5,
		Category: category,
	}

	// Check if vector store is available
	if t.store == nil {
		return map[string]interface{}{
			"success": false,
			"results": []interface{}{},
			"message": "Vector store not available (disabled in config)",
		}, nil
	}

	results, err := t.store.SearchSimilarLearnings(ctx, query, opts)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	// Convert to simple response format
	patterns := make([]map[string]interface{}, len(results))
	for i, r := range results {
		patterns[i] = map[string]interface{}{
			"id":         r.ID.String(),
			"category":   r.Category,
			"content":    r.Content,
			"confidence": r.Confidence,
			"score":      r.Score,
		}
	}

	return map[string]interface{}{
		"found":    len(patterns),
		"patterns": patterns,
	}, nil
}

// ============================================================================
// StoreFailurePatternTool - Store a failure pattern with its fix
// ============================================================================

type StoreFailurePatternTool struct {
	store *vectordb.Store
}

func (t *StoreFailurePatternTool) Name() string { return "store_failure_pattern" }

func (t *StoreFailurePatternTool) Description() string {
	return "Store a failure pattern and its fix. Use this when you learn how to fix a specific type of failure."
}

func (t *StoreFailurePatternTool) Parameters() ToolParameters {
	return ToolParameters{
		Type: "object",
		Properties: map[string]ToolProperty{
			"failure_type": {
				Type:        "string",
				Description: "Type of failure",
				Enum:        []string{"auth_failure", "validation_error", "timeout", "schema_mismatch", "rate_limit", "connection_error"},
			},
			"error_signature": {
				Type:        "string",
				Description: "The error message or pattern that identifies this failure",
			},
			"error_code": {
				Type:        "string",
				Description: "HTTP status code or error code (e.g., '401', '422')",
			},
			"fix_description": {
				Type:        "string",
				Description: "How to fix or avoid this failure",
			},
			"root_cause": {
				Type:        "string",
				Description: "The identified root cause of the failure",
			},
			"endpoint_pattern": {
				Type:        "string",
				Description: "The endpoint pattern where this occurs (e.g., '/users/{id}')",
			},
			"http_method": {
				Type:        "string",
				Description: "HTTP method (GET, POST, etc.)",
			},
		},
		Required: []string{"failure_type", "error_signature", "fix_description"},
	}
}

func (t *StoreFailurePatternTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	pattern := &vectordb.FailurePattern{
		FailureType:     getString(args, "failure_type"),
		ErrorSignature:  getString(args, "error_signature"),
		ErrorCode:       getString(args, "error_code"),
		FixDescription:  getString(args, "fix_description"),
		RootCause:       getString(args, "root_cause"),
		EndpointPattern: getString(args, "endpoint_pattern"),
		HTTPMethod:      getString(args, "http_method"),
	}

	// Check if vector store is available
	if t.store == nil {
		return map[string]interface{}{
			"success": false,
			"message": "Vector store not available (disabled in config)",
		}, nil
	}

	if err := t.store.StoreFailurePattern(ctx, pattern); err != nil {
		return nil, fmt.Errorf("failed to store failure pattern: %w", err)
	}

	return map[string]interface{}{
		"success": true,
		"id":      pattern.ID.String(),
		"message": fmt.Sprintf("Stored failure pattern: %s", pattern.FailureType),
	}, nil
}

// ============================================================================
// SearchFailureFixesTool - Search for fixes for similar failures
// ============================================================================

type SearchFailureFixesTool struct {
	store *vectordb.Store
}

func (t *SearchFailureFixesTool) Name() string { return "search_failure_fixes" }

func (t *SearchFailureFixesTool) Description() string {
	return "Search for fixes to similar failures. Use this when encountering an error to find known solutions."
}

func (t *SearchFailureFixesTool) Parameters() ToolParameters {
	return ToolParameters{
		Type: "object",
		Properties: map[string]ToolProperty{
			"error_description": {
				Type:        "string",
				Description: "Description of the error or failure you're trying to fix",
			},
			"limit": {
				Type:        "integer",
				Description: "Maximum number of results (default: 5)",
			},
		},
		Required: []string{"error_description"},
	}
}

func (t *SearchFailureFixesTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	query, _ := args["error_description"].(string)
	limit := 5
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}

	// Check if vector store is available
	if t.store == nil {
		return map[string]interface{}{
			"success": false,
			"fixes":   []interface{}{},
			"message": "Vector store not available (disabled in config)",
		}, nil
	}

	opts := vectordb.SearchOptions{Limit: limit, MinScore: 0.4}
	results, err := t.store.SearchSimilarFailures(ctx, query, opts)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	fixes := make([]map[string]interface{}, len(results))
	for i, r := range results {
		fixes[i] = map[string]interface{}{
			"id":              r.ID.String(),
			"failure_type":    r.FailureType,
			"error_signature": r.ErrorSignature,
			"fix_description": r.FixDescription,
			"success_rate":    r.FixSuccessRate,
			"score":           r.Score,
		}
	}

	return map[string]interface{}{
		"found": len(fixes),
		"fixes": fixes,
	}, nil
}

// ============================================================================
// StoreSuccessfulStrategyTool - Store a successful test strategy
// ============================================================================

type StoreSuccessfulStrategyTool struct {
	store *vectordb.Store
}

func (t *StoreSuccessfulStrategyTool) Name() string { return "store_successful_strategy" }

func (t *StoreSuccessfulStrategyTool) Description() string {
	return "Store a successful test strategy for future reuse. Use this when a test approach works well."
}

func (t *StoreSuccessfulStrategyTool) Parameters() ToolParameters {
	return ToolParameters{
		Type: "object",
		Properties: map[string]ToolProperty{
			"strategy_type": {
				Type:        "string",
				Description: "Type of strategy",
				Enum:        []string{"auth_flow", "data_generation", "test_sequence", "edge_case_coverage", "error_handling"},
			},
			"strategy_name": {
				Type:        "string",
				Description: "A short name for the strategy",
			},
			"description": {
				Type:        "string",
				Description: "Detailed description of the strategy",
			},
			"strategy_content": {
				Type:        "object",
				Description: "The strategy details (steps, payloads, etc.)",
			},
		},
		Required: []string{"strategy_type", "strategy_name", "description"},
	}
}

func (t *StoreSuccessfulStrategyTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	strategy := &vectordb.SuccessfulStrategy{
		StrategyType: getString(args, "strategy_type"),
		StrategyName: getString(args, "strategy_name"),
		Description:  getString(args, "description"),
	}

	if content, ok := args["strategy_content"].(map[string]interface{}); ok {
		strategy.StrategyContent = content
	}

	// Check if vector store is available
	if t.store == nil {
		return map[string]interface{}{
			"success": false,
			"message": "Vector store not available (disabled in config)",
		}, nil
	}

	if err := t.store.StoreStrategy(ctx, strategy); err != nil {
		return nil, fmt.Errorf("failed to store strategy: %w", err)
	}

	return map[string]interface{}{
		"success": true,
		"id":      strategy.ID.String(),
		"message": fmt.Sprintf("Stored strategy: %s", strategy.StrategyName),
	}, nil
}

// ============================================================================
// SearchStrategiesTool - Search for relevant test strategies
// ============================================================================

type SearchStrategiesTool struct {
	store *vectordb.Store
}

func (t *SearchStrategiesTool) Name() string { return "search_strategies" }

func (t *SearchStrategiesTool) Description() string {
	return "Search for successful test strategies. Use this to find proven approaches for testing."
}

func (t *SearchStrategiesTool) Parameters() ToolParameters {
	return ToolParameters{
		Type: "object",
		Properties: map[string]ToolProperty{
			"query": {
				Type:        "string",
				Description: "Description of what kind of strategy you need",
			},
			"limit": {
				Type:        "integer",
				Description: "Maximum number of results (default: 5)",
			},
		},
		Required: []string{"query"},
	}
}

func (t *SearchStrategiesTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	query, _ := args["query"].(string)
	limit := 5
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}

	opts := vectordb.SearchOptions{Limit: limit, MinScore: 0.4}
	results, err := t.store.SearchSimilarStrategies(ctx, query, opts)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	strategies := make([]map[string]interface{}, len(results))
	for i, r := range results {
		strategies[i] = map[string]interface{}{
			"id":            r.ID.String(),
			"strategy_type": r.StrategyType,
			"strategy_name": r.StrategyName,
			"description":   r.Description,
			"success_rate":  r.SuccessRate,
			"score":         r.Score,
		}
	}

	return map[string]interface{}{
		"found":      len(strategies),
		"strategies": strategies,
	}, nil
}

// ============================================================================
// UpdateLearningConfidenceTool - Update confidence of a learning
// ============================================================================

type UpdateLearningConfidenceTool struct {
	store *vectordb.Store
}

func (t *UpdateLearningConfidenceTool) Name() string { return "update_learning_confidence" }

func (t *UpdateLearningConfidenceTool) Description() string {
	return "Mark a learning as successfully applied, increasing its usage count and confidence."
}

func (t *UpdateLearningConfidenceTool) Parameters() ToolParameters {
	return ToolParameters{
		Type: "object",
		Properties: map[string]ToolProperty{
			"learning_id": {
				Type:        "string",
				Description: "The ID of the learning that was successfully applied",
			},
		},
		Required: []string{"learning_id"},
	}
}

func (t *UpdateLearningConfidenceTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	learningID, _ := args["learning_id"].(string)

	id, err := parseUUID(learningID)
	if err != nil {
		return nil, fmt.Errorf("invalid learning_id: %w", err)
	}

	if err := t.store.IncrementLearningUsage(ctx, id); err != nil {
		return nil, fmt.Errorf("failed to update learning: %w", err)
	}

	return map[string]interface{}{
		"success": true,
		"message": "Learning usage incremented",
	}, nil
}

// ============================================================================
// Helper functions
// ============================================================================

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func parseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

// ============================================================================
// ReadTestSuiteTool - Read a test suite file
// ============================================================================

type ReadTestSuiteTool struct{}

func (t *ReadTestSuiteTool) Name() string { return "read_test_suite" }

func (t *ReadTestSuiteTool) Description() string {
	return "Read a test suite JSON file. Returns the parsed test suite with all test cases, their configurations, and metadata."
}

func (t *ReadTestSuiteTool) Parameters() ToolParameters {
	return ToolParameters{
		Type: "object",
		Properties: map[string]ToolProperty{
			"workflow_id": {
				Type:        "string",
				Description: "The workflow ID to find the test suite for",
			},
			"file_path": {
				Type:        "string",
				Description: "Direct file path to the test suite JSON (alternative to workflow_id)",
			},
		},
		Required: []string{},
	}
}

func (t *ReadTestSuiteTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	workflowID := getString(args, "workflow_id")
	filePath := getString(args, "file_path")

	// Determine file path
	var suitePath string
	if filePath != "" {
		suitePath = filePath
	} else if workflowID != "" {
		suitePath = findOutputFile(SuitesDir, workflowID, "test-suite")
		if suitePath == "" {
			// Also check plans directory for execution plans
			suitePath = findOutputFile(PlansDir, workflowID, "plan")
		}
	} else {
		return nil, fmt.Errorf("either workflow_id or file_path is required")
	}

	if suitePath == "" {
		return nil, fmt.Errorf("test suite not found for workflow_id: %s", workflowID)
	}

	suite, err := LoadTestSuite(suitePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load test suite: %w", err)
	}

	return map[string]interface{}{
		"success":     true,
		"file_path":   suitePath,
		"suite_id":    suite.ID,
		"name":        suite.Name,
		"description": suite.Description,
		"version":     suite.Version,
		"spec_id":     suite.SpecID,
		"base_url":    suite.BaseURL,
		"test_count":  len(suite.Tests),
		"tests":       suite.Tests,
		"statistics":  suite.Statistics,
		"created_at":  suite.CreatedAt,
		"updated_at":  suite.UpdatedAt,
	}, nil
}

// ============================================================================
// ReadTestResultsTool - Read test execution results
// ============================================================================

type ReadTestResultsTool struct{}

func (t *ReadTestResultsTool) Name() string { return "read_test_results" }

func (t *ReadTestResultsTool) Description() string {
	return "Read test execution results. Returns detailed results including pass/fail status, response data, and errors for each test."
}

func (t *ReadTestResultsTool) Parameters() ToolParameters {
	return ToolParameters{
		Type: "object",
		Properties: map[string]ToolProperty{
			"workflow_id": {
				Type:        "string",
				Description: "The workflow ID to find results for",
			},
			"file_path": {
				Type:        "string",
				Description: "Direct file path to the results JSON (alternative to workflow_id)",
			},
		},
		Required: []string{},
	}
}

func (t *ReadTestResultsTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	workflowID := getString(args, "workflow_id")
	filePath := getString(args, "file_path")

	var resultsPath string
	if filePath != "" {
		resultsPath = filePath
	} else if workflowID != "" {
		resultsPath = findOutputFile(ResultsDir, workflowID, "test-results")
		if resultsPath == "" {
			resultsPath = findOutputFile(ResultsDir, workflowID, "results")
		}
	} else {
		return nil, fmt.Errorf("either workflow_id or file_path is required")
	}

	if resultsPath == "" {
		return nil, fmt.Errorf("test results not found for workflow_id: %s", workflowID)
	}

	data, err := os.ReadFile(resultsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read results file: %w", err)
	}

	var results map[string]interface{}
	if err := json.Unmarshal(data, &results); err != nil {
		return nil, fmt.Errorf("failed to parse results JSON: %w", err)
	}

	results["success"] = true
	results["file_path"] = resultsPath
	return results, nil
}

// ============================================================================
// ReadSpecTool - Read the original OpenAPI spec
// ============================================================================

type ReadSpecTool struct{}

func (t *ReadSpecTool) Name() string { return "read_spec" }

func (t *ReadSpecTool) Description() string {
	return "Read the original OpenAPI specification. Returns the parsed spec with all endpoints, schemas, and security definitions."
}

func (t *ReadSpecTool) Parameters() ToolParameters {
	return ToolParameters{
		Type: "object",
		Properties: map[string]ToolProperty{
			"workflow_id": {
				Type:        "string",
				Description: "The workflow ID to find the spec for",
			},
			"file_path": {
				Type:        "string",
				Description: "Direct file path to the spec file (alternative to workflow_id)",
			},
		},
		Required: []string{},
	}
}

func (t *ReadSpecTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	workflowID := getString(args, "workflow_id")
	filePath := getString(args, "file_path")

	var specPath string
	if filePath != "" {
		specPath = filePath
	} else if workflowID != "" {
		specPath = findOutputFile(SpecsDir, workflowID, "")
	} else {
		return nil, fmt.Errorf("either workflow_id or file_path is required")
	}

	if specPath == "" {
		return nil, fmt.Errorf("spec not found for workflow_id: %s", workflowID)
	}

	data, err := os.ReadFile(specPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read spec file: %w", err)
	}

	var spec map[string]interface{}
	if err := json.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("failed to parse spec JSON: %w", err)
	}

	return map[string]interface{}{
		"success":   true,
		"file_path": specPath,
		"spec":      spec,
	}, nil
}

// ============================================================================
// ListOutputFilesTool - List available output files for a workflow
// ============================================================================

type ListOutputFilesTool struct{}

func (t *ListOutputFilesTool) Name() string { return "list_output_files" }

func (t *ListOutputFilesTool) Description() string {
	return "List all available output files for a workflow. Returns categorized list of specs, suites, results, payloads, and other generated files."
}

func (t *ListOutputFilesTool) Parameters() ToolParameters {
	return ToolParameters{
		Type: "object",
		Properties: map[string]ToolProperty{
			"workflow_id": {
				Type:        "string",
				Description: "The workflow ID to list files for",
			},
		},
		Required: []string{"workflow_id"},
	}
}

func (t *ListOutputFilesTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	workflowID := getString(args, "workflow_id")
	if workflowID == "" {
		return nil, fmt.Errorf("workflow_id is required")
	}

	result := map[string]interface{}{
		"workflow_id": workflowID,
		"files":       make(map[string][]string),
	}

	// Define directories to search
	directories := map[string]string{
		"specs":     SpecsDir,
		"suites":    SuitesDir,
		"results":   ResultsDir,
		"payloads":  PayloadsDir,
		"discovery": DiscoveryDir,
		"strategy":  StrategyDir,
		"plans":     PlansDir,
	}

	filesMap := make(map[string][]string)
	totalFiles := 0

	for category, dir := range directories {
		files := findFilesForWorkflow(dir, workflowID)
		if len(files) > 0 {
			filesMap[category] = files
			totalFiles += len(files)
		}
	}

	result["files"] = filesMap
	result["total_files"] = totalFiles
	result["success"] = true

	return result, nil
}

// ============================================================================
// File system helper functions
// ============================================================================

// findOutputFile finds an output file matching the workflow ID and optional suffix
func findOutputFile(dir, workflowID, suffix string) string {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return ""
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}

	// Try different patterns
	patterns := []string{
		fmt.Sprintf("%s-%s.json", workflowID, suffix),
		fmt.Sprintf("%s.json", workflowID),
	}

	if suffix != "" {
		patterns = append(patterns, fmt.Sprintf("%s-%s.json", workflowID, suffix))
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		// Check for exact matches first
		for _, pattern := range patterns {
			if name == pattern {
				return filepath.Join(dir, name)
			}
		}
		// Check for prefix match
		if strings.HasPrefix(name, workflowID) && strings.HasSuffix(name, ".json") {
			return filepath.Join(dir, name)
		}
	}

	return ""
}

// findFilesForWorkflow finds all files matching a workflow ID in a directory
func findFilesForWorkflow(dir, workflowID string) []string {
	var files []string

	if err := os.MkdirAll(dir, 0755); err != nil {
		return files
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return files
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.Contains(name, workflowID) {
			files = append(files, filepath.Join(dir, name))
		}
	}

	return files
}

// createBackup creates a backup of a file before modification
func createBackup(filePath string) (string, error) {
	if err := os.MkdirAll(BackupDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file for backup: %w", err)
	}

	baseName := filepath.Base(filePath)
	timestamp := time.Now().Format("20060102-150405")
	backupName := fmt.Sprintf("%s.%s.bak", baseName, timestamp)
	backupPath := filepath.Join(BackupDir, backupName)

	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write backup: %w", err)
	}

	return backupPath, nil
}

// ============================================================================
// EditTestCaseTool - Modify a specific test case in the suite
// ============================================================================

type EditTestCaseTool struct{}

func (t *EditTestCaseTool) Name() string { return "edit_test_case" }

func (t *EditTestCaseTool) Description() string {
	return "Modify a specific test case in the suite. Can update any field like payload, expected_status, headers, or context settings."
}

func (t *EditTestCaseTool) Parameters() ToolParameters {
	return ToolParameters{
		Type: "object",
		Properties: map[string]ToolProperty{
			"workflow_id": {
				Type:        "string",
				Description: "The workflow ID of the test suite to modify",
			},
			"file_path": {
				Type:        "string",
				Description: "Direct file path to the test suite (alternative to workflow_id)",
			},
			"test_order": {
				Type:        "integer",
				Description: "The order number of the test to modify (1-based index)",
			},
			"test_name": {
				Type:        "string",
				Description: "The name of the test to modify (alternative to test_order)",
			},
			"test_id": {
				Type:        "string",
				Description: "The ID of the test to modify (alternative to test_order/test_name)",
			},
			"modifications": {
				Type:        "object",
				Description: "Object containing fields to modify (e.g., {\"expected_status\": 201, \"payload\": {...}})",
			},
		},
		Required: []string{"modifications"},
	}
}

func (t *EditTestCaseTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	workflowID := getString(args, "workflow_id")
	filePath := getString(args, "file_path")
	testOrder := getInt(args, "test_order")
	testName := getString(args, "test_name")
	testID := getString(args, "test_id")
	modifications, _ := args["modifications"].(map[string]interface{})

	if modifications == nil || len(modifications) == 0 {
		return nil, fmt.Errorf("modifications object is required")
	}

	// Find suite path
	var suitePath string
	if filePath != "" {
		suitePath = filePath
	} else if workflowID != "" {
		suitePath = findOutputFile(SuitesDir, workflowID, "test-suite")
	}

	if suitePath == "" {
		return nil, fmt.Errorf("test suite not found")
	}

	// Load suite
	suite, err := LoadTestSuite(suitePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load test suite: %w", err)
	}

	// Find target test
	var targetIdx int = -1
	for i, test := range suite.Tests {
		if testID != "" && test.ID == testID {
			targetIdx = i
			break
		}
		if testName != "" && test.Name == testName {
			targetIdx = i
			break
		}
		if testOrder > 0 && i+1 == testOrder {
			targetIdx = i
			break
		}
	}

	if targetIdx == -1 {
		return nil, fmt.Errorf("test case not found")
	}

	// Create backup
	backupPath, err := createBackup(suitePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create backup: %w", err)
	}

	// Apply modifications
	test := &suite.Tests[targetIdx]
	applyTestModifications(test, modifications)

	// Save suite
	suite.UpdatedAt = time.Now()
	if _, err := SaveTestSuite(suite, filepath.Dir(suitePath)); err != nil {
		return nil, fmt.Errorf("failed to save modified suite: %w", err)
	}

	return map[string]interface{}{
		"success":       true,
		"message":       fmt.Sprintf("Test case '%s' modified successfully", test.Name),
		"test_id":       test.ID,
		"test_name":     test.Name,
		"backup_path":   backupPath,
		"modifications": modifications,
	}, nil
}

// ============================================================================
// AddTestCaseTool - Add a new test case to the suite
// ============================================================================

type AddTestCaseTool struct{}

func (t *AddTestCaseTool) Name() string { return "add_test_case" }

func (t *AddTestCaseTool) Description() string {
	return "Add a new test case to the suite. Specify the full test case definition and optionally where to insert it."
}

func (t *AddTestCaseTool) Parameters() ToolParameters {
	return ToolParameters{
		Type: "object",
		Properties: map[string]ToolProperty{
			"workflow_id": {
				Type:        "string",
				Description: "The workflow ID of the test suite to modify",
			},
			"file_path": {
				Type:        "string",
				Description: "Direct file path to the test suite (alternative to workflow_id)",
			},
			"test_case": {
				Type:        "object",
				Description: "The full test case definition with name, method, path, payload, expected_status, etc.",
			},
			"insert_after": {
				Type:        "integer",
				Description: "Insert after this test order number (0 or omit to append at end)",
			},
		},
		Required: []string{"test_case"},
	}
}

func (t *AddTestCaseTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	workflowID := getString(args, "workflow_id")
	filePath := getString(args, "file_path")
	testCaseData, _ := args["test_case"].(map[string]interface{})
	insertAfter := getInt(args, "insert_after")

	if testCaseData == nil {
		return nil, fmt.Errorf("test_case object is required")
	}

	// Find suite path
	var suitePath string
	if filePath != "" {
		suitePath = filePath
	} else if workflowID != "" {
		suitePath = findOutputFile(SuitesDir, workflowID, "test-suite")
	}

	if suitePath == "" {
		return nil, fmt.Errorf("test suite not found")
	}

	// Load suite
	suite, err := LoadTestSuite(suitePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load test suite: %w", err)
	}

	// Create backup
	backupPath, err := createBackup(suitePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create backup: %w", err)
	}

	// Create new test case
	newTest := createTestFromMap(testCaseData, len(suite.Tests)+1)

	// Insert at position or append
	if insertAfter > 0 && insertAfter <= len(suite.Tests) {
		// Insert after specified position
		suite.Tests = append(suite.Tests[:insertAfter], append([]PersistedTest{newTest}, suite.Tests[insertAfter:]...)...)
	} else {
		// Append at end
		suite.Tests = append(suite.Tests, newTest)
	}

	// Update statistics
	updateSuiteStatistics(suite)

	// Save suite
	suite.UpdatedAt = time.Now()
	if _, err := SaveTestSuite(suite, filepath.Dir(suitePath)); err != nil {
		return nil, fmt.Errorf("failed to save modified suite: %w", err)
	}

	return map[string]interface{}{
		"success":     true,
		"message":     fmt.Sprintf("Test case '%s' added successfully", newTest.Name),
		"test_id":     newTest.ID,
		"test_name":   newTest.Name,
		"position":    insertAfter + 1,
		"total_tests": len(suite.Tests),
		"backup_path": backupPath,
	}, nil
}

// ============================================================================
// RemoveTestCaseTool - Remove a test case from the suite
// ============================================================================

type RemoveTestCaseTool struct{}

func (t *RemoveTestCaseTool) Name() string { return "remove_test_case" }

func (t *RemoveTestCaseTool) Description() string {
	return "Remove a test case from the suite. Identify the test by order number, name, or ID."
}

func (t *RemoveTestCaseTool) Parameters() ToolParameters {
	return ToolParameters{
		Type: "object",
		Properties: map[string]ToolProperty{
			"workflow_id": {
				Type:        "string",
				Description: "The workflow ID of the test suite to modify",
			},
			"file_path": {
				Type:        "string",
				Description: "Direct file path to the test suite (alternative to workflow_id)",
			},
			"test_order": {
				Type:        "integer",
				Description: "The order number of the test to remove (1-based)",
			},
			"test_name": {
				Type:        "string",
				Description: "The name of the test to remove",
			},
			"test_id": {
				Type:        "string",
				Description: "The ID of the test to remove",
			},
		},
		Required: []string{},
	}
}

func (t *RemoveTestCaseTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	workflowID := getString(args, "workflow_id")
	filePath := getString(args, "file_path")
	testOrder := getInt(args, "test_order")
	testName := getString(args, "test_name")
	testID := getString(args, "test_id")

	// Find suite path
	var suitePath string
	if filePath != "" {
		suitePath = filePath
	} else if workflowID != "" {
		suitePath = findOutputFile(SuitesDir, workflowID, "test-suite")
	}

	if suitePath == "" {
		return nil, fmt.Errorf("test suite not found")
	}

	// Load suite
	suite, err := LoadTestSuite(suitePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load test suite: %w", err)
	}

	// Find target test
	var targetIdx int = -1
	var removedTest PersistedTest
	for i, test := range suite.Tests {
		if testID != "" && test.ID == testID {
			targetIdx = i
			removedTest = test
			break
		}
		if testName != "" && test.Name == testName {
			targetIdx = i
			removedTest = test
			break
		}
		if testOrder > 0 && i+1 == testOrder {
			targetIdx = i
			removedTest = test
			break
		}
	}

	if targetIdx == -1 {
		return nil, fmt.Errorf("test case not found")
	}

	// Create backup
	backupPath, err := createBackup(suitePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create backup: %w", err)
	}

	// Remove test
	suite.Tests = append(suite.Tests[:targetIdx], suite.Tests[targetIdx+1:]...)

	// Update statistics
	updateSuiteStatistics(suite)

	// Save suite
	suite.UpdatedAt = time.Now()
	if _, err := SaveTestSuite(suite, filepath.Dir(suitePath)); err != nil {
		return nil, fmt.Errorf("failed to save modified suite: %w", err)
	}

	return map[string]interface{}{
		"success":      true,
		"message":      fmt.Sprintf("Test case '%s' removed successfully", removedTest.Name),
		"removed_id":   removedTest.ID,
		"removed_name": removedTest.Name,
		"total_tests":  len(suite.Tests),
		"backup_path":  backupPath,
	}, nil
}

// ============================================================================
// UpdateTestContextTool - Update context/variables in the test suite
// ============================================================================

type UpdateTestContextTool struct{}

func (t *UpdateTestContextTool) Name() string { return "update_test_context" }

func (t *UpdateTestContextTool) Description() string {
	return "Update context capture or requirement settings for tests in the suite. Can modify how tests share data."
}

func (t *UpdateTestContextTool) Parameters() ToolParameters {
	return ToolParameters{
		Type: "object",
		Properties: map[string]ToolProperty{
			"workflow_id": {
				Type:        "string",
				Description: "The workflow ID of the test suite to modify",
			},
			"file_path": {
				Type:        "string",
				Description: "Direct file path to the test suite (alternative to workflow_id)",
			},
			"test_order": {
				Type:        "integer",
				Description: "The order number of the test to modify (optional, modifies all if not specified)",
			},
			"test_id": {
				Type:        "string",
				Description: "The ID of the test to modify",
			},
			"context_capture": {
				Type:        "object",
				Description: "New context capture settings (enabled, store_as, fields)",
			},
			"context_required": {
				Type:        "object",
				Description: "New context required settings (type, fields)",
			},
		},
		Required: []string{},
	}
}

func (t *UpdateTestContextTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	workflowID := getString(args, "workflow_id")
	filePath := getString(args, "file_path")
	testOrder := getInt(args, "test_order")
	testID := getString(args, "test_id")
	contextCapture, _ := args["context_capture"].(map[string]interface{})
	contextRequired, _ := args["context_required"].(map[string]interface{})

	if contextCapture == nil && contextRequired == nil {
		return nil, fmt.Errorf("either context_capture or context_required must be provided")
	}

	// Find suite path
	var suitePath string
	if filePath != "" {
		suitePath = filePath
	} else if workflowID != "" {
		suitePath = findOutputFile(SuitesDir, workflowID, "test-suite")
	}

	if suitePath == "" {
		return nil, fmt.Errorf("test suite not found")
	}

	// Load suite
	suite, err := LoadTestSuite(suitePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load test suite: %w", err)
	}

	// Create backup
	backupPath, err := createBackup(suitePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create backup: %w", err)
	}

	// Update tests
	modifiedCount := 0
	for i := range suite.Tests {
		// Check if this test should be modified
		if testID != "" && suite.Tests[i].ID != testID {
			continue
		}
		if testOrder > 0 && i+1 != testOrder {
			continue
		}

		if contextCapture != nil {
			suite.Tests[i].ContextCapture = parseContextCapture(contextCapture)
		}
		if contextRequired != nil {
			suite.Tests[i].ContextRequired = parseContextRequired(contextRequired)
		}
		modifiedCount++

		// If specific test was targeted, stop after modifying it
		if testID != "" || testOrder > 0 {
			break
		}
	}

	// Save suite
	suite.UpdatedAt = time.Now()
	if _, err := SaveTestSuite(suite, filepath.Dir(suitePath)); err != nil {
		return nil, fmt.Errorf("failed to save modified suite: %w", err)
	}

	return map[string]interface{}{
		"success":        true,
		"message":        fmt.Sprintf("Updated context settings for %d test(s)", modifiedCount),
		"modified_count": modifiedCount,
		"backup_path":    backupPath,
	}, nil
}

// ============================================================================
// Additional helper functions for write tools
// ============================================================================

// getInt extracts an integer from a map
func getInt(m map[string]interface{}, key string) int {
	if v, ok := m[key].(float64); ok {
		return int(v)
	}
	if v, ok := m[key].(int); ok {
		return v
	}
	return 0
}

// applyTestModifications applies modifications to a persisted test
func applyTestModifications(test *PersistedTest, modifications map[string]interface{}) {
	if name, ok := modifications["name"].(string); ok {
		test.Name = name
	}
	if desc, ok := modifications["description"].(string); ok {
		test.Description = desc
	}
	if category, ok := modifications["category"].(string); ok {
		test.Category = category
	}
	if method, ok := modifications["method"].(string); ok {
		test.Method = method
	}
	if path, ok := modifications["path"].(string); ok {
		test.Path = path
	}
	if status, ok := modifications["expected_status"].(float64); ok {
		test.ExpectedStatus = int(status)
	}
	if payload, ok := modifications["payload"]; ok {
		test.Payload = payload
	}
	if headers, ok := modifications["headers"].(map[string]interface{}); ok {
		test.Headers = headers
	}
	if pathParams, ok := modifications["path_params"].(map[string]interface{}); ok {
		test.PathParams = pathParams
	}
	if queryParams, ok := modifications["query_params"].(map[string]interface{}); ok {
		test.QueryParams = queryParams
	}
	if skip, ok := modifications["skip"].(bool); ok {
		test.Skip = skip
	}
	if skipReason, ok := modifications["skip_reason"].(string); ok {
		test.SkipReason = skipReason
	}
	if cc, ok := modifications["context_capture"].(map[string]interface{}); ok {
		test.ContextCapture = parseContextCapture(cc)
	}
	if cr, ok := modifications["context_required"].(map[string]interface{}); ok {
		test.ContextRequired = parseContextRequired(cr)
	}
	if tags, ok := modifications["tags"].([]interface{}); ok {
		test.Tags = make([]string, 0, len(tags))
		for _, tag := range tags {
			if s, ok := tag.(string); ok {
				test.Tags = append(test.Tags, s)
			}
		}
	}
}

// createTestFromMap creates a PersistedTest from a map
func createTestFromMap(data map[string]interface{}, order int) PersistedTest {
	test := PersistedTest{
		ID: fmt.Sprintf("test-%d-%d", order, time.Now().UnixNano()%1000000),
	}

	// Apply all fields from the map
	applyTestModifications(&test, data)

	// If no ID was set, use the generated one; otherwise use provided ID
	if id, ok := data["id"].(string); ok && id != "" {
		test.ID = id
	}

	return test
}

// updateSuiteStatistics recalculates statistics for a test suite
func updateSuiteStatistics(suite *TestSuite) {
	suite.Statistics = SuiteStatistics{
		TotalTests: len(suite.Tests),
	}

	for _, test := range suite.Tests {
		switch test.Category {
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
}

// parseContextCapture parses context capture settings from a map
func parseContextCapture(data map[string]interface{}) *ContextCapture {
	cc := &ContextCapture{}

	if enabled, ok := data["enabled"].(bool); ok {
		cc.Enabled = enabled
	}
	if storeAs, ok := data["store_as"].(string); ok {
		cc.StoreAs = storeAs
	}
	if fields, ok := data["fields"].([]interface{}); ok {
		cc.Fields = make([]string, 0, len(fields))
		for _, f := range fields {
			if s, ok := f.(string); ok {
				cc.Fields = append(cc.Fields, s)
			}
		}
	}

	return cc
}

// parseContextRequired parses context required settings from a map
func parseContextRequired(data map[string]interface{}) *ContextRequired {
	cr := &ContextRequired{}

	if t, ok := data["type"].(string); ok {
		cr.Type = t
	}
	if fields, ok := data["fields"].([]interface{}); ok {
		cr.Fields = make([]string, 0, len(fields))
		for _, f := range fields {
			if s, ok := f.(string); ok {
				cr.Fields = append(cr.Fields, s)
			}
		}
	}

	return cr
}

// ============================================================================
// GenerateRecommendationsTool - Analyze results and generate improvement recommendations
// ============================================================================

type GenerateRecommendationsTool struct {
	store *vectordb.Store
}

func (t *GenerateRecommendationsTool) Name() string { return "generate_recommendations" }

func (t *GenerateRecommendationsTool) Description() string {
	return "Analyze test results and generate recommendations for improving test coverage, fixing failures, and optimizing the test suite. Returns actionable suggestions based on patterns and vector DB learnings."
}

func (t *GenerateRecommendationsTool) Parameters() ToolParameters {
	return ToolParameters{
		Type: "object",
		Properties: map[string]ToolProperty{
			"workflow_id": {
				Type:        "string",
				Description: "The workflow ID to analyze results for",
			},
			"file_path": {
				Type:        "string",
				Description: "Direct path to results file (alternative to workflow_id)",
			},
			"focus_area": {
				Type:        "string",
				Description: "Optional focus area for recommendations",
				Enum:        []string{"coverage", "reliability", "performance", "security", "all"},
			},
		},
		Required: []string{},
	}
}

func (t *GenerateRecommendationsTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	workflowID := getString(args, "workflow_id")
	filePath := getString(args, "file_path")
	focusArea := getString(args, "focus_area")
	if focusArea == "" {
		focusArea = "all"
	}

	// Find results file
	var resultsPath string
	if filePath != "" {
		resultsPath = filePath
	} else if workflowID != "" {
		resultsPath = findOutputFile(ResultsDir, workflowID, "test-results")
	}

	if resultsPath == "" {
		return nil, fmt.Errorf("test results not found - provide workflow_id or file_path")
	}

	// Load results
	resultsData, err := os.ReadFile(resultsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read results file: %w", err)
	}

	var results map[string]interface{}
	if err := json.Unmarshal(resultsData, &results); err != nil {
		return nil, fmt.Errorf("failed to parse results: %w", err)
	}

	// Generate recommendations
	recommendations := t.generateFromResults(ctx, results, focusArea)

	return map[string]interface{}{
		"success":              true,
		"recommendation_count": len(recommendations),
		"recommendations":      recommendations,
		"focus_area":           focusArea,
		"analyzed_file":        resultsPath,
	}, nil
}

func (t *GenerateRecommendationsTool) generateFromResults(ctx context.Context, results map[string]interface{}, focusArea string) []map[string]interface{} {
	var recommendations []map[string]interface{}
	recID := 0

	// Extract summary
	summary, _ := results["summary"].(map[string]interface{})
	testResults, _ := results["results"].([]interface{})

	totalTests := getFloatValue(summary, "total_tests")
	passed := getFloatValue(summary, "passed")
	_ = getFloatValue(summary, "failed") // Available for future use
	skipped := getFloatValue(summary, "skipped")

	// Calculate pass rate
	passRate := 0.0
	if totalTests > 0 {
		passRate = (passed / totalTests) * 100
	}

	// Analyze based on focus area
	if focusArea == "all" || focusArea == "reliability" {
		// Low pass rate recommendation
		if passRate < 70 {
			recID++
			recommendations = append(recommendations, map[string]interface{}{
				"id":          fmt.Sprintf("rec-%d", recID),
				"priority":    "high",
				"category":    "reliability",
				"title":       fmt.Sprintf("Low pass rate: %.1f%%", passRate),
				"description": "The test pass rate is below 70%. Review failing tests and fix underlying issues.",
			})
		}

		// High skip rate
		if skipped > totalTests/3 {
			recID++
			recommendations = append(recommendations, map[string]interface{}{
				"id":          fmt.Sprintf("rec-%d", recID),
				"priority":    "medium",
				"category":    "reliability",
				"title":       fmt.Sprintf("%.0f tests skipped", skipped),
				"description": "Many tests are being skipped. Check context dependencies and test ordering.",
			})
		}
	}

	// Analyze individual test failures
	if testResults != nil && (focusArea == "all" || focusArea == "reliability") {
		authFailures := 0
		validationFailures := 0

		for _, tr := range testResults {
			result, ok := tr.(map[string]interface{})
			if !ok {
				continue
			}

			passed, _ := result["passed"].(bool)
			skipped, _ := result["skipped"].(bool)
			statusCode := int(getFloatValue(result, "status_code"))

			if !passed && !skipped {
				if statusCode == 401 || statusCode == 403 {
					authFailures++
				} else if statusCode == 400 || statusCode == 422 {
					validationFailures++
				}
			}
		}

		if authFailures > 0 {
			recID++
			recommendations = append(recommendations, map[string]interface{}{
				"id":          fmt.Sprintf("rec-%d", recID),
				"priority":    "high",
				"category":    "reliability",
				"title":       fmt.Sprintf("%d authentication failures", authFailures),
				"description": "Tests are failing with 401/403 errors. Verify API credentials and token configuration.",
			})
		}

		if validationFailures > 0 {
			recID++
			recommendations = append(recommendations, map[string]interface{}{
				"id":          fmt.Sprintf("rec-%d", recID),
				"priority":    "medium",
				"category":    "reliability",
				"title":       fmt.Sprintf("%d validation failures", validationFailures),
				"description": "Tests are failing with 400/422 errors. Check request payloads match API requirements.",
			})
		}
	}

	// Search vector DB for relevant recommendations
	if t.store != nil && (focusArea == "all" || focusArea == "reliability") {
		memoryRecs := t.searchMemoryForRecommendations(ctx, results)
		recommendations = append(recommendations, memoryRecs...)
	}

	// Coverage recommendations
	if focusArea == "all" || focusArea == "coverage" {
		recID++
		recommendations = append(recommendations, map[string]interface{}{
			"id":          fmt.Sprintf("rec-%d", recID),
			"priority":    "low",
			"category":    "coverage",
			"title":       "Review test coverage",
			"description": "Consider adding edge case tests, error handling scenarios, and security validation tests.",
		})
	}

	return recommendations
}

func (t *GenerateRecommendationsTool) searchMemoryForRecommendations(ctx context.Context, results map[string]interface{}) []map[string]interface{} {
	var recommendations []map[string]interface{}

	// Search for similar failure patterns
	query := "API test failures common fixes"
	opts := vectordb.SearchOptions{Limit: 2, MinScore: 0.5}
	failures, err := t.store.SearchSimilarFailures(ctx, query, opts)
	if err != nil {
		return recommendations
	}

	for _, f := range failures {
		recommendations = append(recommendations, map[string]interface{}{
			"id":          fmt.Sprintf("mem-%s", f.ID.String()[:8]),
			"priority":    "medium",
			"category":    "reliability",
			"title":       fmt.Sprintf("Known pattern: %s", f.FailureType),
			"description": f.FixDescription,
			"source":      "memory",
		})
	}

	return recommendations
}

func getFloatValue(m map[string]interface{}, key string) float64 {
	if v, ok := m[key].(float64); ok {
		return v
	}
	if v, ok := m[key].(int); ok {
		return float64(v)
	}
	return 0
}
