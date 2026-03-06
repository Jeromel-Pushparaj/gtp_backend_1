package autonomous

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"gitlab.com/tekion/development/toc/poc/opentest/pkg/llm"
)

// PlannedExecutor handles AI-planned, deterministic test execution
type PlannedExecutor struct {
	llmClient           *llm.Client
	httpClient          *http.Client
	testContext         *TestContext
	authManager         *AuthManager
	baseURL             string
	smartContextCapture *SmartContextCapture // Enhanced context capture
	dependencyGraph     *DependencyGraph     // Resource dependency tracking
}

// NewPlannedExecutor creates a new planned executor
func NewPlannedExecutor(llmClient *llm.Client, baseURL string) *PlannedExecutor {
	return &PlannedExecutor{
		llmClient:           llmClient,
		httpClient:          &http.Client{Timeout: 30 * time.Second},
		testContext:         NewTestContext(),
		baseURL:             baseURL,
		smartContextCapture: NewSmartContextCapture(),
		dependencyGraph:     NewDependencyGraph(),
	}
}

// SetAuthManager sets the auth manager for authentication
func (pe *PlannedExecutor) SetAuthManager(mgr *AuthManager) {
	pe.authManager = mgr
}

// SetContextMappings sets dynamic context type mappings from LLM analysis
func (pe *PlannedExecutor) SetContextMappings(mappings map[string]string) {
	pe.testContext.SetContextMappings(mappings)
}

// InitializeFromSpec initializes smart context capture and dependency graph from OpenAPI spec
func (pe *PlannedExecutor) InitializeFromSpec(spec map[string]interface{}) {
	if pe.smartContextCapture != nil {
		pe.smartContextCapture.RegisterSchemasFromSpec(spec)
		log.Printf("📊 Initialized smart context capture from OpenAPI spec")
	}
	if pe.dependencyGraph != nil {
		if err := pe.dependencyGraph.BuildFromSpec(spec); err != nil {
			log.Printf("⚠️ Failed to build dependency graph: %v", err)
		}
	}
}

// generateID generates a unique ID for the execution plan
func generateID() string {
	return uuid.New().String()
}

// CreateExecutionPlan builds an execution plan from payloads using deterministic ordering
// This approach doesn't require AI - it uses smart ordering based on HTTP methods and dependencies
func (pe *PlannedExecutor) CreateExecutionPlan(ctx context.Context, specID, workflowID string,
	strategy map[string]interface{}, payloads map[string]interface{}) (*ExecutionPlan, error) {

	log.Printf("🧠 Building execution plan for spec %s", specID)

	// Extract all tests from payloads and build ordered plan
	tests := pe.extractAndOrderTests(payloads)

	// Calculate statistics
	var positiveCount, negativeCount, boundaryCount, contextCount int
	for _, t := range tests {
		switch t.Category {
		case "positive":
			positiveCount++
		case "negative":
			negativeCount++
		case "boundary":
			boundaryCount++
		}
		if t.ContextRequired != nil || t.ContextCapture != nil {
			contextCount++
		}
	}

	plan := &ExecutionPlan{
		ID:         generateID(),
		SpecID:     specID,
		WorkflowID: workflowID,
		CreatedAt:  time.Now(),
		AIReasoning: "Deterministic ordering: POST/CREATE first to establish resources, " +
			"GET/read operations next, PUT/PATCH for updates, DELETE last for cleanup. " +
			"Positive tests first within each endpoint, then negative/boundary tests.",
		Tests: tests,
		Statistics: PlanStatistics{
			TotalTests:       len(tests),
			PositiveTests:    positiveCount,
			NegativeTests:    negativeCount,
			BoundaryTests:    boundaryCount,
			TestsWithContext: contextCount,
		},
	}

	log.Printf("✅ Execution plan created: %d tests (%d positive, %d negative, %d boundary)",
		len(plan.Tests), positiveCount, negativeCount, boundaryCount)

	return plan, nil
}

// extractAndOrderTests extracts tests from payloads and orders them optimally
func (pe *PlannedExecutor) extractAndOrderTests(payloads map[string]interface{}) []PlannedTest {
	var allTests []PlannedTest

	// Define method priority for ordering
	methodPriority := map[string]int{
		"POST":   1, // Create resources first
		"GET":    2, // Read after creation
		"PUT":    3, // Update existing
		"PATCH":  3, // Update existing
		"DELETE": 4, // Cleanup last
	}

	// Define category priority
	categoryPriority := map[string]int{
		"positive": 1, // Positive tests first
		"negative": 2, // Then negative
		"boundary": 3, // Then boundary
	}

	// Extract tests from each endpoint
	for endpointKey, endpointData := range payloads {
		endpoint, ok := endpointData.(map[string]interface{})
		if !ok {
			continue
		}

		// Parse endpoint key - handle both formats:
		// 1. "METHOD /path" (e.g., "POST /pet")
		// 2. "/path" (just path, method comes from test data)
		var defaultMethod, path string
		parts := strings.SplitN(endpointKey, " ", 2)
		if len(parts) == 2 {
			// Format: "METHOD /path"
			defaultMethod = parts[0]
			path = parts[1]
		} else {
			// Format: "/path" - method will come from individual tests
			defaultMethod = ""
			path = endpointKey
		}

		// Extract tests from each category
		for _, category := range []string{"positive", "negative", "boundary"} {
			categoryTests, ok := endpoint[category].([]interface{})
			if !ok {
				continue
			}

			for _, testData := range categoryTests {
				test, ok := testData.(map[string]interface{})
				if !ok {
					continue
				}

				// Get method from test data if not in endpoint key
				method := defaultMethod
				if method == "" {
					if testMethod, ok := test["method"].(string); ok {
						method = testMethod
					} else {
						// Skip if no method found
						continue
					}
				}

				// Get path from test data if specified, otherwise use endpoint path
				testPath := path
				if tp, ok := test["path"].(string); ok && tp != "" {
					testPath = tp
				}

				plannedTest := pe.convertToPlannedTest(test, method, testPath, category)
				plannedTest.MethodPriority = methodPriority[method]
				plannedTest.CategoryPriority = categoryPriority[category]
				allTests = append(allTests, plannedTest)
			}
		}
	}

	// Sort tests by: method priority, then category priority, then name
	sort.Slice(allTests, func(i, j int) bool {
		if allTests[i].MethodPriority != allTests[j].MethodPriority {
			return allTests[i].MethodPriority < allTests[j].MethodPriority
		}
		if allTests[i].CategoryPriority != allTests[j].CategoryPriority {
			return allTests[i].CategoryPriority < allTests[j].CategoryPriority
		}
		return allTests[i].Name < allTests[j].Name
	})

	// Assign order numbers
	for i := range allTests {
		allTests[i].Order = i + 1
	}

	return allTests
}

