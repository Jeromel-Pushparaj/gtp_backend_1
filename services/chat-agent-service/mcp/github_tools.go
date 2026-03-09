package mcp

// GitHub tool argument types for documentation purposes.
// The HTTP-based MCP server uses dynamic tool mapping via mapGitHubToolToEndpoint()
// instead of explicit tool registration.

type ListPullRequestsArgs struct {
	Repo  string `json:"repo" jsonschema:"required,description=Repository name"`
	State string `json:"state" jsonschema:"description=PR state (open, closed, all)"`
}

type GetPullRequestArgs struct {
	Repo   string `json:"repo" jsonschema:"required,description=Repository name"`
	Number string `json:"number" jsonschema:"required,description=PR number"`
}

type ListCommitsArgs struct {
	Repo  string `json:"repo" jsonschema:"required,description=Repository name"`
	Since string `json:"since" jsonschema:"description=ISO 8601 date string to filter commits"`
}

type GetCommitActivityArgs struct {
	Repo string `json:"repo" jsonschema:"required,description=Repository name"`
}

type ListIssuesArgs struct {
	Repo  string `json:"repo" jsonschema:"required,description=Repository name"`
	State string `json:"state" jsonschema:"description=Issue state (open, closed, all)"`
}

type ListIssueCommentsArgs struct {
	Repo   string `json:"repo" jsonschema:"required,description=Repository name"`
	Number string `json:"number" jsonschema:"required,description=Issue number"`
}

type CheckReadmeArgs struct {
	Repo string `json:"repo" jsonschema:"required,description=Repository name"`
}

type ListBranchesArgs struct {
	Repo string `json:"repo" jsonschema:"required,description=Repository name"`
}

type ListOrgMembersArgs struct{}

type ListOrgTeamsArgs struct{}

type GetRepositoryMetricsArgs struct {
	Repo string `json:"repo" jsonschema:"required,description=Repository name"`
}

type GetAllRepositoriesMetricsArgs struct{}

// mapGitHubToolToEndpoint maps GitHub tool names to API endpoints
func mapGitHubToolToEndpoint(toolName string, args map[string]interface{}) (endpoint, method string, params map[string]string, body interface{}, found bool) {
	params = make(map[string]string)

	switch toolName {
	case "list_pull_requests":
		if repo, ok := args["repo"].(string); ok {
			params["repo"] = repo
		}
		if state, ok := args["state"].(string); ok {
			params["state"] = state
		}
		return "/api/v1/github/pulls", "GET", params, nil, true

	case "get_pull_request":
		if repo, ok := args["repo"].(string); ok {
			params["repo"] = repo
		}
		if number, ok := args["number"].(string); ok {
			params["number"] = number
		}
		return "/api/v1/github/pulls/get", "GET", params, nil, true

	case "list_commits":
		if repo, ok := args["repo"].(string); ok {
			params["repo"] = repo
		}
		if since, ok := args["since"].(string); ok {
			params["since"] = since
		}
		return "/api/v1/github/commits", "GET", params, nil, true

	case "get_commit_activity":
		if repo, ok := args["repo"].(string); ok {
			params["repo"] = repo
		}
		return "/api/v1/github/commits/activity", "GET", params, nil, true

	case "list_issues":
		if repo, ok := args["repo"].(string); ok {
			params["repo"] = repo
		}
		if state, ok := args["state"].(string); ok {
			params["state"] = state
		}
		return "/api/v1/github/issues", "GET", params, nil, true

	case "list_issue_comments":
		if repo, ok := args["repo"].(string); ok {
			params["repo"] = repo
		}
		if number, ok := args["number"].(string); ok {
			params["number"] = number
		}
		return "/api/v1/github/issues/comments", "GET", params, nil, true

	case "check_readme":
		if repo, ok := args["repo"].(string); ok {
			params["repo"] = repo
		}
		return "/api/v1/github/readme", "GET", params, nil, true

	case "list_branches":
		if repo, ok := args["repo"].(string); ok {
			params["repo"] = repo
		}
		return "/api/v1/github/branches", "GET", params, nil, true

	case "list_org_members":
		return "/api/v1/github/org/members", "GET", nil, nil, true

	case "list_org_teams":
		return "/api/v1/github/org/teams", "GET", nil, nil, true

	case "get_repository_metrics":
		if repo, ok := args["repo"].(string); ok {
			params["repo"] = repo
		}
		return "/api/v1/github/metrics", "GET", params, nil, true

	case "get_all_repositories_metrics":
		return "/api/v1/github/metrics/all", "GET", nil, nil, true

	default:
		return "", "", nil, nil, false
	}
}