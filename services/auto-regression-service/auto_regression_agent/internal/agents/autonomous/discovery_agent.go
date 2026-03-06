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

// DiscoveryAgent is an autonomous agent that analyzes OpenAPI specs
type DiscoveryAgent struct {
	*Agent
}

// NewDiscoveryAgent creates a new autonomous discovery agent
func NewDiscoveryAgent(
	llmClient *llm.Client,
	eventBus *events.Bus,
	messageBus *events.MessageBus,
	consensus *events.ConsensusEngine,
) *DiscoveryAgent {
	baseAgent := NewAgent(
		"discovery_agent",
		AgentTypeDiscovery,
		[]string{"spec_analysis", "endpoint_extraction", "schema_analysis"},
		llmClient,
		eventBus,
		messageBus,
		consensus,
	)

	return &DiscoveryAgent{
		Agent: baseAgent,
	}
}

// Start starts the discovery agent
func (da *DiscoveryAgent) Start(ctx context.Context) error {
	log.Printf("🔍 Starting Discovery Agent...")

	// Start base agent
	if err := da.Agent.Start(ctx); err != nil {
		return err
	}

	// Subscribe to spec_uploaded events
	go da.listenToSpecUploaded(ctx)

	// Subscribe to consensus requests
	go da.listenToConsensusRequests(ctx)

	log.Printf("✅ Discovery Agent ready and listening")
	return nil
}

// listenToSpecUploaded listens for spec uploaded events
func (da *DiscoveryAgent) listenToSpecUploaded(ctx context.Context) {
	err := da.EventBus.Subscribe(ctx, events.EventTypeSpecUploaded, func(event *events.Event) error {
		log.Printf("🔍 Discovery Agent received spec_uploaded event: workflow=%s", event.WorkflowID)

		da.setState(AgentStateProcessing)
		defer da.setState(AgentStateIdle)

		// Extract spec path from payload
		specPath, ok := event.Payload["spec_path"].(string)
		if !ok {
			return fmt.Errorf("spec_path not found in event payload")
		}

		specID, ok := event.Payload["spec_id"].(string)
		if !ok {
			return fmt.Errorf("spec_id not found in event payload")
		}

		// Analyze spec using AI
		return da.analyzeSpec(ctx, specID, specPath, event.WorkflowID)
	})

	if err != nil && err != context.Canceled {
		log.Printf("Error subscribing to spec_uploaded: %v", err)
	}
}

