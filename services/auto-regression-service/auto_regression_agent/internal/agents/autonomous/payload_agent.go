package autonomous

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"gitlab.com/tekion/development/toc/poc/opentest/internal/events"
	"gitlab.com/tekion/development/toc/poc/opentest/pkg/llm"
)

// PayloadAgent is an autonomous agent that generates test payloads
type PayloadAgent struct {
	*Agent
}

// NewPayloadAgent creates a new autonomous payload generator agent
func NewPayloadAgent(
	llmClient *llm.Client,
	eventBus *events.Bus,
	messageBus *events.MessageBus,
	consensus *events.ConsensusEngine,
) *PayloadAgent {
	baseAgent := NewAgent(
		"payload_agent",
		AgentTypePayload,
		[]string{"payload_generation", "test_data_creation", "schema_validation"},
		llmClient,
		eventBus,
		messageBus,
		consensus,
	)

	return &PayloadAgent{
		Agent: baseAgent,
	}
}

// Start starts the payload agent
func (pa *PayloadAgent) Start(ctx context.Context) error {
	log.Printf("🎲 Starting Payload Generator Agent...")

	// Start base agent
	if err := pa.Agent.Start(ctx); err != nil {
		return err
	}

	// Subscribe to strategy_approved events
	go pa.listenToStrategyApproved(ctx)

	log.Printf("✅ Payload Generator Agent ready and listening")
	return nil
}

// listenToStrategyApproved listens for strategy approved events
func (pa *PayloadAgent) listenToStrategyApproved(ctx context.Context) {
	err := pa.EventBus.Subscribe(ctx, events.EventTypeStrategyApproved, func(event *events.Event) error {
		log.Printf("🎲 Payload Agent received strategy_approved event")

		pa.setState(AgentStateProcessing)
		defer pa.setState(AgentStateIdle)

		specID, ok := event.Payload["spec_id"].(string)
		if !ok {
			return fmt.Errorf("spec_id not found in event payload")
		}

		workflowID, ok := event.Payload["workflow_id"].(string)
		if !ok {
			return fmt.Errorf("workflow_id not found in event payload")
		}

		strategy, ok := event.Payload["strategy"].(map[string]interface{})
		if !ok {
			return fmt.Errorf("strategy not found in event payload")
		}

		// Extract spec_path and base_url
		specPath, _ := event.Payload["spec_path"].(string)
		baseURL, _ := event.Payload["base_url"].(string)

		// Store for later use
		pa.Memory.Store(fmt.Sprintf("spec_path:%s", specID), specPath)
		pa.Memory.Store(fmt.Sprintf("base_url:%s", specID), baseURL)

		// Generate test payloads
		return pa.generatePayloads(ctx, specID, workflowID, strategy)
	})

	if err != nil && err != context.Canceled {
		log.Printf("Error subscribing to strategy_approved: %v", err)
	}
}

