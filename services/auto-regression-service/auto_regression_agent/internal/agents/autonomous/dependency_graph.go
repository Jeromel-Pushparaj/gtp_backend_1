package autonomous

import (
	"fmt"
	"log"
	"regexp"
	"sort"
	"strings"
)

// DependencyGraph tracks resource dependencies between API endpoints
// It builds a graph where:
// - Producers: Endpoints that create resources (typically POST)
// - Consumers: Endpoints that require resources (GET/{id}, PUT/{id}, DELETE/{id})
type DependencyGraph struct {
	// Producers: endpoint -> list of resources it creates
	producers map[string][]ResourceInfo
	// Consumers: endpoint -> list of resources it requires
	consumers map[string][]ResourceRequirement
	// ResourceTypes: resource type -> producer endpoint
	resourceProducers map[string]string
	// Execution order (topologically sorted)
	executionOrder []string
}

// ResourceInfo describes a resource created by an endpoint
type ResourceInfo struct {
	Type       string   // e.g., "pet", "user", "order"
	IDField    string   // e.g., "id", "petId"
	Fields     []string // All fields available (e.g., ["id", "name", "status"])
	PathPrefix string   // e.g., "/pet" for resources at /pet
}

// ResourceRequirement describes a resource required by an endpoint
type ResourceRequirement struct {
	Type         string // e.g., "pet", "user"
	Field        string // e.g., "id", "username"
	ParamName    string // Path param name (e.g., "petId")
	ParamType    string // "path", "query", "body"
	IsOptional   bool   // Whether the requirement is optional
	ProducerHint string // Hint about which endpoint produces this
}

// NewDependencyGraph creates a new dependency graph
func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{
		producers:         make(map[string][]ResourceInfo),
		consumers:         make(map[string][]ResourceRequirement),
		resourceProducers: make(map[string]string),
		executionOrder:    make([]string, 0),
	}
}

// BuildFromSpec builds the dependency graph from an OpenAPI spec
func (dg *DependencyGraph) BuildFromSpec(spec map[string]interface{}) error {
	paths, ok := spec["paths"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("no paths found in spec")
	}

	// First pass: identify all producers (POST endpoints that create resources)
	for path, pathItem := range paths {
		pathData, ok := pathItem.(map[string]interface{})
		if !ok {
			continue
		}

		for method, opData := range pathData {
			methodUpper := strings.ToUpper(method)
			if !isHTTPMethod(method) {
				continue
			}

			endpoint := methodUpper + " " + path
			operation, ok := opData.(map[string]interface{})
			if !ok {
				continue
			}

			// Identify producers (POST endpoints without path params are typically creators)
			if methodUpper == "POST" && !strings.Contains(path, "{") {
				resourceType := inferResourceTypeFromPath(path)
				if resourceType != "" {
					dg.registerProducer(endpoint, ResourceInfo{
						Type:       resourceType,
						IDField:    "id",
						Fields:     dg.extractResponseFields(operation),
						PathPrefix: path,
					})
				}
			}

			// Identify consumers (endpoints with path parameters)
			if strings.Contains(path, "{") {
				requirements := dg.extractRequirements(path, operation)
				if len(requirements) > 0 {
					dg.registerConsumer(endpoint, requirements)
				}
			}
		}
	}

	// Build execution order
	dg.buildExecutionOrder()

	log.Printf("📊 Dependency graph built: %d producers, %d consumers",
		len(dg.producers), len(dg.consumers))

	return nil
}

// registerProducer registers an endpoint as a resource producer
func (dg *DependencyGraph) registerProducer(endpoint string, info ResourceInfo) {
	dg.producers[endpoint] = append(dg.producers[endpoint], info)
	dg.resourceProducers[info.Type] = endpoint
	log.Printf("📦 Registered producer: %s creates '%s'", endpoint, info.Type)
}

// registerConsumer registers an endpoint as a resource consumer
func (dg *DependencyGraph) registerConsumer(endpoint string, requirements []ResourceRequirement) {
	dg.consumers[endpoint] = append(dg.consumers[endpoint], requirements...)
	for _, req := range requirements {
		log.Printf("🔗 Registered consumer: %s requires '%s.%s'", endpoint, req.Type, req.Field)
	}
}

// extractRequirements extracts resource requirements from path parameters
func (dg *DependencyGraph) extractRequirements(path string, operation map[string]interface{}) []ResourceRequirement {
	requirements := make([]ResourceRequirement, 0)

	// Extract path parameters
	paramRegex := regexp.MustCompile(`\{([^}]+)\}`)
	matches := paramRegex.FindAllStringSubmatch(path, -1)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		paramName := match[1]
		resourceType, fieldName := dg.inferResourceFromParam(paramName, path)

		requirements = append(requirements, ResourceRequirement{
			Type:         resourceType,
			Field:        fieldName,
			ParamName:    paramName,
			ParamType:    "path",
			ProducerHint: dg.findProducerForResource(resourceType),
		})
	}

	return requirements
}

