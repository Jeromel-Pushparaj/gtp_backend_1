package autonomous

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"
)

// TestContext manages variables and state across test executions with AI-powered capture and resolution
type TestContext struct {
	mu               sync.RWMutex
	variables        map[string]interface{}            // Stores captured variables (e.g., resource IDs, tokens, etc.)
	contextData      map[string]map[string]interface{} // Stores context by type (e.g., "resource" -> {id: 123, name: "test"})
	primaryContext   map[string]bool                   // Tracks which context types have their primary (first valid) capture
	variableCapture  *VariableCapture                  // AI-powered variable capture (legacy)
	variableResolver *VariableResolver                 // AI-powered variable resolution (legacy)
	contextResolver  *ContextResolver                  // AI-powered context marker resolution (new)
	contextMappings  map[string]string                 // Dynamic context type mappings from LLM analysis (e.g., "resource" -> "pet")
}

// NewTestContext creates a new test context
func NewTestContext() *TestContext {
	return &TestContext{
		variables:       make(map[string]interface{}),
		contextData:     make(map[string]map[string]interface{}),
		primaryContext:  make(map[string]bool),
		contextMappings: make(map[string]string),
	}
}

// SetContextMappings sets the dynamic context type mappings from LLM analysis
// This allows the system to understand API-specific resource type aliases
func (tc *TestContext) SetContextMappings(mappings map[string]string) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.contextMappings = mappings
	if len(mappings) > 0 {
		log.Printf("📋 Loaded %d context type mappings", len(mappings))
		for alias, canonical := range mappings {
			log.Printf("   %s → %s", alias, canonical)
		}
	}
}

// SetAIServices sets the AI-powered services for intelligent variable handling
func (tc *TestContext) SetAIServices(capture *VariableCapture, resolver *VariableResolver) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.variableCapture = capture
	tc.variableResolver = resolver
}

// SetContextResolver sets the AI-powered context resolver
func (tc *TestContext) SetContextResolver(resolver *ContextResolver) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.contextResolver = resolver
}

// Reset clears all variables but preserves AI services
func (tc *TestContext) Reset() {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.variables = make(map[string]interface{})
	tc.contextData = make(map[string]map[string]interface{})
	tc.primaryContext = make(map[string]bool)
}

// StoreContext stores context data by type (e.g., "resource", "user", "order")
// It also stores under normalized canonical names for better cross-reference
// IMPORTANT: Once primary context is captured for a type, it won't be overwritten
// This prevents boundary/negative tests from polluting valid context
func (tc *TestContext) StoreContext(contextType string, data map[string]interface{}) {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	canonicalType := tc.normalizeContextType(contextType)

	// Check if primary context already exists for this type
	if tc.primaryContext[canonicalType] {
		log.Printf("🛡️ Preserved primary context for '%s' (ignoring new data: %v)", canonicalType, data)
		return
	}

	// Store under the original type name
	if tc.contextData[contextType] == nil {
		tc.contextData[contextType] = make(map[string]interface{})
	}
	for key, value := range data {
		tc.contextData[contextType][key] = value
	}

	// Also store under canonical name if different (for better resolution)
	if canonicalType != contextType {
		if tc.contextData[canonicalType] == nil {
			tc.contextData[canonicalType] = make(map[string]interface{})
		}
		for key, value := range data {
			tc.contextData[canonicalType][key] = value
		}
		log.Printf("📦 Stored PRIMARY context: %s (also as %s) = %v", contextType, canonicalType, data)
	} else {
		log.Printf("📦 Stored PRIMARY context: %s = %v", contextType, data)
	}

	// Mark as primary - subsequent stores will be ignored
	tc.primaryContext[canonicalType] = true
}

// normalizeContextType maps various context type names to canonical names
// It first checks dynamic mappings from LLM, then falls back to defaults
// e.g., "resource" -> "pet", "created_user" -> "user"
func (tc *TestContext) normalizeContextType(contextType string) string {
	lower := strings.ToLower(contextType)

	// First check dynamic mappings from LLM analysis (API-specific)
	if canonical, ok := tc.contextMappings[lower]; ok {
		return canonical
	}

	// Fall back to default mappings for common patterns
	defaultMap := map[string]string{
		// Generic resource patterns
		"resource": "resource",
		"item":     "item",
		"entity":   "entity",
		"object":   "object",
		"created":  "created",
		"new":      "new",
		// Common prefixes/suffixes that should strip
		"created_": "",
		"new_":     "",
		// Session-like contexts
		"session":      "session",
		"auth":         "session",
		"login":        "session",
		"token":        "session",
		"user_session": "session",
	}

	if canonical, ok := defaultMap[lower]; ok && canonical != "" {
		return canonical
	}

	// Handle "created_X" or "new_X" patterns - extract the base type
	if strings.HasPrefix(lower, "created_") {
		return strings.TrimPrefix(lower, "created_")
	}
	if strings.HasPrefix(lower, "new_") {
		return strings.TrimPrefix(lower, "new_")
	}

	return contextType
}

