package autonomous

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"

	"gitlab.com/tekion/development/toc/poc/opentest/pkg/llm"
)

// contextMarkerRegex matches {{CONTEXT:type.field}} patterns
var contextMarkerRegex = regexp.MustCompile(`\{\{CONTEXT:([^}]+)\}\}`)

// ContextResolver handles AI-powered resolution of {{CONTEXT:type.field}} markers
type ContextResolver struct {
	llmClient *llm.Client
}

// NewContextResolver creates a new context resolver
func NewContextResolver(llmClient *llm.Client) *ContextResolver {
	return &ContextResolver{
		llmClient: llmClient,
	}
}

// ContextResolveRequest contains the test case and available context
type ContextResolveRequest struct {
	TestCase         map[string]interface{}
	AvailableContext map[string]interface{}
	Endpoint         string
	Method           string
}

// ContextResolveResult contains the resolved test case
type ContextResolveResult struct {
	ResolvedTestCase map[string]interface{}
	ContextUsed      map[string]string
	Resolved         bool
}

// ResolveContextMarkers detects and resolves {{CONTEXT:type.field}} markers
// First tries deterministic resolution, then falls back to AI if needed
func (cr *ContextResolver) ResolveContextMarkers(ctx context.Context, req ContextResolveRequest) (*ContextResolveResult, error) {
	// First, check if test case has any context markers
	testCaseJSON, _ := json.Marshal(req.TestCase)
	testCaseStr := string(testCaseJSON)

	markers := contextMarkerRegex.FindAllStringSubmatch(testCaseStr, -1)

	if len(markers) == 0 {
		// No context markers found, return original test case
		return &ContextResolveResult{
			ResolvedTestCase: req.TestCase,
			ContextUsed:      make(map[string]string),
			Resolved:         true,
		}, nil
	}

	// Extract unique markers
	uniqueMarkers := make(map[string]bool)
	for _, match := range markers {
		if len(match) > 1 {
			uniqueMarkers[match[1]] = true
		}
	}

	log.Printf("🔍 Found context markers: %v", getKeys(uniqueMarkers))

	// STEP 1: Try deterministic resolution first (fast, reliable)
	resolvedJSON, contextUsed, allResolved := cr.resolveDeterministic(testCaseStr, req.AvailableContext)

	if allResolved {
		// All markers resolved deterministically
		var resolvedTestCase map[string]interface{}
		if err := json.Unmarshal([]byte(resolvedJSON), &resolvedTestCase); err != nil {
			return nil, fmt.Errorf("failed to parse resolved test case: %w", err)
		}

		log.Printf("✅ All %d context markers resolved deterministically", len(contextUsed))
		for marker, value := range contextUsed {
			log.Printf("   %s → %s", marker, value)
		}

		return &ContextResolveResult{
			ResolvedTestCase: resolvedTestCase,
			ContextUsed:      contextUsed,
			Resolved:         true,
		}, nil
	}

	// STEP 2: If deterministic resolution partially failed, try AI as fallback
	if cr.llmClient != nil {
		log.Printf("⚠️  Deterministic resolution incomplete, trying AI fallback...")
		return cr.resolveWithAI(ctx, req, uniqueMarkers)
	}

	// STEP 3: If no AI available and not all resolved, return partial result
	var resolvedTestCase map[string]interface{}
	if err := json.Unmarshal([]byte(resolvedJSON), &resolvedTestCase); err != nil {
		return nil, fmt.Errorf("failed to parse partially resolved test case: %w", err)
	}

	log.Printf("⚠️  Only %d of %d context markers resolved (no AI fallback available)", len(contextUsed), len(uniqueMarkers))

	return &ContextResolveResult{
		ResolvedTestCase: resolvedTestCase,
		ContextUsed:      contextUsed,
		Resolved:         false, // Mark as not fully resolved
	}, nil
}