// convertToPlannedTest converts a raw test map to a PlannedTest struct
func (pe *PlannedExecutor) convertToPlannedTest(test map[string]interface{}, method, path, category string) PlannedTest {
	pt := PlannedTest{
		Method:   method,
		Path:     path,
		Category: category,
	}

	// Extract name
	if name, ok := test["name"].(string); ok {
		pt.Name = name
	}

	// Extract description
	if desc, ok := test["description"].(string); ok {
		pt.Description = desc
	}

	// Extract expected status
	if status, ok := test["expected_status"].(float64); ok {
		pt.ExpectedStatus = int(status)
	}

	// Extract payload
	if payload, ok := test["payload"]; ok {
		pt.Payload = payload
	}

	// Extract path params
	if pathParams, ok := test["path_params"].(map[string]interface{}); ok {
		pt.PathParams = pathParams
	}

	// Extract query params
	if queryParams, ok := test["query_params"].(map[string]interface{}); ok {
		pt.QueryParams = queryParams
	}

	// Extract headers
	if headers, ok := test["headers"].(map[string]interface{}); ok {
		pt.Headers = headers
	}

	// Extract context capture
	if cc, ok := test["context_capture"].(map[string]interface{}); ok {
		pt.ContextCapture = &ContextCapture{}
		if enabled, ok := cc["enabled"].(bool); ok {
			pt.ContextCapture.Enabled = enabled
		}
		if storeAs, ok := cc["store_as"].(string); ok {
			pt.ContextCapture.StoreAs = storeAs
		}
		if fields, ok := cc["fields"].([]interface{}); ok {
			for _, f := range fields {
				if fs, ok := f.(string); ok {
					pt.ContextCapture.Fields = append(pt.ContextCapture.Fields, fs)
				}
			}
		}
	}

	// Extract context required
	if cr, ok := test["context_required"].(map[string]interface{}); ok {
		pt.ContextRequired = &ContextRequired{}
		if t, ok := cr["type"].(string); ok {
			pt.ContextRequired.Type = t
		}
		if fields, ok := cr["fields"].([]interface{}); ok {
			for _, f := range fields {
				if fs, ok := f.(string); ok {
					pt.ContextRequired.Fields = append(pt.ContextRequired.Fields, fs)
				}
			}
		}
	}

	// Extract skip on context missing
	if skip, ok := test["skip_on_context_missing"].(bool); ok {
		pt.SkipOnContextMissing = skip
	}

	// Set up context capture for POST methods creating resources
	if method == "POST" {
		resourceType := pe.inferResourceType(path)
		if pt.ContextCapture == nil {
			// Auto-configure context capture for POST endpoints
			if resourceType != "" {
				pt.ContextCapture = &ContextCapture{
					Enabled: true,
					StoreAs: resourceType,
					Fields:  []string{"id"},
				}
			}
		} else if resourceType != "" && isGenericContextType(pt.ContextCapture.StoreAs) {
			// Override generic store_as values with the actual resource type
			// This fixes the issue where LLM uses "resource" instead of "pet", "user", etc.
			pt.ContextCapture.StoreAs = resourceType
		}
	}

	// Set up context requirements for endpoints with path parameters
	if strings.Contains(path, "{") && pt.ContextRequired == nil && category == "positive" {
		resourceType := pe.inferResourceType(path)
		if resourceType != "" {
			pt.ContextRequired = &ContextRequired{
				Type:   resourceType,
				Fields: []string{"id"},
			}
			pt.SkipOnContextMissing = true
			// Update path to use context reference
			pt.Path = pe.addContextToPath(path, resourceType)
		}
	}

	return pt
}

// inferResourceType infers the resource type from the path generically.
// Works with any API: /pet, /customers/{id}, /api/v1/products, etc.
func (pe *PlannedExecutor) inferResourceType(path string) string {
	return InferResourceTypeFromPath(path)
}

// InferResourceTypeFromPath extracts a singular resource type from an API path.
// Exported for use by other packages. Handles production API patterns.
func InferResourceTypeFromPath(path string) string {
	path = strings.TrimPrefix(path, "/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		return ""
	}

	// Find the last non-parameter, non-prefix part
	for i := len(parts) - 1; i >= 0; i-- {
		part := strings.ToLower(parts[i])
		if isSkippablePart(part) {
			continue
		}
		return Singularize(part)
	}

	// Fallback to first valid part
	for _, part := range parts {
		part = strings.ToLower(part)
		if !isSkippablePart(part) {
			return Singularize(part)
		}
	}
	return ""
}

