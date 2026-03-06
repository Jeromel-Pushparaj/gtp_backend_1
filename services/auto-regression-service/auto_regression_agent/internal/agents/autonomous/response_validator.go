package autonomous

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

// ValidationResult contains the results of response validation
type ValidationResult struct {
	Valid    bool              `json:"valid"`
	Errors   []ValidationError `json:"errors,omitempty"`
	Warnings []string          `json:"warnings,omitempty"`
}

// ValidationError represents a single validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Type    string `json:"type"` // required, type_mismatch, pattern, range, enum
}

// ResponseValidator validates API responses against schemas and expectations
type ResponseValidator struct {
	schemas map[string]SchemaDefinition
}

// NewResponseValidator creates a new response validator
func NewResponseValidator(schemas map[string]SchemaDefinition) *ResponseValidator {
	if schemas == nil {
		schemas = make(map[string]SchemaDefinition)
	}
	return &ResponseValidator{schemas: schemas}
}

// ValidateResponse validates a response against expected fields and schema
func (rv *ResponseValidator) ValidateResponse(
	responseBody []byte,
	statusCode int,
	test *PersistedTest,
) *ValidationResult {
	result := &ValidationResult{Valid: true}

	// Parse response body
	var respData interface{}
	if len(responseBody) > 0 {
		if err := json.Unmarshal(responseBody, &respData); err != nil {
			// Not JSON - only an error if we expected JSON validation
			if test.ResponseValidation != nil && test.ResponseValidation.ValidateSchema {
				result.Valid = false
				result.Errors = append(result.Errors, ValidationError{
					Field:   "body",
					Message: "Response is not valid JSON",
					Type:    "parse_error",
				})
			}
			return result
		}
	}

	// Validate expected fields if specified
	if len(test.ExpectedFields) > 0 {
		rv.validateExpectedFields(respData, test.ExpectedFields, result)
	}

	// Validate against schema if specified
	if test.ExpectedSchema != "" && test.ResponseValidation != nil && test.ResponseValidation.ValidateSchema {
		if schema, ok := rv.schemas[test.ExpectedSchema]; ok {
			rv.validateAgainstSchema(respData, &schema, "", test.ResponseValidation, result)
		}
	}

	return result
}

// validateExpectedFields validates specific field expectations
func (rv *ResponseValidator) validateExpectedFields(
	data interface{},
	fields []ExpectedField,
	result *ValidationResult,
) {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "",
			Message: "Response is not an object",
			Type:    "type_mismatch",
		})
		return
	}

	for _, field := range fields {
		value := rv.getValueByPath(dataMap, field.Path)

		// Check required
		if field.Required && value == nil {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   field.Path,
				Message: fmt.Sprintf("Required field '%s' is missing", field.Path),
				Type:    "required",
			})
			continue
		}

		if value == nil {
			continue // Field not present and not required
		}

		// Check type
		if field.Type != "" {
			if !rv.checkType(value, field.Type) {
				result.Valid = false
				result.Errors = append(result.Errors, ValidationError{
					Field:   field.Path,
					Message: fmt.Sprintf("Field '%s' expected type '%s', got '%s'", field.Path, field.Type, reflect.TypeOf(value).String()),
					Type:    "type_mismatch",
				})
			}
		}

		// Check exact value match
		if field.Value != nil {
			if !reflect.DeepEqual(value, field.Value) {
				result.Valid = false
				result.Errors = append(result.Errors, ValidationError{
					Field:   field.Path,
					Message: fmt.Sprintf("Field '%s' expected value '%v', got '%v'", field.Path, field.Value, value),
					Type:    "value_mismatch",
				})
			}
		}

		// Check pattern for strings
		if field.Pattern != "" {
			if strVal, ok := value.(string); ok {
				if matched, _ := regexp.MatchString(field.Pattern, strVal); !matched {
					result.Valid = false
					result.Errors = append(result.Errors, ValidationError{
						Field:   field.Path,
						Message: fmt.Sprintf("Field '%s' does not match pattern '%s'", field.Path, field.Pattern),
						Type:    "pattern",
					})
				}
			}
		}

		// Check string length
		if strVal, ok := value.(string); ok {
			rv.validateStringLength(field.Path, strVal, field.MinLen, field.MaxLen, result)
		}

		// Check numeric range
		if numVal, ok := rv.toFloat64(value); ok {
			rv.validateNumericRange(field.Path, numVal, field.Min, field.Max, result)
		}
	}
}

