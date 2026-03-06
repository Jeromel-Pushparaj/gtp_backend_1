package autonomous

import (
	"testing"
)

func TestDependencyGraph_BuildFromSpec(t *testing.T) {
	dg := NewDependencyGraph()

	// Create a minimal OpenAPI spec
	spec := map[string]interface{}{
		"paths": map[string]interface{}{
			"/pet": map[string]interface{}{
				"post": map[string]interface{}{
					"operationId": "addPet",
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "successful operation",
						},
					},
				},
			},
			"/pet/{petId}": map[string]interface{}{
				"get": map[string]interface{}{
					"operationId": "getPetById",
					"parameters": []interface{}{
						map[string]interface{}{
							"name":     "petId",
							"in":       "path",
							"required": true,
						},
					},
				},
				"delete": map[string]interface{}{
					"operationId": "deletePet",
				},
			},
		},
	}

	err := dg.BuildFromSpec(spec)
	if err != nil {
		t.Fatalf("Failed to build dependency graph: %v", err)
	}

	// Check that POST /pet is registered as a producer
	if !dg.IsProducer("POST /pet") {
		t.Error("Expected POST /pet to be a producer")
	}

	// Check that GET /pet/{petId} is registered as a consumer
	if !dg.IsConsumer("GET /pet/{petId}") {
		t.Error("Expected GET /pet/{petId} to be a consumer")
	}

	// Check that DELETE /pet/{petId} is registered as a consumer
	if !dg.IsConsumer("DELETE /pet/{petId}") {
		t.Error("Expected DELETE /pet/{petId} to be a consumer")
	}

	// Check producer for pet resource
	producer := dg.GetProducerForResource("pet")
	if producer != "POST /pet" {
		t.Errorf("Expected producer 'POST /pet', got '%s'", producer)
	}
}

func TestDependencyGraph_GetRequirements(t *testing.T) {
	dg := NewDependencyGraph()

	spec := map[string]interface{}{
		"paths": map[string]interface{}{
			"/pet": map[string]interface{}{
				"post": map[string]interface{}{},
			},
			"/pet/{petId}": map[string]interface{}{
				"get": map[string]interface{}{},
			},
		},
	}

	err := dg.BuildFromSpec(spec)
	if err != nil {
		t.Fatalf("Failed to build dependency graph: %v", err)
	}

	requirements := dg.GetRequirementsForEndpoint("GET /pet/{petId}")
	if len(requirements) == 0 {
		t.Error("Expected requirements for GET /pet/{petId}")
	}

	// Check that the requirement is for pet.id
	found := false
	for _, req := range requirements {
		if req.Type == "pet" && req.Field == "id" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected requirement for pet.id")
	}
}

func TestDependencyGraph_ExecutionOrder(t *testing.T) {
	dg := NewDependencyGraph()

	spec := map[string]interface{}{
		"paths": map[string]interface{}{
			"/pet": map[string]interface{}{
				"post": map[string]interface{}{},
			},
			"/pet/{petId}": map[string]interface{}{
				"get":    map[string]interface{}{},
				"delete": map[string]interface{}{},
			},
		},
	}

	err := dg.BuildFromSpec(spec)
	if err != nil {
		t.Fatalf("Failed to build dependency graph: %v", err)
	}

	order := dg.GetExecutionOrder()
	if len(order) == 0 {
		t.Error("Expected non-empty execution order")
	}

	// POST /pet should come before GET /pet/{petId} and DELETE /pet/{petId}
	postIndex := -1
	getIndex := -1
	deleteIndex := -1

	for i, ep := range order {
		switch ep {
		case "POST /pet":
			postIndex = i
		case "GET /pet/{petId}":
			getIndex = i
		case "DELETE /pet/{petId}":
			deleteIndex = i
		}
	}

	if postIndex == -1 {
		t.Error("POST /pet not found in execution order")
	}
	if getIndex != -1 && postIndex > getIndex {
		t.Error("POST /pet should come before GET /pet/{petId}")
	}
	if deleteIndex != -1 && postIndex > deleteIndex {
		t.Error("POST /pet should come before DELETE /pet/{petId}")
	}
}

func TestDependencyGraph_GenerateContextMapping(t *testing.T) {
	dg := NewDependencyGraph()

	spec := map[string]interface{}{
		"paths": map[string]interface{}{
			"/pet": map[string]interface{}{
				"post": map[string]interface{}{},
			},
			"/pet/{petId}": map[string]interface{}{
				"get": map[string]interface{}{},
			},
		},
	}

	err := dg.BuildFromSpec(spec)
	if err != nil {
		t.Fatalf("Failed to build dependency graph: %v", err)
	}

	mapping := dg.GenerateContextMapping("GET /pet/{petId}")
	if len(mapping) == 0 {
		t.Error("Expected non-empty context mapping")
	}

	// Check that the mapping contains the expected context reference
	expectedKey := "path_params.petId"
	expectedValue := "{{CONTEXT:pet.id}}"

	if val, ok := mapping[expectedKey]; !ok || val != expectedValue {
		t.Errorf("Expected mapping[%s]=%s, got %v", expectedKey, expectedValue, mapping)
	}
}

