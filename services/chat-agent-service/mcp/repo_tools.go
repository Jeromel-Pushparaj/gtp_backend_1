package mcp

// Repository tool argument types for documentation purposes.
// The HTTP-based MCP server uses dynamic tool mapping via mapRepoToolToEndpoint()
// instead of explicit tool registration.

type FetchReposByOrgArgs struct {
	OrgID string `json:"org_id" jsonschema:"required,description=Organization ID"`
}

type UpdateRepoArgs struct {
	RepoID          string `json:"repo_id" jsonschema:"required,description=Repository ID"`
	JiraProjectKey  string `json:"jira_project_key" jsonschema:"description=Jira project key"`
	EnvironmentName string `json:"environment_name" jsonschema:"description=Environment name"`
}

type FetchGitHubMetricsByRepoArgs struct {
	RepoID string `json:"repo_id" jsonschema:"required,description=Repository ID"`
}

type FetchJiraMetricsByRepoArgs struct {
	RepoID string `json:"repo_id" jsonschema:"required,description=Repository ID"`
}

type FetchSonarMetricsByRepoArgs struct {
	RepoID string `json:"repo_id" jsonschema:"required,description=Repository ID"`
}

// mapRepoToolToEndpoint maps Repository tool names to API endpoints
func mapRepoToolToEndpoint(toolName string, args map[string]interface{}) (endpoint, method string, params map[string]string, body interface{}, found bool) {
	params = make(map[string]string)

	switch toolName {
	case "fetch_repos_by_org":
		if orgID, ok := args["org_id"].(string); ok {
			params["org_id"] = orgID
		}
		return "/api/v1/repos/fetch", "GET", params, nil, true

	case "update_repo":
		if repoID, ok := args["repo_id"].(string); ok {
			params["repo_id"] = repoID
		}
		bodyMap := make(map[string]string)
		if jiraKey, ok := args["jira_project_key"].(string); ok {
			bodyMap["jira_project_key"] = jiraKey
		}
		if envName, ok := args["environment_name"].(string); ok {
			bodyMap["environment_name"] = envName
		}
		return "/api/v1/repos/update", "PUT", params, bodyMap, true

	case "fetch_github_metrics_by_repo":
		if repoID, ok := args["repo_id"].(string); ok {
			params["repo_id"] = repoID
		}
		return "/api/v1/repos/metrics/github", "GET", params, nil, true

	case "fetch_jira_metrics_by_repo":
		if repoID, ok := args["repo_id"].(string); ok {
			params["repo_id"] = repoID
		}
		return "/api/v1/repos/metrics/jira", "GET", params, nil, true

	case "fetch_sonar_metrics_by_repo":
		if repoID, ok := args["repo_id"].(string); ok {
			params["repo_id"] = repoID
		}
		return "/api/v1/repos/metrics/sonar", "GET", params, nil, true

	default:
		return "", "", nil, nil, false
	}
}