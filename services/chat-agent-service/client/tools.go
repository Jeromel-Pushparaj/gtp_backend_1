package client

func GetAvailableTools() []Tool {
	return []Tool{
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "health_check",
				Description: "Check the health status of the GTP backend service",
				Parameters: map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
				},
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "list_pull_requests",
				Description: "List pull requests for a repository",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"repo": map[string]interface{}{
							"type":        "string",
							"description": "Repository name",
						},
						"state": map[string]interface{}{
							"type":        "string",
							"description": "PR state (open, closed, all)",
						},
					},
					"required": []string{"repo"},
				},
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "list_commits",
				Description: "List commits for a repository",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"repo": map[string]interface{}{
							"type":        "string",
							"description": "Repository name",
						},
						"since": map[string]interface{}{
							"type":        "string",
							"description": "Only commits after this date (ISO 8601 format)",
						},
					},
					"required": []string{"repo"},
				},
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "list_issues",
				Description: "List issues for a repository",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"repo": map[string]interface{}{
							"type":        "string",
							"description": "Repository name",
						},
						"state": map[string]interface{}{
							"type":        "string",
							"description": "Issue state (open, closed, all)",
						},
					},
					"required": []string{"repo"},
				},
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "check_readme",
				Description: "Check if a repository has a README file",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"repo": map[string]interface{}{
							"type":        "string",
							"description": "Repository name",
						},
					},
					"required": []string{"repo"},
				},
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "list_branches",
				Description: "List branches for a repository",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"repo": map[string]interface{}{
							"type":        "string",
							"description": "Repository name",
						},
					},
					"required": []string{"repo"},
				},
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "list_org_members",
				Description: "List organization members",
				Parameters: map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
				},
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "list_org_teams",
				Description: "List organization teams",
				Parameters: map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
				},
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "fetch_orgs",
				Description: "Get all organizations",
				Parameters: map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
				},
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "fetch_repos_by_org",
				Description: "Fetch repositories for an organization from GitHub",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"org_id": map[string]interface{}{
							"type":        "string",
							"description": "Organization ID",
						},
					},
					"required": []string{"org_id"},
				},
			},
		},
	}
}

