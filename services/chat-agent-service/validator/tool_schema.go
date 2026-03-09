package validator

import (
	"fmt"
	"time"
)

// FieldSchema defines expected type and validation for a field
type FieldSchema struct {
	Type     string // "string", "number", "boolean", "date"
	Required bool
	Enum     []string // For state fields like "open", "closed", "all"
}

// ToolSchema defines the expected schema for a tool
type ToolSchema struct {
	Fields map[string]FieldSchema
}

// ToolSchemas maps tool names to their expected argument schemas
var ToolSchemas = map[string]ToolSchema{
	"health_check": {
		Fields: map[string]FieldSchema{},
	},
	"list_pull_requests": {
		Fields: map[string]FieldSchema{
			"repo":  {Type: "string", Required: true},
			"state": {Type: "string", Required: false, Enum: []string{"open", "closed", "all"}},
		},
	},
	"list_commits": {
		Fields: map[string]FieldSchema{
			"repo":  {Type: "string", Required: true},
			"since": {Type: "date", Required: false},
		},
	},
	"list_issues": {
		Fields: map[string]FieldSchema{
			"repo":  {Type: "string", Required: true},
			"state": {Type: "string", Required: false, Enum: []string{"open", "closed", "all"}},
		},
	},
	"check_readme": {
		Fields: map[string]FieldSchema{
			"repo": {Type: "string", Required: true},
		},
	},
	"list_branches": {
		Fields: map[string]FieldSchema{
			"repo": {Type: "string", Required: true},
		},
	},
	"list_org_members": {
		Fields: map[string]FieldSchema{},
	},
	"list_org_teams": {
		Fields: map[string]FieldSchema{},
	},
	"fetch_orgs": {
		Fields: map[string]FieldSchema{},
	},
	"fetch_repos_by_org": {
		Fields: map[string]FieldSchema{
			"org_id": {Type: "string", Required: true},
		},
	},
	"get_pull_request": {
		Fields: map[string]FieldSchema{
			"repo":   {Type: "string", Required: true},
			"number": {Type: "string", Required: true},
		},
	},
	"get_commit_activity": {
		Fields: map[string]FieldSchema{
			"repo": {Type: "string", Required: true},
		},
	},
	"list_issue_comments": {
		Fields: map[string]FieldSchema{
			"repo":   {Type: "string", Required: true},
			"number": {Type: "string", Required: true},
		},
	},
	"get_repository_metrics": {
		Fields: map[string]FieldSchema{
			"repo": {Type: "string", Required: true},
		},
	},
	"get_all_repositories_metrics": {
		Fields: map[string]FieldSchema{},
	},
	// Repository Tools
	"update_repo": {
		Fields: map[string]FieldSchema{
			"repo_id":          {Type: "string", Required: true},
			"jira_project_key": {Type: "string", Required: false},
			"environment_name": {Type: "string", Required: false},
		},
	},
	"fetch_github_metrics_by_repo": {
		Fields: map[string]FieldSchema{
			"repo_id": {Type: "string", Required: true},
		},
	},
	"fetch_jira_metrics_by_repo": {
		Fields: map[string]FieldSchema{
			"repo_id": {Type: "string", Required: true},
		},
	},
	"fetch_sonar_metrics_by_repo": {
		Fields: map[string]FieldSchema{
			"repo_id": {Type: "string", Required: true},
		},
	},
	// Jira Tools
	"get_jira_issue_stats": {
		Fields: map[string]FieldSchema{
			"project": {Type: "string", Required: true},
		},
	},
	"get_jira_open_bugs": {
		Fields: map[string]FieldSchema{
			"project": {Type: "string", Required: true},
		},
	},
	"get_jira_open_tasks": {
		Fields: map[string]FieldSchema{
			"project": {Type: "string", Required: true},
		},
	},
	"get_jira_issues_by_assignee": {
		Fields: map[string]FieldSchema{
			"project": {Type: "string", Required: true},
		},
	},
	"get_jira_sprint_stats": {
		Fields: map[string]FieldSchema{
			"project": {Type: "string", Required: true},
		},
	},
	"get_jira_project_metrics": {
		Fields: map[string]FieldSchema{
			"project": {Type: "string", Required: true},
		},
	},
	"search_jira_issues": {
		Fields: map[string]FieldSchema{
			"jql":         {Type: "string", Required: true},
			"max_results": {Type: "string", Required: false},
		},
	},
	// Metrics Tools
	"collect_github_metrics": {
		Fields: map[string]FieldSchema{
			"repo": {Type: "string", Required: true},
		},
	},
	"get_stored_github_metrics": {
		Fields: map[string]FieldSchema{
			"repo": {Type: "string", Required: true},
		},
	},
	"collect_sonar_metrics": {
		Fields: map[string]FieldSchema{
			"repo": {Type: "string", Required: true},
		},
	},
	"get_stored_sonar_metrics": {
		Fields: map[string]FieldSchema{
			"repo": {Type: "string", Required: true},
		},
	},
	// Organization Management tools
	"create_org": {
		Fields: map[string]FieldSchema{
			"name":          {Type: "string", Required: true},
			"github_pat":    {Type: "string", Required: true},
			"sonar_token":   {Type: "string", Required: true},
			"sonar_org_key": {Type: "string", Required: true},
			"jira_token":    {Type: "string", Required: true},
			"jira_domain":   {Type: "string", Required: true},
			"jira_email":    {Type: "string", Required: true},
		},
	},
	// SonarCloud Tools
	"list_secrets": {
		Fields: map[string]FieldSchema{},
	},
	"add_env_secrets": {
		Fields: map[string]FieldSchema{},
	},
	"update_workflows": {
		Fields: map[string]FieldSchema{},
	},
	"full_setup": {
		Fields: map[string]FieldSchema{},
	},
	"fetch_results": {
		Fields: map[string]FieldSchema{},
	},
	"get_sonar_metrics": {
		Fields: map[string]FieldSchema{
			"repo":           {Type: "string", Required: true},
			"include_issues": {Type: "string", Required: false},
		},
	},
	"process_repository": {
		Fields: map[string]FieldSchema{
			"repository_name": {Type: "string", Required: true},
		},
	},
}

