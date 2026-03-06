package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	// Read the debug response file
	data, err := os.ReadFile("debug-response-20260303-145132.json")
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	// Parse as APIResponse
	var apiResp APIResponse
	if err := json.Unmarshal(data, &apiResp); err != nil {
		fmt.Printf("Error parsing as APIResponse: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Successfully parsed API response\n")
	fmt.Printf("Success: %v\n", apiResp.Success)
	fmt.Printf("Message: %s\n", apiResp.Message)
	fmt.Printf("Workflow ID: %s\n", apiResp.WorkflowID)
	fmt.Printf("Spec ID: %s\n", apiResp.SpecID)
	fmt.Printf("\n")
	fmt.Printf("Total tests: %d\n", apiResp.Results.Summary.TotalTests)
	fmt.Printf("Passed: %d\n", apiResp.Results.Summary.Passed)
	fmt.Printf("Failed: %d\n", apiResp.Results.Summary.Failed)
	fmt.Printf("Skipped: %d\n", apiResp.Results.Summary.Skipped)
	fmt.Printf("Results count: %d\n", len(apiResp.Results.Results))

	// Convert to TestResponse format
	testResp := TestResponse{
		Plan:        apiResp.Results.Plan,
		Results:     apiResp.Results.Results,
		Summary:     apiResp.Results.Summary,
		ExecutedAt:  apiResp.Results.ExecutedAt,
		Duration:    apiResp.Results.Duration,
		ContextData: apiResp.Results.ContextData,
	}

	// Test aggregation
	aggregated := aggregateResults(testResp)
	fmt.Printf("\n✓ Successfully aggregated results\n")
	fmt.Printf("Unique test cases: %d\n", len(aggregated.UniqueTestCases))
	fmt.Printf("Tests passed: %d\n", aggregated.TestsPassed)
	fmt.Printf("Tests failed: %d\n", aggregated.TestsFailed)
	fmt.Printf("Tests skipped: %d\n", aggregated.TestsSkipped)
	fmt.Printf("Pass rate: %.2f%%\n", aggregated.PassRate)

	// Print first few test cases
	fmt.Printf("\nFirst 5 test cases:\n")
	for i := 0; i < 5 && i < len(aggregated.UniqueTestCases); i++ {
		tc := aggregated.UniqueTestCases[i]
		status := "✓"
		if !tc.Passed {
			status = "✗"
		}
		if tc.Skipped {
			status = "⊘"
		}
		fmt.Printf("  %s %s - Status: %d (%s)\n",
			status, tc.Name, tc.StatusCode, tc.Category)
	}

	// Convert to JSON
	jsonData, err := json.MarshalIndent(aggregated, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n✓ Successfully converted to JSON (%d bytes)\n", len(jsonData))
}

