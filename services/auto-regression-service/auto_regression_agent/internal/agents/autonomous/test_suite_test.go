package autonomous

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSaveAndLoadTestSuite(t *testing.T) {
	// Create a temporary directory for test suites
	tmpDir := t.TempDir()

	// Create a test suite
	suite := &TestSuite{
		ID:          "test-suite-123",
		Name:        "Test API Suite",
		Description: "A test suite for testing",
		Version:     "1.0.0",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		SpecID:      "spec-123",
		SpecName:    "Test API",
		BaseURL:     "https://api.example.com",
		Tests: []PersistedTest{
			{
				ID:             "test-1",
				Name:           "Create User",
				Description:    "Create a new user",
				Category:       "positive",
				Method:         "POST",
				Path:           "/users",
				Payload:        map[string]interface{}{"name": "John", "email": "john@example.com"},
				ExpectedStatus: 201,
				ResponseValidation: &ResponseValidation{
					ValidateSchema:   true,
					ValidateRequired: true,
					ValidateTypes:    true,
					AllowExtraFields: true,
				},
				ContextCapture: &ContextCapture{
					Enabled: true,
					StoreAs: "user",
					Fields:  []string{"id"},
				},
			},
			{
				ID:             "test-2",
				Name:           "Get User",
				Description:    "Get user by ID",
				Category:       "positive",
				Method:         "GET",
				Path:           "/users/{{CONTEXT:user.id}}",
				ExpectedStatus: 200,
				ContextRequired: &ContextRequired{
					Type:   "user",
					Fields: []string{"id"},
				},
			},
		},
		Statistics: SuiteStatistics{
			TotalTests:    2,
			PositiveTests: 2,
		},
	}

	// Save the suite
	filePath, err := SaveTestSuite(suite, tmpDir)
	if err != nil {
		t.Fatalf("Failed to save test suite: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatalf("Test suite file was not created: %s", filePath)
	}

	// Load the suite
	loaded, err := LoadTestSuite(filePath)
	if err != nil {
		t.Fatalf("Failed to load test suite: %v", err)
	}

	// Verify loaded data
	if loaded.ID != suite.ID {
		t.Errorf("Expected ID %s, got %s", suite.ID, loaded.ID)
	}
	if loaded.Name != suite.Name {
		t.Errorf("Expected Name %s, got %s", suite.Name, loaded.Name)
	}
	if len(loaded.Tests) != len(suite.Tests) {
		t.Errorf("Expected %d tests, got %d", len(suite.Tests), len(loaded.Tests))
	}
	if loaded.Tests[0].Method != "POST" {
		t.Errorf("Expected method POST, got %s", loaded.Tests[0].Method)
	}
	if loaded.Tests[1].Path != "/users/{{CONTEXT:user.id}}" {
		t.Errorf("Expected path with context marker, got %s", loaded.Tests[1].Path)
	}

	t.Logf("✅ Test suite saved to: %s", filePath)
	t.Logf("✅ Test suite loaded successfully with %d tests", len(loaded.Tests))
}

func TestConvertFromExecutionPlan(t *testing.T) {
	// Create an execution plan
	plan := &ExecutionPlan{
		ID:          "plan-123",
		SpecID:      "spec-456",
		WorkflowID:  "workflow-789",
		CreatedAt:   time.Now(),
		AIReasoning: "Test reasoning",
		Tests: []PlannedTest{
			{
				Order:          1,
				Name:           "Create Pet",
				Description:    "Create a new pet",
				Category:       "positive",
				Method:         "POST",
				Path:           "/pet",
				ExpectedStatus: 200,
				ContextCapture: &ContextCapture{
					Enabled: true,
					StoreAs: "pet",
					Fields:  []string{"id"},
				},
			},
			{
				Order:          2,
				Name:           "Invalid Pet",
				Description:    "Create invalid pet",
				Category:       "negative",
				Method:         "POST",
				Path:           "/pet",
				ExpectedStatus: 400,
			},
		},
		Statistics: PlanStatistics{
			TotalTests:    2,
			PositiveTests: 1,
			NegativeTests: 1,
		},
	}

	// Convert to test suite
	suite := ConvertFromExecutionPlan(plan, "Petstore", "https://petstore.example.com")

	if suite.ID != plan.ID {
		t.Errorf("Expected ID %s, got %s", plan.ID, suite.ID)
	}
	if len(suite.Tests) != 2 {
		t.Errorf("Expected 2 tests, got %d", len(suite.Tests))
	}
	if suite.Statistics.PositiveTests != 1 {
		t.Errorf("Expected 1 positive test, got %d", suite.Statistics.PositiveTests)
	}
	if suite.Statistics.NegativeTests != 1 {
		t.Errorf("Expected 1 negative test, got %d", suite.Statistics.NegativeTests)
	}

	t.Logf("✅ Converted execution plan to test suite: %s", suite.Name)
}

func TestSaveTestSuiteCreatesDirectory(t *testing.T) {
	tmpDir := filepath.Join(t.TempDir(), "nested", "suites")

	suite := &TestSuite{
		ID:        "test-dir-creation",
		Name:      "Test",
		CreatedAt: time.Now(),
		Tests:     []PersistedTest{},
	}

	filePath, err := SaveTestSuite(suite, tmpDir)
	if err != nil {
		t.Fatalf("Failed to save test suite: %v", err)
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatalf("Test suite file was not created: %s", filePath)
	}

	t.Logf("✅ Directory created and suite saved: %s", filePath)
}