// validateAgainstSchema validates data against an OpenAPI schema
func (rv *ResponseValidator) validateAgainstSchema(
	data interface{},
	schema *SchemaDefinition,
	path string,
	validation *ResponseValidation,
	result *ValidationResult,
) {
	if data == nil {
		return
	}

	switch schema.Type {
	case "object":
		dataMap, ok := data.(map[string]interface{})
		if !ok {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   path,
				Message: fmt.Sprintf("Expected object at '%s', got %T", path, data),
				Type:    "type_mismatch",
			})
			return
		}

		// Check required fields
		if validation.ValidateRequired {
			for _, reqField := range schema.Required {
				fieldPath := rv.joinPath(path, reqField)
				if _, exists := dataMap[reqField]; !exists {
					result.Valid = false
					result.Errors = append(result.Errors, ValidationError{
						Field:   fieldPath,
						Message: fmt.Sprintf("Required field '%s' is missing", fieldPath),
						Type:    "required",
					})
				}
			}
		}

		// Validate each property
		if validation.ValidateTypes {
			for propName, propSchema := range schema.Properties {
				if propValue, exists := dataMap[propName]; exists {
					fieldPath := rv.joinPath(path, propName)
					rv.validateProperty(propValue, &propSchema, fieldPath, result)
				}
			}
		}

		// Check for extra fields
		if !validation.AllowExtraFields {
			for key := range dataMap {
				if _, defined := schema.Properties[key]; !defined {
					result.Warnings = append(result.Warnings,
						fmt.Sprintf("Extra field '%s' not defined in schema", rv.joinPath(path, key)))
				}
			}
		}

	case "array":
		dataArr, ok := data.([]interface{})
		if !ok {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   path,
				Message: fmt.Sprintf("Expected array at '%s', got %T", path, data),
				Type:    "type_mismatch",
			})
			return
		}

		// Validate array items
		if schema.Items != nil {
			for i, item := range dataArr {
				itemPath := fmt.Sprintf("%s[%d]", path, i)
				rv.validateAgainstSchema(item, schema.Items, itemPath, validation, result)
			}
		}
	}
}

// validateProperty validates a single property against its schema
func (rv *ResponseValidator) validateProperty(value interface{}, prop *PropertyDefinition, path string, result *ValidationResult) {
	if value == nil {
		return
	}

	// Type validation
	if !rv.checkSchemaType(value, prop.Type) {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   path,
			Message: fmt.Sprintf("Field '%s' expected type '%s', got '%T'", path, prop.Type, value),
			Type:    "type_mismatch",
		})
		return
	}

	// String validations
	if strVal, ok := value.(string); ok {
		// Pattern validation
		if prop.Pattern != "" {
			if matched, _ := regexp.MatchString(prop.Pattern, strVal); !matched {
				result.Valid = false
				result.Errors = append(result.Errors, ValidationError{
					Field:   path,
					Message: fmt.Sprintf("Field '%s' does not match pattern '%s'", path, prop.Pattern),
					Type:    "pattern",
				})
			}
		}

		// Length validation
		rv.validateStringLength(path, strVal, prop.MinLength, prop.MaxLength, result)

		// Enum validation
		if len(prop.Enum) > 0 {
			found := false
			for _, enumVal := range prop.Enum {
				if strVal == enumVal {
					found = true
					break
				}
			}
			if !found {
				result.Valid = false
				result.Errors = append(result.Errors, ValidationError{
					Field:   path,
					Message: fmt.Sprintf("Field '%s' value '%s' not in enum %v", path, strVal, prop.Enum),
					Type:    "enum",
				})
			}
		}
	}

	// Numeric validations
	if numVal, ok := rv.toFloat64(value); ok {
		rv.validateNumericRange(path, numVal, prop.Minimum, prop.Maximum, result)
	}
}

// Helper methods

// getValueByPath gets a value from a nested map using dot notation
func (rv *ResponseValidator) getValueByPath(data map[string]interface{}, path string) interface{} {
	parts := strings.Split(path, ".")
	current := interface{}(data)

	for _, part := range parts {
		if currentMap, ok := current.(map[string]interface{}); ok {
			current = currentMap[part]
		} else {
			return nil
		}
	}
	return current
}

// joinPath joins path segments with a dot
func (rv *ResponseValidator) joinPath(base, field string) string {
	if base == "" {
		return field
	}
	return base + "." + field
}

// checkType checks if a value matches the expected type string
func (rv *ResponseValidator) checkType(value interface{}, expectedType string) bool {
	switch expectedType {
	case "string":
		_, ok := value.(string)
		return ok
	case "number", "integer":
		_, ok := rv.toFloat64(value)
		return ok
	case "boolean":
		_, ok := value.(bool)
		return ok
	case "array":
		_, ok := value.([]interface{})
		return ok
	case "object":
		_, ok := value.(map[string]interface{})
		return ok
	default:
		return true
	}
}

// checkSchemaType checks if a value matches the OpenAPI schema type
func (rv *ResponseValidator) checkSchemaType(value interface{}, schemaType string) bool {
	return rv.checkType(value, schemaType)
}

// toFloat64 converts a numeric value to float64
func (rv *ResponseValidator) toFloat64(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case int32:
		return float64(v), true
	case json.Number:
		f, err := v.Float64()
		return f, err == nil
	default:
		return 0, false
	}
}

// validateStringLength validates string length constraints
func (rv *ResponseValidator) validateStringLength(path, value string, minLen, maxLen *int, result *ValidationResult) {
	length := len(value)
	if minLen != nil && length < *minLen {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   path,
			Message: fmt.Sprintf("Field '%s' length %d is less than minimum %d", path, length, *minLen),
			Type:    "range",
		})
	}
	if maxLen != nil && length > *maxLen {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   path,
			Message: fmt.Sprintf("Field '%s' length %d exceeds maximum %d", path, length, *maxLen),
			Type:    "range",
		})
	}
}

// validateNumericRange validates numeric range constraints
func (rv *ResponseValidator) validateNumericRange(path string, value float64, min, max *float64, result *ValidationResult) {
	if min != nil && value < *min {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   path,
			Message: fmt.Sprintf("Field '%s' value %v is less than minimum %v", path, value, *min),
			Type:    "range",
		})
	}
	if max != nil && value > *max {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   path,
			Message: fmt.Sprintf("Field '%s' value %v exceeds maximum %v", path, value, *max),
			Type:    "range",
		})
	}
}
