package mcp

// SonarCloud tool argument types for documentation purposes.
// The HTTP-based MCP server uses dynamic tool mapping via mapSonarCloudToolToEndpoint()
// instead of explicit tool registration.

type ListSecretsArgs struct{}

type AddEnvSecretsArgs struct{}

type UpdateWorkflowsArgs struct{}

type FullSetupArgs struct{}

type FetchResultsArgs struct{}

type GetSonarMetricsArgs struct {
	Repo          string `json:"repo" jsonschema:"required,description=Repository name"`
	IncludeIssues string `json:"include_issues" jsonschema:"description=Set to 'true' to include issue details"`
}

type ProcessRepositoryArgs struct {
	RepositoryName string `json:"repository_name" jsonschema:"required,description=Repository name to process"`
}

// mapSonarCloudToolToEndpoint maps SonarCloud tool names to API endpoints
func mapSonarCloudToolToEndpoint(toolName string, args map[string]interface{}) (endpoint, method string, params map[string]string, body interface{}, found bool) {
	params = make(map[string]string)

	switch toolName {
	case "list_secrets":
		return "/api/v1/secrets/list", "GET", nil, nil, true

	case "add_env_secrets":
		return "/api/v1/secrets/add", "POST", nil, nil, true

	case "update_workflows":
		return "/api/v1/workflows/update", "POST", nil, nil, true

	case "full_setup":
		return "/api/v1/setup/full", "POST", nil, nil, true

	case "fetch_results":
		return "/api/v1/results/fetch", "GET", nil, nil, true

	case "get_sonar_metrics":
		if repo, ok := args["repo"].(string); ok {
			params["repo"] = repo
		}
		if includeIssues, ok := args["include_issues"].(string); ok {
			params["include_issues"] = includeIssues
		}
		return "/api/v1/sonar/metrics", "GET", params, nil, true

	case "process_repository":
		bodyMap := make(map[string]string)
		if repoName, ok := args["repository_name"].(string); ok {
			bodyMap["repository_name"] = repoName
		}
		return "/api/v1/repository/process", "POST", nil, bodyMap, true

	default:
		return "", "", nil, nil, false
	}
}