// isSkippablePart returns true for path segments that are not resource names
func isSkippablePart(part string) bool {
	if part == "" || strings.HasPrefix(part, "{") {
		return true
	}

	// Common path prefixes to skip
	skipPrefixes := map[string]bool{
		"api": true, "rest": true, "graphql": true, "grpc": true,
		"internal": true, "public": true, "private": true, "external": true,
		"admin": true, "management": true, "service": true, "services": true,
		"alpha": true, "beta": true, "stable": true, "legacy": true,
	}
	if skipPrefixes[part] {
		return true
	}

	// Skip version patterns: v1, v2, v10, v2.1, 2024-01-15, etc.
	if len(part) >= 2 && part[0] == 'v' && (part[1] >= '0' && part[1] <= '9') {
		return true
	}
	// Date-based versions: 2024-01-15
	if len(part) >= 4 && part[0] >= '0' && part[0] <= '9' {
		return true
	}

	return false
}

// Singularize converts a plural resource name to singular.
// Handles English pluralization rules including irregular forms.
// Exported for use by other packages.
func Singularize(name string) string {
	if len(name) < 2 {
		return name
	}

	// Handle common irregular plurals
	irregulars := map[string]string{
		"people": "person", "children": "child", "men": "man", "women": "woman",
		"teeth": "tooth", "feet": "foot", "geese": "goose", "mice": "mouse",
		"indices": "index", "vertices": "vertex", "matrices": "matrix",
		"analyses": "analysis", "crises": "crisis", "theses": "thesis",
		"criteria": "criterion", "phenomena": "phenomenon", "data": "datum",
		"media": "medium", "appendices": "appendix", "cacti": "cactus",
		"foci": "focus", "fungi": "fungus", "nuclei": "nucleus",
		"radii": "radius", "stimuli": "stimulus", "syllabi": "syllabus",
		"alumni": "alumnus", "octopi": "octopus",
	}
	if singular, ok := irregulars[name]; ok {
		return singular
	}

	// Handle -ies -> -y (categories -> category, policies -> policy)
	if strings.HasSuffix(name, "ies") && len(name) > 3 {
		return name[:len(name)-3] + "y"
	}

	// Handle -ves -> -f/-fe (leaves -> leaf, knives -> knife)
	if strings.HasSuffix(name, "ves") && len(name) > 3 {
		base := name[:len(name)-3]
		// Common -ves words
		if base == "lea" || base == "li" || base == "kni" || base == "wi" ||
			base == "shel" || base == "hal" || base == "cal" || base == "wol" {
			if base == "li" || base == "kni" || base == "wi" {
				return base + "fe"
			}
			return base + "f"
		}
	}

	// Handle -oes -> -o (heroes -> hero, potatoes -> potato)
	// But not: photos, pianos, videos (already end in -os)
	if strings.HasSuffix(name, "oes") && len(name) > 3 {
		return name[:len(name)-2]
	}

	// Handle -sses -> -ss (classes -> class, addresses -> address)
	if strings.HasSuffix(name, "sses") && len(name) > 4 {
		return name[:len(name)-2]
	}

	// Handle -xes -> -x (boxes -> box, indexes -> index)
	if strings.HasSuffix(name, "xes") && len(name) > 3 {
		return name[:len(name)-2]
	}

	// Handle -ches -> -ch (matches -> match, batches -> batch)
	if strings.HasSuffix(name, "ches") && len(name) > 4 {
		return name[:len(name)-2]
	}

	// Handle -shes -> -sh (dishes -> dish, crashes -> crash)
	if strings.HasSuffix(name, "shes") && len(name) > 4 {
		return name[:len(name)-2]
	}

	// Handle -ses -> -s or -se (buses -> bus, responses -> response)
	if strings.HasSuffix(name, "ses") && len(name) > 3 {
		// Words like "responses", "expenses" -> keep the 'e'
		if strings.HasSuffix(name, "nses") || strings.HasSuffix(name, "rses") {
			return name[:len(name)-1]
		}
		// Words like "buses", "statuses" -> remove 'es'
		return name[:len(name)-2]
	}

	// Handle -zes -> -z (quizzes -> quiz) but be careful
	if strings.HasSuffix(name, "zzes") && len(name) > 4 {
		return name[:len(name)-3]
	}

	// Handle regular -s ending (users -> user, pets -> pet)
	// But don't touch words ending in -ss (class, boss)
	if strings.HasSuffix(name, "s") && !strings.HasSuffix(name, "ss") &&
		!strings.HasSuffix(name, "us") && !strings.HasSuffix(name, "is") {
		return name[:len(name)-1]
	}

	return name
}

// isGenericContextType checks if a context type is a generic placeholder
// that should be replaced with a specific resource type
func isGenericContextType(contextType string) bool {
	generic := map[string]bool{
		"resource": true,
		"item":     true,
		"entity":   true,
		"object":   true,
		"created":  true,
		"response": true,
		"result":   true,
		"data":     true,
	}
	return generic[strings.ToLower(contextType)]
}

// addContextToPath replaces path parameters with context references
// This works generically for any OpenAPI spec by inferring resource types from parameter names
func (pe *PlannedExecutor) addContextToPath(path, resourceType string) string {
	re := regexp.MustCompile(`\{([^}]+)\}`)
	return re.ReplaceAllStringFunc(path, func(match string) string {
		paramName := strings.Trim(match, "{}")

		// Infer context type and field from parameter name
		contextType, fieldName := inferContextFromParam(paramName, resourceType)

		return fmt.Sprintf("{{CONTEXT:%s.%s}}", contextType, fieldName)
	})
}