// resolveDeterministic resolves {{CONTEXT:type.field}} markers directly from available context
// Returns: resolved JSON string, map of resolved markers, and whether all markers were resolved
func (cr *ContextResolver) resolveDeterministic(testCaseJSON string, availableContext map[string]interface{}) (string, map[string]string, bool) {
	contextUsed := make(map[string]string)
	unresolvedCount := 0
	result := testCaseJSON

	// Find all markers and resolve them
	matches := contextMarkerRegex.FindAllStringSubmatch(testCaseJSON, -1)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		fullMarker := match[0] // e.g., {{CONTEXT:pet.id}}
		markerPath := match[1] // e.g., pet.id

		// Skip if already resolved
		if _, exists := contextUsed[fullMarker]; exists {
			continue
		}

		// Parse the marker path (type.field)
		parts := splitMarkerPath(markerPath)
		if len(parts) < 2 {
			log.Printf("⚠️  Invalid marker format: %s", fullMarker)
			unresolvedCount++
			continue
		}

		contextType := parts[0] // e.g., "pet"
		fieldPath := parts[1:]  // e.g., ["id"] or ["category", "id"]

		// Look up the context type
		contextData, ok := availableContext[contextType]
		if !ok {
			log.Printf("⚠️  No context available for type '%s'", contextType)
			unresolvedCount++
			continue
		}

		// Navigate to the field
		value := extractFieldValue(contextData, fieldPath)
		if value == nil {
			log.Printf("⚠️  Field '%s' not found in context '%s'", strings.Join(fieldPath, "."), contextType)
			unresolvedCount++
			continue
		}

		// Convert value to string for replacement
		valueStr := fmt.Sprintf("%v", value)
		contextUsed[fullMarker] = valueStr

		// Replace in the result - handle both quoted and unquoted contexts
		// For strings in JSON, they appear as "{{CONTEXT:...}}" so replace the whole thing
		// For numbers, we might have {{CONTEXT:...}} without quotes
		result = strings.ReplaceAll(result, fmt.Sprintf(`"%s"`, fullMarker), formatJSONValue(value))
		result = strings.ReplaceAll(result, fullMarker, valueStr)
	}

	return result, contextUsed, unresolvedCount == 0
}

// splitMarkerPath splits a marker path like "pet.id" or "order.items.0.id" into parts
func splitMarkerPath(path string) []string {
	return strings.Split(path, ".")
}

// extractFieldValue extracts a nested field value from context data
func extractFieldValue(data interface{}, fieldPath []string) interface{} {
	if len(fieldPath) == 0 {
		return data
	}

	field := fieldPath[0]
	remaining := fieldPath[1:]

	switch v := data.(type) {
	case map[string]interface{}:
		if val, ok := v[field]; ok {
			return extractFieldValue(val, remaining)
		}
		// Try case-insensitive match
		for key, val := range v {
			if strings.EqualFold(key, field) {
				return extractFieldValue(val, remaining)
			}
		}
	case []interface{}:
		// Handle array index (e.g., "0", "1")
		var idx int
		if _, err := fmt.Sscanf(field, "%d", &idx); err == nil {
			if idx >= 0 && idx < len(v) {
				return extractFieldValue(v[idx], remaining)
			}
		}
	}

	return nil
}

// formatJSONValue formats a value for JSON replacement
func formatJSONValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		// Return as quoted string
		return fmt.Sprintf(`"%s"`, v)
	case float64, float32, int, int64, int32:
		// Return as number (no quotes)
		return fmt.Sprintf("%v", v)
	case bool:
		// Return as boolean
		return fmt.Sprintf("%t", v)
	default:
		// For complex types, marshal to JSON
		if b, err := json.Marshal(v); err == nil {
			return string(b)
		}
		return fmt.Sprintf(`"%v"`, v)
	}
}

