package client

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type ToolExecutor struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewToolExecutor(baseURL, apiKey string) *ToolExecutor {
	return &ToolExecutor{
		baseURL:    baseURL,
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}
}

func (t *ToolExecutor) ExecuteTool(toolName string, arguments map[string]interface{}) (string, error) {
	log.Printf("Executing tool: %s with arguments: %v", toolName, arguments)

	// Map tool names to API endpoints
	endpoint, method, params, body := t.mapToolToEndpoint(toolName, arguments)
	if endpoint == "" {
		return "", fmt.Errorf("unknown tool: %s", toolName)
	}

	log.Printf("Calling backend: %s %s%s", method, t.baseURL, endpoint)
	result, err := t.makeRequest(method, endpoint, params, body)
	if err != nil {
		log.Printf("Backend call failed: %v", err)
		return "", err
	}

	log.Printf("Backend call successful, result length: %d", len(result))
	return result, nil
}

func (t *ToolExecutor) makeRequest(method, endpoint string, queryParams map[string]string, body interface{}) (string, error) {
	fullURL := t.baseURL + endpoint

	if len(queryParams) > 0 {
		params := url.Values{}
		for k, v := range queryParams {
			params.Add(k, v)
		}
		fullURL += "?" + params.Encode()
	}

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return "", err
		}
		reqBody = strings.NewReader(string(jsonData))
	}

	req, err := http.NewRequest(method, fullURL, reqBody)
	if err != nil {
		return "", err
	}

	if t.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+t.apiKey)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("API error: %s - %s", resp.Status, string(responseBody))
	}

	return string(responseBody), nil
}

func (t *ToolExecutor) mapToolToEndpoint(toolName string, args map[string]interface{}) (endpoint, method string, params map[string]string, body interface{}) {
	params = make(map[string]string)

	switch toolName {
	case "health_check":
		return "/health", "GET", nil, nil

	// GitHub Tools
	case "list_pull_requests":
		if repo, ok := args["repo"].(string); ok {
			params["repo"] = repo
		}
		if state, ok := args["state"].(string); ok {
			params["state"] = state
		}
		return "/api/v1/github/pulls", "GET", params, nil

	case "list_commits":
		if repo, ok := args["repo"].(string); ok {
			params["repo"] = repo
		}
		if since, ok := args["since"].(string); ok {
			params["since"] = since
		}
		return "/api/v1/github/commits", "GET", params, nil

	case "list_issues":
		if repo, ok := args["repo"].(string); ok {
			params["repo"] = repo
		}
		if state, ok := args["state"].(string); ok {
			params["state"] = state
		}
		return "/api/v1/github/issues", "GET", params, nil

	case "check_readme":
		if repo, ok := args["repo"].(string); ok {
			params["repo"] = repo
		}
		return "/api/v1/github/readme", "GET", params, nil

	case "list_branches":
		if repo, ok := args["repo"].(string); ok {
			params["repo"] = repo
		}
		return "/api/v1/github/branches", "GET", params, nil

	case "list_org_members":
		return "/api/v1/github/org/members", "GET", nil, nil

	case "list_org_teams":
		return "/api/v1/github/org/teams", "GET", nil, nil

	// Organization Tools
	case "fetch_orgs":
		return "/api/v1/orgs", "GET", nil, nil

	case "create_org":
		return "/api/v1/orgs/create", "POST", nil, args

	// Repository Tools
	case "fetch_repos_by_org":
		if orgID, ok := args["org_id"].(string); ok {
			params["org_id"] = orgID
		}
		return "/api/v1/repos/fetch", "GET", params, nil

	// Jira Tools
	case "get_jira_issue_stats":
		if project, ok := args["project"].(string); ok {
			params["project"] = project
		}
		return "/api/v1/jira/issues/stats", "GET", params, nil

	case "get_jira_open_bugs":
		if project, ok := args["project"].(string); ok {
			params["project"] = project
		}
		return "/api/v1/jira/bugs/open", "GET", params, nil

	case "get_jira_open_tasks":
		if project, ok := args["project"].(string); ok {
			params["project"] = project
		}
		return "/api/v1/jira/tasks/open", "GET", params, nil

	case "search_jira_issues":
		if jql, ok := args["jql"].(string); ok {
			params["jql"] = jql
		}
		if maxResults, ok := args["max_results"].(string); ok {
			params["max_results"] = maxResults
		}
		return "/api/v1/jira/issues/search", "GET", params, nil

	default:
		return "", "", nil, nil
	}
}
