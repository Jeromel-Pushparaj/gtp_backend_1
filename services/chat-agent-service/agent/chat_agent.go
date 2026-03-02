package agent

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/jeromelp/gtp_backend_1/services/chat-agent-service/client"
)

type ChatAgent struct {
	groqClient   *client.GroqClient
	toolExecutor *client.ToolExecutor
	tools        []client.Tool
}

func NewChatAgent(groqAPIKey, backendURL, backendAPIKey string) *ChatAgent {
	return &ChatAgent{
		groqClient:   client.NewGroqClient(groqAPIKey),
		toolExecutor: client.NewToolExecutor(backendURL, backendAPIKey),
		tools:        client.GetAvailableTools(),
	}
}

type ChatRequest struct {
	Message string `json:"message"`
}

type ChatResponse struct {
	Response string `json:"response"`
	Error    string `json:"error,omitempty"`
}

func (a *ChatAgent) ProcessMessage(userMessage string) (*ChatResponse, error) {
	messages := []client.ChatMessage{
		{
			Role:    "system",
			Content: "You are a helpful assistant that can access GitHub, Jira, and SonarCloud metrics through various tools. When users ask about repositories, pull requests, issues, or metrics, use the available tools to fetch the information.",
		},
		{
			Role:    "user",
			Content: userMessage,
		},
	}

	// First call to Groq with tools
	resp, err := a.groqClient.Chat(messages, a.tools)
	if err != nil {
		return nil, fmt.Errorf("groq API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from Groq")
	}

	choice := resp.Choices[0]

	log.Printf("Groq response - Finish reason: %s", choice.FinishReason)
	log.Printf("Tool calls in message: %d", len(choice.Message.ToolCalls))

	// Check if the model wants to call tools
	if choice.FinishReason == "tool_calls" && len(choice.Message.ToolCalls) > 0 {
		log.Printf("Processing %d tool calls", len(choice.Message.ToolCalls))

		// Add the assistant's message with tool calls to the conversation
		messages = append(messages, choice.Message)

		// Execute each tool call
		for _, toolCall := range choice.Message.ToolCalls {
			log.Printf("Executing tool: %s with args: %s", toolCall.Function.Name, toolCall.Function.Arguments)

			var args map[string]interface{}
			if toolCall.Function.Arguments != "" {
				if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
					log.Printf("Error parsing tool arguments: %v", err)
					args = make(map[string]interface{})
				}
			} else {
				args = make(map[string]interface{})
			}

			// Execute the tool
			result, err := a.toolExecutor.ExecuteTool(toolCall.Function.Name, args)
			if err != nil {
				log.Printf("Error executing tool %s: %v", toolCall.Function.Name, err)
				result = fmt.Sprintf("Error: %v", err)
			} else {
				log.Printf("Tool %s executed successfully. Result length: %d", toolCall.Function.Name, len(result))
			}

			// Add tool result to messages in the correct format
			messages = append(messages, client.ChatMessage{
				Role:       "tool",
				Content:    result,
				ToolCallID: toolCall.ID,
			})
		}

		log.Printf("Making second call to Groq with tool results")

		// Make another call to Groq with the tool results
		resp, err = a.groqClient.Chat(messages, nil)
		if err != nil {
			return nil, fmt.Errorf("groq API error on second call: %w", err)
		}

		if len(resp.Choices) == 0 {
			return nil, fmt.Errorf("no response from Groq on second call")
		}

		choice = resp.Choices[0]
		log.Printf("Second response received: %s", choice.Message.Content[:min(100, len(choice.Message.Content))])
	} else {
		log.Printf("No tool calls needed, returning direct response")
	}

	return &ChatResponse{
		Response: choice.Message.Content,
	}, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