// resolveWithAI uses AI to intelligently resolve context markers
func (cr *ContextResolver) resolveWithAI(ctx context.Context, req ContextResolveRequest, markers map[string]bool) (*ContextResolveResult, error) {
	testCaseJSON, _ := json.MarshalIndent(req.TestCase, "", "  ")
	contextJSON, _ := json.MarshalIndent(req.AvailableContext, "", "  ")

	markerList := getKeys(markers)

	prompt := fmt.Sprintf(`You are an API testing expert. Resolve context markers in this test case using available context from previous API responses.

**Test Case to Execute:**
%s %s

**Test Case with Context Markers:**
%s

**Context Markers Found:**
%v

**Available Context from Previous Responses:**
%s

**Your Task:**
1. Analyze the test case and identify all {{CONTEXT:type.field}} markers
2. Look at the available context from previous API responses
3. Intelligently match context markers to the appropriate values
4. Replace ALL markers with actual values from the context
5. Return the complete resolved test case

**Context Marker Format:**
- {{CONTEXT:resource.id}} → Use "id" field from "resource" context
- {{CONTEXT:user.email}} → Use "email" field from "user" context
- {{CONTEXT:order.orderId}} → Use "orderId" field from "order" context

**Resolution Rules:**
1. Match the context type (before the dot) to find the right context object
2. Extract the field (after the dot) from that context object
3. If exact match not found, use intelligent matching (e.g., "resourceId" can match "id" from resource context)
4. Preserve data types (numbers stay numbers, strings stay strings)
5. If a marker cannot be resolved, use null and note it in context_used

CRITICAL FORMATTING RULES:
1. Your response MUST be ONLY valid JSON - no explanations, no markdown, no code blocks
2. Do NOT wrap the JSON in markdown code blocks (no backtick markers)
3. Do NOT include any text before or after the JSON
4. Start your response directly with the opening brace {
5. End your response with the closing brace }

Required JSON structure:
{
  "resolved_test_case": {
    "name": "...",
    "method": "%s",
    "path": "%s",
    "path_params": {"id": "actual_value_here"},
    "query_params": {...},
    "payload": {...},
    "expected_status": 200,
    "description": "..."
  },
  "context_used": {
    "{{CONTEXT:resource.id}}": "12345",
    "{{CONTEXT:user.email}}": "test@example.com"
  },
  "resolved": true
}`, req.Method, req.Endpoint, string(testCaseJSON), markerList, string(contextJSON), req.Method, req.Endpoint)

	response, err := cr.llmClient.GenerateCompletion(ctx, prompt, llm.CompletionOptions{
		Temperature: 0.2, // Low temperature for consistent resolution
		MaxTokens:   4096,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to resolve context with AI: %w", err)
	}

	// Parse AI response
	cleanedResponse := stripMarkdownCodeBlocks(response)

	var aiResult struct {
		ResolvedTestCase map[string]interface{} `json:"resolved_test_case"`
		ContextUsed      map[string]string      `json:"context_used"`
		Resolved         bool                   `json:"resolved"`
	}

	if err := json.Unmarshal([]byte(cleanedResponse), &aiResult); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	log.Printf("✅ AI resolved %d context markers", len(aiResult.ContextUsed))
	for marker, value := range aiResult.ContextUsed {
		log.Printf("   %s → %s", marker, value)
	}

	return &ContextResolveResult{
		ResolvedTestCase: aiResult.ResolvedTestCase,
		ContextUsed:      aiResult.ContextUsed,
		Resolved:         aiResult.Resolved,
	}, nil
}

// getKeys returns keys from a map as a slice
func getKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// HasUnresolvedContextMarkers checks if a test case still contains unresolved {{CONTEXT:...}} markers
func HasUnresolvedContextMarkers(testCase map[string]interface{}) (bool, []string) {
	testCaseJSON, err := json.Marshal(testCase)
	if err != nil {
		return false, nil
	}

	markers := contextMarkerRegex.FindAllStringSubmatch(string(testCaseJSON), -1)
	if len(markers) == 0 {
		return false, nil
	}

	unresolvedMarkers := make([]string, 0, len(markers))
	seen := make(map[string]bool)
	for _, match := range markers {
		if len(match) > 0 && !seen[match[0]] {
			unresolvedMarkers = append(unresolvedMarkers, match[0])
			seen[match[0]] = true
		}
	}

	return len(unresolvedMarkers) > 0, unresolvedMarkers
}

// HasUnresolvedMarkersInString checks if a string contains unresolved {{CONTEXT:...}} markers
func HasUnresolvedMarkersInString(s string) bool {
	return contextMarkerRegex.MatchString(s)
}
