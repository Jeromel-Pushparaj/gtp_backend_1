package mcp

import (
	mcp_golang "github.com/metoro-io/mcp-golang"
)

type ListSecretsArgs struct{}

type AddEnvSecretsArgs struct{}

type UpdateWorkflowsArgs struct{}

type FullSetupArgs struct{}

type FetchResultsArgs struct{}

type GetSonarMetricsArgs struct {
	Repo           string `json:"repo" jsonschema:"required,description=Repository name"`
	IncludeIssues  string `json:"include_issues" jsonschema:"description=Set to 'true' to include issue details"`
}

type ProcessRepositoryArgs struct {
	RepositoryName string `json:"repository_name" jsonschema:"required,description=Repository name to process"`
}

func (s *Server) registerSonarCloudTools() error {
	if err := s.mcpServer.RegisterTool("list_secrets", "List all secrets for repositories in the organization", func(args ListSecretsArgs) (*mcp_golang.ToolResponse, error) {
		result, err := s.makeRequest("GET", "/api/v1/secrets/list", nil, nil)
		if err != nil {
			return nil, err
		}
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(result)), nil
	}); err != nil {
		return err
	}

	if err := s.mcpServer.RegisterTool("add_env_secrets", "Add environment secrets to all repositories", func(args AddEnvSecretsArgs) (*mcp_golang.ToolResponse, error) {
		result, err := s.makeRequest("POST", "/api/v1/secrets/add", nil, nil)
		if err != nil {
			return nil, err
		}
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(result)), nil
	}); err != nil {
		return err
	}

	if err := s.mcpServer.RegisterTool("update_workflows", "Update GitHub workflows to use environment secrets", func(args UpdateWorkflowsArgs) (*mcp_golang.ToolResponse, error) {
		result, err := s.makeRequest("POST", "/api/v1/workflows/update", nil, nil)
		if err != nil {
			return nil, err
		}
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(result)), nil
	}); err != nil {
		return err
	}

	if err := s.mcpServer.RegisterTool("full_setup", "Perform full SonarCloud setup for all repositories", func(args FullSetupArgs) (*mcp_golang.ToolResponse, error) {
		result, err := s.makeRequest("POST", "/api/v1/setup/full", nil, nil)
		if err != nil {
			return nil, err
		}
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(result)), nil
	}); err != nil {
		return err
	}

	if err := s.mcpServer.RegisterTool("fetch_results", "Fetch SonarCloud analysis results for all repositories", func(args FetchResultsArgs) (*mcp_golang.ToolResponse, error) {
		result, err := s.makeRequest("GET", "/api/v1/results/fetch", nil, nil)
		if err != nil {
			return nil, err
		}
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(result)), nil
	}); err != nil {
		return err
	}

	if err := s.mcpServer.RegisterTool("get_sonar_metrics", "Get SonarCloud metrics for a specific repository", func(args GetSonarMetricsArgs) (*mcp_golang.ToolResponse, error) {
		params := map[string]string{"repo": args.Repo}
		if args.IncludeIssues != "" {
			params["include_issues"] = args.IncludeIssues
		}
		result, err := s.makeRequest("GET", "/api/v1/sonar/metrics", params, nil)
		if err != nil {
			return nil, err
		}
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(result)), nil
	}); err != nil {
		return err
	}

	if err := s.mcpServer.RegisterTool("process_repository", "Process a single repository for SonarCloud setup", func(args ProcessRepositoryArgs) (*mcp_golang.ToolResponse, error) {
		body := map[string]string{"repository_name": args.RepositoryName}
		result, err := s.makeRequest("POST", "/api/v1/repository/process", nil, body)
		if err != nil {
			return nil, err
		}
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(result)), nil
	}); err != nil {
		return err
	}

	return nil
}

