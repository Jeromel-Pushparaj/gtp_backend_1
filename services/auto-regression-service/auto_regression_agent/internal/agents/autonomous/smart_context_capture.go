package autonomous

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

// SmartContextCapture provides intelligent context capture from API responses
// It handles various response patterns found in real-world APIs:
// - Direct ID in response body ({"id": 123, "name": "test"})
// - Wrapped responses ({"data": {"id": 123}, "status": "success"})
// - Generic responses with ID in request ({"code": 200, "message": "created"})
// - Location header with resource URL
// - Nested IDs ({"pet": {"id": 123}})
type SmartContextCapture struct {
	// Schema information from OpenAPI spec for intelligent field detection
	responseSchemas map[string]ResponseSchemaInfo
	// Known ID field patterns across different APIs
	idFieldPatterns []string
}

// ResponseSchemaInfo contains schema info for a response
type ResponseSchemaInfo struct {
	Endpoint       string   // e.g., "POST /pet"
	ResourceType   string   // e.g., "pet"
	ExpectedFields []string // Fields expected in response
	IDFieldPath    string   // JSON path to ID field (e.g., "data.id", "id")
	WrapperField   string   // If response is wrapped (e.g., "data", "result")
	IDFieldNames   []string // Possible ID field names (e.g., ["id", "petId", "_id"])
}

// SmartCaptureResult contains the result of smart context capture
type SmartCaptureResult struct {
	Captured    map[string]interface{} // Captured fields
	Source      string                 // Where data came from: "response_body", "response_header", "request_payload"
	IDField     string                 // Which field contained the ID
	Success     bool                   // Whether capture was successful
	FailureInfo string                 // Why capture failed (if applicable)
}

// NewSmartContextCapture creates a new smart context capture instance
func NewSmartContextCapture() *SmartContextCapture {
	return &SmartContextCapture{
		responseSchemas: make(map[string]ResponseSchemaInfo),
		idFieldPatterns: []string{
			"id", "Id", "ID", "_id",
			"%sId", "%s_id", "%sID", // petId, pet_id, petID
			"uuid", "UUID", "Uuid",
			"key", "code",
			"identifier", "resourceId",
		},
	}
}

// RegisterSchema registers response schema info for an endpoint
func (scc *SmartContextCapture) RegisterSchema(endpoint string, schema ResponseSchemaInfo) {
	scc.responseSchemas[endpoint] = schema
}

// RegisterSchemasFromSpec registers schemas from an OpenAPI spec
func (scc *SmartContextCapture) RegisterSchemasFromSpec(spec map[string]interface{}) {
	paths, ok := spec["paths"].(map[string]interface{})
	if !ok {
		return
	}

	for path, pathItem := range paths {
		pathData, ok := pathItem.(map[string]interface{})
		if !ok {
			continue
		}

		for method, opData := range pathData {
			if !isHTTPMethod(method) {
				continue
			}

			operation, ok := opData.(map[string]interface{})
			if !ok {
				continue
			}

			endpoint := strings.ToUpper(method) + " " + path
			resourceType := inferResourceTypeFromPath(path)

			schema := ResponseSchemaInfo{
				Endpoint:     endpoint,
				ResourceType: resourceType,
				IDFieldNames: scc.generateIDFieldNames(resourceType),
			}

			// Extract expected fields from response schema
			if responses, ok := operation["responses"].(map[string]interface{}); ok {
				schema.ExpectedFields = scc.extractExpectedFields(responses)
				schema.WrapperField = scc.detectWrapperField(responses)
				schema.IDFieldPath = scc.detectIDFieldPath(responses, resourceType)
			}

			scc.responseSchemas[endpoint] = schema
		}
	}
}

// generateIDFieldNames generates possible ID field names for a resource type
func (scc *SmartContextCapture) generateIDFieldNames(resourceType string) []string {
	names := make([]string, 0, len(scc.idFieldPatterns))
	for _, pattern := range scc.idFieldPatterns {
		if strings.Contains(pattern, "%s") {
			names = append(names, fmt.Sprintf(pattern, resourceType))
			names = append(names, fmt.Sprintf(pattern, strings.Title(resourceType)))
		} else {
			names = append(names, pattern)
		}
	}
	return names
}

// isHTTPMethod checks if a string is an HTTP method
func isHTTPMethod(s string) bool {
	methods := []string{"get", "post", "put", "patch", "delete", "head", "options"}
	s = strings.ToLower(s)
	for _, m := range methods {
		if s == m {
			return true
		}
	}
	return false
}

// inferResourceTypeFromPath extracts resource type from path
func inferResourceTypeFromPath(path string) string {
	path = strings.TrimPrefix(path, "/")
	parts := strings.Split(path, "/")

	// Find the last non-parameter segment
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] != "" && !strings.HasPrefix(parts[i], "{") {
			return strings.TrimSuffix(parts[i], "s") // Remove plural
		}
	}
	return ""
}