// ValidateToolArguments validates tool arguments against their schema
func ValidateToolArguments(toolName string, args map[string]interface{}) error {
	schema, exists := ToolSchemas[toolName]
	if !exists {
		return fmt.Errorf("unknown tool: %s", toolName)
	}

	// Check all fields in schema
	for fieldName, fieldSchema := range schema.Fields {
		value, ok := args[fieldName]

		// Check required fields
		if !ok && fieldSchema.Required {
			return fmt.Errorf("missing required field: %s", fieldName)
		}

		if !ok {
			continue // Optional field not provided
		}

		// Validate field type
		if err := validateFieldType(fieldName, value, fieldSchema); err != nil {
			return err
		}
	}

	return nil
}

func validateFieldType(fieldName string, value interface{}, schema FieldSchema) error {
	switch schema.Type {
	case "string":
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("field '%s' must be string, got %T", fieldName, value)
		}

		// Empty string validation
		if str == "" && schema.Required {
			return fmt.Errorf("field '%s' cannot be empty", fieldName)
		}

		// Enum validation
		if len(schema.Enum) > 0 {
			valid := false
			for _, allowed := range schema.Enum {
				if str == allowed {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("field '%s' must be one of %v, got '%s'", fieldName, schema.Enum, str)
			}
		}

	case "number":
		// JSON unmarshaling converts numbers to float64
		if _, ok := value.(float64); !ok {
			return fmt.Errorf("field '%s' must be number, got %T", fieldName, value)
		}

	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("field '%s' must be boolean, got %T", fieldName, value)
		}

	case "date":
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("field '%s' must be date string (ISO 8601), got %T", fieldName, value)
		}
		// Validate ISO 8601 format
		if _, err := time.Parse(time.RFC3339, str); err != nil {
			return fmt.Errorf("field '%s' must be valid ISO 8601 date, got '%s'", fieldName, str)
		}
	}

	return nil
}
