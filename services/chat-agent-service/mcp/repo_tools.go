package mcp

import (
	mcp_golang "github.com/metoro-io/mcp-golang"
)

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

func (s *Server) registerRepoTools() error {
	if err := s.mcpServer.RegisterTool("fetch_repos_by_org", "Fetch repositories for an organization from GitHub", func(args FetchReposByOrgArgs) (*mcp_golang.ToolResponse, error) {
		params := map[string]string{"org_id": args.OrgID}
		result, err := s.makeRequest("GET", "/api/v1/repos/fetch", params, nil)
		if err != nil {
			return nil, err
		}
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(result)), nil
	}); err != nil {
		return err
	}

	if err := s.mcpServer.RegisterTool("update_repo", "Update repository configuration", func(args UpdateRepoArgs) (*mcp_golang.ToolResponse, error) {
		params := map[string]string{"repo_id": args.RepoID}
		body := make(map[string]string)
		if args.JiraProjectKey != "" {
			body["jira_project_key"] = args.JiraProjectKey
		}
		if args.EnvironmentName != "" {
			body["environment_name"] = args.EnvironmentName
		}
		result, err := s.makeRequest("PUT", "/api/v1/repos/update", params, body)
		if err != nil {
			return nil, err
		}
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(result)), nil
	}); err != nil {
		return err
	}

	if err := s.mcpServer.RegisterTool("fetch_github_metrics_by_repo", "Fetch and store GitHub metrics for a repository", func(args FetchGitHubMetricsByRepoArgs) (*mcp_golang.ToolResponse, error) {
		params := map[string]string{"repo_id": args.RepoID}
		result, err := s.makeRequest("GET", "/api/v1/repos/metrics/github", params, nil)
		if err != nil {
			return nil, err
		}
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(result)), nil
	}); err != nil {
		return err
	}

	if err := s.mcpServer.RegisterTool("fetch_jira_metrics_by_repo", "Fetch and store Jira metrics for a repository", func(args FetchJiraMetricsByRepoArgs) (*mcp_golang.ToolResponse, error) {
		params := map[string]string{"repo_id": args.RepoID}
		result, err := s.makeRequest("GET", "/api/v1/repos/metrics/jira", params, nil)
		if err != nil {
			return nil, err
		}
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(result)), nil
	}); err != nil {
		return err
	}

	if err := s.mcpServer.RegisterTool("fetch_sonar_metrics_by_repo", "Fetch and store SonarCloud metrics for a repository", func(args FetchSonarMetricsByRepoArgs) (*mcp_golang.ToolResponse, error) {
		params := map[string]string{"repo_id": args.RepoID}
		result, err := s.makeRequest("GET", "/api/v1/repos/metrics/sonar", params, nil)
		if err != nil {
			return nil, err
		}
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(result)), nil
	}); err != nil {
		return err
	}

	return nil
}