// CaptureContext intelligently captures context from an API response
// It tries multiple strategies in order of preference
func (scc *SmartContextCapture) CaptureContext(
	endpoint string,
	responseBody []byte,
	responseHeaders http.Header,
	requestPayload interface{},
	fieldsToCapture []string,
) *SmartCaptureResult {
	result := &SmartCaptureResult{
		Captured: make(map[string]interface{}),
		Success:  false,
	}

	// Parse response body
	var respData interface{}
	if len(responseBody) > 0 {
		if err := json.Unmarshal(responseBody, &respData); err != nil {
			log.Printf("⚠️ Failed to parse response body: %v", err)
		}
	}

	// Get schema info for this endpoint
	schema := scc.responseSchemas[endpoint]

	// Strategy 1: Direct field extraction from response body
	if respData != nil {
		captured := scc.extractFromResponse(respData, fieldsToCapture, schema)
		if len(captured) > 0 {
			result.Captured = captured
			result.Source = "response_body"
			result.Success = true
			log.Printf("✅ Captured %d fields from response body", len(captured))
			return result
		}
	}

	// Strategy 2: Unwrap nested response and extract
	if respData != nil && schema.WrapperField != "" {
		if wrapped := scc.unwrapResponse(respData, schema.WrapperField); wrapped != nil {
			captured := scc.extractFromResponse(wrapped, fieldsToCapture, schema)
			if len(captured) > 0 {
				result.Captured = captured
				result.Source = "response_body_unwrapped"
				result.Success = true
				log.Printf("✅ Captured %d fields from unwrapped response", len(captured))
				return result
			}
		}
	}

	// Strategy 3: Extract from Location header (common for POST responses)
	if location := responseHeaders.Get("Location"); location != "" {
		if id := scc.extractIDFromLocationHeader(location); id != "" {
			result.Captured["id"] = id
			result.Source = "response_header"
			result.IDField = "Location"
			result.Success = true
			log.Printf("✅ Captured ID from Location header: %s", id)
			return result
		}
	}

	// Strategy 4: Smart ID detection - look for any ID-like field
	if respData != nil {
		if id, fieldName := scc.findAnyIDField(respData, schema.IDFieldNames); id != nil {
			result.Captured["id"] = id
			result.Source = "response_body"
			result.IDField = fieldName
			result.Success = true
			log.Printf("✅ Smart-detected ID in field '%s': %v", fieldName, id)
			return result
		}
	}

	// Strategy 5: Fallback to request payload (for APIs that don't return created resource)
	if requestPayload != nil {
		captured := scc.extractFromPayload(requestPayload, fieldsToCapture)
		if len(captured) > 0 {
			result.Captured = captured
			result.Source = "request_payload"
			result.Success = true
			log.Printf("✅ Captured %d fields from request payload (fallback)", len(captured))
			return result
		}
	}

	result.FailureInfo = "Could not find required fields in response, headers, or request"
	log.Printf("⚠️ Context capture failed: %s", result.FailureInfo)
	return result
}

// extractFromResponse extracts specified fields from response data
func (scc *SmartContextCapture) extractFromResponse(data interface{}, fields []string, schema ResponseSchemaInfo) map[string]interface{} {
	result := make(map[string]interface{})

	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return result
	}

	for _, field := range fields {
		// Try exact match first
		if val, exists := dataMap[field]; exists {
			result[field] = val
			continue
		}

		// Try case-insensitive match
		for key, val := range dataMap {
			if strings.EqualFold(key, field) {
				result[field] = val
				break
			}
		}

		// Try resource-prefixed field names (e.g., "petId" for "id")
		if field == "id" {
			for _, idField := range schema.IDFieldNames {
				if val, exists := dataMap[idField]; exists {
					result[field] = val
					break
				}
			}
		}
	}

	return result
}

// extractExpectedFields extracts field names from OpenAPI response schema
func (scc *SmartContextCapture) extractExpectedFields(responses map[string]interface{}) []string {
	fields := make([]string, 0)

	// Look for 200 or 201 response
	for code, respData := range responses {
		if code != "200" && code != "201" {
			continue
		}

		resp, ok := respData.(map[string]interface{})
		if !ok {
			continue
		}

		// Navigate to schema properties
		content, ok := resp["content"].(map[string]interface{})
		if !ok {
			continue
		}

		for _, contentType := range content {
			ct, ok := contentType.(map[string]interface{})
			if !ok {
				continue
			}
			schema, ok := ct["schema"].(map[string]interface{})
			if !ok {
				continue
			}
			props, ok := schema["properties"].(map[string]interface{})
			if !ok {
				continue
			}
			for fieldName := range props {
				fields = append(fields, fieldName)
			}
		}
	}

	return fields
}