// analyzeSpec analyzes an OpenAPI spec using AI
func (da *DiscoveryAgent) analyzeSpec(ctx context.Context, specID, specPath, workflowID string) error {
	log.Printf("🔍 Analyzing spec: id=%s, path=%s", specID, specPath)

	// Load spec file
	specData, err := os.ReadFile(specPath)
	if err != nil {
		return fmt.Errorf("failed to read spec file: %w", err)
	}

	// Parse spec to extract base URL and servers
	var specJSON map[string]interface{}
	if err := json.Unmarshal(specData, &specJSON); err != nil {
		return fmt.Errorf("failed to parse spec JSON: %w", err)
	}

	// Extract base URL from spec
	baseURL := ""
	if servers, ok := specJSON["servers"].([]interface{}); ok && len(servers) > 0 {
		if server, ok := servers[0].(map[string]interface{}); ok {
			if url, ok := server["url"].(string); ok {
				baseURL = url
			}
		}
	}

	// Fallback: check for Swagger 2.0 format
	if baseURL == "" {
		if host, ok := specJSON["host"].(string); ok {
			scheme := "https"
			if schemes, ok := specJSON["schemes"].([]interface{}); ok && len(schemes) > 0 {
				if s, ok := schemes[0].(string); ok {
					scheme = s
				}
			}
			basePath := ""
			if bp, ok := specJSON["basePath"].(string); ok {
				basePath = bp
			}
			baseURL = fmt.Sprintf("%s://%s%s", scheme, host, basePath)
		}
	}

	// Store the initially extracted base URL (might be relative)
	initialBaseURL := baseURL
	needsBaseURLResolution := strings.HasPrefix(baseURL, "/")

	if needsBaseURLResolution {
		log.Printf("⚠️  Relative URL detected: %s - will ask AI to determine correct base URL", baseURL)
	} else {
		log.Printf("🔍 Extracted base URL: %s", baseURL)
	}

	// Use LLM to analyze the spec and determine base URL if needed
	baseURLInstruction := ""
	if needsBaseURLResolution {
		baseURLInstruction = fmt.Sprintf("\n6. **Base URL Detection**: The spec has a relative base path '%s'. "+
			"Analyze the spec's info section, servers, host, and any documentation to determine the correct base URL. "+
			"If you cannot determine it from the spec, suggest 'http://localhost:8080%s' as fallback.\n", initialBaseURL, initialBaseURL)
	}

	prompt := fmt.Sprintf("You are an API analysis expert. Analyze this OpenAPI specification and provide:\n\n"+
		"1. **Key Endpoints**: List the most important endpoints\n"+
		"2. **Data Dependencies**: Identify which endpoints depend on others (e.g., POST /users before GET /users/{id})\n"+
		"3. **Security Requirements**: Identify authentication and authorization requirements\n"+
		"4. **Test Priorities**: Recommend which endpoints should be tested first and why\n"+
		"5. **Complexity Assessment**: Rate the complexity of testing this API (1-10)\n"+
		"6. **Context Type Mappings**: Identify ALL main resource types from the API paths and create mappings.\n"+
		"   Extract resource names from paths (e.g., /customers -> customer, /api/v1/products -> product).\n"+
		"   Map generic aliases to specific resources: 'resource', 'item', 'entity', 'object' -> actual type.\n"+
		"   Include domain-specific aliases (e.g., for e-commerce: 'purchase' -> 'order', 'buyer' -> 'customer').%s\n\n"+
		"OpenAPI Spec:\n%s\n\n"+
		"CRITICAL FORMATTING RULES:\n"+
		"1. Your response MUST be ONLY valid JSON - no explanations, no markdown, no code blocks\n"+
		"2. Do NOT wrap the JSON in markdown code blocks (no backtick markers)\n"+
		"3. Do NOT include any text before or after the JSON\n"+
		"4. Start your response directly with the opening brace {\n"+
		"5. End your response with the closing brace }\n\n"+
		"Required JSON structure:\n"+
		"{\n"+
		"  \"base_url\": \"https://api.example.com/v1\",\n"+
		"  \"key_endpoints\": [\"endpoint1\", \"endpoint2\"],\n"+
		"  \"dependencies\": {\"endpoint\": [\"depends_on1\", \"depends_on2\"]},\n"+
		"  \"security\": {\"type\": \"bearer\", \"scopes\": []},\n"+
		"  \"test_priorities\": [{\"endpoint\": \"...\", \"priority\": 1, \"reason\": \"...\"}],\n"+
		"  \"complexity\": 7,\n"+
		"  \"recommendations\": [\"recommendation1\", \"recommendation2\"],\n"+
		"  \"context_type_mappings\": {\n"+
		"    \"resource\": \"<main_resource_from_paths>\",\n"+
		"    \"item\": \"<main_resource_from_paths>\",\n"+
		"    \"entity\": \"<main_resource_from_paths>\",\n"+
		"    \"<domain_alias>\": \"<actual_resource>\"\n"+
		"  }\n"+
		"}", baseURLInstruction, string(specData))

	response, err := da.LLMClient.GenerateCompletion(ctx, prompt, llm.CompletionOptions{
		Temperature: 0.3,
		MaxTokens:   8192, // Increased from 2000 to handle large spec analyses
	})
	if err != nil {
		return fmt.Errorf("failed to analyze spec with LLM: %w", err)
	}

	// Parse LLM response - strip markdown code blocks if present
	cleanedResponse := stripMarkdownCodeBlocks(response)

	var analysis map[string]interface{}
	if err := json.Unmarshal([]byte(cleanedResponse), &analysis); err != nil {
		log.Printf("Warning: failed to parse LLM response as JSON, using raw response: %v", err)
		analysis = map[string]interface{}{
			"raw_analysis": response,
		}
	}

	// Use AI-detected base URL if available and we needed resolution
	if needsBaseURLResolution {
		if aiBaseURL, ok := analysis["base_url"].(string); ok && aiBaseURL != "" {
			baseURL = aiBaseURL
			log.Printf("🤖 AI detected base URL: %s", baseURL)
		} else {
			// Fallback to localhost if AI couldn't determine
			baseURL = "http://localhost:8080" + initialBaseURL
			log.Printf("⚠️  AI couldn't determine base URL, using localhost fallback: %s", baseURL)
		}
	}

	// Store the final base URL in analysis
	analysis["base_url"] = baseURL

	// Store analysis in memory
	da.Memory.Store(fmt.Sprintf("analysis:%s", specID), analysis)

	// Extract security schemes from spec
	var spec map[string]interface{}
	if err := json.Unmarshal(specData, &spec); err == nil {
		// Create auth manager to extract security info
		authMgr := NewAuthManager()
		if err := authMgr.LoadSecuritySchemes(spec); err != nil {
			log.Printf("Warning: failed to load security schemes: %v", err)
		}

		// Get security schemes as map for event payload
		securitySchemes := authMgr.GetSecuritySchemes()
		if len(securitySchemes) > 0 {
			log.Printf("🔐 Found %d security scheme(s) in spec", len(securitySchemes))
		}

		// Store security schemes in memory for later use
		da.Memory.Store(fmt.Sprintf("security_schemes:%s", specID), securitySchemes)
	}

	// Save analysis to file
	outputPath := fmt.Sprintf("./output/discovery/%s-ai-analysis.json", specID)
	if err := os.MkdirAll("./output/discovery", 0755); err != nil {
		return fmt.Errorf("failed to create discovery directory: %w", err)
	}

	analysisData, err := json.MarshalIndent(analysis, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal analysis data: %w", err)
	}

	if err := os.WriteFile(outputPath, analysisData, 0644); err != nil {
		return fmt.Errorf("failed to save analysis file: %w", err)
	}

	log.Printf("✅ Spec analysis complete: %s", outputPath)

	// Publish spec_analyzed event with base URL and spec path
	return da.PublishEvent(ctx, events.EventTypeSpecAnalyzed, map[string]interface{}{
		"spec_id":     specID,
		"workflow_id": workflowID,
		"analysis":    analysis,
		"output_path": outputPath,
		"spec_path":   specPath,
		"base_url":    baseURL,
	})
}

