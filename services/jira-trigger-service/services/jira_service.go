package services

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/keerthanau/go/models"
)

// ProcessCreateIssue handles the main logic for creating issue and adding to sprint
func ProcessCreateIssue(jira *models.JiraAPIClient, projectKey string, req *models.CreateIssueRequest) models.CreateIssueResponse {
	// Step 1: Create the issue
	issueKey, _, err := CreateJiraIssue(jira, projectKey, req)
	if err != nil {
		return models.CreateIssueResponse{
			Success: false,
			Error:   "Failed to create issue: " + err.Error(),
		}
	}

	// Step 2: Get board for the project
	boardID, err := GetBoardID(jira, projectKey)
	if err != nil {
		return models.CreateIssueResponse{
			Success: false,
			Error:   "Failed to get board: " + err.Error(),
		}
	}

	// Step 3: Find or create sprint
	var sprintID int
	var sprintName string

	if req.SprintName != "" {
		// Sprint name provided - find existing sprint
		sprintID, err = FindSprintByName(jira, boardID, req.SprintName)
		if err != nil {
			return models.CreateIssueResponse{
				Success: false,
				Error:   "Failed to find sprint: " + err.Error(),
			}
		}
		sprintName = req.SprintName
		log.Printf("✅ Found existing sprint: %s (ID: %d)", sprintName, sprintID)
	} else {
		// No sprint name - create new sprint with auto-generated name
		sprintName = fmt.Sprintf("Auto Sprint %s", time.Now().Format("2006-01-02 15:04"))
		sprintID, err = CreateSprint(jira, boardID, sprintName)
		if err != nil {
			return models.CreateIssueResponse{
				Success: false,
				Error:   "Failed to create sprint: " + err.Error(),
			}
		}

		// Start the newly created sprint
		err = StartSprint(jira, sprintID)
		if err != nil {
			log.Printf("Warning: Could not start sprint: %v", err)
		}
	}

	// Step 4: Add issue to sprint
	err = AddIssueToSprint(jira, sprintID, issueKey)
	if err != nil {
		return models.CreateIssueResponse{
			Success: false,
			Error:   "Failed to add issue to sprint: " + err.Error(),
		}
	}

	// Success!
	return models.CreateIssueResponse{
		Success:  true,
		IssueKey: issueKey,
		IssueURL: jira.BaseURL + "/browse/" + issueKey,
		SprintID: sprintID,
		Message:  fmt.Sprintf("Issue %s created and added to sprint '%s' successfully", issueKey, sprintName),
	}
}

// CreateJiraIssue creates a new Jira issue
func CreateJiraIssue(jira *models.JiraAPIClient, projectKey string, req *models.CreateIssueRequest) (string, string, error) {
	// Build base fields
	fields := map[string]any{
		"project":   map[string]any{"key": projectKey},
		"summary":   req.Summary,
		"issuetype": map[string]any{"name": req.IssueType},
		"description": map[string]any{
			"type":    "doc",
			"version": 1,
			"content": []map[string]any{
				{
					"type": "paragraph",
					"content": []map[string]any{
						{
							"type": "text",
							"text": req.Description,
						},
					},
				},
			},
		},
	}

	// Add optional fields if provided
	// Handle assignee - prioritize ID, then fall back to name
	if req.AssigneeID != "" {
		fields["assignee"] = map[string]any{"accountId": req.AssigneeID}
	} else if req.AssigneeName != "" {
		// Find user by display name
		accountID, err := FindUserByName(jira, req.AssigneeName)
		if err != nil {
			log.Printf("⚠️  Warning: Could not find user '%s': %v", req.AssigneeName, err)
		} else {
			fields["assignee"] = map[string]any{"accountId": accountID}
			log.Printf("✅ Found user '%s' with ID: %s", req.AssigneeName, accountID)
		}
	}

	if req.Priority != "" {
		fields["priority"] = map[string]any{"name": req.Priority}
		log.Printf("🔧 Setting priority: %s", req.Priority)
	}

	if len(req.Labels) > 0 {
		fields["labels"] = req.Labels
		log.Printf("🔧 Setting labels: %v", req.Labels)
	}

	createBody := map[string]any{"fields": fields}

	// Debug: Log the full request body
	debugJSON, _ := json.MarshalIndent(createBody, "", "  ")
	log.Printf("📤 Sending to Jira API:\n%s", string(debugJSON))

	respBytes, err := jira.Do("POST", "/rest/api/3/issue", createBody)
	if err != nil {
		log.Printf("❌ Error creating issue: %v", err)
		return "", "", err
	}

	// Debug: Log the response from Jira
	log.Printf("📥 Response from Jira: %s", string(respBytes))

	var created struct {
		Key string `json:"key"`
		ID  string `json:"id"`
	}
	json.Unmarshal(respBytes, &created)
	log.Printf("✅ Created issue: %s (Type: %s, Priority: %s)", created.Key, req.IssueType, req.Priority)
	return created.Key, created.ID, nil
}

