package autonomous

import "testing"

func TestIsBoundaryValue(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected bool
	}{
		// Valid values - should NOT be filtered
		{"valid id 101", float64(101), false},
		{"valid id 1", float64(1), false},
		{"valid id 12345", float64(12345), false},
		{"valid id 999", float64(999), false},
		{"valid string", "myuser", false},
		{"valid name", "John Doe", false},

		// Boundary values - SHOULD be filtered
		{"int64 max", float64(9223372036854775807), true},
		{"int64 max+1", float64(9223372036854775808), true},
		{"very large number", float64(9223372036854776000), true},
		{"int32 max", float64(2147483647), true},
		{"negative -1", float64(-1), true},
		{"zero", float64(0), true},
		{"large negative", float64(-999999999), true},
		{"1e18", float64(1e18), true},

		// Edge case strings
		{"empty string", "", true},
		{"invalid string", "invalid", true},
		{"null string", "null", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isBoundaryValue(tt.value)
			if got != tt.expected {
				t.Errorf("isBoundaryValue(%v) = %v, want %v", tt.value, got, tt.expected)
			}
		})
	}
}

func TestFilterBoundaryValues(t *testing.T) {
	input := map[string]interface{}{
		"id":           float64(101),                 // valid
		"name":         "Maximus",                    // valid
		"boundary_id":  float64(9223372036854775808), // boundary - should be filtered
		"zero_id":      float64(0),                   // boundary - should be filtered
		"valid_status": "available",                  // valid
	}

	filtered := filterBoundaryValues(input)

	// Should keep valid values
	if _, ok := filtered["id"]; !ok {
		t.Error("Expected 'id' to be kept")
	}
	if _, ok := filtered["name"]; !ok {
		t.Error("Expected 'name' to be kept")
	}
	if _, ok := filtered["valid_status"]; !ok {
		t.Error("Expected 'valid_status' to be kept")
	}

	// Should filter boundary values
	if _, ok := filtered["boundary_id"]; ok {
		t.Error("Expected 'boundary_id' to be filtered out")
	}
	if _, ok := filtered["zero_id"]; ok {
		t.Error("Expected 'zero_id' to be filtered out")
	}
}

func TestCheckExpectedStatus(t *testing.T) {
	pe := &PlannedExecutor{}

	tests := []struct {
		name     string
		actual   int
		expected int
		want     bool
	}{
		// Exact matches
		{"200 == 200", 200, 200, true},
		{"201 == 201", 201, 201, true},
		{"404 == 404", 404, 404, true},

		// 2xx flexibility - any 2xx matches any 2xx
		{"201 matches expected 200", 201, 200, true},
		{"200 matches expected 201", 200, 201, true},
		{"204 matches expected 200", 204, 200, true},
		{"200 matches expected 204", 200, 204, true},
		{"202 matches expected 200", 202, 200, true},

		// 4xx requires exact match
		{"400 != 404", 400, 404, false},
		{"404 != 400", 404, 400, false},
		{"422 != 400", 422, 400, false},
		{"401 != 403", 401, 403, false},

		// 5xx requires exact match
		{"500 != 502", 500, 502, false},
		{"503 != 500", 503, 500, false},

		// Cross-category always fails
		{"200 != 400", 200, 400, false},
		{"404 != 200", 404, 200, false},
		{"500 != 200", 500, 200, false},

		// Expected 0 means any 2xx
		{"200 with expected 0", 200, 0, true},
		{"201 with expected 0", 201, 0, true},
		{"204 with expected 0", 204, 0, true},
		{"404 with expected 0", 404, 0, false},
		{"500 with expected 0", 500, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pe.checkExpectedStatus(tt.actual, tt.expected)
			if got != tt.want {
				t.Errorf("checkExpectedStatus(%d, %d) = %v, want %v", tt.actual, tt.expected, got, tt.want)
			}
		})
	}
}

func TestSingularize(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Regular plurals
		{"users", "user"},
		{"pets", "pet"},
		{"orders", "order"},
		{"products", "product"},
		{"customers", "customer"},

		// -es endings
		{"addresses", "address"},
		{"classes", "class"},
		{"buses", "bus"},
		{"boxes", "box"},
		{"matches", "match"},
		{"dishes", "dish"},
		{"statuses", "status"},

		// -ies endings
		{"categories", "category"},
		{"policies", "policy"},
		{"entries", "entry"},
		{"companies", "company"},

		// Irregular plurals
		{"people", "person"},
		{"children", "child"},
		{"indices", "index"},
		{"analyses", "analysis"},
		{"criteria", "criterion"},

		// Words ending in -s that shouldn't change
		{"status", "status"},
		{"analysis", "analysis"},

		// Already singular
		{"user", "user"},
		{"pet", "pet"},
		{"order", "order"},

		// -ses special cases
		{"responses", "response"},
		{"expenses", "expense"},

		// Short words
		{"a", "a"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := Singularize(tt.input)
			if result != tt.expected {
				t.Errorf("Singularize(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestInferResourceTypeFromPath(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		// Simple paths
		{"/pets", "pet"},
		{"/users", "user"},
		{"/orders", "order"},

		// With ID parameters
		{"/customers/{id}", "customer"},
		{"/products/{productId}", "product"},
		{"/invoices/{invoiceId}/items/{itemId}", "item"},

		// API versioning
		{"/api/v1/customers", "customer"},
		{"/api/v2/orders/{id}", "order"},
		{"/api/v10/products", "product"},
		{"/rest/v1/users", "user"},

		// Nested resources (returns last)
		{"/customers/{customerId}/orders", "order"},
		{"/api/v1/users/{userId}/addresses/{addressId}", "address"},

		// Common prefixes
		{"/internal/billing/invoices", "invoice"},
		{"/public/api/v1/products", "product"},
		{"/admin/users", "user"},

		// Edge cases
		{"/", ""},
		{"/api/v1", ""},
		{"/{id}", ""},

		// Real-world examples
		{"/store/orders/{orderId}", "order"},
		{"/pet/{petId}/uploadImage", "uploadimage"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := InferResourceTypeFromPath(tt.path)
			if result != tt.expected {
				t.Errorf("InferResourceTypeFromPath(%q) = %q, want %q", tt.path, result, tt.expected)
			}
		})
	}
}
