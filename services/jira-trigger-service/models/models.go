package models

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// JiraAPIClient handles Jira API requests
type JiraAPIClient struct {
	BaseURL string
	Email   string
	Token   string
	HTTP    *http.Client
}

// AuthHeader generates Basic Auth header
func (j *JiraAPIClient) AuthHeader() string {
	raw := j.Email + ":" + j.Token
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(raw))
}

// Do makes HTTP request to Jira API
func (j *JiraAPIClient) Do(method, path string, body any) ([]byte, error) {
	var r io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		r = bytes.NewBuffer(b)
	}
	req, err := http.NewRequest(method, j.BaseURL+path, r)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", j.AuthHeader())
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := j.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	out, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%s %s failed (%d): %s", method, path, resp.StatusCode, string(out))
	}
	return out, nil
}

// CreateIssueRequest represents the incoming request body
type CreateIssueRequest struct {
	// Required fields
	Summary string `json:"summary"` // Issue title (REQUIRED)

	// Optional fields
	ProjectKey   string   `json:"projectKey"`   // Project key (optional, uses config default if empty)
	IssueType    string   `json:"issueType"`    // Type: Task, Story, Bug (optional, default: Task)
	SprintName   string   `json:"sprintName"`   // Name of the sprint (optional - if empty, creates new sprint)
	Description  string   `json:"description"`  // Issue description (optional, default: "Created via API")
	AssigneeID   string   `json:"assigneeId"`   // Jira account ID (optional, takes priority over assigneeName)
	AssigneeName string   `json:"assigneeName"` // Jira display name (optional, e.g., "Ganesh Sriramulu")
	Priority     string   `json:"priority"`     // Priority: Highest, High, Medium, Low, Lowest (optional)
	Labels       []string `json:"labels"`       // Array of labels (optional)
}

// CreateIssueResponse represents the API response
type CreateIssueResponse struct {
	Success  bool   `json:"success"`
	IssueKey string `json:"issueKey,omitempty"`
	IssueURL string `json:"issueUrl,omitempty"`
	SprintID int    `json:"sprintId,omitempty"`
	Message  string `json:"message,omitempty"`
	Error    string `json:"error,omitempty"`
}