// inferContextFromParam infers the context type and field name from a path parameter
// Examples:
//   - "petId" -> ("pet", "id")
//   - "userId" -> ("user", "id")
//   - "orderId" -> ("order", "id")
//   - "username" -> ("user", "username")
//   - "itemName" -> ("item", "name")
//   - "id" -> (resourceType, "id")
func inferContextFromParam(paramName, fallbackResourceType string) (contextType, fieldName string) {
	paramLower := strings.ToLower(paramName)

	// Check for common "*Id" pattern (e.g., petId, userId, orderId)
	if strings.HasSuffix(paramLower, "id") && len(paramName) > 2 {
		// Extract resource name from parameter (e.g., "petId" -> "pet")
		resourceName := paramName[:len(paramName)-2]
		// Handle case variations: "petId" -> "pet", "PetId" -> "Pet" -> "pet"
		resourceName = strings.ToLower(resourceName[:1]) + resourceName[1:]
		if len(resourceName) > 1 {
			resourceName = strings.ToLower(resourceName)
		}
		return resourceName, "id"
	}

	// Check for common "*Name" pattern (e.g., username, itemName)
	if strings.HasSuffix(paramLower, "name") && len(paramName) > 4 {
		// Extract resource name from parameter (e.g., "userName" -> "user")
		resourceName := paramName[:len(paramName)-4]
		resourceName = strings.ToLower(resourceName)
		if resourceName == "" {
			resourceName = fallbackResourceType
		}
		return resourceName, paramName // Use full param name as field (e.g., "username")
	}

	// Check if the param name itself is a common identifier
	if paramLower == "id" {
		return fallbackResourceType, "id"
	}
	if paramLower == "name" {
		return fallbackResourceType, "name"
	}
	if paramLower == "username" {
		return "user", "username"
	}
	if paramLower == "email" {
		return "user", "email"
	}

	// For unknown patterns, use the param name as both context type and field
	// This handles cases like {slug}, {code}, {token}
	return fallbackResourceType, paramName
}

// ExecutePlan executes the plan deterministically without AI calls
func (pe *PlannedExecutor) ExecutePlan(ctx context.Context, plan *ExecutionPlan) (*ExecutionResult, error) {
	log.Printf("⚡ Deterministic Execution Phase: Executing %d tests", len(plan.Tests))

	startTime := time.Now()
	results := make([]PlannedTestResult, 0, len(plan.Tests))
	pe.testContext.Reset()

	passed, failed, skipped := 0, 0, 0

	for i, test := range plan.Tests {
		log.Printf("▶️  [%d/%d] %s %s - %s", i+1, len(plan.Tests), test.Method, test.Path, test.Name)

		result := pe.executeTest(ctx, test)
		results = append(results, result)

		if result.Skipped {
			skipped++
			log.Printf("⏭️  Skipped: %s", result.SkipReason)
		} else if result.Passed {
			passed++
			log.Printf("✅ Passed (status=%d, time=%v)", result.StatusCode, result.ResponseTime)
		} else {
			failed++
			log.Printf("❌ Failed (expected=%d, got=%d): %s", test.ExpectedStatus, result.StatusCode, result.Error)
		}
	}

	duration := time.Since(startTime)
	passRate := 0.0
	if len(plan.Tests) > 0 {
		passRate = float64(passed) / float64(len(plan.Tests)) * 100
	}

	summary := ExecutionSummary{
		TotalTests: len(plan.Tests),
		Passed:     passed,
		Failed:     failed,
		Skipped:    skipped,
		PassRate:   passRate,
		TotalTime:  duration,
		LLMCalls:   1, // Only the planning call
	}

	log.Printf("📊 Execution complete: %d/%d passed (%.1f%%) in %v with %d LLM call(s)",
		passed, len(plan.Tests), passRate, duration, summary.LLMCalls)

	return &ExecutionResult{
		Plan:        plan,
		Results:     results,
		Summary:     summary,
		ExecutedAt:  startTime,
		Duration:    duration,
		ContextData: pe.testContext.GetAllContext(),
	}, nil
}

// executeTest runs a single test deterministically
func (pe *PlannedExecutor) executeTest(ctx context.Context, test PlannedTest) PlannedTestResult {
	result := PlannedTestResult{Test: test}
	startTime := time.Now()

	// First, resolve context markers in path_params values
	resolvedPathParams := pe.resolveContextInPathParams(test.PathParams)

	// Then substitute path parameters (e.g., {petId} -> 123)
	resolvedPath := pe.substitutePathParams(test.Path, resolvedPathParams)

	// Then resolve any remaining context markers in the path itself
	resolvedPath = pe.resolveContextInString(resolvedPath)

	// Store the resolved path in the result for analysis
	result.ResolvedPath = resolvedPath

	// Debug: Log the resolved path if it was different from original
	if resolvedPath != test.Path {
		log.Printf("   🔗 Resolved path: %s → %s", test.Path, resolvedPath)
	}

	// Check if context resolution failed (still has markers)
	if HasUnresolvedMarkersInString(resolvedPath) {
		// Always skip tests with unresolved context markers - they can't be executed
		result.Skipped = true
		result.SkipReason = fmt.Sprintf("Unresolved context in path: %s", resolvedPath)
		return result
	}

	// Resolve context in payload
	var resolvedPayload interface{}
	if test.Payload != nil {
		resolvedPayload = pe.resolveContextInPayload(test.Payload)
		// Store the resolved payload in the result for analysis
		result.ResolvedPayload = resolvedPayload
	}

	// Build full URL with query parameters
	fullURL := pe.baseURL + resolvedPath
	if len(test.QueryParams) > 0 {
		fullURL = pe.addQueryParams(fullURL, test.QueryParams)
	}

	// Execute HTTP request
	statusCode, responseBody, respHeaders, err := pe.doHTTPRequest(ctx, test.Method, fullURL, resolvedPayload, test.Headers)
	result.ResponseTime = time.Since(startTime)

	if err != nil {
		result.Error = err.Error()
		result.Passed = false
		return result
	}

	result.StatusCode = statusCode

	// Parse response body
	var respData interface{}
	if len(responseBody) > 0 {
		if err := json.Unmarshal(responseBody, &respData); err == nil {
			result.ResponseBody = respData
		}
	}

	// Check if test passed
	result.Passed = pe.checkExpectedStatus(statusCode, test.ExpectedStatus)

	// Capture context if enabled and response was successful
	if test.ContextCapture != nil && test.ContextCapture.Enabled && statusCode >= 200 && statusCode < 300 {
		endpoint := test.Method + " " + test.Path
		captured := pe.captureContextWithSmartFallback(endpoint, test.ContextCapture, responseBody, respHeaders, test.Payload)
		result.ContextCaptured = captured
	}

	return result
}

