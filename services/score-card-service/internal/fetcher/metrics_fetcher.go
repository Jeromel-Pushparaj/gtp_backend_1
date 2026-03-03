package fetcher

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/jeromelp/gtp_backen_1/service/score-card-service/internal/models"
)

// MetricsFetcher fetches metrics from external APIs
type MetricsFetcher struct {
	baseURL    string
	httpClient *http.Client
}

// NewMetricsFetcher creates a new metrics fetcher
func NewMetricsFetcher(baseURL string) *MetricsFetcher {
	return &MetricsFetcher{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// APIResponse represents the standard API response format
type APIResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
	Error   string          `json:"error"`
}

// FetchGitHubMetrics fetches GitHub metrics for a repository
func (mf *MetricsFetcher) FetchGitHubMetrics(repo string) (*models.GitHubMetrics, error) {
	url := fmt.Sprintf("%s/api/v1/github/metrics?repo=%s", mf.baseURL, repo)
	
	resp, err := mf.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch GitHub metrics: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !apiResp.Success {
		return nil, fmt.Errorf("API error: %s", apiResp.Error)
	}

	var metrics models.GitHubMetrics
	if err := json.Unmarshal(apiResp.Data, &metrics); err != nil {
		return nil, fmt.Errorf("failed to parse GitHub metrics: %w", err)
	}

	return &metrics, nil
}

// FetchSonarMetrics fetches SonarCloud metrics for a repository
func (mf *MetricsFetcher) FetchSonarMetrics(repo string) (*models.SonarMetrics, error) {
	url := fmt.Sprintf("%s/api/v1/sonar/metrics?repo=%s&include_issues=true", mf.baseURL, repo)
	
	resp, err := mf.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Sonar metrics: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !apiResp.Success {
		// SonarCloud might not be configured, return nil without error
		return nil, nil
	}

	var metrics models.SonarMetrics
	if err := json.Unmarshal(apiResp.Data, &metrics); err != nil {
		return nil, fmt.Errorf("failed to parse Sonar metrics: %w", err)
	}

	return &metrics, nil
}

// FetchJiraMetrics fetches Jira metrics for a project
func (mf *MetricsFetcher) FetchJiraMetrics(projectKey string) (*models.JiraMetrics, error) {
	url := fmt.Sprintf("%s/api/v1/jira/metrics?project=%s", mf.baseURL, projectKey)
	
	resp, err := mf.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Jira metrics: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !apiResp.Success {
		// Jira might not be configured, return nil without error
		return nil, nil
	}

	var metrics models.JiraMetrics
	if err := json.Unmarshal(apiResp.Data, &metrics); err != nil {
		return nil, fmt.Errorf("failed to parse Jira metrics: %w", err)
	}

	return &metrics, nil
}

// FetchAllMetrics fetches all metrics for a service
func (mf *MetricsFetcher) FetchAllMetrics(serviceName, jiraProjectKey string) (*models.CombinedMetrics, error) {
	combined := &models.CombinedMetrics{
		ServiceName: serviceName,
		CollectedAt: time.Now(),
	}

	// Fetch GitHub metrics
	github, err := mf.FetchGitHubMetrics(serviceName)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch GitHub metrics: %w", err)
	}
	combined.GitHub = github

	// Fetch SonarCloud metrics (optional)
	sonar, _ := mf.FetchSonarMetrics(serviceName)
	combined.Sonar = sonar

	// Fetch Jira metrics (optional)
	if jiraProjectKey != "" {
		jira, _ := mf.FetchJiraMetrics(jiraProjectKey)
		combined.Jira = jira
	}

	return combined, nil
}

