package autonomous

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"gitlab.com/tekion/development/toc/poc/opentest/internal/events"
	"gitlab.com/tekion/development/toc/poc/opentest/pkg/llm"
)

// DesignerAgent is an autonomous agent that designs test strategies
type DesignerAgent struct {
	*Agent
}

// NewDesignerAgent creates a new autonomous test designer agent
func NewDesignerAgent(
	llmClient *llm.Client,
	eventBus *events.Bus,
	messageBus *events.MessageBus,
	consensus *events.ConsensusEngine,
) *DesignerAgent {
	baseAgent := NewAgent(
		"designer_agent",
		AgentTypeDesigner,
		[]string{"test_strategy", "test_planning", "scenario_design"},
		llmClient,
		eventBus,
		messageBus,
		consensus,
	)

	return &DesignerAgent{
		Agent: baseAgent,
	}
}

// Start starts the designer agent
func (da *DesignerAgent) Start(ctx context.Context) error {
	log.Printf("🎨 Starting Test Designer Agent...")

	// Start base agent
	if err := da.Agent.Start(ctx); err != nil {
		return err
	}

	// Subscribe to spec_analyzed events
	go da.listenToSpecAnalyzed(ctx)

	// Subscribe to consensus requests
	go da.listenToConsensusRequests(ctx)

	log.Printf("✅ Test Designer Agent ready and listening")
	return nil
}

// listenToSpecAnalyzed listens for spec analyzed events
func (da *DesignerAgent) listenToSpecAnalyzed(ctx context.Context) {
	err := da.EventBus.Subscribe(ctx, events.EventTypeSpecAnalyzed, func(event *events.Event) error {
		log.Printf("🎨 Designer Agent received spec_analyzed event")

		da.setState(AgentStateProcessing)
		defer da.setState(AgentStateIdle)

		// Extract analysis from payload
		analysis, ok := event.Payload["analysis"].(map[string]interface{})
		if !ok {
			return fmt.Errorf("analysis not found in event payload")
		}

		specID, ok := event.Payload["spec_id"].(string)
		if !ok {
			return fmt.Errorf("spec_id not found in event payload")
		}

		workflowID, ok := event.Payload["workflow_id"].(string)
		if !ok {
			return fmt.Errorf("workflow_id not found in event payload")
		}

		// Extract spec_path and base_url
		specPath, _ := event.Payload["spec_path"].(string)
		baseURL, _ := event.Payload["base_url"].(string)

		// Store for later use
		da.Memory.Store(fmt.Sprintf("spec_path:%s", specID), specPath)
		da.Memory.Store(fmt.Sprintf("base_url:%s", specID), baseURL)

		// Design test strategy
		return da.designTestStrategy(ctx, specID, workflowID, analysis)
	})

	if err != nil && err != context.Canceled {
		log.Printf("Error subscribing to spec_analyzed: %v", err)
	}
}