// generatePayloads generates test payloads based on strategy
func (pa *PayloadAgent) generatePayloads(ctx context.Context, specID, workflowID string, strategy map[string]interface{}) error {
	log.Printf("🎲 Generating test payloads for spec: %s", specID)

	// Load spec file to get schemas
	specPath := fmt.Sprintf("./output/specs/%s.json", specID)
	specData, err := os.ReadFile(specPath)
	if err != nil {
		return fmt.Errorf("failed to read spec file: %w", err)
	}

	var spec map[string]interface{}
	if err := json.Unmarshal(specData, &spec); err != nil {
		return fmt.Errorf("failed to parse spec: %w", err)
	}

	// Try to parse raw_strategy if strategy parsing failed earlier
	workingStrategy := strategy
	if rawStrategy, ok := strategy["raw_strategy"].(string); ok {
		log.Printf("📋 Parsing raw_strategy from failed JSON parse...")
		var parsed map[string]interface{}
		if err := json.Unmarshal([]byte(rawStrategy), &parsed); err == nil {
			workingStrategy = parsed
			log.Printf("✅ Successfully parsed raw_strategy")
		} else {
			log.Printf("Warning: failed to parse raw_strategy: %v", err)
		}
	}

	// Extract execution order from strategy
	executionOrder, ok := workingStrategy["execution_order"].([]interface{})
	if !ok {
		log.Printf("📋 No execution_order in strategy, extracting from phases...")
		executionOrder = pa.extractEndpointsFromPhases(workingStrategy)
	}

	if len(executionOrder) == 0 {
		log.Printf("📋 No endpoints found in phases, extracting from OpenAPI spec...")
		executionOrder = pa.extractEndpointsFromSpec(spec)
	}

	log.Printf("📋 Will generate payloads for %d endpoints", len(executionOrder))

	// Generate payloads for each endpoint
	payloads := make(map[string]interface{})

	for _, endpoint := range executionOrder {
		endpointStr, ok := endpoint.(string)
		if !ok {
			continue
		}

		log.Printf("🎲 Generating payloads for endpoint: %s", endpointStr)

		// Use LLM to generate realistic test data
		payloadData, err := pa.generateEndpointPayloads(ctx, endpointStr, spec)
		if err != nil {
			log.Printf("⚠️  LLM failed for %s: %v, using fallback generator", endpointStr, err)
			// Use fallback generator when LLM fails
			payloadData = pa.generateFallbackPayloads(endpointStr, spec)
		}

		payloads[endpointStr] = payloadData
	}

	// Store payloads in memory
	pa.Memory.Store(fmt.Sprintf("payloads:%s", specID), payloads)

	// Save payloads to file
	outputPath := fmt.Sprintf("./output/payloads/%s-test-payloads.json", specID)
	if err := os.MkdirAll("./output/payloads", 0755); err != nil {
		return fmt.Errorf("failed to create payloads directory: %w", err)
	}

	payloadJSON, err := json.MarshalIndent(payloads, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal payloads data: %w", err)
	}

	if err := os.WriteFile(outputPath, payloadJSON, 0644); err != nil {
		return fmt.Errorf("failed to save payloads file: %w", err)
	}

	log.Printf("✅ Test payloads generated: %s", outputPath)

	// Retrieve spec_path and base_url from memory
	specPathVal, _ := pa.Memory.Recall(fmt.Sprintf("spec_path:%s", specID))
	baseURLVal, _ := pa.Memory.Recall(fmt.Sprintf("base_url:%s", specID))

	var originalSpecPath, baseURL string
	if specPathVal != nil {
		originalSpecPath, _ = specPathVal.(string)
	}
	if baseURLVal != nil {
		baseURL, _ = baseURLVal.(string)
	}

	// Publish payloads_ready event with spec_path and base_url
	return pa.PublishEvent(ctx, events.EventTypePayloadsReady, map[string]interface{}{
		"spec_id":     specID,
		"workflow_id": workflowID,
		"payloads":    payloads,
		"output_path": outputPath,
		"spec_path":   originalSpecPath,
		"base_url":    baseURL,
	})
}

