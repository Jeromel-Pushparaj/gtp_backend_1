package executor

import (
	"fmt"
	"regexp"
	"sync"

	"gitlab.com/tekion/development/toc/poc/opentest/pkg/domain/execution"
)

// VariableResolver resolves variables in test cases
type VariableResolver struct {
	variables map[string]interface{}
	mu        sync.RWMutex
}

// NewVariableResolver creates a new variable resolver
func NewVariableResolver() *VariableResolver {
	return &VariableResolver{
		variables: make(map[string]interface{}),
	}
}

// SetVariables sets multiple variables
func (vr *VariableResolver) SetVariables(vars map[string]interface{}) {
	vr.mu.Lock()
	defer vr.mu.Unlock()

	for key, value := range vars {
		vr.variables[key] = value
	}
}

// SetVariable sets a single variable
func (vr *VariableResolver) SetVariable(key string, value interface{}) {
	vr.mu.Lock()
	defer vr.mu.Unlock()

	vr.variables[key] = value
}

// GetVariable gets a variable value
func (vr *VariableResolver) GetVariable(key string) (interface{}, bool) {
	vr.mu.RLock()
	defer vr.mu.RUnlock()

	value, exists := vr.variables[key]
	return value, exists
}

// ResolveTest resolves variables in a test case
func (vr *VariableResolver) ResolveTest(test execution.TestCase) (execution.TestCase, error) {
	vr.mu.RLock()
	defer vr.mu.RUnlock()

	resolved := test

	// Resolve endpoint
	endpoint, err := vr.resolveString(test.Endpoint)
	if err != nil {
		return resolved, fmt.Errorf("failed to resolve endpoint: %w", err)
	}
	resolved.Endpoint = endpoint

	// Resolve headers
	if test.Headers != nil {
		resolvedHeaders := make(map[string]string)
		for key, value := range test.Headers {
			resolvedValue, err := vr.resolveString(value)
			if err != nil {
				return resolved, fmt.Errorf("failed to resolve header %s: %w", key, err)
			}
			resolvedHeaders[key] = resolvedValue
		}
		resolved.Headers = resolvedHeaders
	}

	// Resolve path parameters
	if test.PathParams != nil {
		resolvedParams := make(map[string]string)
		for key, value := range test.PathParams {
			resolvedValue, err := vr.resolveString(value)
			if err != nil {
				return resolved, fmt.Errorf("failed to resolve path param %s: %w", key, err)
			}
			resolvedParams[key] = resolvedValue
		}
		resolved.PathParams = resolvedParams
	}

	// Resolve query parameters
	if test.QueryParams != nil {
		resolvedParams := make(map[string]string)
		for key, value := range test.QueryParams {
			resolvedValue, err := vr.resolveString(value)
			if err != nil {
				return resolved, fmt.Errorf("failed to resolve query param %s: %w", key, err)
			}
			resolvedParams[key] = resolvedValue
		}
		resolved.QueryParams = resolvedParams
	}

	// Resolve payload
	if test.Payload != nil {
		resolvedPayload, err := vr.resolveMap(test.Payload)
		if err != nil {
			return resolved, fmt.Errorf("failed to resolve payload: %w", err)
		}
		resolved.Payload = resolvedPayload
	}

	return resolved, nil
}

// resolveString resolves variables in a string
func (vr *VariableResolver) resolveString(str string) (string, error) {
	// Pattern: ${variable_name}
	re := regexp.MustCompile(`\$\{([^}]+)\}`)

	result := re.ReplaceAllStringFunc(str, func(match string) string {
		// Extract variable name
		varName := match[2 : len(match)-1] // Remove ${ and }

		value, exists := vr.variables[varName]
		if !exists {
			return match // Keep original if variable not found
		}

		// Convert value to string
		return fmt.Sprintf("%v", value)
	})

	return result, nil
}

// resolveMap resolves variables in a map
func (vr *VariableResolver) resolveMap(m map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for key, value := range m {
		resolvedValue, err := vr.resolveValue(value)
		if err != nil {
			return nil, err
		}
		result[key] = resolvedValue
	}

	return result, nil
}

// resolveValue resolves variables in any value type
func (vr *VariableResolver) resolveValue(value interface{}) (interface{}, error) {
	switch v := value.(type) {
	case string:
		return vr.resolveString(v)
	case map[string]interface{}:
		return vr.resolveMap(v)
	case []interface{}:
		return vr.resolveSlice(v)
	default:
		return value, nil
	}
}

// resolveSlice resolves variables in a slice
func (vr *VariableResolver) resolveSlice(slice []interface{}) ([]interface{}, error) {
	result := make([]interface{}, len(slice))

	for i, value := range slice {
		resolvedValue, err := vr.resolveValue(value)
		if err != nil {
			return nil, err
		}
		result[i] = resolvedValue
	}

	return result, nil
}
