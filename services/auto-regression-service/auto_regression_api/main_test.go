package main

import (
	"encoding/json"
	"os"
	"testing"
)

func TestParseExampleResponse(t *testing.T) {
	// Read the example test results file
	data, err := os.ReadFile("ca7576a9-28bf-45e5-bda1-d77e7d4b25ff-test-results.json")
	if err != nil {
		t.Fatalf("Error reading file: %v", err)
	}

	// Try to parse it
	var testResp TestResponse
	if err := json.Unmarshal(data, &testResp); err != nil {
		t.Fatalf("Error parsing JSON: %v", err)
	}

	t.Logf("✓ Successfully parsed test response")
	t.Logf("Total tests: %d", testResp.Summary.TotalTests)
	t.Logf("Passed: %d", testResp.Summary.Passed)
	t.Logf("Failed: %d", testResp.Summary.Failed)
	t.Logf("Skipped: %d", testResp.Summary.Skipped)
	t.Logf("Results count: %d", len(testResp.Results))

	// Verify counts match
	if testResp.Summary.TotalTests != 32 {
		t.Errorf("Expected 32 total tests, got %d", testResp.Summary.TotalTests)
	}

	if len(testResp.Results) != 32 {
		t.Errorf("Expected 32 results, got %d", len(testResp.Results))
	}
}

func TestAggregateResults(t *testing.T) {
	// Read the example test results file
	data, err := os.ReadFile("ca7576a9-28bf-45e5-bda1-d77e7d4b25ff-test-results.json")
	if err != nil {
		t.Fatalf("Error reading file: %v", err)
	}

	var testResp TestResponse
	if err := json.Unmarshal(data, &testResp); err != nil {
		t.Fatalf("Error parsing JSON: %v", err)
	}

	// Test aggregation
	aggregated := aggregateResults(testResp)

	t.Logf("✓ Successfully aggregated results")
	t.Logf("Unique test cases: %d", len(aggregated.UniqueTestCases))
	t.Logf("Tests passed: %d", aggregated.TestsPassed)
	t.Logf("Tests failed: %d", aggregated.TestsFailed)
	t.Logf("Tests skipped: %d", aggregated.TestsSkipped)

	// Verify aggregation
	if aggregated.TotalTests != 32 {
		t.Errorf("Expected 32 total tests, got %d", aggregated.TotalTests)
	}

	if aggregated.TestsPassed != 9 {
		t.Errorf("Expected 9 passed tests, got %d", aggregated.TestsPassed)
	}

	if aggregated.TestsFailed != 19 {
		t.Errorf("Expected 19 failed tests, got %d", aggregated.TestsFailed)
	}

	if aggregated.TestsSkipped != 4 {
		t.Errorf("Expected 4 skipped tests, got %d", aggregated.TestsSkipped)
	}

	if len(aggregated.UniqueTestCases) != 32 {
		t.Errorf("Expected 32 unique test cases, got %d", len(aggregated.UniqueTestCases))
	}

	// Print first few test cases
	t.Log("\nFirst 3 test cases:")
	for i := 0; i < 3 && i < len(aggregated.UniqueTestCases); i++ {
		tc := aggregated.UniqueTestCases[i]
		t.Logf("  %d. %s - Status: %d, Passed: %v, Skipped: %v",
			i+1, tc.Name, tc.StatusCode, tc.Passed, tc.Skipped)
	}

	// Convert to JSON to verify it's valid
	jsonData, err := json.MarshalIndent(aggregated, "", "  ")
	if err != nil {
		t.Fatalf("Error marshaling aggregated response: %v", err)
	}

	t.Logf("\n✓ Successfully converted to JSON (%d bytes)", len(jsonData))
}

