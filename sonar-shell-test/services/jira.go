package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	
	"strings"
	"time"

	"sonar-automation/models"
)

// JiraService wraps the Jira API client
type JiraService struct {
	baseURL string
	email   string
	token   string
	client  *http.Client
}

// NewJiraService creates a new Jira API service
func NewJiraService(domain, email, token string) *JiraService {
	return &JiraService{
		baseURL: fmt.Sprintf("https://%s/rest/api/3", domain),
		email:   email,
		token:   token,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

// makeRequest makes an authenticated request to Jira API
func (js *JiraService) makeRequest(method, endpoint string, body io.Reader) (*http.Response, error) {
	url := fmt.Sprintf("%s%s", js.baseURL, endpoint)

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(js.email, js.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	return js.client.Do(req)
}

// SearchIssues searches for issues using JQL
func (js *JiraService) SearchIssues(jql string, maxResults int) ([]models.JiraIssue, error) {
	// Use POST method with the new /search/jql endpoint as per Jira API v3
	endpoint := "/search/jql"

	// Prepare request body
	requestBody := map[string]interface{}{
		"jql":        jql,
		"maxResults": maxResults,
		"fields":     []string{"summary", "description", "issuetype", "status", "priority", "assignee", "reporter", "created", "updated", "resolutiondate", "customfield_10016", "labels", "components"},
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	resp, err := js.makeRequest("POST", endpoint, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return nil, fmt.Errorf("failed to search issues: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Jira API error (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Issues []struct {
			Key    string `json:"key"`
			ID     string `json:"id"`
			Fields struct {
				Summary     string      `json:"summary"`
				Description interface{} `json:"description"` // Can be string or object (ADF format)
				IssueType   struct {
					Name string `json:"name"`
				} `json:"issuetype"`
				Status struct {
					Name string `json:"name"`
				} `json:"status"`
				Priority *struct {
					Name string `json:"name"`
				} `json:"priority"`
				Assignee *struct {
					DisplayName string `json:"displayName"`
				} `json:"assignee"`
				Reporter struct {
					DisplayName string `json:"displayName"`
				} `json:"reporter"`
				Created        string   `json:"created"`
				Updated        string   `json:"updated"`
				ResolutionDate *string  `json:"resolutiondate"`
				StoryPoints    *float64 `json:"customfield_10016"` // Story points field
				Labels         []string `json:"labels"`
				Components     []struct {
					Name string `json:"name"`
				} `json:"components"`
			} `json:"fields"`
		} `json:"issues"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var issues []models.JiraIssue
	for _, item := range result.Issues {
		issue := models.JiraIssue{
			Key:       item.Key,
			ID:        item.ID,
			Summary:   item.Fields.Summary,
			IssueType: item.Fields.IssueType.Name,
			Status:    item.Fields.Status.Name,
			Reporter:  item.Fields.Reporter.DisplayName,
			Labels:    item.Fields.Labels,
		}

		// Handle description - can be string or object (ADF format)
		if item.Fields.Description != nil {
			switch desc := item.Fields.Description.(type) {
			case string:
				issue.Description = desc
			case map[string]interface{}:
				// It's an Atlassian Document Format object, just use empty string or summary
				issue.Description = "" // Could extract text from ADF if needed
			default:
				issue.Description = ""
			}
		}

		if item.Fields.Priority != nil {
			issue.Priority = item.Fields.Priority.Name
		}

		if item.Fields.Assignee != nil {
			issue.Assignee = item.Fields.Assignee.DisplayName
		}

		if item.Fields.StoryPoints != nil {
			issue.StoryPoints = *item.Fields.StoryPoints
		}

		// Parse dates
		if created, err := time.Parse(time.RFC3339, item.Fields.Created); err == nil {
			issue.Created = created
		}
		if updated, err := time.Parse(time.RFC3339, item.Fields.Updated); err == nil {
			issue.Updated = updated
		}
		if item.Fields.ResolutionDate != nil {
			if resolved, err := time.Parse(time.RFC3339, *item.Fields.ResolutionDate); err == nil {
				issue.Resolved = &resolved
			}
		}

		// Extract component names
		for _, comp := range item.Fields.Components {
			issue.Components = append(issue.Components, comp.Name)
		}

		issues = append(issues, issue)
	}

	return issues, nil
}

// GetIssuesByType gets issues filtered by type
func (js *JiraService) GetIssuesByType(projectKey, issueType string, maxResults int) ([]models.JiraIssue, error) {
	jql := fmt.Sprintf("project = %s AND issuetype = %s ORDER BY created DESC", projectKey, issueType)
	return js.SearchIssues(jql, maxResults)
}

// GetOpenBugs gets all open bugs for a project
func (js *JiraService) GetOpenBugs(projectKey string) ([]models.JiraIssue, error) {
	jql := fmt.Sprintf("project = %s AND issuetype = Bug AND status != Done AND status != Closed ORDER BY priority DESC, created DESC", projectKey)
	return js.SearchIssues(jql, 100)
}

// GetOpenTasks gets all open tasks for a project
func (js *JiraService) GetOpenTasks(projectKey string) ([]models.JiraIssue, error) {
	jql := fmt.Sprintf("project = %s AND issuetype = Task AND status != Done AND status != Closed ORDER BY created DESC", projectKey)
	return js.SearchIssues(jql, 100)
}

// GetIssuesByAssignee gets issues grouped by assignee
func (js *JiraService) GetIssuesByAssignee(projectKey string) (map[string][]models.JiraIssue, error) {
	jql := fmt.Sprintf("project = %s ORDER BY assignee, created DESC", projectKey)
	issues, err := js.SearchIssues(jql, 500)
	if err != nil {
		return nil, err
	}

	byAssignee := make(map[string][]models.JiraIssue)
	for _, issue := range issues {
		assignee := issue.Assignee
		if assignee == "" {
			assignee = "Unassigned"
		}
		byAssignee[assignee] = append(byAssignee[assignee], issue)
	}

	return byAssignee, nil
}

// GetIssueStats calculates issue statistics for a project
func (js *JiraService) GetIssueStats(projectKey string) (*models.JiraIssueStats, error) {
	jql := fmt.Sprintf("project = %s", projectKey)
	issues, err := js.SearchIssues(jql, 1000)
	if err != nil {
		return nil, err
	}

	stats := &models.JiraIssueStats{}
	stats.TotalIssues = len(issues)

	var totalResolveTime float64
	var resolvedCount int

	for _, issue := range issues {
		// Count by status
		status := strings.ToLower(issue.Status)
		if strings.Contains(status, "open") || strings.Contains(status, "to do") {
			stats.OpenIssues++
		} else if strings.Contains(status, "in progress") || strings.Contains(status, "progress") {
			stats.InProgressIssues++
		} else if strings.Contains(status, "done") || strings.Contains(status, "closed") || strings.Contains(status, "resolved") {
			stats.ClosedIssues++
		}

		// Count by type
		issueType := strings.ToLower(issue.IssueType)
		if strings.Contains(issueType, "bug") {
			stats.Bugs++
		} else if strings.Contains(issueType, "task") {
			stats.Tasks++
		} else if strings.Contains(issueType, "story") {
			stats.Stories++
		} else if strings.Contains(issueType, "epic") {
			stats.Epics++
		}

		// Calculate average time to resolve
		if issue.Resolved != nil {
			duration := issue.Resolved.Sub(issue.Created)
			totalResolveTime += duration.Hours()
			resolvedCount++
		}
	}

	if resolvedCount > 0 {
		stats.AvgTimeToResolve = totalResolveTime / float64(resolvedCount)
	}

	return stats, nil
}

// GetProjectBoards gets all boards for a project
func (js *JiraService) GetProjectBoards(projectKey string) ([]models.JiraBoard, error) {
	// Extract domain from baseURL (e.g., "https://domain.atlassian.net/rest/api/3" -> "domain.atlassian.net")
	domain := strings.TrimPrefix(js.baseURL, "https://")
	domain = strings.TrimSuffix(domain, "/rest/api/3")

	// Use agile API to get boards filtered by project
	url := fmt.Sprintf("https://%s/rest/agile/1.0/board?projectKeyOrId=%s", domain, projectKey)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(js.email, js.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := js.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get boards: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Jira API error (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Values []struct {
			ID       int64  `json:"id"`
			Name     string `json:"name"`
			Type     string `json:"type"`
			Location struct {
				ProjectKey string `json:"projectKey"`
			} `json:"location"`
		} `json:"values"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var boards []models.JiraBoard
	for _, item := range result.Values {
		board := models.JiraBoard{
			ID:   item.ID,
			Name: item.Name,
			Type: item.Type,
		}
		board.Location.ProjectKey = item.Location.ProjectKey
		boards = append(boards, board)
	}

	return boards, nil
}

// GetSprints gets sprints for a board
func (js *JiraService) GetSprints(boardID int) ([]models.JiraSprint, error) {
	// Extract domain from baseURL
	domain := strings.TrimPrefix(js.baseURL, "https://")
	domain = strings.TrimSuffix(domain, "/rest/api/3")

	// Use agile API for sprints
	url := fmt.Sprintf("https://%s/rest/agile/1.0/board/%d/sprint", domain, boardID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(js.email, js.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := js.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get sprints: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Jira API error (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Values []struct {
			ID            int64  `json:"id"`
			Name          string `json:"name"`
			State         string `json:"state"`
			StartDate     string `json:"startDate"`
			EndDate       string `json:"endDate"`
			CompleteDate  string `json:"completeDate"`
			Goal          string `json:"goal"`
		} `json:"values"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var sprints []models.JiraSprint
	for _, item := range result.Values {
		sprint := models.JiraSprint{
			ID:    item.ID,
			Name:  item.Name,
			State: item.State,
			Goal:  item.Goal,
		}

		if item.StartDate != "" {
			if startDate, err := time.Parse(time.RFC3339, item.StartDate); err == nil {
				sprint.StartDate = &startDate
			}
		}
		if item.EndDate != "" {
			if endDate, err := time.Parse(time.RFC3339, item.EndDate); err == nil {
				sprint.EndDate = &endDate
			}
		}
		if item.CompleteDate != "" {
			if completeDate, err := time.Parse(time.RFC3339, item.CompleteDate); err == nil {
				sprint.CompleteDate = &completeDate
			}
		}

		sprints = append(sprints, sprint)
	}

	return sprints, nil
}

// GetSprintStats calculates sprint statistics for all boards in a project
func (js *JiraService) GetSprintStats(projectKey string) (*models.JiraSprintStats, error) {
	// Get all boards for the project
	boards, err := js.GetProjectBoards(projectKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get project boards: %w", err)
	}

	stats := &models.JiraSprintStats{}
	var totalDuration float64
	var durationCount int
	var allSprints []models.JiraSprint

	// Aggregate sprints from all boards
	for _, board := range boards {
		sprints, err := js.GetSprints(int(board.ID))
		if err != nil {
			// Log error but continue with other boards
			continue
		}
		allSprints = append(allSprints, sprints...)
	}

	stats.TotalSprints = len(allSprints)

	for _, sprint := range allSprints {
		if sprint.State == "active" {
			stats.ActiveSprints++
		} else if sprint.State == "closed" {
			stats.CompletedSprints++

			// Calculate sprint duration
			if sprint.StartDate != nil && sprint.EndDate != nil {
				duration := sprint.EndDate.Sub(*sprint.StartDate)
				totalDuration += duration.Hours() / 24 // Convert to days
				durationCount++
			}
		}
	}

	if durationCount > 0 {
		stats.AvgSprintDuration = totalDuration / float64(durationCount)
	}

	return stats, nil
}

// GetProjectMetrics gets comprehensive metrics for a project across all boards
func (js *JiraService) GetProjectMetrics(projectKey string) (*models.JiraProjectMetrics, error) {
	metrics := &models.JiraProjectMetrics{
		ProjectKey:  projectKey,
		CollectedAt: time.Now(),
	}

	// Get issue stats
	issueStats, err := js.GetIssueStats(projectKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue stats: %w", err)
	}
	metrics.IssueStats = *issueStats

	// Get sprint stats across all boards
	sprintStats, err := js.GetSprintStats(projectKey)
	if err == nil {
		metrics.SprintStats = *sprintStats
	}

	// Get assignee stats
	issuesByAssignee, err := js.GetIssuesByAssignee(projectKey)
	if err == nil {
		for assignee, issues := range issuesByAssignee {
			assigneeStats := models.JiraAssigneeStats{
				Assignee:     assignee,
				TotalIssues:  len(issues),
			}

			var totalResolveTime float64
			var resolvedCount int

			for _, issue := range issues {
				status := strings.ToLower(issue.Status)
				if strings.Contains(status, "done") || strings.Contains(status, "closed") {
					assigneeStats.ClosedIssues++
				} else {
					assigneeStats.OpenIssues++
				}

				issueType := strings.ToLower(issue.IssueType)
				if strings.Contains(issueType, "bug") {
					assigneeStats.Bugs++
				} else if strings.Contains(issueType, "task") {
					assigneeStats.Tasks++
				}

				if issue.Resolved != nil {
					duration := issue.Resolved.Sub(issue.Created)
					totalResolveTime += duration.Hours()
					resolvedCount++
				}
			}

			if resolvedCount > 0 {
				assigneeStats.AvgTimeToResolve = totalResolveTime / float64(resolvedCount)
			}

			metrics.AssigneeStats = append(metrics.AssigneeStats, assigneeStats)
		}
	}

	return metrics, nil
}