// normalizeContextTypeStatic is a static version for use outside of TestContext (e.g., tests)
func normalizeContextType(contextType string) string {
	tc := &TestContext{contextMappings: make(map[string]string)}
	return tc.normalizeContextType(contextType)
}

// GetContext retrieves context data by type
func (tc *TestContext) GetContext(contextType string) map[string]interface{} {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	if data, ok := tc.contextData[contextType]; ok {
		return data
	}
	return nil
}

// GetContextWithFallback tries to find context by type with smart fallback strategies
// It tries: exact match, normalized type, dynamic LLM mappings, aliases, and finally any context with the field
func (tc *TestContext) GetContextWithFallback(contextType string, fieldName string) (map[string]interface{}, string) {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	// 1. Try exact match first
	if data, ok := tc.contextData[contextType]; ok {
		return data, contextType
	}

	// 2. Try normalized/lowercase version
	normalizedType := strings.ToLower(contextType)
	if data, ok := tc.contextData[normalizedType]; ok {
		return data, normalizedType
	}

	// 3. Try canonical name from dynamic LLM mappings
	canonicalType := tc.normalizeContextType(contextType)
	if canonicalType != contextType {
		if data, ok := tc.contextData[canonicalType]; ok {
			log.Printf("   🔄 Context fallback: '%s' → '%s' (via dynamic mapping)", contextType, canonicalType)
			return data, canonicalType
		}
	}

	// 4. Try aliases based on dynamic mappings (reverse lookup)
	aliases := tc.getContextAliases(contextType)
	for _, alias := range aliases {
		if data, ok := tc.contextData[alias]; ok {
			log.Printf("   🔄 Context fallback: '%s' → '%s' (via alias)", contextType, alias)
			return data, alias
		}
	}

	// 5. Try to find ANY context that has the requested field
	// This handles cases where LLM uses completely different naming
	// Prefer specific context types over generic ones (resource, item, etc.)
	var fallbackData map[string]interface{}
	var fallbackType string
	for ctxType, data := range tc.contextData {
		if _, hasField := data[fieldName]; hasField {
			// Skip generic context types if we're looking for a specific one
			if isGenericContextTypeName(ctxType) {
				// Store as fallback but keep looking for specific types
				if fallbackData == nil {
					fallbackData = data
					fallbackType = ctxType
				}
				continue
			}
			log.Printf("   🔄 Fallback: Found field '%s' in context '%s' (requested: '%s')", fieldName, ctxType, contextType)
			return data, ctxType
		}
	}

	// Return generic fallback if no specific type found
	if fallbackData != nil {
		log.Printf("   🔄 Fallback: Found field '%s' in generic context '%s' (requested: '%s')", fieldName, fallbackType, contextType)
		return fallbackData, fallbackType
	}

	return nil, ""
}

// isGenericContextTypeName checks if a context type name is generic
func isGenericContextTypeName(contextType string) bool {
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

// getContextAliases returns aliases for a context type using dynamic LLM mappings
// It finds all aliases that map to the same canonical type
func (tc *TestContext) getContextAliases(contextType string) []string {
	normalizedType := strings.ToLower(contextType)
	var aliases []string

	// First, find the canonical type for the requested context type
	canonicalType := tc.normalizeContextType(contextType)

	// Then find all other types that map to the same canonical type
	for alias, canonical := range tc.contextMappings {
		if canonical == canonicalType && alias != normalizedType {
			aliases = append(aliases, alias)
		}
	}

	// Also add common generic fallbacks
	genericFallbacks := []string{"resource", "item", "entity", "object", "created", "response"}
	for _, fallback := range genericFallbacks {
		if fallback != normalizedType && fallback != canonicalType {
			aliases = append(aliases, fallback)
		}
	}

	return aliases
}

// getContextAliases is a static version for backward compatibility (used in tests)
func getContextAliases(contextType string) []string {
	tc := &TestContext{contextMappings: make(map[string]string)}
	return tc.getContextAliases(contextType)
}

// GetAllContext returns all stored context data
func (tc *TestContext) GetAllContext() map[string]interface{} {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	allContext := make(map[string]interface{})
	for contextType, data := range tc.contextData {
		allContext[contextType] = data
	}
	return allContext
}

// GetAllContextTypes returns all context type names
func (tc *TestContext) GetAllContextTypes() []string {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	types := make([]string, 0, len(tc.contextData))
	for contextType := range tc.contextData {
		types = append(types, contextType)
	}
	return types
}

// Set stores a variable in the context
func (tc *TestContext) Set(key string, value interface{}) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.variables[key] = value
}

