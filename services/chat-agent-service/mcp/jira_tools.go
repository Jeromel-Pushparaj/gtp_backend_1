package mcp

// Jira tool argument types for documentation purposes.
// The HTTP-based MCP server uses dynamic tool mapping via mapJiraToolToEndpoint()
// instead of explicit tool registration.

type GetJiraIssueStatsArgs struct {
	Project string `json:"project" jsonschema:"required,description=Jira project key"`
}

type GetJiraOpenBugsArgs struct {
	Project string `json:"project" jsonschema:"required,description=Jira project key"`
}

type GetJiraOpenTasksArgs struct {
	Project string `json:"project" jsonschema:"required,description=Jira project key"`
}

type GetJiraIssuesByAssigneeArgs struct {
	Project string `json:"project" jsonschema:"required,description=Jira project key"`
}

type GetJiraSprintStatsArgs struct {
	Project string `json:"project" jsonschema:"required,description=Jira project key"`
}

type GetJiraProjectMetricsArgs struct {
	Project string `json:"project" jsonschema:"required,description=Jira project key"`
}

type SearchJiraIssuesArgs struct {
	JQL        string `json:"jql" jsonschema:"required,description=JQL query string"`
	MaxResults string `json:"max_results" jsonschema:"description=Maximum number of results"`
}

// mapJiraToolToEndpoint maps Jira tool names to API endpoints
func mapJiraToolToEndpoint(toolName string, args map[string]interface{}) (endpoint, method string, params map[string]string, body interface{}, found bool) {
	params = make(map[string]string)

	switch toolName {
	case "get_jira_issue_stats":
		if project, ok := args["project"].(string); ok {
			params["project"] = project
		}
		return "/api/v1/jira/issues/stats", "GET", params, nil, true

	case "get_jira_open_bugs":
		if project, ok := args["project"].(string); ok {
			params["project"] = project
		}
		return "/api/v1/jira/bugs/open", "GET", params, nil, true

	case "get_jira_open_tasks":
		if project, ok := args["project"].(string); ok {
			params["project"] = project
		}
		return "/api/v1/jira/tasks/open", "GET", params, nil, true

	case "get_jira_issues_by_assignee":
		if project, ok := args["project"].(string); ok {
			params["project"] = project
		}
		return "/api/v1/jira/issues/by-assignee", "GET", params, nil, true

	case "get_jira_sprint_stats":
		if project, ok := args["project"].(string); ok {
			params["project"] = project
		}
		return "/api/v1/jira/sprints/stats", "GET", params, nil, true

	case "get_jira_project_metrics":
		if project, ok := args["project"].(string); ok {
			params["project"] = project
		}
		return "/api/v1/jira/metrics", "GET", params, nil, true

	case "search_jira_issues":
		if jql, ok := args["jql"].(string); ok {
			params["jql"] = jql
		}
		if maxResults, ok := args["max_results"].(string); ok {
			params["max_results"] = maxResults
		}
		return "/api/v1/jira/issues/search", "GET", params, nil, true

	default:
		return "", "", nil, nil, false
	}
}