// designTestStrategy designs a comprehensive test strategy
func (da *DesignerAgent) designTestStrategy(ctx context.Context, specID, workflowID string, analysis map[string]interface{}) error {
	log.Printf("🎨 Designing test strategy for spec: %s", specID)

	// Convert analysis to JSON for LLM
	analysisJSON, _ := json.MarshalIndent(analysis, "", "  ")

	// Use LLM to design test strategy with context requirement markers
	prompt := fmt.Sprintf("You are a test strategy expert. Based on the following API analysis, design a comprehensive test strategy.\n\n"+
		"API Analysis:\n%s\n\n"+
		"Design a test strategy that includes:\n"+
		"1. **Test Phases**: Break down testing into phases (e.g., smoke tests, functional tests, integration tests)\n"+
		"2. **Test Scenarios**: List specific test scenarios for each endpoint WITH context requirements\n"+
		"3. **Context Requirements**: For EACH endpoint, specify what data context it needs or creates\n"+
		"4. **Execution Order**: Optimal order to execute tests (creators before consumers)\n"+
		"5. **Success Criteria**: What defines a successful test run\n\n"+
		"**CRITICAL: Context Requirement Markers**\n"+
		"For EACH endpoint in scenarios, you MUST analyze and provide context_requirements:\n"+
		"- **needs_context**: boolean - does this endpoint need data from previous responses?\n"+
		"- **creates_context**: boolean - does this endpoint create data that others will use?\n"+
		"- **context_type**: string - type of resource (e.g., \"resource\", \"user\", \"order\")\n"+
		"- **depends_on**: string - which endpoint creates the data this needs (if needs_context=true)\n"+
		"- **context_fields_needed**: array - which fields from previous responses are needed\n"+
		"- **context_fields_created**: array - which fields this endpoint creates for others\n"+
		"- **context_mapping**: object - map of where to inject context (e.g., {\"path_params.id\": \"{{CONTEXT:resource.id}}\"})\n\n"+
		"**Examples:**\n"+
		"POST /api/resources - Creates context:\n"+
		"  {\"needs_context\": false, \"creates_context\": true, \"context_type\": \"resource\", \"context_fields_created\": [\"id\", \"name\"]}\n\n"+
		"GET /api/resources/{id} - Needs context:\n"+
		"  {\"needs_context\": true, \"creates_context\": false, \"context_type\": \"resource\", \"depends_on\": \"POST /api/resources\", \"context_fields_needed\": [\"id\"], \"context_mapping\": {\"path_params.id\": \"{{CONTEXT:resource.id}}\"}}\n\n"+
		"PUT /api/resources/{id} - Needs AND creates context:\n"+
		"  {\"needs_context\": true, \"creates_context\": true, \"context_type\": \"resource\", \"depends_on\": \"POST /api/resources\", \"context_fields_needed\": [\"id\"], \"context_fields_created\": [\"id\", \"name\"], \"context_mapping\": {\"path_params.id\": \"{{CONTEXT:resource.id}}\"}}\n\n"+
		"CRITICAL FORMATTING RULES:\n"+
		"1. Your response MUST be ONLY valid JSON - no explanations, no markdown, no code blocks\n"+
		"2. Do NOT wrap the JSON in markdown code blocks (no backtick markers)\n"+
		"3. Do NOT include any text before or after the JSON\n"+
		"4. Start your response directly with the opening brace {\n"+
		"5. End your response with the closing brace }\n\n"+
		"Required JSON structure:\n"+
		"{\n"+
		"  \"phases\": [\n"+
		"    {\"name\": \"smoke\", \"description\": \"...\", \"endpoints\": [\"...\"]}\n"+
		"  ],\n"+
		"  \"scenarios\": [\n"+
		"    {\n"+
		"      \"endpoint\": \"POST /api/resources\",\n"+
		"      \"scenarios\": [\"create valid resource\", \"create duplicate resource\"],\n"+
		"      \"context_requirements\": {\n"+
		"        \"needs_context\": false,\n"+
		"        \"creates_context\": true,\n"+
		"        \"context_type\": \"resource\",\n"+
		"        \"context_fields_created\": [\"id\", \"name\"]\n"+
		"      }\n"+
		"    },\n"+
		"    {\n"+
		"      \"endpoint\": \"GET /api/resources/{id}\",\n"+
		"      \"scenarios\": [\"get existing resource\", \"get non-existent resource\"],\n"+
		"      \"context_requirements\": {\n"+
		"        \"needs_context\": true,\n"+
		"        \"creates_context\": false,\n"+
		"        \"context_type\": \"resource\",\n"+
		"        \"depends_on\": \"POST /api/resources\",\n"+
		"        \"context_fields_needed\": [\"id\"],\n"+
		"        \"context_mapping\": {\"path_params.id\": \"{{CONTEXT:resource.id}}\"}\n"+
		"      }\n"+
		"    }\n"+
		"  ],\n"+
		"  \"execution_order\": [\"POST /api/resources\", \"GET /api/resources/{id}\"],\n"+
		"  \"success_criteria\": {\n"+
		"    \"min_coverage\": 80,\n"+
		"    \"max_failures\": 5\n"+
		"  }\n"+
		"}", string(analysisJSON))

	response, err := da.LLMClient.GenerateCompletion(ctx, prompt, llm.CompletionOptions{
		Temperature: 0.4,
		MaxTokens:   8192, // Increased from 3000 to handle large test strategies
	})
	if err != nil {
		return fmt.Errorf("failed to design strategy with LLM: %w", err)
	}

	// Parse LLM response - strip markdown code blocks if present
	cleanedResponse := stripMarkdownCodeBlocks(response)

	var strategy map[string]interface{}
	if err := json.Unmarshal([]byte(cleanedResponse), &strategy); err != nil {
		log.Printf("Warning: failed to parse LLM response as JSON: %v", err)
		strategy = map[string]interface{}{
			"raw_strategy": response,
		}
	}

	// Store strategy in memory
	da.Memory.Store(fmt.Sprintf("strategy:%s", specID), strategy)

	// Save strategy to file
	outputPath := fmt.Sprintf("./output/strategy/%s-test-strategy.json", specID)
	if err := os.MkdirAll("./output/strategy", 0755); err != nil {
		return fmt.Errorf("failed to create strategy directory: %w", err)
	}

	strategyData, err := json.MarshalIndent(strategy, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal strategy data: %w", err)
	}

	if err := os.WriteFile(outputPath, strategyData, 0644); err != nil {
		return fmt.Errorf("failed to save strategy file: %w", err)
	}

	log.Printf("✅ Test strategy designed: %s", outputPath)

	// Request consensus from other agents
	log.Printf("🗳️  Requesting consensus on test strategy...")

	decision, err := da.RequestConsensus(ctx,
		"Should we approve this test strategy?",
		[]string{"approve", "reject", "modify"},
		map[string]interface{}{
			"spec_id":  specID,
			"strategy": strategy,
		},
	)

	if err != nil {
		log.Printf("Warning: consensus failed, proceeding anyway: %v", err)
		decision = "approve" // Default to approve if consensus fails
	}

	log.Printf("✅ Consensus decision: %s", decision)

	// Retrieve spec_path and base_url from memory
	specPath, _ := da.Memory.Recall(fmt.Sprintf("spec_path:%s", specID))
	baseURL, _ := da.Memory.Recall(fmt.Sprintf("base_url:%s", specID))

	// Publish strategy approved event with spec_path and base_url
	return da.PublishEvent(ctx, events.EventTypeStrategyApproved, map[string]interface{}{
		"spec_id":          specID,
		"workflow_id":      workflowID,
		"strategy":         strategy,
		"output_path":      outputPath,
		"consensus_result": decision,
		"spec_path":        specPath,
		"base_url":         baseURL,
	})
}

// listenToConsensusRequests listens for consensus requests
func (da *DesignerAgent) listenToConsensusRequests(ctx context.Context) {
	err := da.EventBus.Subscribe(ctx, events.EventTypeConsensusRequest, func(event *events.Event) error {
		log.Printf("🗳️  Designer Agent received consensus request")

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
func (da *DesignerAgent) provideConsensusVote(ctx context.Context, requestID, question string, context map[string]interface{}) error {
	log.Printf("🗳️  Designer Agent voting on: %s", question)

	// Use LLM to analyze and vote
	prompt := fmt.Sprintf("You are a test design expert. You've been asked to vote on the following question:\n\n"+
		"Question: %s\n\n"+
		"Context: %v\n\n"+
		"Based on your expertise in test strategy and planning, provide your vote and reasoning.\n\n"+
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
