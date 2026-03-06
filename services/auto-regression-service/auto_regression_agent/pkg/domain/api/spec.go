package api

import "time"

// Spec represents an OpenAPI specification
type Spec struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	ServiceName string    `json:"service_name"`
	TeamID      string    `json:"team_id"`
	Content     []byte    `json:"content"` // Raw OpenAPI spec content
	Format      string    `json:"format"`  // yaml or json
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Endpoint represents a single API endpoint extracted from spec
type Endpoint struct {
	ID          string            `json:"id"`
	SpecID      string            `json:"spec_id"`
	Path        string            `json:"path"`
	Method      string            `json:"method"`
	OperationID string            `json:"operation_id"`
	Summary     string            `json:"summary"`
	Description string            `json:"description"`
	Tags        []string          `json:"tags"`
	Parameters  []Parameter       `json:"parameters"`
	RequestBody *RequestBody      `json:"request_body,omitempty"`
	Responses   map[string]Response `json:"responses"`
	Security    []SecurityRequirement `json:"security,omitempty"`
	Deprecated  bool              `json:"deprecated"`
}

// Parameter represents an API parameter
type Parameter struct {
	Name        string      `json:"name"`
	In          string      `json:"in"` // query, header, path, cookie
	Description string      `json:"description"`
	Required    bool        `json:"required"`
	Schema      *Schema     `json:"schema,omitempty"`
	Example     interface{} `json:"example,omitempty"`
}

// RequestBody represents request body definition
type RequestBody struct {
	Description string               `json:"description"`
	Required    bool                 `json:"required"`
	Content     map[string]MediaType `json:"content"`
}

// Response represents response definition
type Response struct {
	Description string               `json:"description"`
	Content     map[string]MediaType `json:"content,omitempty"`
	Headers     map[string]Header    `json:"headers,omitempty"`
}

// MediaType represents media type definition
type MediaType struct {
	Schema   *Schema     `json:"schema,omitempty"`
	Example  interface{} `json:"example,omitempty"`
	Examples map[string]Example `json:"examples,omitempty"`
}

// Schema represents JSON schema
type Schema struct {
	Type                 string              `json:"type,omitempty"`
	Format               string              `json:"format,omitempty"`
	Properties           map[string]*Schema  `json:"properties,omitempty"`
	Required             []string            `json:"required,omitempty"`
	Items                *Schema             `json:"items,omitempty"`
	Enum                 []interface{}       `json:"enum,omitempty"`
	Minimum              *float64            `json:"minimum,omitempty"`
	Maximum              *float64            `json:"maximum,omitempty"`
	MinLength            *int                `json:"minLength,omitempty"`
	MaxLength            *int                `json:"maxLength,omitempty"`
	Pattern              string              `json:"pattern,omitempty"`
	AdditionalProperties interface{}         `json:"additionalProperties,omitempty"`
	Ref                  string              `json:"$ref,omitempty"`
}

// Header represents response header
type Header struct {
	Description string      `json:"description"`
	Schema      *Schema     `json:"schema,omitempty"`
	Example     interface{} `json:"example,omitempty"`
}

// Example represents an example value
type Example struct {
	Summary     string      `json:"summary"`
	Description string      `json:"description"`
	Value       interface{} `json:"value"`
}

// SecurityRequirement represents security requirement
type SecurityRequirement struct {
	Name   string   `json:"name"`
	Scopes []string `json:"scopes"`
}

// SpecMetadata contains extracted metadata from spec
type SpecMetadata struct {
	SpecID           string    `json:"spec_id"`
	Title            string    `json:"title"`
	Version          string    `json:"version"`
	BaseURL          string    `json:"base_url"`
	TotalEndpoints   int       `json:"total_endpoints"`
	EndpointsByTag   map[string]int `json:"endpoints_by_tag"`
	SecuritySchemes  []string  `json:"security_schemes"`
	ExtractedAt      time.Time `json:"extracted_at"`
}

// EndpointComplexity represents complexity analysis of an endpoint
type EndpointComplexity struct {
	EndpointID       string  `json:"endpoint_id"`
	ComplexityScore  int     `json:"complexity_score"`
	ParameterCount   int     `json:"parameter_count"`
	RequiredParams   int     `json:"required_params"`
	ResponseCount    int     `json:"response_count"`
	HasRequestBody   bool    `json:"has_request_body"`
	HasSecurity      bool    `json:"has_security"`
	Priority         string  `json:"priority"` // critical, high, medium, low
}