// contextMarkerPattern matches {{CONTEXT:type.field}} patterns
var contextMarkerPattern = regexp.MustCompile(`\{\{CONTEXT:([^}]+)\}\}`)

// resolveContextInString resolves {{CONTEXT:type.field}} markers in a string
func (pe *PlannedExecutor) resolveContextInString(s string) string {
	return contextMarkerPattern.ReplaceAllStringFunc(s, func(match string) string {
		// Extract the marker path (e.g., "pet.id" from "{{CONTEXT:pet.id}}")
		parts := contextMarkerPattern.FindStringSubmatch(match)
		if len(parts) < 2 {
			return match
		}

		markerPath := parts[1]
		pathParts := strings.Split(markerPath, ".")
		if len(pathParts) < 2 {
			return match
		}

		contextType := pathParts[0]
		fieldName := pathParts[1]
		fieldPath := pathParts[1:]

		// Look up the context with smart fallback
		contextData, actualType := pe.testContext.GetContextWithFallback(contextType, fieldName)
		if contextData == nil {
			log.Printf("   ⚠️ Context not found for type '%s' (available: %v)", contextType, pe.testContext.GetAllContextTypes())
			return match // Keep original if context not found
		}

		// Log if we used a fallback
		if actualType != contextType {
			log.Printf("   🔄 Context fallback: '%s' → '%s'", contextType, actualType)
		}

		// Navigate to the field
		value := extractFieldValue(contextData, fieldPath)
		if value == nil {
			log.Printf("   ⚠️ Field '%v' not found in context '%s' (available fields: %v)", fieldPath, actualType, getMapKeys(contextData))
			return match // Keep original if field not found
		}

		// Format the value properly - avoid scientific notation for large integers
		resolved := formatContextValue(value)
		log.Printf("   ✅ Resolved {{CONTEXT:%s}} → %s", markerPath, resolved)
		return resolved
	})
}

// formatContextValue formats a context value for string interpolation
// It avoids scientific notation for large integers (common issue with JSON float64)
func formatContextValue(value interface{}) string {
	switch v := value.(type) {
	case float64:
		// Check if it's actually an integer (no decimal part)
		// Use math.Trunc to check, and format with %.0f for large integers
		// that might overflow int64
		if v == math.Trunc(v) {
			// Use %.0f to avoid scientific notation for large integers
			return fmt.Sprintf("%.0f", v)
		}
		// For actual floats, use full precision
		return fmt.Sprintf("%f", v)
	case int64:
		return fmt.Sprintf("%d", v)
	case int:
		return fmt.Sprintf("%d", v)
	case string:
		// Check if the string looks like scientific notation for an integer
		// e.g., "9.223372036854776e+18" or "9.9999999e+07"
		return convertScientificNotationString(v)
	default:
		return fmt.Sprintf("%v", value)
	}
}

// convertScientificNotationString converts a string in scientific notation to a regular integer string
// if it represents a whole number, otherwise returns the original string
func convertScientificNotationString(s string) string {
	// Try to parse as float64 to detect scientific notation
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		// Not a number, return as-is
		return s
	}

	// Check if it's a whole number (no fractional part)
	if f == math.Trunc(f) && (strings.Contains(s, "e") || strings.Contains(s, "E")) {
		// It's scientific notation for an integer, convert to regular format
		return fmt.Sprintf("%.0f", f)
	}

	// Return original string for non-scientific notation or actual floats
	return s
}

