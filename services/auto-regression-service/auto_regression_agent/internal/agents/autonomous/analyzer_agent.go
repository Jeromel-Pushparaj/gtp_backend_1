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

// AnalyzerAgent is an autonomous agent that analyzes test results
type AnalyzerAgent struct {
	*Agent
}

// NewAnalyzerAgent creates a new autonomous analyzer agent
func NewAnalyzerAgent(
	llmClient *llm.Client,
	eventBus *events.Bus,
	messageBus *events.MessageBus,
	consensus *events.ConsensusEngine,
) *AnalyzerAgent {
	baseAgent := NewAgent(
		"analyzer_agent",
		AgentTypeAnalyzer,
		[]string{"result_analysis", "pattern_detection", "feedback_generation"},
		llmClient,
		eventBus,
		messageBus,
		consensus,
	)

	return &AnalyzerAgent{
		Agent: baseAgent,
	}
}

// Start starts the analyzer agent
func (aa *AnalyzerAgent) Start(ctx context.Context) error {
	log.Printf("📊 Starting Analyzer Agent...")

	// Start base agent
	if err := aa.Agent.Start(ctx); err != nil {
		return err
	}

	// Subscribe to tests_complete events
	go aa.listenToTestsComplete(ctx)

	// Subscribe to consensus requests
	go aa.listenToConsensusRequests(ctx)

	log.Printf("✅ Analyzer Agent ready and listening")
	return nil
}

// listenToTestsComplete listens for tests complete events
func (aa *AnalyzerAgent) listenToTestsComplete(ctx context.Context) {
	err := aa.EventBus.Subscribe(ctx, events.EventTypeTestsComplete, func(event *events.Event) error {
		log.Printf("📊 Analyzer Agent received tests_complete event")

		aa.setState(AgentStateProcessing)
		defer aa.setState(AgentStateIdle)

		specID, ok := event.Payload["spec_id"].(string)
		if !ok {
			return fmt.Errorf("spec_id not found in event payload")
		}

		workflowID, ok := event.Payload["workflow_id"].(string)
		if !ok {
			return fmt.Errorf("workflow_id not found in event payload")
		}

		summary, ok := event.Payload["summary"].(map[string]interface{})
		if !ok {
			return fmt.Errorf("summary not found in event payload")
		}

		results, ok := event.Payload["results"].(map[string]interface{})
		if !ok {
			return fmt.Errorf("results not found in event payload")
		}

		// Analyze results
		return aa.analyzeResults(ctx, specID, workflowID, summary, results)
	})

	if err != nil && err != context.Canceled {
		log.Printf("Error subscribing to tests_complete: %v", err)
	}
}

// analyzeResults analyzes test results and provides feedback
func (aa *AnalyzerAgent) analyzeResults(ctx context.Context, specID, workflowID string, summary, results map[string]interface{}) error {
	log.Printf("📊 Analyzing test results for spec: %s", specID)

	// Convert results to JSON for LLM
	resultsJSON, _ := json.MarshalIndent(map[string]interface{}{
		"summary": summary,
		"results": results,
	}, "", "  ")

	// Use LLM to analyze results
	prompt := fmt.Sprintf(`You are a test analysis expert. Analyze these test execution results and provide insights:

Test Results:
%s

Provide analysis including:
1. **Overall Assessment**: How well did the tests perform?
2. **Failure Patterns**: Are there any patterns in the failures?
3. **Recommendations**: What should be improved?
4. **Feedback for Agents**:
   - Discovery Agent: Any issues with endpoint analysis?
   - Designer Agent: Was the test strategy effective?
   - Payload Agent: Were the test payloads appropriate?
   - Executor Agent: Any execution issues?

Respond in JSON format:
{
  "assessment": "...",
  "patterns": ["pattern1", "pattern2"],
  "recommendations": ["rec1", "rec2"],
  "feedback": {
    "discovery_agent": "...",
    "designer_agent": "...",
    "payload_agent": "...",
    "executor_agent": "..."
  },
  "confidence": 0.85
}`, string(resultsJSON))

	response, err := aa.LLMClient.GenerateCompletion(ctx, prompt, llm.CompletionOptions{
		Temperature: 0.3,
		MaxTokens:   8192, // Increased from 2000 to handle large result analyses
	})
	if err != nil {
		return fmt.Errorf("failed to analyze results with LLM: %w", err)
	}

	// Parse LLM response - strip markdown code blocks if present
	cleanedResponse := stripMarkdownCodeBlocks(response)

	var analysis map[string]interface{}
	if err := json.Unmarshal([]byte(cleanedResponse), &analysis); err != nil {
		log.Printf("Warning: failed to parse LLM response as JSON: %v", err)
		analysis = map[string]interface{}{
			"raw_analysis": response,
		}
	}

	// Store analysis in memory
	aa.Memory.Store(fmt.Sprintf("analysis:%s", specID), analysis)

	// Save analysis to file
	outputPath := fmt.Sprintf("./output/analysis/%s-result-analysis.json", specID)
	if err := os.MkdirAll("./output/analysis", 0755); err != nil {
		return fmt.Errorf("failed to create analysis directory: %w", err)
	}

	analysisData, err := json.MarshalIndent(analysis, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal analysis data: %w", err)
	}

	if err := os.WriteFile(outputPath, analysisData, 0644); err != nil {
		return fmt.Errorf("failed to save analysis file: %w", err)
	}

	log.Printf("✅ Result analysis complete: %s", outputPath)

	// Send feedback to other agents
	aa.sendFeedbackToAgents(ctx, analysis)

	// Publish analysis_complete event
	return aa.PublishEvent(ctx, events.EventTypeAnalysisComplete, map[string]interface{}{
		"spec_id":     specID,
		"workflow_id": workflowID,
		"analysis":    analysis,
		"output_path": outputPath,
	})
}