// detectWrapperField detects if responses are wrapped (e.g., {"data": {...}})
func (scc *SmartContextCapture) detectWrapperField(responses map[string]interface{}) string {
	commonWrappers := []string{"data", "result", "response", "payload", "body", "content"}

	for code, respData := range responses {
		if code != "200" && code != "201" {
			continue
		}

		resp, ok := respData.(map[string]interface{})
		if !ok {
			continue
		}

		content, ok := resp["content"].(map[string]interface{})
		if !ok {
			continue
		}

		for _, contentType := range content {
			ct, ok := contentType.(map[string]interface{})
			if !ok {
				continue
			}
			schema, ok := ct["schema"].(map[string]interface{})
			if !ok {
				continue
			}
			props, ok := schema["properties"].(map[string]interface{})
			if !ok {
				continue
			}

			// Check if only wrapper field exists
			if len(props) == 1 || len(props) == 2 {
				for fieldName := range props {
					for _, wrapper := range commonWrappers {
						if strings.EqualFold(fieldName, wrapper) {
							return fieldName
						}
					}
				}
			}
		}
	}

	return ""
}

// detectIDFieldPath detects the path to the ID field in response
func (scc *SmartContextCapture) detectIDFieldPath(responses map[string]interface{}, resourceType string) string {
	idFields := []string{"id", resourceType + "Id", resourceType + "_id", "_id"}

	for code, respData := range responses {
		if code != "200" && code != "201" {
			continue
		}

		resp, ok := respData.(map[string]interface{})
		if !ok {
			continue
		}

		content, ok := resp["content"].(map[string]interface{})
		if !ok {
			continue
		}

		for _, contentType := range content {
			ct, ok := contentType.(map[string]interface{})
			if !ok {
				continue
			}
			schema, ok := ct["schema"].(map[string]interface{})
			if !ok {
				continue
			}
			props, ok := schema["properties"].(map[string]interface{})
			if !ok {
				continue
			}

			// Check for ID field at top level
			for _, idField := range idFields {
				if _, exists := props[idField]; exists {
					return idField
				}
			}
		}
	}

	return "id" // Default to "id"
}

// unwrapResponse extracts data from wrapper field
func (scc *SmartContextCapture) unwrapResponse(data interface{}, wrapperField string) interface{} {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil
	}

	if wrapped, exists := dataMap[wrapperField]; exists {
		return wrapped
	}

	// Try case-insensitive match
	for key, val := range dataMap {
		if strings.EqualFold(key, wrapperField) {
			return val
		}
	}

	return nil
}

// extractIDFromLocationHeader extracts ID from Location header URL
func (scc *SmartContextCapture) extractIDFromLocationHeader(location string) string {
	// Handle common patterns:
	// /api/resources/123
	// /resources/123
	// https://api.example.com/resources/123

	// Remove query string if present
	if idx := strings.Index(location, "?"); idx != -1 {
		location = location[:idx]
	}

	// Split by / and get the last segment
	parts := strings.Split(strings.TrimSuffix(location, "/"), "/")
	if len(parts) > 0 {
		lastPart := parts[len(parts)-1]
		// Verify it looks like an ID (not empty, not a path segment)
		if lastPart != "" && !strings.HasPrefix(lastPart, "{") {
			return lastPart
		}
	}

	return ""
}

// findAnyIDField searches for any ID-like field in the response
func (scc *SmartContextCapture) findAnyIDField(data interface{}, idFieldNames []string) (interface{}, string) {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, ""
	}

	// First, try the configured ID field names
	for _, fieldName := range idFieldNames {
		if val, exists := dataMap[fieldName]; exists {
			return val, fieldName
		}
	}

	// Second, try common ID patterns with case-insensitive matching
	commonPatterns := []string{"id", "_id", "uuid", "key"}
	for key, val := range dataMap {
		keyLower := strings.ToLower(key)
		for _, pattern := range commonPatterns {
			if keyLower == pattern || strings.HasSuffix(keyLower, pattern) {
				return val, key
			}
		}
	}

	// Third, look for nested objects that might contain ID
	for key, val := range dataMap {
		if nestedMap, ok := val.(map[string]interface{}); ok {
			if nestedID, nestedField := scc.findAnyIDField(nestedMap, idFieldNames); nestedID != nil {
				return nestedID, key + "." + nestedField
			}
		}
	}

	return nil, ""
}

// extractFromPayload extracts fields from request payload (fallback)
func (scc *SmartContextCapture) extractFromPayload(payload interface{}, fields []string) map[string]interface{} {
	result := make(map[string]interface{})

	// Convert payload to map
	var payloadMap map[string]interface{}
	switch p := payload.(type) {
	case map[string]interface{}:
		payloadMap = p
	default:
		// Try JSON marshal/unmarshal
		if jsonBytes, err := json.Marshal(payload); err == nil {
			json.Unmarshal(jsonBytes, &payloadMap)
		}
	}

	if payloadMap == nil {
		return result
	}

	for _, field := range fields {
		if val, exists := payloadMap[field]; exists {
			// Skip unresolved context markers
			if strVal, isStr := val.(string); isStr && strings.Contains(strVal, "{{CONTEXT:") {
				continue
			}
			result[field] = val
		}
	}

	return result
}