// generateEndpointPayloads generates payloads for a specific endpoint
func (pa *PayloadAgent) generateEndpointPayloads(ctx context.Context, endpoint string, spec map[string]interface{}) (map[string]interface{}, error) {
	// Extract schema for this endpoint from spec
	// Parse endpoint to get method and path (e.g., "POST /users" -> method: POST, path: /users)
	parts := strings.Fields(endpoint)
	method := "GET"
	path := endpoint
	if len(parts) >= 2 {
		method = parts[0]
		path = parts[1]
	}

	// Extract actual schema from OpenAPI spec
	schema := pa.extractSchemaFromSpec(spec, path, strings.ToLower(method))
	schemaJSON := ""
	if schema != nil {
		schemaBytes, _ := json.MarshalIndent(schema, "  ", "  ")
		schemaJSON = string(schemaBytes)
	}

	// Extract parameters (path, query, headers) from OpenAPI spec
	parameters := pa.extractParametersFromSpec(spec, path, strings.ToLower(method))
	parametersJSON := ""
	if parameters != nil {
		paramBytes, _ := json.MarshalIndent(parameters, "  ", "  ")
		parametersJSON = string(paramBytes)
	}

	// Build prompt with actual schema and parameters
	schemaSection := ""
	if schemaJSON != "" {
		schemaSection = fmt.Sprintf(`

**OpenAPI Request Body Schema:**
%s

IMPORTANT: Generate payloads that EXACTLY match this schema. Use the correct field names, types, and constraints.`, schemaJSON)
	}

	parametersSection := ""
	if parametersJSON != "" {
		parametersSection = fmt.Sprintf(`

**OpenAPI Parameters (Path, Query, Headers):**
%s

IMPORTANT: Generate realistic values for ALL required parameters. Include them in the test case structure.`, parametersJSON)
	}

	prompt := fmt.Sprintf("You are a test data generation expert. Generate realistic test payloads for the following API endpoint:\n\n"+
		"Endpoint: %s\n"+
		"Method: %s\n"+
		"Path: %s%s%s\n\n"+
		"Generate 3 types of test cases:\n"+
		"1. **Positive Cases**: Valid data that should succeed (3 examples)\n"+
		"2. **Negative Cases**: Invalid data that should fail (3 examples)\n"+
		"3. **Boundary Cases**: Edge cases and limits (2 examples)\n\n"+
		"For each test case, provide:\n"+
		"- name: descriptive name\n"+
		"- method: HTTP method (e.g., \"POST\", \"GET\")\n"+
		"- path: endpoint path (e.g., \"/api/resources/{id}\")\n"+
		"- path_params: object with path parameter values - REQUIRED if path has {params}\n"+
		"- query_params: object with query parameter values (e.g., {\"status\": \"active\", \"limit\": \"10\"}) - optional\n"+
		"- headers: object with custom headers (e.g., {\"X-Custom-Header\": \"value\"}) - optional, auth headers handled separately\n"+
		"- payload: the actual JSON payload (null for GET requests, must match schema if provided)\n"+
		"- expected_status: expected HTTP status code\n"+
		"- description: what this test validates\n"+
		"- context_capture: (for POST/PUT that create resources) specify which fields to capture for later use\n"+
		"- context_required: (for GET/DELETE/PUT that use resources) specify which context is needed\n\n"+
		"**CONTEXT MARKER RULES (SIMPLIFIED):**\n"+
		"Use simple context markers for data dependencies:\n\n"+
		"**For endpoints that CREATE resources (POST/PUT):**\n"+
		"- Use realistic hardcoded values in payload\n"+
		"- Add context_capture to specify what to save:\n"+
		"  \"context_capture\": {\"enabled\": true, \"fields\": [\"id\", \"name\"], \"store_as\": \"resource\"}\n\n"+
		"**For endpoints that USE resources (GET/DELETE/PUT with IDs):**\n"+
		"- Use context markers: {{CONTEXT:type.field}}\n"+
		"- Examples:\n"+
		"  * {{CONTEXT:resource.id}} - ID from resource context\n"+
		"  * {{CONTEXT:user.email}} - Email from user context\n"+
		"  * {{CONTEXT:order.orderId}} - Order ID from order context\n"+
		"- Add context_required to specify what's needed:\n"+
		"  \"context_required\": {\"type\": \"resource\", \"fields\": [\"id\"]}\n\n"+
		"**Context markers work in:**\n"+
		"- path_params: {\"id\": \"{{CONTEXT:resource.id}}\"}\n"+
		"- query_params: {\"userId\": \"{{CONTEXT:user.id}}\"}\n"+
		"- headers: {\"X-Resource-ID\": \"{{CONTEXT:resource.id}}\"}\n"+
		"- payload: {\"resourceId\": \"{{CONTEXT:resource.id}}\"}\n\n"+
		"CRITICAL FORMATTING RULES:\n"+
		"1. Your response MUST be ONLY valid JSON - no explanations, no markdown, no code blocks\n"+
		"2. Do NOT wrap the JSON in markdown code blocks (no backtick markers)\n"+
		"3. Do NOT include any text before or after the JSON\n"+
		"4. Start your response directly with the opening brace {\n"+
		"5. End your response with the closing brace }\n\n"+
		"Required JSON structure:\n"+
		"{\n"+
		"  \"positive\": [\n"+
		"    {\n"+
		"      \"name\": \"...\",\n"+
		"      \"method\": \"%s\",\n"+
		"      \"path\": \"%s\",\n"+
		"      \"path_params\": {\"id\": \"{{CONTEXT:resource.id}}\"},\n"+
		"      \"query_params\": {\"key\": \"value\"},\n"+
		"      \"headers\": {\"X-Header\": \"value\"},\n"+
		"      \"payload\": {...},\n"+
		"      \"expected_status\": 200,\n"+
		"      \"description\": \"...\",\n"+
		"      \"context_capture\": {\"enabled\": true, \"fields\": [\"id\"], \"store_as\": \"resource\"},\n"+
		"      \"context_required\": {\"type\": \"resource\", \"fields\": [\"id\"]}\n"+
		"    }\n"+
		"  ],\n"+
		"  \"negative\": [\n"+
		"    {\"name\": \"...\", \"method\": \"%s\", \"path\": \"%s\", \"path_params\": {\"id\": \"{{CONTEXT:resource.id}}\"}, \"payload\": {...}, \"expected_status\": 400, \"description\": \"...\", \"context_required\": {\"type\": \"resource\", \"fields\": [\"id\"]}}\n"+
		"  ],\n"+
		"  \"boundary\": [\n"+
		"    {\"name\": \"...\", \"method\": \"%s\", \"path\": \"%s\", \"path_params\": {\"id\": \"{{CONTEXT:resource.id}}\"}, \"payload\": {...}, \"expected_status\": 200, \"description\": \"...\", \"context_required\": {\"type\": \"resource\", \"fields\": [\"id\"]}}\n"+
		"  ]\n"+
		"}\n\n"+
		"EXAMPLES:\n\n"+
		"**POST /api/resources (Creates context):**\n"+
		"{\n"+
		"  \"name\": \"Create new resource\",\n"+
		"  \"method\": \"POST\",\n"+
		"  \"path\": \"/api/resources\",\n"+
		"  \"payload\": {\"name\": \"Test Resource\", \"status\": \"active\"},\n"+
		"  \"expected_status\": 201,\n"+
		"  \"description\": \"Create a new resource\",\n"+
		"  \"context_capture\": {\"enabled\": true, \"fields\": [\"id\", \"name\"], \"store_as\": \"resource\"}\n"+
		"}\n\n"+
		"**GET /api/resources/{id} (Uses context):**\n"+
		"{\n"+
		"  \"name\": \"Get resource by ID\",\n"+
		"  \"method\": \"GET\",\n"+
		"  \"path\": \"/api/resources/{id}\",\n"+
		"  \"path_params\": {\"id\": \"{{CONTEXT:resource.id}}\"},\n"+
		"  \"expected_status\": 200,\n"+
		"  \"description\": \"Retrieve existing resource\",\n"+
		"  \"context_required\": {\"type\": \"resource\", \"fields\": [\"id\"]}\n"+
		"}\n\n"+
		"**PUT /api/resources/{id} (Uses AND creates context):**\n"+
		"{\n"+
		"  \"name\": \"Update resource\",\n"+
		"  \"method\": \"PUT\",\n"+
		"  \"path\": \"/api/resources/{id}\",\n"+
		"  \"path_params\": {\"id\": \"{{CONTEXT:resource.id}}\"},\n"+
		"  \"payload\": {\"name\": \"Updated Resource\", \"status\": \"inactive\"},\n"+
		"  \"expected_status\": 200,\n"+
		"  \"description\": \"Update existing resource\",\n"+
		"  \"context_required\": {\"type\": \"resource\", \"fields\": [\"id\"]},\n"+
		"  \"context_capture\": {\"enabled\": true, \"fields\": [\"id\", \"name\"], \"store_as\": \"resource\"}\n"+
		"}", endpoint, method, path, schemaSection, parametersSection, method, path, method, path, method, path)

	response, err := pa.LLMClient.GenerateCompletion(ctx, prompt, llm.CompletionOptions{
		Temperature: 0.6,
		MaxTokens:   8192, // Groq supports up to 8192 tokens for most models
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate payloads with LLM: %w", err)
	}

	// Log response length for debugging
	log.Printf("📊 Payload generation response length: %d characters for endpoint %s %s", len(response), method, path)

	// Parse LLM response - strip markdown code blocks if present
	cleanedResponse := stripMarkdownCodeBlocks(response)

	var payloadData map[string]interface{}
	if err := json.Unmarshal([]byte(cleanedResponse), &payloadData); err != nil {
		log.Printf("Warning: failed to parse LLM response as JSON for %s: %v", endpoint, err)
		log.Printf("Response was: %s", response)
		return map[string]interface{}{
			"raw_response": response,
		}, nil
	}

	return payloadData, nil
}

// extractSchemaFromSpec extracts the request body schema for a specific endpoint
func (pa *PayloadAgent) extractSchemaFromSpec(spec map[string]interface{}, path, method string) map[string]interface{} {
	// Navigate: spec.paths[path][method].requestBody.content["application/json"].schema

	paths, ok := spec["paths"].(map[string]interface{})
	if !ok {
		log.Printf("Warning: no paths found in spec")
		return nil
	}

	pathItem, ok := paths[path].(map[string]interface{})
	if !ok {
		log.Printf("Warning: path %s not found in spec", path)
		return nil
	}

	operation, ok := pathItem[method].(map[string]interface{})
	if !ok {
		log.Printf("Warning: method %s not found for path %s", method, path)
		return nil
	}

	// Try to get request body schema
	requestBody, ok := operation["requestBody"].(map[string]interface{})
	if !ok {
		// No request body (e.g., GET requests)
		log.Printf("Info: no requestBody for %s %s", method, path)

		// Try to get response schema instead for reference
		responses, ok := operation["responses"].(map[string]interface{})
		if !ok {
			return nil
		}

		// Get 200 response schema
		response200, ok := responses["200"].(map[string]interface{})
		if !ok {
			return nil
		}

		content, ok := response200["content"].(map[string]interface{})
		if !ok {
			return nil
		}

		appJSON, ok := content["application/json"].(map[string]interface{})
		if !ok {
			return nil
		}

		schema, ok := appJSON["schema"].(map[string]interface{})
		if !ok {
			return nil
		}

		return schema
	}

	// Extract request body schema
	content, ok := requestBody["content"].(map[string]interface{})
	if !ok {
		log.Printf("Warning: no content in requestBody for %s %s", method, path)
		return nil
	}

	appJSON, ok := content["application/json"].(map[string]interface{})
	if !ok {
		// Try other content types
		for contentType, contentData := range content {
			if contentMap, ok := contentData.(map[string]interface{}); ok {
				if schema, ok := contentMap["schema"].(map[string]interface{}); ok {
					log.Printf("Info: using schema from content-type: %s", contentType)
					return schema
				}
			}
		}
		log.Printf("Warning: no application/json in content for %s %s", method, path)
		return nil
	}

	schema, ok := appJSON["schema"].(map[string]interface{})
	if !ok {
		log.Printf("Warning: no schema in application/json for %s %s", method, path)
		return nil
	}

	return schema
}

// extractParametersFromSpec extracts path, query, and header parameters from OpenAPI spec
func (pa *PayloadAgent) extractParametersFromSpec(spec map[string]interface{}, path, method string) map[string]interface{} {
	paths, ok := spec["paths"].(map[string]interface{})
	if !ok {
		return nil
	}

	pathItem, ok := paths[path].(map[string]interface{})
	if !ok {
		return nil
	}

	operation, ok := pathItem[method].(map[string]interface{})
	if !ok {
		return nil
	}

	// Extract parameters array
	parameters, ok := operation["parameters"].([]interface{})
	if !ok {
		// No parameters defined
		return map[string]interface{}{
			"path_params":   []interface{}{},
			"query_params":  []interface{}{},
			"header_params": []interface{}{},
		}
	}

	pathParams := []interface{}{}
	queryParams := []interface{}{}
	headerParams := []interface{}{}

	for _, param := range parameters {
		paramMap, ok := param.(map[string]interface{})
		if !ok {
			continue
		}

		paramIn, _ := paramMap["in"].(string)
		paramName, _ := paramMap["name"].(string)
		paramRequired, _ := paramMap["required"].(bool)
		paramSchema := paramMap["schema"]
		paramDescription, _ := paramMap["description"].(string)

		paramInfo := map[string]interface{}{
			"name":        paramName,
			"required":    paramRequired,
			"schema":      paramSchema,
			"description": paramDescription,
		}

		switch paramIn {
		case "path":
			pathParams = append(pathParams, paramInfo)
		case "query":
			queryParams = append(queryParams, paramInfo)
		case "header":
			headerParams = append(headerParams, paramInfo)
		}
	}

	return map[string]interface{}{
		"path_params":   pathParams,
		"query_params":  queryParams,
		"header_params": headerParams,
	}
}

// extractEndpointsFromPhases extracts endpoint list from phases in strategy
func (pa *PayloadAgent) extractEndpointsFromPhases(strategy map[string]interface{}) []interface{} {
	endpoints := []interface{}{}
	seen := make(map[string]bool)

	phases, ok := strategy["phases"].([]interface{})
	if !ok {
		return endpoints
	}

	for _, phase := range phases {
		phaseMap, ok := phase.(map[string]interface{})
		if !ok {
			continue
		}

		phaseEndpoints, ok := phaseMap["endpoints"].([]interface{})
		if !ok {
			continue
		}

		for _, ep := range phaseEndpoints {
			epStr, ok := ep.(string)
			if !ok {
				continue
			}
			// Skip flow patterns like "POST /pet -> POST /store/order"
			if strings.Contains(epStr, "->") {
				continue
			}
			if !seen[epStr] {
				seen[epStr] = true
				endpoints = append(endpoints, ep)
			}
		}
	}

	log.Printf("📋 Extracted %d endpoints from phases", len(endpoints))
	return endpoints
}

// extractEndpointsFromSpec extracts endpoints directly from OpenAPI spec paths
func (pa *PayloadAgent) extractEndpointsFromSpec(spec map[string]interface{}) []interface{} {
	endpoints := []interface{}{}

	paths, ok := spec["paths"].(map[string]interface{})
	if !ok {
		return endpoints
	}

	// Define method priority for ordering (creators first, deleters last)
	methodOrder := map[string]int{
		"post":   1,
		"put":    2,
		"patch":  3,
		"get":    4,
		"delete": 5,
	}

	type endpointInfo struct {
		endpoint string
		priority int
	}
	var orderedEndpoints []endpointInfo

	for path, pathItem := range paths {
		pathItemMap, ok := pathItem.(map[string]interface{})
		if !ok {
			continue
		}

		for method := range pathItemMap {
			methodLower := strings.ToLower(method)
			if priority, isMethod := methodOrder[methodLower]; isMethod {
				ep := strings.ToUpper(method) + " " + path
				orderedEndpoints = append(orderedEndpoints, endpointInfo{endpoint: ep, priority: priority})
			}
		}
	}

	// Sort by priority
	for i := 0; i < len(orderedEndpoints); i++ {
		for j := i + 1; j < len(orderedEndpoints); j++ {
			if orderedEndpoints[i].priority > orderedEndpoints[j].priority {
				orderedEndpoints[i], orderedEndpoints[j] = orderedEndpoints[j], orderedEndpoints[i]
			}
		}
	}

	for _, ep := range orderedEndpoints {
		endpoints = append(endpoints, ep.endpoint)
	}

	log.Printf("📋 Extracted %d endpoints from OpenAPI spec", len(endpoints))
	return endpoints
}

// generateFallbackPayloads generates basic test payloads without LLM when AI is unavailable
func (pa *PayloadAgent) generateFallbackPayloads(endpoint string, spec map[string]interface{}) map[string]interface{} {
	log.Printf("🔧 Generating fallback payloads for: %s", endpoint)

	parts := strings.SplitN(endpoint, " ", 2)
	if len(parts) != 2 {
		return map[string]interface{}{"error": "invalid endpoint format"}
	}
	method := parts[0]
	path := parts[1]

	// Get endpoint info from spec
	paths, _ := spec["paths"].(map[string]interface{})
	pathItem, _ := paths[path].(map[string]interface{})
	operation, _ := pathItem[strings.ToLower(method)].(map[string]interface{})

	// Generate basic test cases based on method type
	positive := make([]map[string]interface{}, 0)
	negative := make([]map[string]interface{}, 0)
	boundary := make([]map[string]interface{}, 0)

	// Extract path parameters
	pathParams := make(map[string]interface{})
	if strings.Contains(path, "{") {
		// Extract parameter names from path
		paramRegex := strings.NewReplacer("{", "", "}", "")
		for _, part := range strings.Split(path, "/") {
			if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
				paramName := paramRegex.Replace(part)
				// Use context marker for IDs, or generate sample value
				if strings.Contains(strings.ToLower(paramName), "id") {
					pathParams[paramName] = "{{CONTEXT:" + strings.TrimSuffix(paramName, "Id") + ".id}}"
				} else {
					pathParams[paramName] = "test_" + paramName
				}
			}
		}
	}

	// Generate positive test case
	positiveTest := map[string]interface{}{
		"name":            "Basic " + method + " request",
		"method":          method,
		"path":            path,
		"expected_status": pa.getExpectedStatus(method),
		"description":     "Fallback: Basic positive test for " + endpoint,
	}
	if len(pathParams) > 0 {
		positiveTest["path_params"] = pathParams
		positiveTest["context_required"] = map[string]interface{}{
			"type":   "resource",
			"fields": []string{"id"},
		}
	}

	// Add payload for POST/PUT/PATCH
	if method == "POST" || method == "PUT" || method == "PATCH" {
		positiveTest["payload"] = pa.generateSamplePayload(operation)
		if method == "POST" {
			positiveTest["context_capture"] = map[string]interface{}{
				"enabled":  true,
				"fields":   []string{"id"},
				"store_as": pa.extractResourceName(path),
			}
		}
	}
	positive = append(positive, positiveTest)

	// Generate negative test case (invalid input)
	negativeTest := map[string]interface{}{
		"name":            "Invalid " + method + " request",
		"method":          method,
		"path":            path,
		"expected_status": 400,
		"description":     "Fallback: Negative test with invalid data",
	}
	if len(pathParams) > 0 {
		invalidParams := make(map[string]interface{})
		for k := range pathParams {
			invalidParams[k] = "invalid_value_!@#"
		}
		negativeTest["path_params"] = invalidParams
	}
	if method == "POST" || method == "PUT" || method == "PATCH" {
		negativeTest["payload"] = map[string]interface{}{}
	}
	negative = append(negative, negativeTest)

	// Generate boundary test case
	boundaryTest := map[string]interface{}{
		"name":            "Boundary " + method + " request",
		"method":          method,
		"path":            path,
		"expected_status": pa.getExpectedStatus(method),
		"description":     "Fallback: Boundary test with edge values",
	}
	if len(pathParams) > 0 {
		boundaryParams := make(map[string]interface{})
		for k := range pathParams {
			if strings.Contains(strings.ToLower(k), "id") {
				boundaryParams[k] = 1 // Minimum valid ID
			} else {
				boundaryParams[k] = "a" // Single character
			}
		}
		boundaryTest["path_params"] = boundaryParams
	}
	boundary = append(boundary, boundaryTest)

	return map[string]interface{}{
		"positive": positive,
		"negative": negative,
		"boundary": boundary,
		"fallback": true,
	}
}

