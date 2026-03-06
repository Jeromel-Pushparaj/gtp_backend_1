package agent

import "time"

// LLMRequest represents a request to an LLM
type LLMRequest struct {
	Prompt      string  `json:"prompt"`
	Temperature float64 `json:"temperature"`
	MaxTokens   int     `json:"max_tokens"`
}

// LLMResponse represents a response from an LLM
type LLMResponse struct {
	Content      string `json:"content"`
	FinishReason string `json:"finish_reason"`
	TokensUsed   int    `json:"tokens_used"`
	Model        string `json:"model"`
}

// DiscoveryResult represents the output of spec discovery
type DiscoveryResult struct {
	SpecID           string                `json:"spec_id"`
	Title            string                `json:"title"`
	Version          string                `json:"version"`
	BaseURL          string                `json:"base_url"`
	Endpoints        []DiscoveredEndpoint  `json:"endpoints"`
	CreatorEndpoints []CreatorEndpoint     `json:"creator_endpoints"`
	Schemas          map[string]SchemaInfo `json:"schemas"`
	SecuritySchemes  []SecurityScheme      `json:"security_schemes"`
	DiscoveredAt     time.Time             `json:"discovered_at"`
	ValidationErrors []string              `json:"validation_errors,omitempty"`
}

// DiscoveredEndpoint represents a discovered API endpoint
type DiscoveredEndpoint struct {
	ID              string                  `json:"id"`
	Path            string                  `json:"path"`
	Method          string                  `json:"method"`
	OperationID     string                  `json:"operation_id"`
	Summary         string                  `json:"summary"`
	Description     string                  `json:"description"`
	Tags            []string                `json:"tags"`
	Parameters      []ParameterInfo         `json:"parameters"`
	RequestBody     *RequestBodyInfo        `json:"request_body,omitempty"`
	Responses       map[string]ResponseInfo `json:"responses"`
	Security        []SecurityRequirement   `json:"security,omitempty"`
	Deprecated      bool                    `json:"deprecated"`
	IsCreator       bool                    `json:"is_creator"`                 // Detected as creator endpoint
	CreatesResource string                  `json:"creates_resource,omitempty"` // Resource type created
}

// ParameterInfo represents parameter metadata
type ParameterInfo struct {
	Name        string      `json:"name"`
	In          string      `json:"in"` // query, header, path, cookie
	Description string      `json:"description"`
	Required    bool        `json:"required"`
	Schema      SchemaInfo  `json:"schema"`
	Example     interface{} `json:"example,omitempty"`
}

// RequestBodyInfo represents request body metadata
type RequestBodyInfo struct {
	Description string                 `json:"description"`
	Required    bool                   `json:"required"`
	ContentType string                 `json:"content_type"` // Primary content type (e.g., application/json)
	Schema      SchemaInfo             `json:"schema"`
	Examples    map[string]ExampleInfo `json:"examples,omitempty"`
}

// ResponseInfo represents response metadata
type ResponseInfo struct {
	StatusCode  string                `json:"status_code"`
	Description string                `json:"description"`
	ContentType string                `json:"content_type,omitempty"`
	Schema      SchemaInfo            `json:"schema,omitempty"`
	Headers     map[string]HeaderInfo `json:"headers,omitempty"`
}

// SchemaInfo represents JSON schema metadata
type SchemaInfo struct {
	Type        string                `json:"type"` // string, number, integer, boolean, array, object
	Format      string                `json:"format,omitempty"`
	Properties  map[string]SchemaInfo `json:"properties,omitempty"`
	Required    []string              `json:"required,omitempty"`
	Items       *SchemaInfo           `json:"items,omitempty"` // For arrays
	Enum        []interface{}         `json:"enum,omitempty"`
	Pattern     string                `json:"pattern,omitempty"`
	MinLength   *uint64               `json:"min_length,omitempty"`
	MaxLength   *uint64               `json:"max_length,omitempty"`
	Minimum     *float64              `json:"minimum,omitempty"`
	Maximum     *float64              `json:"maximum,omitempty"`
	Description string                `json:"description,omitempty"`
	Example     interface{}           `json:"example,omitempty"`
	Ref         string                `json:"ref,omitempty"` // $ref reference
}

// HeaderInfo represents HTTP header metadata
type HeaderInfo struct {
	Description string     `json:"description"`
	Schema      SchemaInfo `json:"schema"`
	Required    bool       `json:"required"`
}

// ExampleInfo represents an example value
type ExampleInfo struct {
	Summary     string      `json:"summary"`
	Description string      `json:"description"`
	Value       interface{} `json:"value"`
}

// CreatorEndpoint represents an endpoint that creates resources
type CreatorEndpoint struct {
	EndpointID      string   `json:"endpoint_id"`
	Path            string   `json:"path"`
	Method          string   `json:"method"`
	ResourceType    string   `json:"resource_type"`     // e.g., "Resource", "User", "Order"
	IDFieldName     string   `json:"id_field_name"`     // e.g., "id", "resourceId", "userId"
	IDFieldLocation string   `json:"id_field_location"` // "response.body", "response.header"
	IDFieldPath     string   `json:"id_field_path"`     // JSON path to ID field
	RequiredFields  []string `json:"required_fields"`   // Required request fields
	Confidence      float64  `json:"confidence"`        // 0.0 to 1.0
}

// SecurityScheme represents authentication/authorization scheme
type SecurityScheme struct {
	Name         string `json:"name"`
	Type         string `json:"type"`             // apiKey, http, oauth2, openIdConnect
	Scheme       string `json:"scheme,omitempty"` // For http type: bearer, basic
	In           string `json:"in,omitempty"`     // For apiKey: query, header, cookie
	BearerFormat string `json:"bearer_format,omitempty"`
}

// SecurityRequirement represents security requirement for an endpoint
type SecurityRequirement struct {
	Name   string   `json:"name"`
	Scopes []string `json:"scopes,omitempty"`
}

// DiscoveryMetrics represents discovery performance metrics
type DiscoveryMetrics struct {
	TotalEndpoints   int           `json:"total_endpoints"`
	CreatorEndpoints int           `json:"creator_endpoints"`
	SchemasExtracted int           `json:"schemas_extracted"`
	ValidationErrors int           `json:"validation_errors"`
	ProcessingTime   time.Duration `json:"processing_time"`
	ParsingTime      time.Duration `json:"parsing_time"`
	AnalysisTime     time.Duration `json:"analysis_time"`
}
