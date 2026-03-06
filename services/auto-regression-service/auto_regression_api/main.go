package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// Request structures
type TestRequest struct {
	GithubURL string `json:"github_url"`
	PATToken  string `json:"pat_token"`
	Branch    string `json:"branch"`
}

// Response structures from the test API
type TestResult struct {
	Test struct {
		Order          int    `json:"order"`
		Name           string `json:"name"`
		Description    string `json:"description"`
		Category       string `json:"category"`
		Method         string `json:"method"`
		Path           string `json:"path"`
		ExpectedStatus int    `json:"expected_status"`
	} `json:"test"`
	ResolvedPath string `json:"resolved_path"`
	Passed       bool   `json:"passed"`
	StatusCode   int    `json:"status_code"`
	ResponseTime int64  `json:"response_time"`
	Skipped      bool   `json:"skipped"`
	SkipReason   string `json:"skip_reason,omitempty"`
}

// API wrapper response
type APIResponse struct {
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	WorkflowID string `json:"workflow_id"`
	SpecID     string `json:"spec_id"`
	Results    struct {
		Plan struct {
			ID         string `json:"id"`
			SpecID     string `json:"spec_id"`
			WorkflowID string `json:"workflow_id"`
			CreatedAt  string `json:"created_at"`
		} `json:"plan"`
		Results    []TestResult `json:"results"`
		Summary    struct {
			TotalTests int     `json:"total_tests"`
			Passed     int     `json:"passed"`
			Failed     int     `json:"failed"`
			Skipped    int     `json:"skipped"`
			PassRate   float64 `json:"pass_rate"`
			TotalTime  int64   `json:"total_time"`
			LLMCalls   int     `json:"llm_calls"`
		} `json:"summary"`
		ExecutedAt  string                 `json:"executed_at"`
		Duration    int64                  `json:"duration"`
		ContextData map[string]interface{} `json:"context_data"`
	} `json:"results"`
}

// For backward compatibility with file-based responses
type TestResponse struct {
	Plan struct {
		ID         string `json:"id"`
		SpecID     string `json:"spec_id"`
		WorkflowID string `json:"workflow_id"`
		CreatedAt  string `json:"created_at"`
	} `json:"plan"`
	Results    []TestResult `json:"results"`
	Summary    struct {
		TotalTests int     `json:"total_tests"`
		Passed     int     `json:"passed"`
		Failed     int     `json:"failed"`
		Skipped    int     `json:"skipped"`
		PassRate   float64 `json:"pass_rate"`
		TotalTime  int64   `json:"total_time"`
		LLMCalls   int     `json:"llm_calls"`
	} `json:"summary"`
	ExecutedAt  string                 `json:"executed_at"`
	Duration    int64                  `json:"duration"`
	ContextData map[string]interface{} `json:"context_data"`
}

// Aggregated response structure
type TestCaseResult struct {
	Name        string `json:"name"`
	StatusCode  int    `json:"status_code"`
	Passed      bool   `json:"passed"`
	Skipped     bool   `json:"skipped"`
	Category    string `json:"category"`
	Method      string `json:"method"`
	Path        string `json:"path"`
}

