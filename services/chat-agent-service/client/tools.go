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
				Name:        "get_pull_request",
				Description: "Get details of a specific pull request",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"repo": map[string]interface{}{
							"type":        "string",
							"description": "Repository name",
						},
						"number": map[string]interface{}{
							"type":        "string",
							"description": "Pull request number",
						},
					},
					"required": []string{"repo", "number"},
				},
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "get_commit_activity",
				Description: "Get commit activity analysis for a repository",
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
				Name:        "list_issue_comments",
				Description: "List comments for a specific issue",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"repo": map[string]interface{}{
							"type":        "string",
							"description": "Repository name",
						},
						"number": map[string]interface{}{
							"type":        "string",
							"description": "Issue number",
						},
					},
					"required": []string{"repo", "number"},
				},
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "get_repository_metrics",
				Description: "Get comprehensive metrics for a specific repository",
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
				Name:        "get_all_repositories_metrics",
				Description: "Get metrics for all repositories in the organization",
				Parameters: map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
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
				Name:        "create_org",
				Description: "Create a new organization with GitHub, SonarCloud, and Jira credentials",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name": map[string]interface{}{
							"type":        "string",
							"description": "Organization name",
						},
						"github_pat": map[string]interface{}{
							"type":        "string",
							"description": "GitHub Personal Access Token",
						},
						"sonar_token": map[string]interface{}{
							"type":        "string",
							"description": "SonarCloud token",
						},
						"sonar_org_key": map[string]interface{}{
							"type":        "string",
							"description": "SonarCloud organization key",
						},
						"jira_token": map[string]interface{}{
							"type":        "string",
							"description": "Jira API token",
						},
						"jira_domain": map[string]interface{}{
							"type":        "string",
							"description": "Jira domain (e.g., company.atlassian.net)",
						},
						"jira_email": map[string]interface{}{
							"type":        "string",
							"description": "Jira user email",
						},
					},
					"required": []string{"name", "github_pat", "sonar_token", "sonar_org_key", "jira_token", "jira_domain", "jira_email"},
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
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "update_repo",
				Description: "Update repository configuration (Jira project key, environment name)",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"repo_id": map[string]interface{}{
							"type":        "string",
							"description": "Repository ID",
						},
						"jira_project_key": map[string]interface{}{
							"type":        "string",
							"description": "Jira project key (optional)",
						},
						"environment_name": map[string]interface{}{
							"type":        "string",
							"description": "Environment name (optional)",
						},
					},
					"required": []string{"repo_id"},
				},
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "fetch_github_metrics_by_repo",
				Description: "Fetch and store GitHub metrics for a repository",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"repo_id": map[string]interface{}{
							"type":        "string",
							"description": "Repository ID",
						},
					},
					"required": []string{"repo_id"},
				},
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "fetch_jira_metrics_by_repo",
				Description: "Fetch and store Jira metrics for a repository",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"repo_id": map[string]interface{}{
							"type":        "string",
							"description": "Repository ID",
						},
					},
					"required": []string{"repo_id"},
				},
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "fetch_sonar_metrics_by_repo",
				Description: "Fetch and store SonarCloud metrics for a repository",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"repo_id": map[string]interface{}{
							"type":        "string",
							"description": "Repository ID",
						},
					},
					"required": []string{"repo_id"},
				},
			},
		},
		// Jira Tools
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "get_jira_issue_stats",
				Description: "Get issue statistics for a Jira project",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"project": map[string]interface{}{
							"type":        "string",
							"description": "Jira project key (e.g., SCRUM)",
						},
					},
					"required": []string{"project"},
				},
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "get_jira_open_bugs",
				Description: "Get open bugs for a Jira project",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"project": map[string]interface{}{
							"type":        "string",
							"description": "Jira project key",
						},
					},
					"required": []string{"project"},
				},
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "get_jira_open_tasks",
				Description: "Get open tasks for a Jira project",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"project": map[string]interface{}{
							"type":        "string",
							"description": "Jira project key",
						},
					},
					"required": []string{"project"},
				},
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "get_jira_issues_by_assignee",
				Description: "Get issues grouped by assignee for a Jira project",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"project": map[string]interface{}{
							"type":        "string",
							"description": "Jira project key",
						},
					},
					"required": []string{"project"},
				},
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "get_jira_sprint_stats",
				Description: "Get sprint statistics for a Jira project",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"project": map[string]interface{}{
							"type":        "string",
							"description": "Jira project key",
						},
					},
					"required": []string{"project"},
				},
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "get_jira_project_metrics",
				Description: "Get comprehensive project metrics for a Jira project",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"project": map[string]interface{}{
							"type":        "string",
							"description": "Jira project key",
						},
					},
					"required": []string{"project"},
				},
			},
		},

		{
			Type: "function",
			Function: ToolFunction{
				Name:        "search_jira_issues",
				Description: "Search for Jira issues using JQL query",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"jql": map[string]interface{}{
							"type":        "string",
							"description": "JQL query string",
						},
						"max_results": map[string]interface{}{
							"type":        "string",
							"description": "Maximum number of results (default: 50)",
						},
					},
					"required": []string{"jql"},
				},
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "collect_github_metrics",
				Description: "Collect and store GitHub metrics for a repository",
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
				Name:        "get_stored_github_metrics",
				Description: "Get stored GitHub metrics for a repository",
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
				Name:        "collect_sonar_metrics",
				Description: "Collect and store SonarCloud metrics for a repository",
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
				Name:        "get_stored_sonar_metrics",
				Description: "Get stored SonarCloud metrics for a repository",
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
		// SonarCloud Tools
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "list_secrets",
				Description: "List all secrets for repositories in the organization",
				Parameters: map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
				},
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "add_env_secrets",
				Description: "Add environment secrets to all repositories",
				Parameters: map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
				},
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "update_workflows",
				Description: "Update GitHub workflows to use environment secrets",
				Parameters: map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
				},
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "full_setup",
				Description: "Perform full SonarCloud setup for all repositories",
				Parameters: map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
				},
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "fetch_results",
				Description: "Fetch SonarCloud analysis results for all repositories",
				Parameters: map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
				},
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "get_sonar_metrics",
				Description: "Get SonarCloud metrics for a specific repository",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"repo": map[string]interface{}{
							"type":        "string",
							"description": "Repository name",
						},
						"include_issues": map[string]interface{}{
							"type":        "string",
							"description": "Set to 'true' to include issue details (optional)",
						},
					},
					"required": []string{"repo"},
				},
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "process_repository",
				Description: "Process a single repository for SonarCloud setup",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"repository_name": map[string]interface{}{
							"type":        "string",
							"description": "Repository name to process",
						},
					},
					"required": []string{"repository_name"},
				},
			},
		},
	}
}
