package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"sonar-automation/models"
)

// SonarCloudService wraps the SonarCloud API client
type SonarCloudService struct {
	baseURL string
	token   string
	orgKey  string
	client  *http.Client
}

// NewSonarCloudService creates a new SonarCloud API service
func NewSonarCloudService(token, orgKey string) *SonarCloudService {
	return &SonarCloudService{
		baseURL: "https://sonarcloud.io/api",
		token:   token,
		orgKey:  orgKey,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

// ProjectExists checks if a project exists in SonarCloud
func (sc *SonarCloudService) ProjectExists(projectKey string) (bool, error) {
	endpoint := fmt.Sprintf("%s/projects/search?organization=%s&projects=%s",
		sc.baseURL,
		url.QueryEscape(sc.orgKey),
		url.QueryEscape(projectKey))

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return false, err
	}

	req.SetBasicAuth(sc.token, "")

	resp, err := sc.client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return false, nil
	}

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Components []struct {
			Key string `json:"key"`
		} `json:"components"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, err
	}

	return len(result.Components) > 0, nil
}

// CreateProject creates a new project in SonarCloud
func (sc *SonarCloudService) CreateProject(projectKey, projectName string) error {
	endpoint := fmt.Sprintf("%s/projects/create", sc.baseURL)

	data := url.Values{}
	data.Set("organization", sc.orgKey)
	data.Set("project", projectKey)
	data.Set("name", projectName)

	req, err := http.NewRequest("POST", endpoint, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(sc.token, "")

	resp, err := sc.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create project (HTTP %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// SetMainBranch sets the main branch for a project
func (sc *SonarCloudService) SetMainBranch(projectKey, branchName string) error {
	endpoint := fmt.Sprintf("%s/project_branches/rename", sc.baseURL)

	data := url.Values{}
	data.Set("project", projectKey)
	data.Set("name", branchName)

	req, err := http.NewRequest("POST", endpoint, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(sc.token, "")

	resp, err := sc.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 400 might mean branch already set, which is fine
	if resp.StatusCode != 200 && resp.StatusCode != 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to set main branch (HTTP %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetProjectAnalyses gets the analysis history for a project
func (sc *SonarCloudService) GetProjectAnalyses(projectKey string) ([]models.Analysis, error) {
	endpoint := fmt.Sprintf("%s/project_analyses/search?project=%s", sc.baseURL, url.QueryEscape(projectKey))

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(sc.token, "")

	resp, err := sc.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get analyses (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Analyses []models.Analysis `json:"analyses"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Analyses, nil
}

// GetQualityGateStatus gets the quality gate status for a project
func (sc *SonarCloudService) GetQualityGateStatus(projectKey string) (*models.QualityGateStatus, error) {
	endpoint := fmt.Sprintf("%s/qualitygates/project_status?projectKey=%s", sc.baseURL, url.QueryEscape(projectKey))

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(sc.token, "")

	resp, err := sc.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get quality gate status (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		ProjectStatus models.QualityGateStatus `json:"projectStatus"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result.ProjectStatus, nil
}

// GetProjectMeasures gets measures for a project
func (sc *SonarCloudService) GetProjectMeasures(projectKey string) ([]models.Measure, error) {
	metrics := []string{
		"bugs", "vulnerabilities", "code_smells",
		"coverage", "duplicated_lines_density",
		"ncloc", "sqale_rating", "reliability_rating", "security_rating",
	}

	endpoint := fmt.Sprintf("%s/measures/component?component=%s&metricKeys=%s",
		sc.baseURL,
		url.QueryEscape(projectKey),
		url.QueryEscape(joinStrings(metrics, ",")))

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(sc.token, "")

	resp, err := sc.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get measures (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Component struct {
			Measures []models.Measure `json:"measures"`
		} `json:"component"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Component.Measures, nil
}

// GetIssues gets issues for a project
func (sc *SonarCloudService) GetIssues(projectKey string, pageSize int) (*models.IssuesResponse, error) {
	endpoint := fmt.Sprintf("%s/issues/search?componentKeys=%s&ps=%d",
		sc.baseURL,
		url.QueryEscape(projectKey),
		pageSize)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(sc.token, "")

	resp, err := sc.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get issues (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var result models.IssuesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Helper function to join strings
func joinStrings(strs []string, sep string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}

