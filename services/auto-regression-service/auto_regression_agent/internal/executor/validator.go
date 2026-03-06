package executor

import (
	"fmt"
	"reflect"

	"github.com/xeipuuv/gojsonschema"
	"gitlab.com/tekion/development/toc/poc/opentest/pkg/domain/execution"
)

// Validator validates HTTP responses against assertions
type Validator struct{}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{}
}

// Validate validates all assertions against the HTTP response
func (v *Validator) Validate(assertions execution.Assertions, response *HTTPResponse) []execution.AssertionResult {
	results := make([]execution.AssertionResult, 0)

	// Validate status code
	results = append(results, v.validateStatusCode(assertions.StatusCode, response.StatusCode))

	// Validate response time
	if assertions.ResponseTime > 0 {
		results = append(results, v.validateResponseTime(assertions.ResponseTime, response.ResponseTime))
	}

	// Validate headers
	if len(assertions.Headers) > 0 {
		results = append(results, v.validateHeaders(assertions.Headers, response.Headers)...)
	}

	// Validate JSON schema
	if assertions.JSONSchema != nil {
		results = append(results, v.validateJSONSchema(assertions.JSONSchema, response.Body))
	}

	// Validate JSON path assertions
	if len(assertions.JSONPath) > 0 {
		results = append(results, v.validateJSONPath(assertions.JSONPath, response.Body)...)
	}

	// Validate contains fields
	if len(assertions.ContainsFields) > 0 {
		results = append(results, v.validateContainsFields(assertions.ContainsFields, response.Body)...)
	}

	return results
}

// validateStatusCode validates the HTTP status code with flexible matching.
// For 2xx success codes: any 2xx is acceptable (200, 201, 204 are all valid success responses)
// For 4xx/5xx error codes: exact match required (400 vs 404 vs 422 have different meanings)
func (v *Validator) validateStatusCode(expected int, actual int) execution.AssertionResult {
	passed := isStatusCodeMatch(expected, actual)
	message := ""
	if !passed {
		message = fmt.Sprintf("Expected status code %d, got %d", expected, actual)
	}

	return execution.AssertionResult{
		Type:     "status_code",
		Expected: expected,
		Actual:   actual,
		Passed:   passed,
		Message:  message,
	}
}

// isStatusCodeMatch checks if actual status code matches expected with flexible matching
func isStatusCodeMatch(expected, actual int) bool {
	// If expected is 0, any 2xx is considered success
	if expected == 0 {
		return actual >= 200 && actual < 300
	}

	// For 2xx success codes, accept any 2xx response
	// This handles cases like expecting 200 but getting 201 Created
	if expected >= 200 && expected < 300 {
		return actual >= 200 && actual < 300
	}

	// For error codes (4xx, 5xx), require exact match
	// Different error codes have different meanings
	return actual == expected
}

// validateResponseTime validates the response time
func (v *Validator) validateResponseTime(maxTime int, actualTime int64) execution.AssertionResult {
	passed := actualTime <= int64(maxTime)
	message := ""
	if !passed {
		message = fmt.Sprintf("Response time %dms exceeded maximum %dms", actualTime, maxTime)
	}

	return execution.AssertionResult{
		Type:     "response_time",
		Expected: maxTime,
		Actual:   actualTime,
		Passed:   passed,
		Message:  message,
	}
}

// validateHeaders validates response headers
func (v *Validator) validateHeaders(expected map[string]string, actual map[string]string) []execution.AssertionResult {
	results := make([]execution.AssertionResult, 0)

	for key, expectedValue := range expected {
		actualValue, exists := actual[key]
		passed := exists && actualValue == expectedValue
		message := ""

		if !exists {
			message = fmt.Sprintf("Header %s not found in response", key)
		} else if actualValue != expectedValue {
			message = fmt.Sprintf("Header %s: expected %s, got %s", key, expectedValue, actualValue)
		}

		results = append(results, execution.AssertionResult{
			Type:     "header",
			Expected: map[string]string{key: expectedValue},
			Actual:   map[string]string{key: actualValue},
			Passed:   passed,
			Message:  message,
		})
	}

	return results
}

// validateJSONSchema validates response body against JSON schema
func (v *Validator) validateJSONSchema(schema map[string]interface{}, body map[string]interface{}) execution.AssertionResult {
	schemaLoader := gojsonschema.NewGoLoader(schema)
	documentLoader := gojsonschema.NewGoLoader(body)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return execution.AssertionResult{
			Type:     "json_schema",
			Expected: schema,
			Actual:   body,
			Passed:   false,
			Message:  fmt.Sprintf("Schema validation error: %v", err),
		}
	}

	passed := result.Valid()
	message := ""
	if !passed {
		errors := make([]string, 0)
		for _, err := range result.Errors() {
			errors = append(errors, err.String())
		}
		message = fmt.Sprintf("Schema validation failed: %v", errors)
	}

	return execution.AssertionResult{
		Type:     "json_schema",
		Expected: schema,
		Actual:   body,
		Passed:   passed,
		Message:  message,
	}
}

// validateJSONPath validates JSON path assertions
func (v *Validator) validateJSONPath(assertions map[string]interface{}, body map[string]interface{}) []execution.AssertionResult {
	results := make([]execution.AssertionResult, 0)

	for path, expectedValue := range assertions {
		actualValue, err := extractJSONPath(body, path)
		if err != nil {
			results = append(results, execution.AssertionResult{
				Type:     "json_path",
				Expected: expectedValue,
				Actual:   nil,
				Passed:   false,
				Message:  fmt.Sprintf("Failed to extract path %s: %v", path, err),
			})
			continue
		}

		passed := reflect.DeepEqual(expectedValue, actualValue)
		message := ""
		if !passed {
			message = fmt.Sprintf("Path %s: expected %v, got %v", path, expectedValue, actualValue)
		}

		results = append(results, execution.AssertionResult{
			Type:     "json_path",
			Expected: expectedValue,
			Actual:   actualValue,
			Passed:   passed,
			Message:  message,
		})
	}

	return results
}

// validateContainsFields validates that response contains specific fields
func (v *Validator) validateContainsFields(fields []string, body map[string]interface{}) []execution.AssertionResult {
	results := make([]execution.AssertionResult, 0)

	for _, field := range fields {
		_, exists := body[field]
		message := ""
		if !exists {
			message = fmt.Sprintf("Field %s not found in response", field)
		}

		results = append(results, execution.AssertionResult{
			Type:     "contains_field",
			Expected: field,
			Actual:   exists,
			Passed:   exists,
			Message:  message,
		})
	}

	return results
}
