package mcp

// Organization tool argument types for documentation purposes.
// The HTTP-based MCP server uses dynamic tool mapping via mapOrgToolToEndpoint()
// instead of explicit tool registration.

type FetchOrgsArgs struct{}

type CreateOrgArgs struct {
	Name        string `json:"name" jsonschema:"required,description=Organization name"`
	GithubPAT   string `json:"github_pat" jsonschema:"required,description=GitHub Personal Access Token"`
	SonarToken  string `json:"sonar_token" jsonschema:"required,description=SonarCloud token"`
	SonarOrgKey string `json:"sonar_org_key" jsonschema:"required,description=SonarCloud organization key"`
	JiraToken   string `json:"jira_token" jsonschema:"required,description=Jira API token"`
	JiraDomain  string `json:"jira_domain" jsonschema:"required,description=Jira domain (e.g., company.atlassian.net)"`
	JiraEmail   string `json:"jira_email" jsonschema:"required,description=Jira user email"`
}

// mapOrgToolToEndpoint maps Organization tool names to API endpoints
func mapOrgToolToEndpoint(toolName string, args map[string]interface{}) (endpoint, method string, params map[string]string, body interface{}, found bool) {
	params = make(map[string]string)

	switch toolName {
	case "fetch_orgs":
		return "/api/v1/orgs", "GET", nil, nil, true

	case "create_org":
		return "/api/v1/orgs/create", "POST", nil, args, true

	default:
		return "", "", nil, nil, false
	}
}