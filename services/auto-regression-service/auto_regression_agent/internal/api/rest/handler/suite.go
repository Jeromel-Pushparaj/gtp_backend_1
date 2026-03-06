package handler

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"gitlab.com/tekion/development/toc/poc/opentest/internal/agents/autonomous"
	"gitlab.com/tekion/development/toc/poc/opentest/pkg/llm"
)

// SuiteHandler handles test suite operations
type SuiteHandler struct {
	suitesDir string
	llmClient *llm.Client
}

// NewSuiteHandler creates a new suite handler
func NewSuiteHandler(llmClient *llm.Client) *SuiteHandler {
	return &SuiteHandler{
		suitesDir: "./test_suites", // Default directory for test suites
		llmClient: llmClient,
	}
}

// ListSuites lists all saved test suites
func (h *SuiteHandler) ListSuites(c *gin.Context) {
	suites := []map[string]interface{}{}

	// Check if suites directory exists
	if _, err := os.Stat(h.suitesDir); os.IsNotExist(err) {
		c.JSON(http.StatusOK, gin.H{
			"suites": suites,
			"count":  0,
		})
		return
	}

	// Read all suite files
	files, err := os.ReadDir(h.suitesDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to read suites directory: " + err.Error(),
		})
		return
	}

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		filePath := filepath.Join(h.suitesDir, file.Name())
		suite, err := autonomous.LoadTestSuite(filePath)
		if err != nil {
			continue // Skip invalid files
		}

		suites = append(suites, map[string]interface{}{
			"id":         suite.ID,
			"name":       suite.Name,
			"spec_name":  suite.SpecName,
			"test_count": len(suite.Tests),
			"created_at": suite.CreatedAt,
			"updated_at": suite.UpdatedAt,
			"file_path":  filePath,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"suites": suites,
		"count":  len(suites),
	})
}

// RunSuiteRequest is the request body for running a saved test suite
type RunSuiteRequest struct {
	SuiteID string `json:"suite_id" binding:"required"`
	BaseURL string `json:"base_url,omitempty"` // Optional override
}

// RunSuite runs a saved test suite
func (h *SuiteHandler) RunSuite(c *gin.Context) {
	var req RunSuiteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body: " + err.Error(),
		})
		return
	}

	// Find the suite file
	suitePath := filepath.Join(h.suitesDir, req.SuiteID+"-test-suite.json")
	if _, err := os.Stat(suitePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Test suite not found: " + req.SuiteID,
		})
		return
	}

	// Create a PlannedExecutor to run the suite
	baseURL := req.BaseURL
	if baseURL == "" {
		// Load suite to get default base URL
		suite, err := autonomous.LoadTestSuite(suitePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to load test suite: " + err.Error(),
			})
			return
		}
		baseURL = suite.BaseURL
	}

	executor := autonomous.NewPlannedExecutor(h.llmClient, baseURL)
	result, err := executor.RunSavedSuite(c.Request.Context(), suitePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to run test suite: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"suite_id":    req.SuiteID,
		"summary":     result.Summary,
		"executed_at": result.ExecutedAt,
		"duration":    result.Duration.String(),
		"results":     result.Results,
	})
}

// GetSuite gets details of a specific test suite
func (h *SuiteHandler) GetSuite(c *gin.Context) {
	suiteID := c.Param("suite_id")

	suitePath := filepath.Join(h.suitesDir, suiteID+"-test-suite.json")
	suite, err := autonomous.LoadTestSuite(suitePath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Test suite not found: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, suite)
}

// DeleteSuite deletes a saved test suite
func (h *SuiteHandler) DeleteSuite(c *gin.Context) {
	suiteID := c.Param("suite_id")

	suitePath := filepath.Join(h.suitesDir, suiteID+"-test-suite.json")
	if _, err := os.Stat(suitePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Test suite not found",
		})
		return
	}

	if err := os.Remove(suitePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete test suite: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Test suite deleted successfully",
	})
}