type AggregatedResponse struct {
	UniqueTestCases []TestCaseResult `json:"unique_test_cases"`
	TotalTests      int              `json:"total_tests"`
	TestsPassed     int              `json:"tests_passed"`
	TestsFailed     int              `json:"tests_failed"`
	TestsSkipped    int              `json:"tests_skipped"`
	PassRate        float64          `json:"pass_rate"`
	ExecutedAt      string           `json:"executed_at"`
	Duration        int64            `json:"duration_ns"`
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func handleTestRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req TestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.GithubURL == "" || req.PATToken == "" {
		http.Error(w, "github_url and pat_token are required", http.StatusBadRequest)
		return
	}

	// Prepare request to the test API
	testAPIURL := "http://localhost:8080/api/v1/github/test"
	requestBody, err := json.Marshal(req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal request: %v", err), http.StatusInternalServerError)
		return
	}

	// Call the test API
	log.Printf("Calling test API with URL: %s, Branch: %s", req.GithubURL, req.Branch)
	resp, err := http.Post(testAPIURL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to call test API: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read response: %v", err), http.StatusInternalServerError)
		return
	}

	// Check if the API call was successful
	if resp.StatusCode != http.StatusOK {
		http.Error(w, fmt.Sprintf("Test API returned error: %s (status: %d)", string(body), resp.StatusCode), resp.StatusCode)
		return
	}

	// Save raw response to file for debugging
	timestamp := time.Now().Format("20060102-150405")
	debugFile := fmt.Sprintf("debug-response-%s.json", timestamp)
	if err := os.WriteFile(debugFile, body, 0644); err != nil {
		log.Printf("Warning: Failed to save debug file: %v", err)
	} else {
		log.Printf("Saved raw API response to: %s", debugFile)
	}

	// Log the raw response for debugging
	log.Printf("Raw API response (first 1000 chars): %s", string(body[:min(1000, len(body))]))

	// First, try to parse as a generic map to see the structure
	var rawResp map[string]interface{}
	if err := json.Unmarshal(body, &rawResp); err != nil {
		log.Printf("Failed to parse as generic JSON. Error: %v", err)
		http.Error(w, fmt.Sprintf("Invalid JSON response from test API: %v", err), http.StatusInternalServerError)
		return
	}

	// Check if there's an error in the response
	if errMsg, ok := rawResp["error"].(string); ok {
		log.Printf("Test API returned error: %s", errMsg)
		http.Error(w, fmt.Sprintf("Test API error: %s", errMsg), http.StatusBadRequest)
		return
	}

	// Check if this is the new API response format (with success field)
	var testResp TestResponse
	if success, ok := rawResp["success"].(bool); ok && success {
		// New API format: parse as APIResponse
		var apiResp APIResponse
		if err := json.Unmarshal(body, &apiResp); err != nil {
			log.Printf("Failed to parse as APIResponse. Error: %v", err)
			http.Error(w, fmt.Sprintf("Failed to parse API response: %v", err), http.StatusInternalServerError)
			return
		}

		// Extract the nested test response
		testResp.Plan = apiResp.Results.Plan
		testResp.Results = apiResp.Results.Results
		testResp.Summary = apiResp.Results.Summary
		testResp.ExecutedAt = apiResp.Results.ExecutedAt
		testResp.Duration = apiResp.Results.Duration
		testResp.ContextData = apiResp.Results.ContextData

		log.Printf("Parsed new API format. Success: %v, Message: %s", apiResp.Success, apiResp.Message)
	} else {
		// Old format: parse directly as TestResponse
		if err := json.Unmarshal(body, &testResp); err != nil {
			log.Printf("Failed to parse response into TestResponse struct. Error: %v", err)
			log.Printf("Response structure: %+v", rawResp)

			// Try to provide helpful error message
			if results, ok := rawResp["results"]; ok {
				log.Printf("Results field type: %T", results)
				log.Printf("Results value: %+v", results)
			}

			http.Error(w, fmt.Sprintf("Failed to parse test response: %v. Check logs for details.", err), http.StatusInternalServerError)
			return
		}
	}

	log.Printf("Successfully parsed response. Total tests: %d, Passed: %d, Failed: %d, Skipped: %d",
		testResp.Summary.TotalTests, testResp.Summary.Passed, testResp.Summary.Failed, testResp.Summary.Skipped)

	// Aggregate results
	aggregated := aggregateResults(testResp)

	// Send response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(aggregated); err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully sent aggregated response with %d test cases", len(aggregated.UniqueTestCases))
}

func aggregateResults(testResp TestResponse) AggregatedResponse {
	uniqueTestCases := make([]TestCaseResult, 0, len(testResp.Results))

	for _, result := range testResp.Results {
		testCase := TestCaseResult{
			Name:       result.Test.Name,
			StatusCode: result.StatusCode,
			Passed:     result.Passed,
			Skipped:    result.Skipped,
			Category:   result.Test.Category,
			Method:     result.Test.Method,
			Path:       result.ResolvedPath,
		}
		uniqueTestCases = append(uniqueTestCases, testCase)
	}

	return AggregatedResponse{
		UniqueTestCases: uniqueTestCases,
		TotalTests:      testResp.Summary.TotalTests,
		TestsPassed:     testResp.Summary.Passed,
		TestsFailed:     testResp.Summary.Failed,
		TestsSkipped:    testResp.Summary.Skipped,
		PassRate:        testResp.Summary.PassRate,
		ExecutedAt:      testResp.ExecutedAt,
		Duration:        testResp.Duration,
	}
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	})
}

func main() {
	http.HandleFunc("/api/v1/test/aggregate", handleTestRequest)
	http.HandleFunc("/health", healthCheck)

	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8092"
	}
	portWithColon := ":" + port

	log.Printf("Starting Auto-Regression API Service on port %s", port)
	log.Printf("Endpoint: POST http://localhost:%s/api/v1/test/aggregate", port)
	log.Printf("Health check: GET http://localhost:%s/health", port)

	if err := http.ListenAndServe(portWithColon, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

