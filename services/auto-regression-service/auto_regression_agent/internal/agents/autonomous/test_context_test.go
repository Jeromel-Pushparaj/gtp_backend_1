package autonomous

import (
	"testing"
)

func TestNormalizeContextTypeWithDynamicMappings(t *testing.T) {
	tc := NewTestContext()

	// Set up Petstore-like mappings (as would be returned by LLM)
	tc.SetContextMappings(map[string]string{
		"resource":    "pet",
		"animal":      "pet",
		"item":        "pet",
		"account":     "user",
		"profile":     "user",
		"purchase":    "order",
		"transaction": "order",
		"image":       "upload",
		"file":        "upload",
	})

	tests := []struct {
		input    string
		expected string
	}{
		{"resource", "pet"},
		{"animal", "pet"},
		{"created_pet", "pet"}, // Uses prefix stripping
		{"account", "user"},
		{"created_user", "user"}, // Uses prefix stripping
		{"purchase", "order"},
		{"transaction", "order"},
		{"image", "upload"},
		{"pet", "pet"},         // Already canonical
		{"user", "user"},       // Already canonical
		{"unknown", "unknown"}, // Unknown types pass through
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := tc.normalizeContextType(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeContextType(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNormalizeContextTypeWithoutMappings(t *testing.T) {
	// Test the default behavior without API-specific mappings
	tc := NewTestContext()

	tests := []struct {
		input    string
		expected string
	}{
		{"created_pet", "pet"},     // Strips "created_" prefix
		{"new_user", "user"},       // Strips "new_" prefix
		{"created_order", "order"}, // Strips "created_" prefix
		{"pet", "pet"},             // Pass-through
		{"resource", "resource"},   // Pass-through (no dynamic mapping)
		{"unknown", "unknown"},     // Pass-through
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := tc.normalizeContextType(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeContextType(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetContextWithFallback(t *testing.T) {
	tc := NewTestContext()

	// Set up Petstore-like mappings (as would be returned by LLM)
	tc.SetContextMappings(map[string]string{
		"resource": "pet",
		"animal":   "pet",
	})

	// Store context as "resource" (LLM inconsistent naming)
	tc.StoreContext("resource", map[string]interface{}{
		"id":   12345,
		"name": "Buddy",
	})

	// Test 1: Exact match for "resource" should work
	data, actualType := tc.GetContextWithFallback("resource", "id")
	if data == nil {
		t.Fatal("Expected to find context for 'resource', got nil")
	}
	if actualType != "resource" && actualType != "pet" {
		t.Errorf("Expected actualType to be 'resource' or 'pet', got %q", actualType)
	}
	if data["id"] != 12345 {
		t.Errorf("Expected id=12345, got %v", data["id"])
	}

	// Test 2: Looking for "pet" should find "resource" via canonical normalization
	// Since StoreContext also stores under canonical name
	data2, actualType2 := tc.GetContextWithFallback("pet", "id")
	if data2 == nil {
		t.Fatal("Expected to find context for 'pet' via fallback, got nil")
	}
	if data2["id"] != 12345 {
		t.Errorf("Expected id=12345, got %v", data2["id"])
	}
	t.Logf("Successfully found 'pet' context via type %q", actualType2)

	// Test 3: Looking for "animal" should find the pet context via aliases
	data3, actualType3 := tc.GetContextWithFallback("animal", "id")
	if data3 == nil {
		t.Fatal("Expected to find context for 'animal' via fallback, got nil")
	}
	if data3["id"] != 12345 {
		t.Errorf("Expected id=12345, got %v", data3["id"])
	}
	t.Logf("Successfully found 'animal' context via type %q", actualType3)
}

func TestGetContextWithFallbackFieldSearch(t *testing.T) {
	tc := NewTestContext()

	// Store with a completely non-standard name
	tc.StoreContext("weird_name", map[string]interface{}{
		"order_id": 99999,
		"status":   "pending",
	})

	// Test: Looking for any context type with "order_id" field should find it
	data, actualType := tc.GetContextWithFallback("order", "order_id")
	if data == nil {
		t.Fatal("Expected to find context via field search, got nil")
	}
	if actualType != "weird_name" {
		t.Errorf("Expected actualType to be 'weird_name', got %q", actualType)
	}
	if data["order_id"] != 99999 {
		t.Errorf("Expected order_id=99999, got %v", data["order_id"])
	}
}

func TestStoreContextNormalization(t *testing.T) {
	tc := NewTestContext()

	// Set up Petstore-like mappings (as would be returned by LLM)
	tc.SetContextMappings(map[string]string{
		"resource": "pet",
	})

	// Store as "resource" - should also create "pet" entry via normalization
	tc.StoreContext("resource", map[string]interface{}{
		"id":   777,
		"name": "TestPet",
	})

	// Direct lookup for "pet" should work (via canonical normalization)
	petData := tc.GetContext("pet")
	if petData == nil {
		t.Fatal("Expected direct lookup for 'pet' to work after storing as 'resource'")
	}
	if petData["id"] != 777 {
		t.Errorf("Expected id=777, got %v", petData["id"])
	}

	// Direct lookup for "resource" should also work
	resourceData := tc.GetContext("resource")
	if resourceData == nil {
		t.Fatal("Expected direct lookup for 'resource' to work")
	}
	if resourceData["id"] != 777 {
		t.Errorf("Expected id=777, got %v", resourceData["id"])
	}
}

func TestGetContextAliases(t *testing.T) {
	// Test that getContextAliases returns generic aliases for all types
	// These are common patterns that might be used across different APIs
	aliases := getContextAliases("pet")

	// The static aliases should include common patterns like "resource", "item", etc.
	t.Logf("Aliases for 'pet': %v", aliases)

	// Should have some generic aliases
	if len(aliases) == 0 {
		t.Error("Expected some generic aliases for 'pet'")
	}

	// Check that common generic aliases are present
	hasResource := false
	for _, a := range aliases {
		if a == "resource" {
			hasResource = true
			break
		}
	}
	if !hasResource {
		t.Errorf("Expected 'resource' to be a generic alias, got %v", aliases)
	}
}