// getExpectedStatus returns the expected HTTP status for a method
func (pa *PayloadAgent) getExpectedStatus(method string) int {
	switch method {
	case "POST":
		return 201
	case "DELETE":
		return 200
	default:
		return 200
	}
}

// generateSamplePayload generates a basic sample payload from operation schema
func (pa *PayloadAgent) generateSamplePayload(operation map[string]interface{}) map[string]interface{} {
	payload := make(map[string]interface{})

	// Try to extract from requestBody (OpenAPI 3.x)
	if requestBody, ok := operation["requestBody"].(map[string]interface{}); ok {
		if content, ok := requestBody["content"].(map[string]interface{}); ok {
			if jsonContent, ok := content["application/json"].(map[string]interface{}); ok {
				if schema, ok := jsonContent["schema"].(map[string]interface{}); ok {
					return pa.generateFromSchema(schema)
				}
			}
		}
	}

	// Try to extract from parameters (OpenAPI 2.x / Swagger)
	if params, ok := operation["parameters"].([]interface{}); ok {
		for _, p := range params {
			param, ok := p.(map[string]interface{})
			if !ok {
				continue
			}
			if param["in"] == "body" {
				if schema, ok := param["schema"].(map[string]interface{}); ok {
					return pa.generateFromSchema(schema)
				}
			}
		}
	}

	// Default minimal payload
	payload["name"] = "test_item"
	payload["status"] = "active"
	return payload
}

