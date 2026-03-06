package client

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
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

// GetJiraMetrics fetches Jira metrics for a repository from sonar-shell-test
func (c *SonarClient) GetJiraMetrics(repoID int64) (*models.JiraMetrics, error) {
	url := fmt.Sprintf("%s/api/v1/repos/metrics/jira?repo_id=%d", c.baseURL, repoID)

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

	// Convert data to metrics
	dataBytes, err := json.Marshal(apiResp.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	var metrics models.JiraMetrics
	if err := json.Unmarshal(dataBytes, &metrics); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metrics: %w", err)
	}

	return &metrics, nil
}

// GetPullRequests fetches pull requests for a repository from sonar-shell-test
func (c *SonarClient) GetPullRequests(repoName, state string) ([]models.PullRequest, error) {
	url := fmt.Sprintf("%s/api/v1/github/pulls?repo=%s&state=%s", c.baseURL, repoName, state)

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

	// Convert data to pull requests
	dataBytes, err := json.Marshal(apiResp.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	var prs []models.PullRequest
	if err := json.Unmarshal(dataBytes, &prs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal pull requests: %w", err)
	}

	return prs, nil
}

// GetIssues fetches GitHub issues for a repository from sonar-shell-test
func (c *SonarClient) GetIssues(repoName, state string) ([]models.Issue, error) {
	url := fmt.Sprintf("%s/api/v1/github/issues?repo=%s&state=%s", c.baseURL, repoName, state)

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

	// Convert data to issues
	dataBytes, err := json.Marshal(apiResp.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	var issues []models.Issue
	if err := json.Unmarshal(dataBytes, &issues); err != nil {
		return nil, fmt.Errorf("failed to unmarshal issues: %w", err)
	}

	return issues, nil
}

// GetJiraIssues fetches Jira issues for a project from sonar-shell-test
func (c *SonarClient) GetJiraIssues(projectKey string, maxResults int) ([]models.JiraIssue, error) {
	jql := fmt.Sprintf("project=%s AND status!=Done", projectKey)
	encodedJQL := url.QueryEscape(jql)
	apiURL := fmt.Sprintf("%s/api/v1/jira/issues/search?jql=%s&max_results=%d", c.baseURL, encodedJQL, maxResults)

	req, err := http.NewRequest("GET", apiURL, nil)
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

	// Convert data to Jira issues
	dataBytes, err := json.Marshal(apiResp.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	var issues []models.JiraIssue
	if err := json.Unmarshal(dataBytes, &issues); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Jira issues: %w", err)
	}

	return issues, nil
}

// GetJiraOpenBugs fetches open bugs for a Jira project from sonar-shell-test
func (c *SonarClient) GetJiraOpenBugs(projectKey string) ([]models.JiraIssue, error) {
	apiURL := fmt.Sprintf("%s/api/v1/jira/bugs/open?project=%s", c.baseURL, projectKey)

	req, err := http.NewRequest("GET", apiURL, nil)
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

	// Convert data to Jira issues
	dataBytes, err := json.Marshal(apiResp.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	var bugs []models.JiraIssue
	if err := json.Unmarshal(dataBytes, &bugs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Jira bugs: %w", err)
	}

	return bugs, nil
}

// GetJiraOpenTasks fetches open tasks for a Jira project from sonar-shell-test
func (c *SonarClient) GetJiraOpenTasks(projectKey string) ([]models.JiraIssue, error) {
	apiURL := fmt.Sprintf("%s/api/v1/jira/tasks/open?project=%s", c.baseURL, projectKey)

	req, err := http.NewRequest("GET", apiURL, nil)
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

	// Convert data to Jira issues
	dataBytes, err := json.Marshal(apiResp.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	var tasks []models.JiraIssue
	if err := json.Unmarshal(dataBytes, &tasks); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Jira tasks: %w", err)
	}

	return tasks, nil
}

func (c *SonarClient) GetEvaluationMetrics(repoID int64, repoName string) (*models.EvaluationMetrics, error) {
	metrics := &models.EvaluationMetrics{
		ServiceName: repoName,
	}

	sonarMetrics, err := c.GetSonarMetrics(repoID)
	if err != nil {
		log.Printf("Warning: failed to fetch SonarCloud metrics for repo %d: %v", repoID, err)
	} else {
		metrics.Coverage = sonarMetrics.Coverage                             // data.coverage
		metrics.CodeSmells = int(sonarMetrics.CodeSmells)                    // data.code_smells
		metrics.Vulnerabilities = int(sonarMetrics.Vulnerabilities)          // data.vulnerabilities
		metrics.DuplicatedLinesDensity = sonarMetrics.DuplicatedLinesDensity // data.duplicated_lines_density
	}

	githubMetrics, err := c.GetGitHubMetrics(repoID)
	if err != nil {
		log.Printf("Warning: failed to fetch GitHub metrics for repo %d: %v", repoID, err)
	} else {
		if githubMetrics.HasReadme {
			metrics.HasReadme = 1 // data.has_readme (boolean -> int)
		} else {
			metrics.HasReadme = 0
		}
	}

	jiraMetrics, err := c.GetJiraMetrics(repoID)
	if err != nil {
		log.Printf("Warning: failed to fetch Jira metrics for repo %d: %v", repoID, err)
	} else {
		// Convert avg_time_to_resolve (in hours) to days and round
		metrics.MTTR = int(jiraMetrics.AvgTimeToResolve / 24) // data.avg_time_to_resolve
	}

	metrics.DeploymentFrequency = 0

	return metrics, nil
}

// GetSonarMetrics fetches SonarCloud metrics from database
func (c *SonarClient) GetSonarMetrics(repoID int64) (*models.SonarMetrics, error) {
	url := fmt.Sprintf("%s/api/v1/repos/metrics/sonar?repo_id=%d", c.baseURL, repoID)

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
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var apiResp struct {
		Success bool                `json:"success"`
		Data    models.SonarMetrics `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !apiResp.Success {
		return nil, fmt.Errorf("API returned success=false")
	}

	return &apiResp.Data, nil
}