// inferResourceFromParam infers resource type and field from parameter name
func (dg *DependencyGraph) inferResourceFromParam(paramName, path string) (string, string) {
	paramLower := strings.ToLower(paramName)

	// Pattern: {petId} -> pet.id
	if strings.HasSuffix(paramLower, "id") && len(paramName) > 2 {
		resourceType := strings.ToLower(paramName[:len(paramName)-2])
		return resourceType, "id"
	}

	// Pattern: {id} -> infer from path
	if paramLower == "id" {
		resourceType := inferResourceTypeFromPath(path)
		return resourceType, "id"
	}

	// Pattern: {username} -> user.username
	if strings.HasSuffix(paramLower, "name") && len(paramName) > 4 {
		resourceType := strings.ToLower(paramName[:len(paramName)-4])
		if resourceType == "" {
			resourceType = inferResourceTypeFromPath(path)
		}
		return resourceType, paramName
	}

	// Pattern: {username} as is -> user.username
	if paramLower == "username" {
		return "user", "username"
	}

	// Fallback: use path to infer resource type
	resourceType := inferResourceTypeFromPath(path)
	return resourceType, paramName
}

// findProducerForResource finds the producer endpoint for a resource type
func (dg *DependencyGraph) findProducerForResource(resourceType string) string {
	if producer, exists := dg.resourceProducers[resourceType]; exists {
		return producer
	}
	return ""
}

// extractResponseFields extracts field names from response schema
func (dg *DependencyGraph) extractResponseFields(operation map[string]interface{}) []string {
	fields := []string{"id"} // Always include id as default

	responses, ok := operation["responses"].(map[string]interface{})
	if !ok {
		return fields
	}

	// Look for 200 or 201 response
	for code, respData := range responses {
		if code != "200" && code != "201" {
			continue
		}

		resp, ok := respData.(map[string]interface{})
		if !ok {
			continue
		}

		content, ok := resp["content"].(map[string]interface{})
		if !ok {
			continue
		}

		for _, ct := range content {
			ctData, ok := ct.(map[string]interface{})
			if !ok {
				continue
			}
			schema, ok := ctData["schema"].(map[string]interface{})
			if !ok {
				continue
			}
			props, ok := schema["properties"].(map[string]interface{})
			if !ok {
				continue
			}
			for field := range props {
				fields = append(fields, field)
			}
		}
	}

	return fields
}

// buildExecutionOrder builds a topologically sorted execution order
func (dg *DependencyGraph) buildExecutionOrder() {
	// Collect all endpoints
	allEndpoints := make(map[string]bool)
	for ep := range dg.producers {
		allEndpoints[ep] = true
	}
	for ep := range dg.consumers {
		allEndpoints[ep] = true
	}

	// Build dependency edges
	// producer -> consumers (consumers depend on producers)
	dependencies := make(map[string][]string)
	for consumer, requirements := range dg.consumers {
		for _, req := range requirements {
			if producer := req.ProducerHint; producer != "" {
				dependencies[consumer] = append(dependencies[consumer], producer)
			}
		}
	}

	// Topological sort using Kahn's algorithm
	inDegree := make(map[string]int)
	for ep := range allEndpoints {
		inDegree[ep] = len(dependencies[ep])
	}

	// Start with endpoints that have no dependencies (producers)
	queue := make([]string, 0)
	for ep := range allEndpoints {
		if inDegree[ep] == 0 {
			queue = append(queue, ep)
		}
	}

	// Sort queue for deterministic order
	sort.Strings(queue)

	// Process queue
	for len(queue) > 0 {
		// Pop first element
		current := queue[0]
		queue = queue[1:]
		dg.executionOrder = append(dg.executionOrder, current)

		// Find all endpoints that depend on this one
		for ep, deps := range dependencies {
			for _, dep := range deps {
				if dep == current {
					inDegree[ep]--
					if inDegree[ep] == 0 {
						queue = append(queue, ep)
						sort.Strings(queue)
					}
				}
			}
		}
	}
}

// GetExecutionOrder returns the topologically sorted execution order
func (dg *DependencyGraph) GetExecutionOrder() []string {
	return dg.executionOrder
}

// GetProducerForResource returns the producer endpoint for a resource type
func (dg *DependencyGraph) GetProducerForResource(resourceType string) string {
	return dg.resourceProducers[resourceType]
}

// GetRequirementsForEndpoint returns the resource requirements for an endpoint
func (dg *DependencyGraph) GetRequirementsForEndpoint(endpoint string) []ResourceRequirement {
	return dg.consumers[endpoint]
}

// GetResourcesCreatedBy returns the resources created by an endpoint
func (dg *DependencyGraph) GetResourcesCreatedBy(endpoint string) []ResourceInfo {
	return dg.producers[endpoint]
}

// IsProducer returns true if the endpoint is a resource producer
func (dg *DependencyGraph) IsProducer(endpoint string) bool {
	_, exists := dg.producers[endpoint]
	return exists
}

// IsConsumer returns true if the endpoint is a resource consumer
func (dg *DependencyGraph) IsConsumer(endpoint string) bool {
	_, exists := dg.consumers[endpoint]
	return exists
}

// GenerateContextMapping generates context mapping for a consumer endpoint
func (dg *DependencyGraph) GenerateContextMapping(endpoint string) map[string]string {
	mapping := make(map[string]string)

	requirements := dg.consumers[endpoint]
	for _, req := range requirements {
		key := fmt.Sprintf("%s_params.%s", req.ParamType, req.ParamName)
		value := fmt.Sprintf("{{CONTEXT:%s.%s}}", req.Type, req.Field)
		mapping[key] = value
	}

	return mapping
}
