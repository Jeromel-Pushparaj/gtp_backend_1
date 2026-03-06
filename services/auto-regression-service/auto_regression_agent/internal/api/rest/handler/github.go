package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gitlab.com/tekion/development/toc/poc/opentest/internal/orchestration"
	"gitlab.com/tekion/development/toc/poc/opentest/pkg/domain/workflow"
)

// GitHubHandler handles GitHub integration requests
type GitHubHandler struct {
	orchestrator *orchestration.Orchestrator
}

// NewGitHubHandler creates a new GitHub handler
func NewGitHubHandler(orch *orchestration.Orchestrator) *GitHubHandler {
	return &GitHubHandler{
		orchestrator: orch,
	}
}

// GitHubTestRequest represents the request to trigger tests from GitHub
type GitHubTestRequest struct {
	GitHubURL string `json:"github_url" binding:"required"` // e.g., "https://github.com/owner/repo"
	PATToken  string `json:"pat_token" binding:"required"`  // GitHub Personal Access Token
	Branch    string `json:"branch"`                        // Optional, defaults to "main"
}

// GitHubTestResponse represents the response after running tests
type GitHubTestResponse struct {
	Success    bool                   `json:"success"`
	Message    string                 `json:"message"`
	WorkflowID string                 `json:"workflow_id,omitempty"`
	SpecID     string                 `json:"spec_id,omitempty"`
	Results    map[string]interface{} `json:"results,omitempty"`
	Error      string                 `json:"error,omitempty"`
}

// TriggerTest handles POST /api/v1/github/test
func (h *GitHubHandler) TriggerTest(c *gin.Context) {
	var req GitHubTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, GitHubTestResponse{
			Success: false,
			Error:   fmt.Sprintf("Invalid request: %v", err),
		})
		return
	}

	// Default branch
	if req.Branch == "" {
		req.Branch = "main"
	}

	log.Printf("📥 GitHub test request: repo=%s, branch=%s", req.GitHubURL, req.Branch)

	// Step 1: Fetch OpenAPI spec from GitHub
	specContent, specFormat, err := h.fetchOpenAPISpecFromGitHub(req.GitHubURL, req.Branch, req.PATToken)
	if err != nil {
		c.JSON(http.StatusBadRequest, GitHubTestResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to fetch OpenAPI spec: %v", err),
			Message: "OpenAPI spec (openAPISpec.json or openAPISpec.yml) not found in repository root",
		})
		return
	}

	log.Printf("✅ Found OpenAPI spec: format=%s, size=%d bytes", specFormat, len(specContent))

	// Step 2: Generate spec ID
	specID := uuid.New().String()
	teamID := "github-integration"

	// Step 3: Process the spec and create workflow
	ctx := context.Background()
	result, err := h.orchestrator.ProcessSpec(
		ctx,
		specID,
		teamID,
		specContent,
		specFormat,
		workflow.RunModeFull, // Run mode
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, GitHubTestResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to process spec: %v", err),
		})
		return
	}

	log.Printf("✅ Workflow created: id=%s, spec_id=%s", result.WorkflowID, result.SpecID)

	// Step 4: Wait for tests to complete (with timeout)
	timeout := 5 * time.Minute
	testResults, err := h.waitForTestCompletion(ctx, result.SpecID, timeout)
	if err != nil {
		c.JSON(http.StatusRequestTimeout, GitHubTestResponse{
			Success:    false,
			WorkflowID: result.WorkflowID,
			SpecID:     result.SpecID,
			Error:      fmt.Sprintf("Test execution timeout or failed: %v", err),
		})
		return
	}

	log.Printf("✅ Tests completed: workflow=%s, results available", result.WorkflowID)

	// Step 5: Return results
	c.JSON(http.StatusOK, GitHubTestResponse{
		Success:    true,
		Message:    "Tests completed successfully",
		WorkflowID: result.WorkflowID,
		SpecID:     result.SpecID,
		Results:    testResults,
	})
}