// Get retrieves a variable from the context
func (tc *TestContext) Get(key string) (interface{}, bool) {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	value, exists := tc.variables[key]
	return value, exists
}

// GetString retrieves a variable as a string
func (tc *TestContext) GetString(key string) (string, bool) {
	value, exists := tc.Get(key)
	if !exists {
		return "", false
	}

	switch v := value.(type) {
	case string:
		return v, true
	case float64:
		return fmt.Sprintf("%.0f", v), true
	case int:
		return fmt.Sprintf("%d", v), true
	default:
		return fmt.Sprintf("%v", v), true
	}
}

// CaptureFromResponse extracts important fields from API response using AI
func (tc *TestContext) CaptureFromResponse(ctx context.Context, endpoint, method string, statusCode int, responseBody []byte) error {
	// Parse response as JSON
	var responseData map[string]interface{}
	if err := json.Unmarshal(responseBody, &responseData); err != nil {
		// Not JSON or empty response, skip capture
		return nil
	}

	// If AI services are available, use intelligent capture
	if tc.variableCapture != nil {
		result, err := tc.variableCapture.CaptureVariables(ctx, CaptureRequest{
			Endpoint:     endpoint,
			Method:       method,
			ResponseBody: responseData,
			StatusCode:   statusCode,
		})
		if err != nil {
			log.Printf("Warning: AI variable capture failed, falling back to basic capture: %v", err)
		} else if len(result.Variables) > 0 {
			// Store all captured variables
			for key, value := range result.Variables {
				tc.Set(key, value)
				log.Printf("🤖 AI captured: %s = %v", key, value)
			}
			return nil
		}
	}

	// Fallback: Basic capture for common patterns
	tc.basicCapture(endpoint, responseData)
	return nil
}

// CaptureContextFromResponse captures context based on context_capture specification in test case
func (tc *TestContext) CaptureContextFromResponse(ctx context.Context, testCase map[string]interface{}, responseBody []byte) error {
	// Check if test case has context_capture specification
	contextCapture, ok := testCase["context_capture"].(map[string]interface{})
	if !ok {
		return nil // No context capture specified
	}

	// Check if capture is enabled
	enabled, _ := contextCapture["enabled"].(bool)
	if !enabled {
		return nil
	}

	// Get context type and fields to capture
	contextType, _ := contextCapture["store_as"].(string)
	if contextType == "" {
		return nil
	}

	fieldsToCapture, ok := contextCapture["fields"].([]interface{})
	if !ok || len(fieldsToCapture) == 0 {
		return nil
	}

	// Parse response as JSON
	var responseData map[string]interface{}
	if err := json.Unmarshal(responseBody, &responseData); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract specified fields
	capturedData := make(map[string]interface{})
	for _, fieldInterface := range fieldsToCapture {
		field, ok := fieldInterface.(string)
		if !ok {
			continue
		}

		if value, exists := responseData[field]; exists {
			capturedData[field] = value
		}
	}

	if len(capturedData) > 0 {
		tc.StoreContext(contextType, capturedData)
		log.Printf("✅ Captured context from response: %s = %v", contextType, capturedData)
	}

	return nil
}

