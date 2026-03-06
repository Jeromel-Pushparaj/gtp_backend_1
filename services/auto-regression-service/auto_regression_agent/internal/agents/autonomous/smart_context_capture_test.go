package autonomous

import (
	"net/http"
	"testing"
)

func TestSmartContextCapture_DirectIDCapture(t *testing.T) {
	scc := NewSmartContextCapture()

	// Test direct ID capture from response body
	responseBody := []byte(`{"id": 123, "name": "test-pet", "status": "available"}`)
	headers := http.Header{}

	result := scc.CaptureContext(
		"POST /pet",
		responseBody,
		headers,
		nil,
		[]string{"id", "name"},
	)

	if !result.Success {
		t.Errorf("Expected success, got failure: %s", result.FailureInfo)
	}

	if result.Source != "response_body" {
		t.Errorf("Expected source 'response_body', got '%s'", result.Source)
	}

	if id, ok := result.Captured["id"]; !ok || id != float64(123) {
		t.Errorf("Expected id=123, got %v", result.Captured["id"])
	}

	if name, ok := result.Captured["name"]; !ok || name != "test-pet" {
		t.Errorf("Expected name='test-pet', got %v", result.Captured["name"])
	}
}

func TestSmartContextCapture_WrappedResponse(t *testing.T) {
	scc := NewSmartContextCapture()

	// Register schema with wrapper field
	scc.RegisterSchema("POST /pet", ResponseSchemaInfo{
		Endpoint:     "POST /pet",
		ResourceType: "pet",
		WrapperField: "data",
		IDFieldNames: []string{"id", "petId"},
	})

	// Test wrapped response
	responseBody := []byte(`{"data": {"id": 456, "name": "wrapped-pet"}, "status": "success"}`)
	headers := http.Header{}

	result := scc.CaptureContext(
		"POST /pet",
		responseBody,
		headers,
		nil,
		[]string{"id"},
	)

	if !result.Success {
		t.Errorf("Expected success, got failure: %s", result.FailureInfo)
	}

	if result.Source != "response_body_unwrapped" {
		t.Errorf("Expected source 'response_body_unwrapped', got '%s'", result.Source)
	}

	if id, ok := result.Captured["id"]; !ok || id != float64(456) {
		t.Errorf("Expected id=456, got %v", result.Captured["id"])
	}
}

func TestSmartContextCapture_LocationHeader(t *testing.T) {
	scc := NewSmartContextCapture()

	// Test Location header extraction
	responseBody := []byte(`{"code": 200, "message": "created"}`)
	headers := http.Header{}
	headers.Set("Location", "/api/pets/789")

	result := scc.CaptureContext(
		"POST /pet",
		responseBody,
		headers,
		nil,
		[]string{"id"},
	)

	if !result.Success {
		t.Errorf("Expected success, got failure: %s", result.FailureInfo)
	}

	if result.Source != "response_header" {
		t.Errorf("Expected source 'response_header', got '%s'", result.Source)
	}

	if id, ok := result.Captured["id"]; !ok || id != "789" {
		t.Errorf("Expected id='789', got %v", result.Captured["id"])
	}
}

func TestSmartContextCapture_SmartIDDetection(t *testing.T) {
	scc := NewSmartContextCapture()

	// Register schema with ID field names
	scc.RegisterSchema("POST /pet", ResponseSchemaInfo{
		Endpoint:     "POST /pet",
		ResourceType: "pet",
		IDFieldNames: []string{"id", "petId", "_id"},
	})

	// Test smart ID detection with non-standard field name
	// When looking for "petId" directly, it should be found in response
	responseBody := []byte(`{"petId": 999, "name": "smart-pet"}`)
	headers := http.Header{}

	result := scc.CaptureContext(
		"POST /pet",
		responseBody,
		headers,
		nil,
		[]string{"petId"}, // Looking for "petId" which exists in response
	)

	if !result.Success {
		t.Errorf("Expected success, got failure: %s", result.FailureInfo)
	}

	if result.Source != "response_body" {
		t.Errorf("Expected source 'response_body', got '%s'", result.Source)
	}

	if id, ok := result.Captured["petId"]; !ok || id != float64(999) {
		t.Errorf("Expected petId=999, got %v", result.Captured["petId"])
	}
}

func TestSmartContextCapture_FallbackToPayload(t *testing.T) {
	scc := NewSmartContextCapture()

	// Test fallback to request payload
	responseBody := []byte(`{"code": 200, "message": "success"}`)
	headers := http.Header{}
	payload := map[string]interface{}{
		"name":   "payload-pet",
		"status": "available",
	}

	result := scc.CaptureContext(
		"POST /pet",
		responseBody,
		headers,
		payload,
		[]string{"name", "status"},
	)

	if !result.Success {
		t.Errorf("Expected success, got failure: %s", result.FailureInfo)
	}

	if result.Source != "request_payload" {
		t.Errorf("Expected source 'request_payload', got '%s'", result.Source)
	}

	if name, ok := result.Captured["name"]; !ok || name != "payload-pet" {
		t.Errorf("Expected name='payload-pet', got %v", result.Captured["name"])
	}
}
