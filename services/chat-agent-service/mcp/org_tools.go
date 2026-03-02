package mcp

import (
	mcp_golang "github.com/metoro-io/mcp-golang"
)

type FetchOrgsArgs struct{}

type CreateOrgArgs struct {
	Name         string `json:"name" jsonschema:"required,description=Organization name"`
	GithubPAT    string `json:"github_pat" jsonschema:"required,description=GitHub Personal Access Token"`
	SonarToken   string `json:"sonar_token" jsonschema:"required,description=SonarCloud token"`
	SonarOrgKey  string `json:"sonar_org_key" jsonschema:"required,description=SonarCloud organization key"`
	JiraToken    string `json:"jira_token" jsonschema:"required,description=Jira API token"`
	JiraDomain   string `json:"jira_domain" jsonschema:"required,description=Jira domain (e.g., company.atlassian.net)"`
	JiraEmail    string `json:"jira_email" jsonschema:"required,description=Jira user email"`
}

func (s *Server) registerOrgTools() error {
	if err := s.mcpServer.RegisterTool("fetch_orgs", "Get all organizations", func(args FetchOrgsArgs) (*mcp_golang.ToolResponse, error) {
		result, err := s.makeRequest("GET", "/api/v1/orgs", nil, nil)
		if err != nil {
			return nil, err
		}
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(result)), nil
	}); err != nil {
		return err
	}

	if err := s.mcpServer.RegisterTool("create_org", "Create a new organization", func(args CreateOrgArgs) (*mcp_golang.ToolResponse, error) {
		body := map[string]string{
			"name":           args.Name,
			"github_pat":     args.GithubPAT,
			"sonar_token":    args.SonarToken,
			"sonar_org_key":  args.SonarOrgKey,
			"jira_token":     args.JiraToken,
			"jira_domain":    args.JiraDomain,
			"jira_email":     args.JiraEmail,
		}
		result, err := s.makeRequest("POST", "/api/v1/orgs/create", nil, body)
		if err != nil {
			return nil, err
		}
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(result)), nil
	}); err != nil {
		return err
	}

	return nil
}