// resolveContextInPayload resolves {{CONTEXT:type.field}} markers in a payload
// It preserves the original type of the resolved value (int, float, bool, etc.)
func (pe *PlannedExecutor) resolveContextInPayload(payload interface{}) interface{} {
	switch v := payload.(type) {
	case string:
		// Check if the entire string is a context marker (for type preservation)
		if contextMarkerPattern.MatchString(v) {
			// Try to resolve and preserve type
			resolved := pe.resolveContextWithType(v)
			return resolved
		}
		return pe.resolveContextInString(v)
	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, val := range v {
			result[key] = pe.resolveContextInPayload(val)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, val := range v {
			result[i] = pe.resolveContextInPayload(val)
		}
		return result
	default:
		return payload
	}
}

// resolveContextWithType resolves a context marker and preserves the original type
// If the string is exactly "{{CONTEXT:type.field}}", it returns the actual value with its type
// If it's a string with embedded markers, it returns a string
func (pe *PlannedExecutor) resolveContextWithType(s string) interface{} {
	// Check if the entire string is a single context marker
	matches := contextMarkerPattern.FindAllStringSubmatch(s, -1)
	if len(matches) == 1 && matches[0][0] == s {
		// Entire string is one marker - preserve type
		markerPath := matches[0][1]
		pathParts := strings.Split(markerPath, ".")
		if len(pathParts) < 2 {
			return s // Return original if invalid format
		}

		contextType := pathParts[0]
		fieldName := pathParts[1]
		fieldPath := pathParts[1:]

		// Look up the context with smart fallback
		contextData, _ := pe.testContext.GetContextWithFallback(contextType, fieldName)
		if contextData == nil {
			return s // Keep original if context not found
		}

		// Navigate to the field and return with original type
		value := extractFieldValue(contextData, fieldPath)
		if value == nil {
			return s // Keep original if field not found
		}

		// Return the value with its original type (int, float, bool, string, etc.)
		return value
	}

	// Multiple markers or embedded in text - resolve to string
	return pe.resolveContextInString(s)
}

// substitutePathParams replaces {paramName} placeholders with actual values from pathParams
func (pe *PlannedExecutor) substitutePathParams(path string, pathParams map[string]interface{}) string {
	if pathParams == nil || len(pathParams) == 0 {
		return path
	}

	result := path
	for paramName, paramValue := range pathParams {
		placeholder := fmt.Sprintf("{%s}", paramName)
		// Use formatContextValue to avoid scientific notation for large integers
		valueStr := formatContextValue(paramValue)
		// URL-encode the value to handle special characters like !@#$%^&*
		encodedValue := url.PathEscape(valueStr)
		result = strings.Replace(result, placeholder, encodedValue, -1)
	}

	return result
}

// resolveContextInPathParams resolves context markers in path parameter values
func (pe *PlannedExecutor) resolveContextInPathParams(pathParams map[string]interface{}) map[string]interface{} {
	if pathParams == nil || len(pathParams) == 0 {
		return pathParams
	}

	resolved := make(map[string]interface{})
	for key, value := range pathParams {
		// Use formatContextValue to avoid scientific notation for large integers
		valueStr := formatContextValue(value)
		resolvedStr := pe.resolveContextInString(valueStr)
		resolved[key] = resolvedStr
	}
	return resolved
}

// addQueryParams adds query parameters to a URL
func (pe *PlannedExecutor) addQueryParams(baseURL string, queryParams map[string]interface{}) string {
	if queryParams == nil || len(queryParams) == 0 {
		return baseURL
	}

	// Parse the base URL to handle existing query params
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		// If parsing fails, fall back to simple concatenation
		separator := "?"
		if strings.Contains(baseURL, "?") {
			separator = "&"
		}
		var params []string
		for key, value := range queryParams {
			// Use formatContextValue to avoid scientific notation, then resolve context
			valueStr := formatContextValue(value)
			valueStr = pe.resolveContextInString(valueStr)
			params = append(params, fmt.Sprintf("%s=%s", url.QueryEscape(key), url.QueryEscape(valueStr)))
		}
		return baseURL + separator + strings.Join(params, "&")
	}

	// Add query params to existing ones
	q := parsedURL.Query()
	for key, value := range queryParams {
		// Use formatContextValue to avoid scientific notation, then resolve context
		valueStr := formatContextValue(value)
		valueStr = pe.resolveContextInString(valueStr)
		q.Set(key, valueStr)
	}
	parsedURL.RawQuery = q.Encode()

	return parsedURL.String()
}

// doHTTPRequest performs an HTTP request and returns status code, body, headers, and error
func (pe *PlannedExecutor) doHTTPRequest(ctx context.Context, method, url string, payload interface{}, headers map[string]interface{}) (int, []byte, http.Header, error) {
	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return 0, nil, nil, fmt.Errorf("failed to marshal payload: %w", err)
		}
		body = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set default headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Set custom headers with context resolution
	for key, value := range headers {
		if strVal, ok := value.(string); ok {
			// Resolve any context markers in header values
			resolvedVal := pe.resolveContextInString(strVal)
			req.Header.Set(key, resolvedVal)
		}
	}

	// Add auth headers if available
	if pe.authManager != nil {
		authHeaders, err := pe.authManager.GetAuthHeaders(ctx, nil)
		if err == nil {
			for key, value := range authHeaders {
				req.Header.Set(key, value)
			}
		}
	}

	resp, err := pe.httpClient.Do(req)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, resp.Header, fmt.Errorf("failed to read response: %w", err)
	}

	return resp.StatusCode, responseBody, resp.Header, nil
}

// checkExpectedStatus checks if the actual status matches expected with flexible matching.
// For 2xx success codes: any 2xx is acceptable (200, 201, 204 are all valid success responses)
// For 4xx/5xx error codes: exact match required (400 vs 404 vs 422 have different meanings)
func (pe *PlannedExecutor) checkExpectedStatus(actual, expected int) bool {
	// If expected is 0, any 2xx is considered success
	if expected == 0 {
		return actual >= 200 && actual < 300
	}

	// For 2xx success codes, accept any 2xx response
	// This handles cases like expecting 200 but getting 201 Created
	if expected >= 200 && expected < 300 {
		return actual >= 200 && actual < 300
	}

	// For error codes (4xx, 5xx), require exact match
	// Different error codes have different meanings:
	// - 400 Bad Request vs 404 Not Found vs 422 Unprocessable Entity
	// - 500 Internal Error vs 502 Bad Gateway vs 503 Service Unavailable
	return actual == expected
}