// generateFromSchema generates sample data from a JSON schema
func (pa *PayloadAgent) generateFromSchema(schema map[string]interface{}) map[string]interface{} {
	payload := make(map[string]interface{})

	properties, ok := schema["properties"].(map[string]interface{})
	if !ok {
		return payload
	}

	for name, propDef := range properties {
		prop, ok := propDef.(map[string]interface{})
		if !ok {
			continue
		}

		propType, _ := prop["type"].(string)
		switch propType {
		case "string":
			if format, ok := prop["format"].(string); ok {
				switch format {
				case "email":
					payload[name] = "test@example.com"
				case "date":
					payload[name] = "2024-01-01"
				case "date-time":
					payload[name] = "2024-01-01T00:00:00Z"
				case "uri", "url":
					payload[name] = "https://example.com"
				default:
					payload[name] = "test_" + name
				}
			} else {
				payload[name] = "test_" + name
			}
		case "integer", "number":
			payload[name] = 1
		case "boolean":
			payload[name] = true
		case "array":
			payload[name] = []interface{}{}
		case "object":
			payload[name] = map[string]interface{}{}
		default:
			payload[name] = "test_value"
		}
	}

	return payload
}

// extractResourceName extracts the resource name from a path generically.
// Uses the shared InferResourceTypeFromPath function for consistency.
func (pa *PayloadAgent) extractResourceName(path string) string {
	resourceType := InferResourceTypeFromPath(path)
	if resourceType == "" {
		return "resource"
	}
	return resourceType
}