// sendFeedbackToAgents sends feedback to other agents
func (aa *AnalyzerAgent) sendFeedbackToAgents(ctx context.Context, analysis map[string]interface{}) {
	feedback, ok := analysis["feedback"].(map[string]interface{})
	if !ok {
		log.Printf("No feedback found in analysis")
		return
	}

	// Send feedback to each agent
	agents := []string{"discovery_agent", "designer_agent", "payload_agent", "executor_agent"}

	for _, agent := range agents {
		if feedbackText, ok := feedback[agent].(string); ok {
			err := aa.SendMessage(ctx, agent, "Test Result Feedback", feedbackText, map[string]interface{}{
				"type":     "feedback",
				"analysis": analysis,
			})

			if err != nil {
				log.Printf("Warning: failed to send feedback to %s: %v", agent, err)
			} else {
				log.Printf("📨 Sent feedback to %s", agent)
			}
		}
	}

	// Also broadcast general insights
	recommendations, ok := analysis["recommendations"].([]interface{})
	if ok && len(recommendations) > 0 {
		aa.BroadcastMessage(ctx, "Test Analysis Insights", "Analysis complete with recommendations", map[string]interface{}{
			"recommendations": recommendations,
			"analysis":        analysis,
		})
	}
}

// listenToConsensusRequests listens for consensus requests
func (aa *AnalyzerAgent) listenToConsensusRequests(ctx context.Context) {
	err := aa.EventBus.Subscribe(ctx, events.EventTypeConsensusRequest, func(event *events.Event) error {
		log.Printf("🗳️  Analyzer Agent received consensus request")

		requestID, ok := event.Payload["request_id"].(string)
		if !ok {
			return fmt.Errorf("request_id not found in event payload")
		}

		question, ok := event.Payload["question"].(string)
		if !ok {
			return fmt.Errorf("question not found in event payload")
		}

		// Use LLM to provide expert opinion
		return aa.provideConsensusVote(ctx, requestID, question, event.Payload)
	})

	if err != nil && err != context.Canceled {
		log.Printf("Error subscribing to consensus requests: %v", err)
	}
}

// provideConsensusVote provides a vote for a consensus request
func (aa *AnalyzerAgent) provideConsensusVote(ctx context.Context, requestID, question string, context map[string]interface{}) error {
	log.Printf("🗳️  Analyzer Agent voting on: %s", question)

	// Use LLM to analyze and vote
	prompt := fmt.Sprintf("You are a test analysis expert. You've been asked to vote on the following question:\n\n"+
		"Question: %s\n\n"+
		"Context: %v\n\n"+
		"Based on your expertise in test result analysis and quality assessment, provide your vote and reasoning.\n\n"+
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

	response, err := aa.LLMClient.GenerateCompletion(ctx, prompt, llm.CompletionOptions{
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
	return aa.SubmitVote(ctx, requestID, voteData.Vote, voteData.Confidence, voteData.Reasoning)
}
