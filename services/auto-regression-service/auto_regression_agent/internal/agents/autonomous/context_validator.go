package autonomous

import (
	"fmt"
	"log"
	"regexp"
	"strings"
)

// ContextValidator validates that required context is available before test execution
type ContextValidator struct {
	testContext     *TestContext
	dependencyGraph *DependencyGraph
}

// ValidationResult contains the result of context validation
type ContextValidationResult struct {
	Valid           bool                 // Whether all required context is available
	MissingContext  []MissingContextInfo // List of missing context items
	AvailableFields map[string][]string  // Available fields per resource type
	Suggestions     []string             // Suggestions for fixing missing context
}

// MissingContextInfo describes a missing context item
type MissingContextInfo struct {
	ResourceType string // e.g., "pet"
	Field        string // e.g., "id"
	RequiredBy   string // e.g., "GET /pet/{petId}"
	Marker       string // e.g., "{{CONTEXT:pet.id}}"
}

// NewContextValidator creates a new context validator
func NewContextValidator(testContext *TestContext, dependencyGraph *DependencyGraph) *ContextValidator {
	return &ContextValidator{
		testContext:     testContext,
		dependencyGraph: dependencyGraph,
	}
}

// ValidateForEndpoint validates that all required context is available for an endpoint
func (cv *ContextValidator) ValidateForEndpoint(endpoint string) *ContextValidationResult {
	result := &ContextValidationResult{
		Valid:           true,
		MissingContext:  make([]MissingContextInfo, 0),
		AvailableFields: make(map[string][]string),
		Suggestions:     make([]string, 0),
	}

	// Get requirements from dependency graph
	if cv.dependencyGraph != nil {
		requirements := cv.dependencyGraph.GetRequirementsForEndpoint(endpoint)
		for _, req := range requirements {
			if !cv.hasContext(req.Type, req.Field) {
				result.Valid = false
				result.MissingContext = append(result.MissingContext, MissingContextInfo{
					ResourceType: req.Type,
					Field:        req.Field,
					RequiredBy:   endpoint,
					Marker:       fmt.Sprintf("{{CONTEXT:%s.%s}}", req.Type, req.Field),
				})

				// Add suggestion
				if producer := req.ProducerHint; producer != "" {
					result.Suggestions = append(result.Suggestions,
						fmt.Sprintf("Run '%s' first to create %s", producer, req.Type))
				}
			}
		}
	}

	// Collect available fields
	if cv.testContext != nil {
		for resourceType, data := range cv.testContext.GetAllContext() {
			if dataMap, ok := data.(map[string]interface{}); ok {
				fields := make([]string, 0, len(dataMap))
				for field := range dataMap {
					fields = append(fields, field)
				}
				result.AvailableFields[resourceType] = fields
			}
		}
	}

	return result
}

// ValidatePath validates that all context markers in a path can be resolved
func (cv *ContextValidator) ValidatePath(path string) *ContextValidationResult {
	result := &ContextValidationResult{
		Valid:           true,
		MissingContext:  make([]MissingContextInfo, 0),
		AvailableFields: make(map[string][]string),
		Suggestions:     make([]string, 0),
	}

	// Find all context markers in the path
	markerPattern := regexp.MustCompile(`\{\{CONTEXT:([^.}]+)\.([^}]+)\}\}`)
	matches := markerPattern.FindAllStringSubmatch(path, -1)

	for _, match := range matches {
		if len(match) < 3 {
			continue
		}
		resourceType := match[1]
		field := match[2]

		if !cv.hasContext(resourceType, field) {
			result.Valid = false
			result.MissingContext = append(result.MissingContext, MissingContextInfo{
				ResourceType: resourceType,
				Field:        field,
				RequiredBy:   path,
				Marker:       match[0],
			})

			// Add suggestion based on dependency graph
			if cv.dependencyGraph != nil {
				if producer := cv.dependencyGraph.GetProducerForResource(resourceType); producer != "" {
					result.Suggestions = append(result.Suggestions,
						fmt.Sprintf("Run '%s' first to create %s", producer, resourceType))
				}
			}
		}
	}

	return result
}

// hasContext checks if a specific context field is available
func (cv *ContextValidator) hasContext(resourceType, field string) bool {
	if cv.testContext == nil {
		return false
	}

	data := cv.testContext.GetContext(resourceType)
	if data == nil {
		return false
	}

	_, exists := data[field]
	return exists
}

// LogValidationResult logs the validation result for debugging
func (cv *ContextValidator) LogValidationResult(result *ContextValidationResult, endpoint string) {
	if result.Valid {
		log.Printf("✅ Context validation passed for %s", endpoint)
		return
	}

	log.Printf("⚠️ Context validation failed for %s:", endpoint)
	for _, missing := range result.MissingContext {
		log.Printf("   - Missing: %s.%s (marker: %s)", missing.ResourceType, missing.Field, missing.Marker)
	}
	if len(result.Suggestions) > 0 {
		log.Printf("   Suggestions:")
		for _, suggestion := range result.Suggestions {
			log.Printf("   - %s", suggestion)
		}
	}
	if len(result.AvailableFields) > 0 {
		log.Printf("   Available context:")
		for resourceType, fields := range result.AvailableFields {
			log.Printf("   - %s: [%s]", resourceType, strings.Join(fields, ", "))
		}
	}
}