// listenToConsensusRequests listens for consensus requests
func (da *DiscoveryAgent) listenToConsensusRequests(ctx context.Context) {
	err := da.EventBus.Subscribe(ctx, events.EventTypeConsensusRequest, func(event *events.Event) error {
		log.Printf("🗳️  Discovery Agent received consensus request")

		requestID, ok := event.Payload["request_id"].(string)
		if !ok {
			return fmt.Errorf("request_id not found in event payload")
		}

		question, ok := event.Payload["question"].(string)
		if !ok {
			return fmt.Errorf("question not found in event payload")
		}

		// Use LLM to provide expert opinion
		return da.provideConsensusVote(ctx, requestID, question, event.Payload)
	})

	if err != nil && err != context.Canceled {
		log.Printf("Error subscribing to consensus requests: %v", err)
	}
}

// provideConsensusVote provides a vote for a consensus request
func (da *DiscoveryAgent) provideConsensusVote(ctx context.Context, requestID, question string, context map[string]interface{}) error {
	log.Printf("🗳️  Discovery Agent voting on: %s", question)

	// Use LLM to analyze and vote
	prompt := fmt.Sprintf("You are an API discovery expert. You've been asked to vote on the following question:\n\n"+
		"Question: %s\n\n"+
		"Context: %v\n\n"+
		"Based on your expertise in API analysis and testing, provide your vote and reasoning.\n\n"+
		"CRITICAL FORMATTING RULES:\n"+
		"1. Your response MUST be ONLY valid JSON - no explanations, no markdown, no code blocks\n"+
		"2. Do NOT wrap the JSON in markdown code blocks (no backtick markers)\n"+
		"3. Do NOT include any text before or after the JSON\n"+
		"4. Start your response directly with the opening brace {\n"+
		"5. End your response with the closing brace }\n\n"+
		"Required JSON structure:\n"+
		"{\n"+
		"  \"vote\": \"your_vote_here\",\n"+
		"  \"confidence\": 0.85,\n"+
		"  \"reasoning\": \"your detailed reasoning here\"\n"+
		"}", question, context)

	response, err := da.LLMClient.GenerateCompletion(ctx, prompt, llm.CompletionOptions{
		Temperature: 0.3,
		MaxTokens:   1000, // Increased from 500 to handle larger contexts
	})
	if err != nil {
		return fmt.Errorf("failed to generate vote: %w", err)
	}

	// Parse response - strip markdown code blocks if present
	cleanedResponse := stripMarkdownCodeBlocks(response)

	var voteData struct {
		Vote       string  `json:"vote"`
		Confidence float64 `json:"confidence"`
		Reasoning  string  `json:"reasoning"`
	}

	if err := json.Unmarshal([]byte(cleanedResponse), &voteData); err != nil {
		log.Printf("Warning: failed to parse vote response (response length: %d chars), using default", len(cleanedResponse))
		log.Printf("Response preview: %s...", cleanedResponse[:min(200, len(cleanedResponse))])
		voteData.Vote = "approve"
		voteData.Confidence = 0.5
		voteData.Reasoning = "Unable to parse LLM response"
	}

	// Submit vote
	return da.SubmitVote(ctx, requestID, voteData.Vote, voteData.Confidence, voteData.Reasoning)
}
