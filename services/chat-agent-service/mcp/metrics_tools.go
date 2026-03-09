package mcp

// Metrics tool argument types for documentation purposes.
// The HTTP-based MCP server uses dynamic tool mapping via mapMetricsToolToEndpoint()
// instead of explicit tool registration.

type CollectGitHubMetricsArgs struct {
	Repo string `json:"repo" jsonschema:"required,description=Repository name"`
}

type GetStoredGitHubMetricsArgs struct {
	Repo string `json:"repo" jsonschema:"required,description=Repository name"`
}

type CollectSonarMetricsArgs struct {
	Repo string `json:"repo" jsonschema:"required,description=Repository name"`
}

type GetStoredSonarMetricsArgs struct {
	Repo string `json:"repo" jsonschema:"required,description=Repository name"`
}

// mapMetricsToolToEndpoint maps Metrics tool names to API endpoints
func mapMetricsToolToEndpoint(toolName string, args map[string]interface{}) (endpoint, method string, params map[string]string, body interface{}, found bool) {
	params = make(map[string]string)

	switch toolName {
	case "collect_github_metrics":
		if repo, ok := args["repo"].(string); ok {
			params["repo"] = repo
		}
		return "/api/v1/metrics/github/collect", "POST", params, nil, true

	case "get_stored_github_metrics":
		if repo, ok := args["repo"].(string); ok {
			params["repo"] = repo
		}
		return "/api/v1/metrics/github/stored", "GET", params, nil, true

	case "collect_sonar_metrics":
		if repo, ok := args["repo"].(string); ok {
			params["repo"] = repo
		}
		return "/api/v1/metrics/sonar/collect", "POST", params, nil, true

	case "get_stored_sonar_metrics":
		if repo, ok := args["repo"].(string); ok {
			params["repo"] = repo
		}
		return "/api/v1/metrics/sonar/stored", "GET", params, nil, true

	default:
		return "", "", nil, nil, false
	}
}