// basicCapture provides fallback variable capture without AI
func (tc *TestContext) basicCapture(endpoint string, responseData map[string]interface{}) {
	// Generic pattern-based capture for common ID fields
	// Extract resource type from endpoint (e.g., /api/v1/users/{id} -> users)
	pathParts := strings.Split(strings.Trim(endpoint, "/"), "/")
	var resourceType string
	for _, part := range pathParts {
		// Skip common prefixes and find the resource name
		if part != "" && part != "api" && !strings.HasPrefix(part, "v") && !strings.Contains(part, "{") {
			resourceType = strings.TrimSuffix(part, "s") // Remove plural 's' if present
			break
		}
	}

	// Capture ID field if present
	if id, ok := responseData["id"]; ok {
		// Store with resource-specific key if we identified the resource
		if resourceType != "" {
			tc.Set(resourceType+"Id", id)
			tc.Set("last"+strings.Title(resourceType)+"Id", id)
		}
		// Always store generic keys as fallback
		tc.Set("id", id)
		tc.Set("lastId", id)
	}

	// Capture common identifier fields
	commonFields := []string{"uuid", "name", "username", "email", "code", "key", "token"}
	for _, field := range commonFields {
		if value, ok := responseData[field]; ok {
			tc.Set(field, value)
			if resourceType != "" {
				tc.Set(resourceType+strings.Title(field), value)
			}
		}
	}

	// Capture session tokens or auth headers
	if token, ok := responseData["token"]; ok {
		tc.Set("authToken", token)
	}
	if sessionId, ok := responseData["sessionId"]; ok {
		tc.Set("sessionId", sessionId)
	}
}

// ResolveTestCase uses AI to intelligently resolve variables in a test case
func (tc *TestContext) ResolveTestCase(ctx context.Context, endpoint, method string, pathParams, queryParams map[string]string, payload map[string]interface{}) (map[string]string, map[string]string, map[string]interface{}, error) {
	// If AI resolver is available, use it
	if tc.variableResolver != nil {
		tc.mu.RLock()
		availableVars := make(map[string]interface{})
		for k, v := range tc.variables {
			availableVars[k] = v
		}
		tc.mu.RUnlock()

		result, err := tc.variableResolver.ResolveVariables(ctx, ResolveRequest{
			Endpoint:           endpoint,
			Method:             method,
			PathParams:         pathParams,
			QueryParams:        queryParams,
			Payload:            payload,
			AvailableVariables: availableVars,
		})
		if err != nil {
			log.Printf("Warning: AI variable resolution failed, falling back to basic resolution: %v", err)
		} else if result.Resolved {
			log.Printf("🤖 AI resolved variables for %s %s", method, endpoint)
			return result.PathParams, result.QueryParams, result.Payload, nil
		}
	}

	// Fallback: Use basic variable resolution
	resolvedPathParams := make(map[string]string)
	for k, v := range pathParams {
		resolvedPathParams[k] = tc.ResolveVariables(v).(string)
	}

	resolvedQueryParams := make(map[string]string)
	for k, v := range queryParams {
		resolvedQueryParams[k] = tc.ResolveVariables(v).(string)
	}

	resolvedPayload := tc.ResolveVariables(payload).(map[string]interface{})

	return resolvedPathParams, resolvedQueryParams, resolvedPayload, nil
}

// ResolveVariables replaces variable placeholders with actual values
// Supports formats: {{varName}}, ${varName}, $varName
func (tc *TestContext) ResolveVariables(input interface{}) interface{} {
	switch v := input.(type) {
	case string:
		return tc.resolveString(v)
	case map[string]interface{}:
		return tc.resolveMap(v)
	case []interface{}:
		return tc.resolveSlice(v)
	default:
		return input
	}
}

// resolveString resolves variables in a string
func (tc *TestContext) resolveString(s string) string {
	// Match {{varName}}, ${varName}, or $varName patterns
	patterns := []string{
		`\{\{([a-zA-Z0-9_]+)\}\}`, // {{varName}}
		`\$\{([a-zA-Z0-9_]+)\}`,   // ${varName}
		`\$([a-zA-Z0-9_]+)`,       // $varName
	}

	result := s
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		result = re.ReplaceAllStringFunc(result, func(match string) string {
			// Extract variable name
			varName := re.FindStringSubmatch(match)[1]

			// Get value from context
			if value, exists := tc.GetString(varName); exists {
				return value
			}

			// Return original if not found
			return match
		})
	}

	return result
}

// resolveMap resolves variables in a map
func (tc *TestContext) resolveMap(m map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for key, value := range m {
		result[key] = tc.ResolveVariables(value)
	}
	return result
}

// resolveSlice resolves variables in a slice
func (tc *TestContext) resolveSlice(slice []interface{}) []interface{} {
	result := make([]interface{}, len(slice))
	for i, value := range slice {
		result[i] = tc.ResolveVariables(value)
	}
	return result
}
