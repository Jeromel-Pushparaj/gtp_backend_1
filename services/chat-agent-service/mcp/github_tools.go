package mcp

import (
	mcp_golang "github.com/metoro-io/mcp-golang"
)

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
	Repo  string `json:"repo" jsonschema:"required,description=Repository name"`
	Issue string `json:"issue" jsonschema:"required,description=Issue number"`
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

func (s *Server) registerGitHubTools() error {
	if err := s.mcpServer.RegisterTool("list_pull_requests", "List pull requests for a repository", func(args ListPullRequestsArgs) (*mcp_golang.ToolResponse, error) {
		params := map[string]string{"repo": args.Repo}
		if args.State != "" {
			params["state"] = args.State
		}
		result, err := s.makeRequest("GET", "/api/v1/github/pulls", params, nil)
		if err != nil {
			return nil, err
		}
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(result)), nil
	}); err != nil {
		return err
	}

	if err := s.mcpServer.RegisterTool("get_pull_request", "Get details of a specific pull request", func(args GetPullRequestArgs) (*mcp_golang.ToolResponse, error) {
		params := map[string]string{"repo": args.Repo, "number": args.Number}
		result, err := s.makeRequest("GET", "/api/v1/github/pulls/get", params, nil)
		if err != nil {
			return nil, err
		}
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(result)), nil
	}); err != nil {
		return err
	}

	if err := s.mcpServer.RegisterTool("list_commits", "List commits for a repository", func(args ListCommitsArgs) (*mcp_golang.ToolResponse, error) {
		params := map[string]string{"repo": args.Repo}
		if args.Since != "" {
			params["since"] = args.Since
		}
		result, err := s.makeRequest("GET", "/api/v1/github/commits", params, nil)
		if err != nil {
			return nil, err
		}
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(result)), nil
	}); err != nil {
		return err
	}

	if err := s.mcpServer.RegisterTool("get_commit_activity", "Get commit activity analysis for a repository", func(args GetCommitActivityArgs) (*mcp_golang.ToolResponse, error) {
		params := map[string]string{"repo": args.Repo}
		result, err := s.makeRequest("GET", "/api/v1/github/commits/activity", params, nil)
		if err != nil {
			return nil, err
		}
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(result)), nil
	}); err != nil {
		return err
	}

	if err := s.mcpServer.RegisterTool("list_issues", "List issues for a repository", func(args ListIssuesArgs) (*mcp_golang.ToolResponse, error) {
		params := map[string]string{"repo": args.Repo}
		if args.State != "" {
			params["state"] = args.State
		}
		result, err := s.makeRequest("GET", "/api/v1/github/issues", params, nil)
		if err != nil {
			return nil, err
		}
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(result)), nil
	}); err != nil {
		return err
	}

	if err := s.mcpServer.RegisterTool("list_issue_comments", "List comments for a specific issue", func(args ListIssueCommentsArgs) (*mcp_golang.ToolResponse, error) {
		params := map[string]string{"repo": args.Repo, "issue": args.Issue}
		result, err := s.makeRequest("GET", "/api/v1/github/issues/comments", params, nil)
		if err != nil {
			return nil, err
		}
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(result)), nil
	}); err != nil {
		return err
	}

	if err := s.mcpServer.RegisterTool("check_readme", "Check if a repository has a README file", func(args CheckReadmeArgs) (*mcp_golang.ToolResponse, error) {
		params := map[string]string{"repo": args.Repo}
		result, err := s.makeRequest("GET", "/api/v1/github/readme", params, nil)
		if err != nil {
			return nil, err
		}
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(result)), nil
	}); err != nil {
		return err
	}

	if err := s.mcpServer.RegisterTool("list_branches", "List branches for a repository", func(args ListBranchesArgs) (*mcp_golang.ToolResponse, error) {
		params := map[string]string{"repo": args.Repo}
		result, err := s.makeRequest("GET", "/api/v1/github/branches", params, nil)
		if err != nil {
			return nil, err
		}
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(result)), nil
	}); err != nil {
		return err
	}

	if err := s.mcpServer.RegisterTool("list_org_members", "List organization members", func(args ListOrgMembersArgs) (*mcp_golang.ToolResponse, error) {
		result, err := s.makeRequest("GET", "/api/v1/github/org/members", nil, nil)
		if err != nil {
			return nil, err
		}
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(result)), nil
	}); err != nil {
		return err
	}

	if err := s.mcpServer.RegisterTool("list_org_teams", "List organization teams", func(args ListOrgTeamsArgs) (*mcp_golang.ToolResponse, error) {
		result, err := s.makeRequest("GET", "/api/v1/github/org/teams", nil, nil)
		if err != nil {
			return nil, err
		}
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(result)), nil
	}); err != nil {
		return err
	}

	if err := s.mcpServer.RegisterTool("get_repository_metrics", "Get comprehensive metrics for a specific repository", func(args GetRepositoryMetricsArgs) (*mcp_golang.ToolResponse, error) {
		params := map[string]string{"repo": args.Repo}
		result, err := s.makeRequest("GET", "/api/v1/github/metrics", params, nil)
		if err != nil {
			return nil, err
		}
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(result)), nil
	}); err != nil {
		return err
	}

	if err := s.mcpServer.RegisterTool("get_all_repositories_metrics", "Get metrics for all repositories in the organization", func(args GetAllRepositoriesMetricsArgs) (*mcp_golang.ToolResponse, error) {
		result, err := s.makeRequest("GET", "/api/v1/github/metrics/all", nil, nil)
		if err != nil {
			return nil, err
		}
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(result)), nil
	}); err != nil {
		return err
	}

	return nil
}