// captureContextWithFallback captures fields from response, with intelligent fallback to request payload
// This handles APIs that return generic responses instead of the created resource (common pattern)
func (pe *PlannedExecutor) captureContextWithFallback(capture *ContextCapture, responseBody []byte, requestPayload interface{}) map[string]interface{} {
	captured := make(map[string]interface{})
	missingFields := make([]string, 0)

	// First, try to capture from response body
	var respData map[string]interface{}
	if err := json.Unmarshal(responseBody, &respData); err != nil {
		log.Printf("⚠️ Failed to parse response for context capture: %v", err)
		respData = make(map[string]interface{})
	}

	// Try to capture each field from response
	for _, field := range capture.Fields {
		if value, ok := respData[field]; ok {
			captured[field] = value
		} else {
			missingFields = append(missingFields, field)
		}
	}

	// If we're missing fields, try to capture from request payload (common for POST endpoints
	// that return generic responses like {"code": 200, "message": "success"})
	if len(missingFields) > 0 && requestPayload != nil {
		var payloadData map[string]interface{}

		// Convert payload to map
		switch p := requestPayload.(type) {
		case map[string]interface{}:
			payloadData = p
		default:
			// Try JSON marshal/unmarshal to convert
			if jsonBytes, err := json.Marshal(requestPayload); err == nil {
				json.Unmarshal(jsonBytes, &payloadData)
			}
		}

		if payloadData != nil {
			for _, field := range missingFields {
				if value, ok := payloadData[field]; ok {
					// Don't capture unresolved context markers - they're not actual values
					if strVal, isString := value.(string); isString && contextMarkerPattern.MatchString(strVal) {
						log.Printf("⚠️ Skipping capture of %s from request payload - contains unresolved context marker: %s", field, strVal)
						continue
					}
					captured[field] = value
					log.Printf("📦 Captured %s from request payload (response didn't contain it)", field)
				}
			}
		}
	}

	if len(captured) > 0 {
		pe.testContext.StoreContext(capture.StoreAs, captured)
		log.Printf("📦 Stored context: %s = %v", capture.StoreAs, captured)
	} else {
		log.Printf("⚠️ No context captured for %s - fields not found in response or request", capture.StoreAs)
	}

	return captured
}

// captureContextWithSmartFallback uses SmartContextCapture for intelligent context extraction
// It handles various API response patterns including wrapped responses, Location headers, and nested IDs
func (pe *PlannedExecutor) captureContextWithSmartFallback(endpoint string, capture *ContextCapture, responseBody []byte, respHeaders http.Header, requestPayload interface{}) map[string]interface{} {
	// Use SmartContextCapture if available
	if pe.smartContextCapture != nil {
		result := pe.smartContextCapture.CaptureContext(
			endpoint,
			responseBody,
			respHeaders,
			requestPayload,
			capture.Fields,
		)

		if result.Success && len(result.Captured) > 0 {
			// Filter out boundary/edge case values before storing
			filtered := filterBoundaryValues(result.Captured)
			if len(filtered) > 0 {
				pe.testContext.StoreContext(capture.StoreAs, filtered)
				log.Printf("📦 Smart captured context: %s = %v (source: %s)", capture.StoreAs, filtered, result.Source)
				return filtered
			} else {
				log.Printf("⚠️ Skipped context capture - all values appear to be boundary/edge cases: %v", result.Captured)
				return nil
			}
		}

		// Log failure info for debugging
		if !result.Success {
			log.Printf("⚠️ Smart context capture failed: %s", result.FailureInfo)
		}
	}

	// Fallback to legacy capture method
	return pe.captureContextWithFallback(capture, responseBody, requestPayload)
}

// filterBoundaryValues removes values that appear to be boundary/edge case test values
// These values shouldn't be captured as valid context for subsequent tests
func filterBoundaryValues(data map[string]interface{}) map[string]interface{} {
	filtered := make(map[string]interface{})

	for key, value := range data {
		if !isBoundaryValue(value) {
			filtered[key] = value
		} else {
			log.Printf("   🚫 Filtered boundary value: %s = %v", key, value)
		}
	}

	return filtered
}

// isBoundaryValue checks if a value is likely a boundary/edge case test value
func isBoundaryValue(value interface{}) bool {
	switch v := value.(type) {
	case float64:
		return isNumericBoundary(v)
	case int:
		return isNumericBoundary(float64(v))
	case int64:
		return isNumericBoundary(float64(v))
	case string:
		// Check for obviously invalid test strings
		lower := strings.ToLower(v)
		if lower == "" || lower == "invalid" || lower == "test" || lower == "null" {
			return true
		}
		// Check for extremely long strings (likely boundary tests)
		if len(v) > 1000 {
			return true
		}
	}
	return false
}

// isNumericBoundary checks if a number is a known boundary/edge case value
func isNumericBoundary(v float64) bool {
	// Int64 max/min boundaries
	const int64Max = 9223372036854775807
	const int64Min = -9223372036854775808
	const int32Max = 2147483647
	const int32Min = -2147483648

	// Check for boundary values (with some tolerance for floating point)
	boundaries := []float64{
		int64Max, int64Min, int64Max + 1, int64Min - 1,
		int32Max, int32Min, int32Max + 1, int32Min - 1,
		0, -1, -999999999,
		1e18, -1e18, // Very large numbers
	}

	for _, b := range boundaries {
		if v == b || (v > 1e15 && v > b*0.99 && v < b*1.01) {
			return true
		}
	}

	// Values over 10^15 are suspicious (likely boundary tests)
	if v > 1e15 || v < -1e15 {
		return true
	}

	return false
}