// GetBoardID gets the board ID for a project
func GetBoardID(jira *models.JiraAPIClient, projectKey string) (int, error) {
	path := "/rest/agile/1.0/board?projectKeyOrId=" + url.QueryEscape(projectKey)
	respBytes, err := jira.Do("GET", path, nil)
	if err != nil {
		return 0, err
	}

	var boards struct {
		Values []struct {
			ID int `json:"id"`
		} `json:"values"`
	}
	json.Unmarshal(respBytes, &boards)
	if len(boards.Values) == 0 {
		return 0, fmt.Errorf("no boards found for project")
	}
	log.Printf("✅ Found board ID: %d", boards.Values[0].ID)
	return boards.Values[0].ID, nil
}

// CreateSprint creates a new sprint
func CreateSprint(jira *models.JiraAPIClient, boardID int, sprintName string) (int, error) {
	startDate := time.Now().Format("2006-01-02T15:04:05.000Z")
	endDate := time.Now().AddDate(0, 0, 14).Format("2006-01-02T15:04:05.000Z")

	body := map[string]any{
		"name":          sprintName,
		"startDate":     startDate,
		"endDate":       endDate,
		"originBoardId": boardID,
	}

	respBytes, err := jira.Do("POST", "/rest/agile/1.0/sprint", body)
	if err != nil {
		return 0, err
	}

	var sprint struct {
		ID int `json:"id"`
	}
	json.Unmarshal(respBytes, &sprint)
	log.Printf("✅ Created sprint: %s (ID: %d)", sprintName, sprint.ID)
	return sprint.ID, nil
}

// StartSprint starts a sprint
func StartSprint(jira *models.JiraAPIClient, sprintID int) error {
	path := fmt.Sprintf("/rest/agile/1.0/sprint/%d", sprintID)
	body := map[string]any{"state": "active"}
	_, err := jira.Do("POST", path, body)
	if err == nil {
		log.Printf("✅ Started sprint ID: %d", sprintID)
	}
	return err
}

// AddIssueToSprint adds an issue to a sprint
func AddIssueToSprint(jira *models.JiraAPIClient, sprintID int, issueKey string) error {
	path := fmt.Sprintf("/rest/agile/1.0/sprint/%d/issue", sprintID)
	body := map[string]any{"issues": []string{issueKey}}
	_, err := jira.Do("POST", path, body)
	if err == nil {
		log.Printf("✅ Added issue %s to sprint %d", issueKey, sprintID)
	}
	return err
}

// FindSprintByName finds a sprint by name (searches active and future sprints)
func FindSprintByName(jira *models.JiraAPIClient, boardID int, sprintName string) (int, error) {
	// Get all sprints for the board (active and future)
	path := fmt.Sprintf("/rest/agile/1.0/board/%d/sprint?state=active,future", boardID)
	respBytes, err := jira.Do("GET", path, nil)
	if err != nil {
		return 0, err
	}

	var sprints struct {
		Values []struct {
			ID    int    `json:"id"`
			Name  string `json:"name"`
			State string `json:"state"`
		} `json:"values"`
	}
	json.Unmarshal(respBytes, &sprints)

	// Search for sprint by name
	for _, sprint := range sprints.Values {
		if sprint.Name == sprintName {
			return sprint.ID, nil
		}
	}

	return 0, fmt.Errorf("sprint '%s' not found (searched active and future sprints)", sprintName)
}

// FindUserByName finds a user by display name and returns their account ID
func FindUserByName(jira *models.JiraAPIClient, displayName string) (string, error) {
	// Search for user by query (searches display name and email)
	path := fmt.Sprintf("/rest/api/3/user/search?query=%s", url.QueryEscape(displayName))
	respBytes, err := jira.Do("GET", path, nil)
	if err != nil {
		return "", err
	}

	var users []struct {
		AccountID   string `json:"accountId"`
		DisplayName string `json:"displayName"`
		Active      bool   `json:"active"`
	}
	json.Unmarshal(respBytes, &users)

	// Search for exact match (case-insensitive)
	for _, user := range users {
		if user.Active && user.DisplayName == displayName {
			return user.AccountID, nil
		}
	}

	// If no exact match, try partial match
	for _, user := range users {
		if user.Active {
			return user.AccountID, nil
		}
	}

	return "", fmt.Errorf("user '%s' not found", displayName)
}
