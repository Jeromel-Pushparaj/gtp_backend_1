package mcp

import (
	mcp_golang "github.com/metoro-io/mcp-golang"
)

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

func (s *Server) registerMetricsTools() error {
	if err := s.mcpServer.RegisterTool("collect_github_metrics", "Collect and store GitHub metrics for a repository", func(args CollectGitHubMetricsArgs) (*mcp_golang.ToolResponse, error) {
		params := map[string]string{"repo": args.Repo}
		result, err := s.makeRequest("POST", "/api/v1/metrics/github/collect", params, nil)
		if err != nil {
			return nil, err
		}
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(result)), nil
	}); err != nil {
		return err
	}

	if err := s.mcpServer.RegisterTool("get_stored_github_metrics", "Get stored GitHub metrics for a repository", func(args GetStoredGitHubMetricsArgs) (*mcp_golang.ToolResponse, error) {
		params := map[string]string{"repo": args.Repo}
		result, err := s.makeRequest("GET", "/api/v1/metrics/github/stored", params, nil)
		if err != nil {
			return nil, err
		}
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(result)), nil
	}); err != nil {
		return err
	}

	if err := s.mcpServer.RegisterTool("collect_sonar_metrics", "Collect and store SonarCloud metrics for a repository", func(args CollectSonarMetricsArgs) (*mcp_golang.ToolResponse, error) {
		params := map[string]string{"repo": args.Repo}
		result, err := s.makeRequest("POST", "/api/v1/metrics/sonar/collect", params, nil)
		if err != nil {
			return nil, err
		}
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(result)), nil
	}); err != nil {
		return err
	}

	if err := s.mcpServer.RegisterTool("get_stored_sonar_metrics", "Get stored SonarCloud metrics for a repository", func(args GetStoredSonarMetricsArgs) (*mcp_golang.ToolResponse, error) {
		params := map[string]string{"repo": args.Repo}
		result, err := s.makeRequest("GET", "/api/v1/metrics/sonar/stored", params, nil)
		if err != nil {
			return nil, err
		}
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(result)), nil
	}); err != nil {
		return err
	}

	return nil
}