// RunSavedSuite runs a saved test suite file without needing AI regeneration
func (pe *PlannedExecutor) RunSavedSuite(ctx context.Context, suitePath string) (*SuiteExecutionResult, error) {
	log.Printf("📂 Loading test suite from: %s", suitePath)

	suite, err := LoadTestSuite(suitePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load test suite: %w", err)
	}

	log.Printf("✅ Loaded suite: %s (%d tests)", suite.Name, len(suite.Tests))

	// Create validator with schemas from suite
	validator := NewResponseValidator(suite.Schemas)

	// Override base URL if set in suite
	if suite.BaseURL != "" {
		pe.baseURL = suite.BaseURL
	}

	return pe.executeSuite(ctx, suite, validator)
}

// executeSuite executes a test suite with validation
func (pe *PlannedExecutor) executeSuite(ctx context.Context, suite *TestSuite, validator *ResponseValidator) (*SuiteExecutionResult, error) {
	log.Printf("⚡ Executing test suite: %s (%d tests)", suite.Name, len(suite.Tests))

	startTime := time.Now()
	results := make([]PersistedTestResult, 0, len(suite.Tests))
	pe.testContext.Reset()

	passed, failed, skipped, validationErrors := 0, 0, 0, 0

	for i, test := range suite.Tests {
		if test.Skip {
			results = append(results, PersistedTestResult{
				Test:       test,
				Skipped:    true,
				SkipReason: test.SkipReason,
			})
			skipped++
			log.Printf("⏭️  [%d/%d] Skipped: %s - %s", i+1, len(suite.Tests), test.Name, test.SkipReason)
			continue
		}

		log.Printf("▶️  [%d/%d] %s %s - %s", i+1, len(suite.Tests), test.Method, test.Path, test.Name)

		result := pe.executePersistedTest(ctx, &test, validator)
		results = append(results, result)

		if result.Skipped {
			skipped++
			log.Printf("⏭️  Skipped: %s", result.SkipReason)
		} else if result.Passed {
			if result.Validation != nil && !result.Validation.Valid {
				validationErrors++
				log.Printf("⚠️  Passed (status OK) but validation failed: %d errors", len(result.Validation.Errors))
			} else {
				passed++
				log.Printf("✅ Passed (status=%d, time=%v)", result.StatusCode, result.ResponseTime)
			}
		} else {
			failed++
			log.Printf("❌ Failed (expected=%d, got=%d): %s", test.ExpectedStatus, result.StatusCode, result.Error)
		}
	}

	duration := time.Since(startTime)
	totalExecuted := len(suite.Tests) - skipped
	passRate := 0.0
	if totalExecuted > 0 {
		passRate = float64(passed) / float64(totalExecuted) * 100
	}

	summary := SuiteExecutionSummary{
		TotalTests:       len(suite.Tests),
		Passed:           passed,
		Failed:           failed,
		Skipped:          skipped,
		ValidationErrors: validationErrors,
		PassRate:         passRate,
		TotalTime:        duration,
	}

	log.Printf("📊 Suite execution complete: %d/%d passed (%.1f%%) in %v",
		passed, totalExecuted, passRate, duration)

	return &SuiteExecutionResult{
		Suite:       suite,
		Results:     results,
		Summary:     summary,
		ExecutedAt:  startTime,
		Duration:    duration,
		ContextData: pe.testContext.GetAllContext(),
	}, nil
}

// executePersistedTest runs a single persisted test with validation
func (pe *PlannedExecutor) executePersistedTest(ctx context.Context, test *PersistedTest, validator *ResponseValidator) PersistedTestResult {
	result := PersistedTestResult{Test: *test}
	startTime := time.Now()

	// Resolve context markers in path params
	resolvedPathParams := pe.resolveContextInPathParams(test.PathParams)

	// Substitute path parameters
	resolvedPath := pe.substitutePathParams(test.Path, resolvedPathParams)

	// Resolve remaining context markers in path
	resolvedPath = pe.resolveContextInString(resolvedPath)

	// Store the resolved path in the result for analysis
	result.ResolvedPath = resolvedPath

	// Check for unresolved markers
	if HasUnresolvedMarkersInString(resolvedPath) {
		result.Skipped = true
		result.SkipReason = fmt.Sprintf("Unresolved context in path: %s", resolvedPath)
		return result
	}

	// Resolve context in payload
	var resolvedPayload interface{}
	if test.Payload != nil {
		resolvedPayload = pe.resolveContextInPayload(test.Payload)
		// Store the resolved payload in the result for analysis
		result.ResolvedPayload = resolvedPayload
	}

	// Build full URL
	fullURL := pe.baseURL + resolvedPath
	if len(test.QueryParams) > 0 {
		fullURL = pe.addQueryParams(fullURL, test.QueryParams)
	}

	// Execute HTTP request
	statusCode, responseBody, respHeaders, err := pe.doHTTPRequest(ctx, test.Method, fullURL, resolvedPayload, test.Headers)
	result.ResponseTime = time.Since(startTime)

	if err != nil {
		result.Error = err.Error()
		result.Passed = false
		return result
	}

	result.StatusCode = statusCode

	// Parse response body
	if len(responseBody) > 0 {
		var respData interface{}
		if err := json.Unmarshal(responseBody, &respData); err == nil {
			result.ResponseBody = respData
		}
	}

	// Check status code
	result.Passed = pe.checkExpectedStatus(statusCode, test.ExpectedStatus)

	// Validate response if status code passed
	if result.Passed && validator != nil {
		result.Validation = validator.ValidateResponse(responseBody, statusCode, test)
	}

	// Capture context if configured
	if test.ContextCapture != nil && test.ContextCapture.Enabled && statusCode >= 200 && statusCode < 300 {
		endpoint := test.Method + " " + test.Path
		captured := pe.captureContextWithSmartFallback(endpoint, test.ContextCapture, responseBody, respHeaders, test.Payload)
		result.ContextCaptured = captured
	}

	return result
}
