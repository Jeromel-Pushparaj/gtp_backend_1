package mcp

import (
	mcp_golang "github.com/metoro-io/mcp-golang"
)

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

func (s *Server) registerJiraTools() error {
	if err := s.mcpServer.RegisterTool("get_jira_issue_stats", "Get issue statistics for a Jira project", func(args GetJiraIssueStatsArgs) (*mcp_golang.ToolResponse, error) {
		params := map[string]string{"project": args.Project}
		result, err := s.makeRequest("GET", "/api/v1/jira/issues/stats", params, nil)
		if err != nil {
			return nil, err
		}
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(result)), nil
	}); err != nil {
		return err
	}

	if err := s.mcpServer.RegisterTool("get_jira_open_bugs", "Get open bugs for a Jira project", func(args GetJiraOpenBugsArgs) (*mcp_golang.ToolResponse, error) {
		params := map[string]string{"project": args.Project}
		result, err := s.makeRequest("GET", "/api/v1/jira/bugs/open", params, nil)
		if err != nil {
			return nil, err
		}
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(result)), nil
	}); err != nil {
		return err
	}

	if err := s.mcpServer.RegisterTool("get_jira_open_tasks", "Get open tasks for a Jira project", func(args GetJiraOpenTasksArgs) (*mcp_golang.ToolResponse, error) {
		params := map[string]string{"project": args.Project}
		result, err := s.makeRequest("GET", "/api/v1/jira/tasks/open", params, nil)
		if err != nil {
			return nil, err
		}
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(result)), nil
	}); err != nil {
		return err
	}

	if err := s.mcpServer.RegisterTool("get_jira_issues_by_assignee", "Get issues grouped by assignee", func(args GetJiraIssuesByAssigneeArgs) (*mcp_golang.ToolResponse, error) {
		params := map[string]string{"project": args.Project}
		result, err := s.makeRequest("GET", "/api/v1/jira/issues/by-assignee", params, nil)
		if err != nil {
			return nil, err
		}
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(result)), nil
	}); err != nil {
		return err
	}

	if err := s.mcpServer.RegisterTool("get_jira_sprint_stats", "Get sprint statistics for a Jira project", func(args GetJiraSprintStatsArgs) (*mcp_golang.ToolResponse, error) {
		params := map[string]string{"project": args.Project}
		result, err := s.makeRequest("GET", "/api/v1/jira/sprints/stats", params, nil)
		if err != nil {
			return nil, err
		}
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(result)), nil
	}); err != nil {
		return err
	}

	if err := s.mcpServer.RegisterTool("get_jira_project_metrics", "Get comprehensive project metrics", func(args GetJiraProjectMetricsArgs) (*mcp_golang.ToolResponse, error) {
		params := map[string]string{"project": args.Project}
		result, err := s.makeRequest("GET", "/api/v1/jira/metrics", params, nil)
		if err != nil {
			return nil, err
		}
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(result)), nil
	}); err != nil {
		return err
	}

	if err := s.mcpServer.RegisterTool("search_jira_issues", "Search for Jira issues using JQL", func(args SearchJiraIssuesArgs) (*mcp_golang.ToolResponse, error) {
		params := map[string]string{"jql": args.JQL}
		if args.MaxResults != "" {
			params["max_results"] = args.MaxResults
		}
		result, err := s.makeRequest("GET", "/api/v1/jira/issues/search", params, nil)
		if err != nil {
			return nil, err
		}
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(result)), nil
	}); err != nil {
		return err
	}

	return nil
}

