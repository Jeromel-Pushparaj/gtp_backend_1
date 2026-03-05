package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/jeromelp/gtp_backend_1/services/service-catalogue/models"
)

// SonarClient handles communication with sonar-shell-test API
type SonarClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewSonarClient creates a new sonar-shell-test API client
func NewSonarClient(baseURL, apiKey string) *SonarClient {
	return &SonarClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// APIResponse represents the standard API response from sonar-shell-test
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Error   string      `json:"error"`
}

// GetOrganizations fetches all organizations from sonar-shell-test
func (c *SonarClient) GetOrganizations() ([]models.Organization, error) {
	url := fmt.Sprintf("%s/api/v1/orgs", c.baseURL)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if c.apiKey != "" {
		req.Header.Set("X-API-Key", c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !apiResp.Success {
		return nil, fmt.Errorf("API returned error: %s", apiResp.Error)
	}

	// Convert data to organizations
	dataBytes, err := json.Marshal(apiResp.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	var orgs []models.Organization
	if err := json.Unmarshal(dataBytes, &orgs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal organizations: %w", err)
	}

	return orgs, nil
}

// FetchRepositoriesByOrg fetches repositories for an organization from sonar-shell-test
func (c *SonarClient) FetchRepositoriesByOrg(orgID int64) ([]models.Repository, error) {
	url := fmt.Sprintf("%s/api/v1/repos/fetch?org_id=%d", c.baseURL, orgID)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if c.apiKey != "" {
		req.Header.Set("X-API-Key", c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !apiResp.Success {
		return nil, fmt.Errorf("API returned error: %s", apiResp.Error)
	}

	// Convert data to repositories
	dataBytes, err := json.Marshal(apiResp.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	var repos []models.Repository
	if err := json.Unmarshal(dataBytes, &repos); err != nil {
		return nil, fmt.Errorf("failed to unmarshal repositories: %w", err)
	}

	return repos, nil
}

// GetGitHubMetrics fetches GitHub metrics for a repository from sonar-shell-test
func (c *SonarClient) GetGitHubMetrics(repoID int64) (*models.GitHubMetrics, error) {
	url := fmt.Sprintf("%s/api/v1/repos/metrics/github?repo_id=%d", c.baseURL, repoID)
	
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if c.apiKey != "" {
		req.Header.Set("X-API-Key", c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !apiResp.Success {
		return nil, fmt.Errorf("API returned error: %s", apiResp.Error)
	}

	// Convert data to metrics
	dataBytes, err := json.Marshal(apiResp.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	var metrics models.GitHubMetrics
	if err := json.Unmarshal(dataBytes, &metrics); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metrics: %w", err)
	}

	return &metrics, nil
}