// fetchOpenAPISpecFromGitHub fetches OpenAPI spec from GitHub repository
func (h *GitHubHandler) fetchOpenAPISpecFromGitHub(repoURL, branch, token string) ([]byte, string, error) {
	// Extract owner and repo from URL
	owner, repo, err := parseGitHubURL(repoURL)
	if err != nil {
		return nil, "", err
	}

	// Try both JSON and YAML formats with different naming conventions
	specFiles := []string{
		"openapi-spec.json",
		"openAPISpec.json",
		"openapi-spec.yml",
		"openAPISpec.yml",
		"openapi-spec.yaml",
		"openAPISpec.yaml",
		"openapi.json",
		"openapi.yml",
		"openapi.yaml",
	}

	for _, filename := range specFiles {
		// GitHub API URL for raw file content
		apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s?ref=%s",
			owner, repo, filename, branch)

		log.Printf("🔍 Checking for %s in %s/%s (branch: %s)", filename, owner, repo, branch)

		// Create request
		req, err := http.NewRequest("GET", apiURL, nil)
		if err != nil {
			continue
		}

		// Add authentication
		// Support both classic tokens (ghp_) and fine-grained tokens (github_pat_)
		authHeader := "token " + token
		if strings.HasPrefix(token, "github_pat_") {
			authHeader = "Bearer " + token
		}
		req.Header.Set("Authorization", authHeader)
		req.Header.Set("Accept", "application/vnd.github.v3.raw")

		// Execute request
		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("⚠️  Error fetching %s: %v", filename, err)
			continue
		}
		defer resp.Body.Close()

		// Check if file exists
		if resp.StatusCode == http.StatusOK {
			content, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, "", fmt.Errorf("failed to read file content: %w", err)
			}

			format := "json"
			if strings.HasSuffix(filename, ".yml") || strings.HasSuffix(filename, ".yaml") {
				format = "yaml"
			}

			log.Printf("✅ Found %s (%d bytes)", filename, len(content))
			return content, format, nil
		}

		if resp.StatusCode == http.StatusNotFound {
			log.Printf("❌ %s not found", filename)
			continue
		}

		// Other error
		log.Printf("⚠️  Unexpected status %d for %s", resp.StatusCode, filename)
	}

	return nil, "", fmt.Errorf("OpenAPI spec not found in repository root (tried: %v)", specFiles)
}

// parseGitHubURL extracts owner and repo from GitHub URL
func parseGitHubURL(url string) (owner, repo string, err error) {
	// Remove trailing slashes and .git
	url = strings.TrimSuffix(strings.TrimSpace(url), "/")
	url = strings.TrimSuffix(url, ".git")

	// Handle different URL formats:
	// - https://github.com/owner/repo
	// - git@github.com:owner/repo
	// - github.com/owner/repo

	if strings.Contains(url, "github.com/") {
		parts := strings.Split(url, "github.com/")
		if len(parts) != 2 {
			return "", "", fmt.Errorf("invalid GitHub URL format")
		}

		pathParts := strings.Split(parts[1], "/")
		if len(pathParts) < 2 {
			return "", "", fmt.Errorf("invalid GitHub URL: missing owner or repo")
		}

		return pathParts[0], pathParts[1], nil
	}

	if strings.Contains(url, "github.com:") {
		parts := strings.Split(url, "github.com:")
		if len(parts) != 2 {
			return "", "", fmt.Errorf("invalid GitHub URL format")
		}

		pathParts := strings.Split(parts[1], "/")
		if len(pathParts) < 2 {
			return "", "", fmt.Errorf("invalid GitHub URL: missing owner or repo")
		}

		return pathParts[0], pathParts[1], nil
	}

	return "", "", fmt.Errorf("unsupported GitHub URL format")
}

// extractRepoName extracts a simple repo name from GitHub URL
func extractRepoName(url string) string {
	owner, repo, err := parseGitHubURL(url)
	if err != nil {
		return "unknown"
	}
	return fmt.Sprintf("%s/%s", owner, repo)
}

// waitForTestCompletion waits for test execution to complete and returns results
func (h *GitHubHandler) waitForTestCompletion(ctx context.Context, specID string, timeout time.Duration) (map[string]interface{}, error) {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	log.Printf("⏳ Waiting for spec %s tests to complete (timeout: %v)", specID, timeout)

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			// Check if timeout exceeded
			if time.Now().After(deadline) {
				return nil, fmt.Errorf("timeout waiting for test completion")
			}

			// Try to read results file
			resultsPath := fmt.Sprintf("./output/results/%s-test-results.json", specID)
			content, err := readResultsFile(resultsPath)
			if err == nil {
				log.Printf("✅ Results file found: %s", resultsPath)
				return content, nil
			}

			// Continue waiting
			log.Printf("⏳ Still waiting for results... (checked %s)", resultsPath)
		}
	}
}

// readResultsFile reads and parses the test results JSON file
func readResultsFile(path string) (map[string]interface{}, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var results map[string]interface{}
	if err := json.Unmarshal(content, &results); err != nil {
		return nil, fmt.Errorf("failed to parse results: %w", err)
	}

	return results, nil
}

