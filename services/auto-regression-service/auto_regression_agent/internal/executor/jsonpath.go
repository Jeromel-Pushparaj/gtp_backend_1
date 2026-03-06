package executor

import (
	"fmt"
	"strconv"
	"strings"
)

// extractJSONPath extracts a value from a JSON object using a simple path notation
// Supports:
// - Simple keys: "name"
// - Nested keys: "user.name"
// - Array indices: "users[0].name"
func extractJSONPath(data map[string]interface{}, path string) (interface{}, error) {
	if path == "" {
		return nil, fmt.Errorf("empty path")
	}

	// Split path by dots, but handle array indices
	parts := parseJSONPath(path)
	
	var current interface{} = data
	
	for i, part := range parts {
		// Check if this part has an array index
		if strings.Contains(part, "[") {
			// Extract key and index
			key, index, err := parseArrayAccess(part)
			if err != nil {
				return nil, fmt.Errorf("invalid array access at %s: %w", part, err)
			}
			
			// Get the array
			currentMap, ok := current.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("expected object at path segment %d, got %T", i, current)
			}
			
			arrayValue, exists := currentMap[key]
			if !exists {
				return nil, fmt.Errorf("key %s not found", key)
			}
			
			// Access array element
			arraySlice, ok := arrayValue.([]interface{})
			if !ok {
				return nil, fmt.Errorf("expected array at %s, got %T", key, arrayValue)
			}
			
			if index < 0 || index >= len(arraySlice) {
				return nil, fmt.Errorf("array index %d out of bounds for %s (length: %d)", index, key, len(arraySlice))
			}
			
			current = arraySlice[index]
		} else {
			// Simple key access
			currentMap, ok := current.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("expected object at path segment %d, got %T", i, current)
			}
			
			value, exists := currentMap[part]
			if !exists {
				return nil, fmt.Errorf("key %s not found", part)
			}
			
			current = value
		}
	}
	
	return current, nil
}

// parseJSONPath parses a JSON path into parts
// Example: "user.addresses[0].city" -> ["user", "addresses[0]", "city"]
func parseJSONPath(path string) []string {
	parts := make([]string, 0)
	current := ""
	inBracket := false
	
	for _, ch := range path {
		if ch == '.' && !inBracket {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			if ch == '[' {
				inBracket = true
			} else if ch == ']' {
				inBracket = false
			}
			current += string(ch)
		}
	}
	
	if current != "" {
		parts = append(parts, current)
	}
	
	return parts
}

// parseArrayAccess parses array access notation
// Example: "users[0]" -> ("users", 0, nil)
func parseArrayAccess(part string) (string, int, error) {
	openBracket := strings.Index(part, "[")
	closeBracket := strings.Index(part, "]")
	
	if openBracket == -1 || closeBracket == -1 {
		return "", 0, fmt.Errorf("invalid array access syntax")
	}
	
	key := part[:openBracket]
	indexStr := part[openBracket+1 : closeBracket]
	
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		return "", 0, fmt.Errorf("invalid array index: %w", err)
	}
	
	return key, index, nil
}

