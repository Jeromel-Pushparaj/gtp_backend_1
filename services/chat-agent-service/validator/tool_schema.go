package validator

import (
	"fmt"
	"time"
)

// FieldSchema defines expected type and validation for a field
type FieldSchema struct {
	Type     string   // "string", "number", "boolean", "date"
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